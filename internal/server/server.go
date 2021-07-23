package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func New(optFuncs ...configOption) *server {

	s := &server{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.buildRouter(),
	}

	return s
}

func (s *server) Run() error {
	s.logger.WithField("service", "Server").Infof("Starting on Port %d", s.port)
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

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(s.authorization)
			r.Get("/items", s.handleGetUserItems)
			r.Post("/items", s.handlePostUserItems)
			r.Get("/items/{itemID}/accounts", s.handleGetItemAccounts)
			r.Get("/items/{itemID}/accounts/{accountID}", s.handleGetItemAccount)
			r.Get("/items/{itemID}", s.handleGetUserItem)
			r.Delete("/items/{itemID}", s.handleDeleteUserItem)

			r.Get("/items/{itemID}/accounts/{accountID}/transactions", s.handleGetAccountTransactions)
			r.Put("/items/{itemID}/accounts/{accountID}/transactions", s.handleUpdateTransactions)

			r.Get("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}", s.handleGetAccountTransaction)
			r.Patch("/items/{itemID}/accounts/{accountID}/transactions/{transactionID}", nil)
		})

		r.Route("/external", func(r chi.Router) {
			r.Route("/plaid", func(r chi.Router) {
				r.Route("/v1", func(r chi.Router) {
					r.Post("/webhook", s.handlePlaidPostV1Webhook)

					r.Group(func(r chi.Router) {
						r.Use(s.authorization)

						r.Get("/link/token", s.handlePlaidGetLinkToken)
					})

				})
			})
			r.Route("/auth0", func(r chi.Router) {
				r.Route("/v1", func(r chi.Router) {
					r.Post("/login", s.handleAuth0PostEmailExchange)
					r.Post("/exchange", s.handleAuth0PostCodeExchange)
				})
			})
		})
	})
	return r
}

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
