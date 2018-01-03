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
	h := HTTP{
		balancer: bl,
		config:   cfg,
	}
	return &h
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

// ServeRequest passes request to upstream server
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

	s.Work <- upstream.Request{
		Done: done,
		F: func(c context.Context, uri string) error {
			resp, err := ht.proxyPass(c, uri, r)
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

func (ht *HTTP) proxyPass(c context.Context, uri string, r *http.Request) (*http.Response, error) {
	req, err := ht.wrapRequest(uri, r)
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

	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, errors.New(
			resp.StatusCode,
			resp.Status,
			"upstream server unavailable",
		)
	}

	return resp, nil
}

func (ht *HTTP) wrapRequest(uri string, r *http.Request) (*http.Request, error) {
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

	if ht.config.passHeaders != nil {
		for h, v := range ht.config.passHeaders {
			req.Header.Add(h, v)
		}
	}

	req.Header.Add("Connection", "Close")
	req.Header.Add("X-Real-IP", r.RemoteAddr)
	req.Header.Add("X-Forwarded-Host", r.Host)

	return req, nil
}
