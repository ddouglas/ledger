package server

import (
	"net/http"

	"github.com/ddouglas/ledger/internal/auth"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/ddouglas/ledger/internal/user"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type configOption func(s *server)

type server struct {
	port             uint
	auth0ServerToken string
	logger           *logrus.Logger
	auth             auth.Service
	importer         importer.Service
	gateway          gateway.Service
	newrelic         *newrelic.Application
	user             user.Service

	server *http.Server
}

func WithNewrelic(app *newrelic.Application) configOption {
	return func(s *server) {
		s.newrelic = app
	}
}

func WithPort(port uint) configOption {
	return func(s *server) {
		s.port = port
	}
}

func WithAuth0ServerToken(token string) configOption {
	return func(s *server) {
		s.auth0ServerToken = token
	}
}

func WithLogger(logger *logrus.Logger) configOption {
	return func(s *server) {
		s.logger = logger
	}
}

func WithAuth(auth auth.Service) configOption {
	return func(s *server) {
		s.auth = auth
	}
}

func WithGateway(gateway gateway.Service) configOption {
	return func(s *server) {
		s.gateway = gateway
	}
}

func WithUser(user user.Service) configOption {
	return func(s *server) {
		s.user = user
	}
}

func WithImporter(importer importer.Service) configOption {
	return func(s *server) {
		s.importer = importer
	}
}
