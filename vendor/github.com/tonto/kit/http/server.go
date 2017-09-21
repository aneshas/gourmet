// Package server provides common http server functionality
package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

// NewServer creates new http server instance
func NewServer(opts ...ServerOption) *Server {
	srv := Server{
		httpServer: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		mux: mux.NewRouter().StrictSlash(true),
	}

	if srv.notFoundHandler != nil {
		srv.mux.NotFoundHandler = srv.notFoundHandler
	}

	srv.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	srv.httpServer.Handler = srv.mux

	for _, o := range opts {
		o(&srv)
	}

	if srv.logger == nil {
		srv.logger = log.New(os.Stdout, "http ", log.Ldate|log.Ltime|log.Llongfile)
	}

	if srv.tlsEnabled() {
		srv.httpServer.TLSConfig = &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		}
	}

	return &srv
}

// Server represents http server implementation
type Server struct {
	httpServer      *http.Server
	logger          *log.Logger
	certFile        string
	keyFile         string
	mux             *mux.Router
	notFoundHandler http.Handler
}

// Run will start a server listening on a given port
func (s *Server) Run(port int) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	s.httpServer.Addr = addr

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	var err error

	go func() {
		s.logger.Printf("Starting server at: %s", s.httpServer.Addr)
		if s.tlsEnabled() {
			err = s.runTLS()
		} else {
			err = s.httpServer.ListenAndServe()
		}
	}()

	if err != nil {
		return err
	}

	<-stop

	s.logger.Println("Server shutting down...")
	err = s.Stop()
	if err != nil {
		return err
	}

	s.logger.Println("Server stopped.")

	return nil
}

func (s *Server) tlsEnabled() bool {
	return (s.certFile != "" && s.keyFile != "")
}

func (s *Server) runTLS() error {
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + r.Host + r.URL.String()
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}),
	}
	go func() { log.Fatal(srv.ListenAndServe()) }()
	return s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
}

// Stop attempts to gracefully shutdown the server
func (s *Server) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}

// RegisterServices registers given http Services with
// the server and sets up routes
func (s *Server) RegisterServices(svcs ...Service) error {
	if svcs == nil {
		return fmt.Errorf("no services provided")
	}

	for _, svc := range svcs {
		err := s.RegisterService(svc)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterService registers a given http Service with
// the server and sets up routes
func (s *Server) RegisterService(svc Service) error {
	endpoints := svc.Endpoints()

	if endpoints == nil {
		return fmt.Errorf("service has no endpoints defined")
	}

	for path, endpoint := range endpoints {
		ep := "/" + path
		if path == "/" {
			ep = ""
		}
		p := fmt.Sprintf("/%s%s", svc.Prefix(), ep)
		route := s.mux.Handle(p, endpoint.Handler)
		if endpoint.Methods != nil {
			route.Methods(endpoint.Methods...)
		}
	}

	return nil
}
