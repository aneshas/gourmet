package http

import (
	"log"

	"github.com/tonto/kit/http/middleware"
)

// ServerOption is used for setting up
// server configuration
type ServerOption func(*Server)

// WithMiddleware represents server option for setting up
// pre request middlewares
func WithMiddleware(m ...middleware.Adapter) ServerOption {
	return func(s *Server) {
		s.httpServer.Handler = middleware.Adapt(s.httpServer.Handler, m...)
	}
}

// WithLogger is used for setting up server logger
func WithLogger(l *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

func WithTLS(cert, key string) ServerOption {
	return func(s *Server) {
		s.certFile = cert
		s.keyFile = key
		s.tls = true
	}
}