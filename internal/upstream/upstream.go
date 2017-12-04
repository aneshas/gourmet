// Package upstream provides upstream request handling
package upstream

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	// ErrPanic represents panic error
	ErrPanic = errors.New("request terminated due to system panic")

	// ErrPassiveHealthCheck represents passive health check fail
	ErrPassiveHealthCheck = errors.New("request failed passive health check timeout")
)

// Request represents upstream request
type Request struct {
	F    func(context.Context, string) error
	Done chan error
}

// ServerConfig represents upstream server configuration
type ServerConfig struct {
	weight        int
	maxFail       int
	failTimeout   time.Duration
	queueBufferSz int
}

// NewServer creates new upstream server instance
// and starts queue handler
func NewServer(uri string, opts ...ServerOption) *Server {
	cfg := ServerConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	h := Server{
		available: true,
		Work:      make(chan Request, cfg.queueBufferSz),
		uri:       uri,
		config:    cfg,
	}

	return &h
}

// Server represents upstream server abstraction
// It holds server properties and maintains a request queue
type Server struct {
	Work      chan Request
	uri       string
	currFail  int32
	config    ServerConfig
	available bool
}

// Available returns a bool indicating wether
// a server is available to receive requests
func (s *Server) Available() bool { return s.available }

// Weight returns weight assigned to upstream server
func (s *Server) Weight() int { return s.config.weight }

// Run runs a server
// It is designed to be run async and closed by sending to c chan
func (s *Server) Run(c chan struct{}) {
	ticker := time.NewTicker(s.config.failTimeout)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.currFail >= int32(s.config.maxFail) {
				s.available = false
			} else {
				s.available = true
			}
			atomic.StoreInt32(&s.currFail, 0)
		case r := <-s.Work:
			// This timeout should probably be a lot shorter and configurable
			ctx, cancel := context.WithTimeout(context.Background(), s.config.failTimeout)
			defer cancel()

			err := r.F(ctx, s.uri)
			if err != nil {
				atomic.AddInt32(&s.currFail, 1)
			}

			r.Done <- err
		case <-c:
			return
		}
	}
}
