package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/account"
	"github.com/ddouglas/ledger/internal/auth"
	"github.com/ddouglas/ledger/internal/cache"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/ddouglas/ledger/internal/item"
	"github.com/ddouglas/ledger/internal/mysql"
	"github.com/ddouglas/ledger/internal/server"
	"github.com/ddouglas/ledger/internal/transaction"
	"github.com/ddouglas/ledger/internal/user"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/plaid/plaid-go/plaid"

	driver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	cfg    *config
	logger *logrus.Logger
)

type core struct {
	logger   *logrus.Logger
	redis    *redis.Client
	newrelic *newrelic.Application
	repos    *repositories
	gateway  gateway.Service
	s3       *s3.Client
}

type repositories struct {
	account     ledger.AccountRepository
	health      ledger.HealthRepository
	institution ledger.InstitutionRepository
	item        ledger.ItemRepository
	transaction ledger.TransactionRepository
	user        ledger.UserRepository
	webhook     ledger.WebhookRepository
}

func init() {
	buildConfig()
	buildLogger()
}

func main() {

	app := cli.NewApp()
	app.Name = "Ledger CLI"
	app.Usage = "Manages Ledger Services"
	app.Commands = []*cli.Command{
		{
			Name:   "server",
			Usage:  "starts the ledger api",
			Action: actionAPI,
		},
		{
			Name:   "worker",
			Usage:  "starts the ledger worker, which handles various background tasks such as processing plaid webhooks and sending notifications",
			Action: actionWorker,
		},
		// {
		// 	Name:   "s3",
		// 	Action: actionS3Upload,
		// },
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

}

func buildCore() *core {
	return &core{
		logger:   logger,
		redis:    buildRedis(),
		newrelic: buildNewRelic(),
		repos:    buildRepositories(),
		gateway:  buildGateway(),
		s3:       buildS3(),
	}
}

func buildAWSConfig() aws.Config {
	awsConf, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.Spaces.ClientID,
				cfg.Spaces.ClientSecret,
				"",
			),
		),
		awsConfig.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL: cfg.Spaces.Endpoint,
				}, nil
			}

			return aws.Endpoint{}, nil
		})),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to load aws configuration: %s", err))
	}

	return awsConf
}

func buildS3() *s3.Client {
	return s3.NewFromConfig(buildAWSConfig())
}

func buildNewRelic() *newrelic.Application {

	entry := logger.WithField("service", "NewRelic")

	app, err := newrelic.NewApplication(newrelic.ConfigFromEnvironment())
	if err != nil {
		logger.WithError(err).Panicf("failed to build newrelic application")
	}

	entry.Info("Waiting For Connection")
	defer entry.Info("Connected")
	err = app.WaitForConnection(time.Second * 20)
	if err != nil {
		entry.WithError(err).Panicf("Connection Failed")
	}

	return app

}

func buildRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		MaxRetries:         5,
		IdleTimeout:        time.Second * 10,
		IdleCheckFrequency: time.Second * 5,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Panicf("failed to ping redis server: %s", err)
	}

	return redisClient
}

func buildRepositories() *repositories {

	m := cfg.MySQL

	config := driver.Config{
		User:                 m.User,
		Passwd:               m.Pass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", m.Host, m.Port),
		DBName:               m.DB,
		Loc:                  time.UTC,
		Timeout:              time.Second,
		ReadTimeout:          time.Second,
		WriteTimeout:         time.Second,
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Panicf("[MySQL Connect] Failed to connect to mysql server: %s", err)
	}

	db.SetConnMaxIdleTime(time.Second * 5)
	db.SetConnMaxLifetime(time.Second * 30)
	db.SetMaxOpenConns(100)

	err = db.Ping()
	if err != nil {
		log.Panicf("[MySQL Connect] Failed to ping mysql server: %s", err)
	}

	dbx := sqlx.NewDb(db, "mysql")

	return &repositories{
		account:     mysql.NewAccountRepository(dbx),
		health:      mysql.NewHealthRepository(dbx),
		institution: mysql.NewInstitutionRepository(dbx),
		item:        mysql.NewItemRepository(dbx),
		transaction: mysql.NewTransactionRepository(dbx),
		user:        mysql.NewUserRepository(dbx),
		webhook:     mysql.NewWebhookRepository(dbx),
	}

}

func buildGateway() gateway.Service {

	var plaidEnv plaid.Environment
	switch cfg.Plaid.Environment {
	case "production":
		plaidEnv = plaid.Production
	case "development":
		plaidEnv = plaid.Development
	default:
		plaidEnv = plaid.Sandbox
	}

	c, err := plaid.NewClient(plaid.ClientOptions{
		ClientID:    cfg.Plaid.ClientID,
		Secret:      cfg.Plaid.ClientSecret,
		Environment: plaidEnv,
		HTTPClient:  http.DefaultClient,
	})
	if err != nil {
		logger.WithError(err).Panic("failed to configure plaid client")
	}

	return gateway.New(
		gateway.WithPlaidClient(c),
		gateway.WithLanguage("en"),
		gateway.WithCountryCode("US"),
		gateway.WithProducts("auth", "transactions"),
		gateway.WithWebhook(cfg.Plaid.Webhook),
		gateway.WithLogger(logger),
	)

}

func actionAPI(c *cli.Context) error {

	core := buildCore()
	client := &http.Client{
		Transport: newrelic.NewRoundTripper(nil),
	}
	cache := cache.New(core.redis)
	oauth2 := oauth2Config()
	user := user.New(
		user.WithUserRepository(core.repos.user),
	)

	auth := auth.New(
		cache,
		client,
		oauth2,
		auth.WithJWKSURI(cfg.Auth0.JWKSURI),
		auth.WithAudience(cfg.Auth0.Audience),
		auth.WithIssuer(cfg.Auth0.Issuer),
	)

	accounts := account.New(
		account.WithAccountRepository(core.repos.account),
	)

	item := item.New(
		item.WithAccount(core.repos.account),
		item.WithGateway(core.gateway),
		item.WithInstitutionRepository(core.repos.institution),
		item.WithItemRepository(core.repos.item),
	)

	transaction := transaction.New(
		transaction.WithCache(cache),
		transaction.WithLogger(core.logger),
		transaction.WithS3(core.s3, cfg.Spaces.Bucket),
		transaction.WithTransactionRepository(core.repos.transaction),
	)

	importer := importer.New(
		importer.WithRedis(core.redis),
		importer.WithGateway(core.gateway),
		importer.WithLogger(core.logger),
		importer.WithNewrelic(core.newrelic),
		importer.WithAccounts(accounts),
		importer.WithItems(item),
		importer.WithTransactions(transaction),
	)

	server := server.New(
		server.WithAuth(auth),
		server.WithAuth0ServerToken(cfg.Auth0.ServerToken),
		server.WithGateway(core.gateway),
		server.WithImporter(importer),
		server.WithLogger(logger),
		server.WithNewrelic(core.newrelic),
		server.WithPort(cfg.API.Port),
		server.WithUser(user),
		server.WithAccounts(accounts),
		server.WithItems(item),
		server.WithTransactions(transaction),
	)

	// Channel to listen for errors generated by api server
	serverErrors := make(chan error, 1)

	// Channel to listen for interrupts and to run a graceful shutdown
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Start up our server
	go func() {
		serverErrors <- server.Run()
	}()

	// Blocking until read from channel(s)
	select {
	case err := <-serverErrors:
		core.logger.Fatalf("error starting server: %v", err.Error())

	case <-osSignals:
		core.logger.Println("starting server shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.GracefullyShutdown(ctx)
		if err != nil {
			core.logger.Fatalf("error trying to shutdown http server: %v", err.Error())
		}

	}

	return nil

}

func actionWorker(c *cli.Context) error {

	core := buildCore()

	accounts := account.New(
		account.WithAccountRepository(core.repos.account),
	)

	item := item.New(
		item.WithAccount(core.repos.account),
		item.WithGateway(core.gateway),
		item.WithInstitutionRepository(core.repos.institution),
		item.WithItemRepository(core.repos.item),
	)

	transaction := transaction.New(
		transaction.WithLogger(core.logger),
		transaction.WithTransactionRepository(core.repos.transaction),
	)

	importer := importer.New(
		importer.WithRedis(core.redis),
		importer.WithGateway(core.gateway),
		importer.WithLogger(core.logger),
		importer.WithNewrelic(core.newrelic),
		importer.WithAccounts(accounts),
		importer.WithItems(item),
		importer.WithTransactions(transaction),
		importer.WithWebhookRepository(core.repos.webhook),
	)

	ctx, cancel := context.WithCancel(context.Background())

	go importer.Run(ctx)

	// Channel to listen for interrupts and to run a graceful shutdown
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	<-osSignals
	core.logger.Println("starting worker shutdown...")
	cancel()

	return nil

}

// func actionS3Upload(c *cli.Context) error {

// 	var ctx = context.Background()

// 	core := buildCore()

// 	// file, err := os.Open("assets/example.jpg")
// 	// if err != nil {
// 	// 	core.logger.WithError(err).Fatal("failed to open example file")
// 	// }

// 	// defer file.Close()

// 	// obj := s3.PutObjectInput{
// 	// 	Bucket:      aws.String("onetwentyseven"),
// 	// 	Key:         aws.String("example.jpg"),
// 	// 	Body:        file,
// 	// 	ContentType: aws.String("image/jpeg"),
// 	// }

// 	// _, err = core.s3.PutObject(context.Background(), &obj)
// 	// if err != nil {
// 	// 	core.logger.WithError(err).Fatal("failed to upload file")
// 	// }

// 	// input := &s3.GetObjectInput{
// 	// 	Bucket: aws.String("onetwentyseven"),
// 	// 	Key:    aws.String("example.jpg"),
// 	// }

// 	// file, err := core.s3.GetObject(ctx, input)
// 	// if err != nil {
// 	// 	core.logger.WithError(err).Fatal("failed to fetch file")
// 	// }

// 	// data, err := io.ReadAll(file.Body)
// 	// if err != nil {
// 	// 	core.logger.WithError(err).Fatal("failed to read file")
// 	// }

// 	// newFile, _ := os.Create("example2.jpg")
// 	// newFile.Write(data)
// 	// newFile.Close()

// 	// presignClient := s3.NewPresignClient(core.s3)

// 	// req, err := presignClient.PresignGetObject(ctx, input)
// 	// if err != nil {
// 	// 	core.logger.WithError(err).Fatal("failed to fetch presigned url")
// 	// }

// 	// fmt.Println(req.SignedHeader)

// 	return nil
// }
