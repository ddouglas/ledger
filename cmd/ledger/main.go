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
	"github.com/ddouglas/ledger/internal/server/gql/dataloaders"
	"github.com/ddouglas/ledger/internal/transaction"
	"github.com/ddouglas/ledger/internal/user"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/plaid/plaid-go/plaid"
	"github.com/robfig/cron/v3"

	driver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	cfg    *config
	logger *logrus.Logger
	dbx    *sqlx.DB
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
	item        ledger.ItemRepository
	migrations  ledger.MigrationRepository
	plaid       ledger.PlaidRepository
	transaction ledger.TransactionRepository
	user        ledger.UserRepository
	webhook     ledger.WebhookRepository
	merchant    ledger.MerchantRepository
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
		{
			Name:  "migrate",
			Usage: "Manage Application DB Migrations",
			Subcommands: []*cli.Command{
				{
					Name:   "create",
					Usage:  "create a new migration",
					Action: actionCreateMigration,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "name",
							Required: true,
							Usage:    "the name of the migration. Filename with be ${datetime}_${name}.sql",
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize cli")
	}

}

func buildCore() *core {
	r := buildRedis()
	repos := buildRepositories()

	return &core{
		logger:   logger,
		redis:    r,
		newrelic: buildNewRelic(),
		repos:    buildRepositories(),
		gateway:  buildGateway(r, repos),
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

	dbx = sqlx.NewDb(db, "mysql")

	return &repositories{
		account:     mysql.NewAccountRepository(dbx),
		health:      mysql.NewHealthRepository(dbx),
		item:        mysql.NewItemRepository(dbx),
		migrations:  mysql.NewMigrationRepostory(dbx),
		plaid:       mysql.NewPlaidRepository(dbx),
		transaction: mysql.NewTransactionRepository(dbx),
		user:        mysql.NewUserRepository(dbx),
		webhook:     mysql.NewWebhookRepository(dbx),
		merchant:    mysql.NewMerchantRepository(dbx),
	}

}

func buildGateway(r *redis.Client, repos *repositories) gateway.Service {

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

	cache := cache.New(r)

	return gateway.New(
		gateway.WithPlaidClient(c),
		gateway.WithLanguage("en"),
		gateway.WithCountryCode("US"),
		gateway.WithProducts("auth", "transactions"),
		gateway.WithWebhook(cfg.Plaid.Webhook),
		gateway.WithLogger(logger),
		gateway.WithCache(cache),
		gateway.WithPlaidRepository(repos.plaid),
	)

}

func actionAPI(c *cli.Context) error {

	core := buildCore()

	if cfg.MySQL.Migrate {
		runMigrations(core)
	}

	client := &http.Client{
		Transport: newrelic.NewRoundTripper(http.DefaultTransport),
	}
	cache := cache.New(core.redis)
	oauth2 := oauth2Config()
	user := user.New(
		core.repos.user,
	)

	auth := auth.New(
		cache,
		client,
		oauth2,
		cfg.Auth0.JWKSURI,
		cfg.Auth0.Audience,
		cfg.Auth0.Issuer,
	)

	accounts := account.New(
		core.repos.account,
	)

	item := item.New(
		core.repos.account,
		core.gateway,
		core.repos.item,
		core.repos.plaid,
	)

	transaction := transaction.New(
		core.s3,
		core.logger,
		core.gateway,
		cache,
		cfg.Spaces.Bucket,
		core.repos.transaction,
		core.repos.merchant,
	)

	importer := importer.New(
		core.newrelic,
		core.logger,
		core.redis,
		core.gateway,
		accounts,
		item,
		transaction,
		core.repos.webhook,
	)

	loaders := dataloaders.New(item)

	server := server.New(
		cfg.API.Port,
		core.newrelic,
		logger,
		auth,
		loaders,
		core.gateway,
		user,
		importer,
		accounts,
		item,
		transaction,
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
		core.repos.account,
	)

	item := item.New(
		core.repos.account,
		core.gateway,
		core.repos.item,
		core.repos.plaid,
	)
	cache := cache.New(core.redis)

	transaction := transaction.New(
		core.s3,
		core.logger,
		core.gateway,
		cache,
		cfg.Spaces.Bucket,
		core.repos.transaction,
		core.repos.merchant,
	)

	importer := importer.New(
		core.newrelic,
		core.logger,
		core.redis,
		core.gateway,
		accounts,
		item,
		transaction,
		core.repos.webhook,
	)

	ctx, cancel := context.WithCancel(context.Background())

	crn := cron.New()
	id, err := crn.AddFunc("@midnight", func() {
		core.gateway.ImportCategories(ctx)
	})
	if err != nil {
		core.logger.WithError(err).Fatal("failed to add import categories job to cron scheduler. exiting go routing")
	}
	core.logger.WithField("id", id).Debug("successfully added import categories job to cron scheduler")

	id, err = crn.AddFunc("@midnight", func() {
		core.gateway.ImportInstitutions(ctx)
	})
	if err != nil {
		core.logger.WithError(err).Fatal("failed to add import institutions job to cron scheduler. exiting go routing")
	}
	core.logger.WithField("id", id).Debug("successfully added import institutions job to cron scheduler")

	core.logger.Info("starting cron...")
	crn.Start()

	core.logger.Info("starting importer...")
	go importer.Run(ctx)

	// Channel to listen for interrupts and to run a graceful shutdown
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	core.logger.Info("worker processes launched successfully")
	<-osSignals
	core.logger.Println("starting worker shutdown...")
	cancel()

	return nil

}
