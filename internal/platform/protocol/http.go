package protocol

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/errors"
	"github.com/tonto/gourmet/internal/upstream"
)

// NewHTTP creates new HTTP instance
func NewHTTP(bl balancer.Balancer, opts ...HTTPOption) *HTTP {
	cfg := Config{}
	for _, o := range opts {
		o(&cfg)
	}
	s := HTTP{
		balancer: bl,
		config:   cfg,
	}
	return &s
}

// HTTP represents http upstream pass
type HTTP struct {
	balancer balancer.Balancer
	config   Config
}

// Config represents http configuration
type Config struct {
	passHeaders    map[string]string
	requestTimeout time.Duration
}

// ServeRequest implement ProtocolHandler ingress interface
func (ht *HTTP) ServeRequest(r *http.Request) (*http.Response, error) {
	var response *http.Response

	s, err := ht.balancer.NextServer()
	if err != nil {
		return nil, errors.New(
			http.StatusServiceUnavailable,
			http.StatusText(http.StatusServiceUnavailable),
			err.Error(),
		)
	}

	done := make(chan error)

	s.Enqueue <- upstream.Request{
		Done: done,
		F: func(c context.Context, uri string) error {
			resp, err := proxyPass(c, uri, r)
			if err != nil {
				return err
			}
			response = resp
			return nil
		},
	}

	err = <-done
	return response, err
}

func proxyPass(c context.Context, uri string, r *http.Request) (*http.Response, error) {
	req, err := wrapRequest(uri, r)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: 10 * time.Second,
		// TODO - Use client timeout from config
	}

	resp, err := client.Do(req.WithContext(c))
	if err != nil {
		return nil, errors.New(
			http.StatusBadGateway,
			http.StatusText(http.StatusBadGateway),
			err.Error(),
		)
	}

	return resp, nil
}

func wrapRequest(uri string, r *http.Request) (*http.Request, error) {
	// TODO - detect protocol from location http_pass xxx_pass
	uuri := "http://" + strings.TrimRight(uri, "/") + r.URL.Path
	if r.URL.RawQuery != "" {
		uuri += "?" + r.URL.RawQuery
	}
	req, err := http.NewRequest(r.Method, uuri, r.Body)
	if err != nil {
		return nil, err
	}

	for h, v := range r.Header {
		if v != nil && len(v) > 0 && v[0] != "" {
			req.Header.Add(h, v[0])
		}
	}

	req.Header.Add("Connection", "Close")
	req.Header.Add("X-Real-IP", r.RemoteAddr)
	req.Header.Add("X-Forwarded-Host", r.Host)

	return req, nil
}
