package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/ddouglas/ledger/internal/account"
	"github.com/ddouglas/ledger/internal/auth"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/ddouglas/ledger/internal/item"
	resolvers "github.com/ddouglas/ledger/internal/server/gql"
	"github.com/ddouglas/ledger/internal/server/gql/dataloaders"
	"github.com/ddouglas/ledger/internal/server/gql/generated"
	"github.com/ddouglas/ledger/internal/transaction"
	"github.com/ddouglas/ledger/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type server struct {
	port        uint
	logger      *logrus.Logger
	auth        auth.Service
	loaders     dataloaders.Service
	importer    importer.Service
	gateway     gateway.Service
	newrelic    *newrelic.Application
	user        user.Service
	account     account.Service
	item        item.Service
	transaction transaction.Service

	server *http.Server
}

func New(
	port uint,
	newrelic *newrelic.Application,
	logger *logrus.Logger,

	auth auth.Service,
	loaders dataloaders.Service,
	gateway gateway.Service,
	user user.Service,
	importer importer.Service,
	account account.Service,
	item item.Service,
	transaction transaction.Service,

) *server {

	s := &server{
		newrelic:    newrelic,
		port:        port,
		logger:      logger,
		auth:        auth,
		loaders:     loaders,
		gateway:     gateway,
		user:        user,
		importer:    importer,
		account:     account,
		item:        item,
		transaction: transaction,
	}

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.buildRouter(),
	}

	return s
}

func (s *server) Run() error {
	s.logger.WithField("service", "server").Infof("Starting on Port %d", s.port)
	return s.server.ListenAndServe()
}

// GracefullyShutdown gracefully shuts down the HTTP API.
func (s *server) GracefullyShutdown(ctx context.Context) error {
	s.logger.Info("attempting to shutdown server gracefully")
	return s.server.Shutdown(ctx)
}

func (s *server) buildRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		s.requestLogger(s.logger),
		s.cors,
		middleware.SetHeader("content-type", "application/json"),
	)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/playground", playground.Handler("GraphQL playground", "/graphql"))

	r.Post("/external/plaid/v1/webhook", s.handlePlaidPostV1Webhook)
	r.Post("/external/plaid/v1/link/token", s.handlePlaidPostLinkToken)

	r.Post("/external/auth0/v1/exchange", s.handleAuth0PostCodeExchange)

	r.Group(func(r chi.Router) {
		r.Use(s.authorization)
		r.Get("/retool/auth", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/items", s.handleGetUserItems)
		// r.Post("/items", s.handlePostUserItems)
		r.Get("/items/{itemID}/accounts", s.handleGetItemAccounts)
		r.Get("/items/{itemID}/accounts/{accountID}", s.handleGetItemAccount)
		r.Get("/items/{itemID}", s.handleGetUserItem)
		r.Delete("/items/{itemID}", s.handleDeleteUserItem)

		r.Get("/items/{itemID}/accounts/{accountID}/transactions", s.handleGetAccountTransactions)
		r.Put("/items/{itemID}/accounts/{accountID}/transactions", s.handleUpdateTransactions)

		r.Get("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}", s.handleGetAccountTransaction)
		r.Patch("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}", s.handlePatchAccountTransaction)

		r.Get("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}/receipt", s.handleGetAccountTransactionReceiptURL)
		r.Post("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}/receipt", s.handlePostAccountTransactionReceipt)
		r.Delete("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}/receipt", s.handleDeleteAccountTransactionReceipt)

		// ##### GraphQL Handler #####
		handler := handler.New(
			generated.NewExecutableSchema(
				generated.Config{
					Resolvers: resolvers.New(
						s.logger,
						s.account,
						s.gateway,
						s.item,
						s.loaders,
						s.transaction,
					),
				},
			),
		)
		handler.AddTransport(transport.POST{})
		handler.AddTransport(transport.MultipartForm{})
		handler.Use(extension.Introspection{})
		handler.SetQueryCache(lru.New(1000))
		handler.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New(100),
		})
		r.Handle("/graphql", handler)
	})

	return r
}

// func PrintRoutes(r chi.Routes) {
// 	var printRoutes func(parentPattern string, r chi.Routes)
// 	printRoutes = func(parentPattern string, r chi.Routes) {
// 		rts := r.Routes()
// 		parentPattern = strings.TrimSuffix(parentPattern, "/*")
// 		for _, rt := range rts {
// 			if rt.SubRoutes == nil {
// 				fmt.Println(parentPattern, "+", rt.Pattern)
// 			} else {
// 				pat := rt.Pattern
// 				subRoutes := rt.SubRoutes
// 				printRoutes(parentPattern+pat, subRoutes)
// 			}
// 		}
// 	}
// 	printRoutes("", r)
// }

func closeRequestBody(ctx context.Context, r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		newrelic.FromContext(ctx).NoticeError(err)
	}
}

func (s *server) writeResponse(ctx context.Context, w http.ResponseWriter, code int, data interface{}) {

	if code != http.StatusOK {
		w.WriteHeader(code)
	}

	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (s *server) writeError(ctx context.Context, w http.ResponseWriter, code int, err error) {

	// If err is not nil, actually pass in a map so that the output to the wire is {"error": "text...."} else just let it fall through
	if err != nil {
		LogEntrySetField(ctx, "error", err)
		s.writeResponse(ctx, w, code, map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	s.writeResponse(ctx, w, code, nil)

}
