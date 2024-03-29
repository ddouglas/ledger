package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ddouglas/ledger/internal"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Cors middleware to allow frontend consumption
func (s *server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "600")

		// Just return for OPTIONS
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var ctx = r.Context()

		authHeader := r.Header.Get("authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			s.writeError(ctx, w, http.StatusUnauthorized, fmt.Errorf("missing or invalid token"))
			return
		}

		var prefixes = []string{`bearer `, `Bearer `}
		for _, prefix := range prefixes {
			authHeader = strings.TrimPrefix(authHeader, prefix)
		}
		token, err := s.auth.ValidateToken(ctx, authHeader)
		if err != nil {
			s.writeError(ctx, w, http.StatusUnauthorized, fmt.Errorf("failed to validate token: %w", err))
			return
		}

		user, err := s.user.UserFromToken(ctx, token)
		if err != nil {
			s.writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		ctx = context.WithValue(ctx, internal.CtxUser, user)
		ctx = context.WithValue(ctx, internal.CtxToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// NewStructuredLogger is a constructor for creating a request logger middleware
func (s *server) requestLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&structuredLogger{logger})
}
