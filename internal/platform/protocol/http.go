package protocol

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tonto/gourmet/internal/balancer"
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

// ServerHTTP implements http.Handler
func (ht *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := context.Background()
	c, cancel := context.WithTimeout(c, 200*time.Millisecond)
	defer cancel()

	h := ht.balancer.NextServer()
	done := make(chan error, 1)

	h.Enqueue <- &upstream.Request{
		Done: done,
		F: func(uri string) error {
			resp, err := proxyPass(uri, r)
			if err != nil {
				return err
			}
			writeBack(w, resp)
			return nil
		},
	}

	select {
	case err := <-done:
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
	case <-c.Done():
		fmt.Fprintf(w, c.Err().Error())
	}
}

func proxyPass(uri string, r *http.Request) (*http.Response, error) {
	req, err := wrapRequest(uri, r)
	if err != nil {
		return nil, err
	}

	// TODO - don't use default client
	return http.DefaultClient.Do(req)
}

func wrapRequest(uri string, r *http.Request) (*http.Request, error) {
	// TODO - detect protocol from location http_pass xxx_pass
	uuri := "http://" + strings.TrimRight(uri, "/") + r.URL.Path
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

func writeBack(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()
	// TODO - write some headers also
	io.Copy(w, resp.Body)
	// TODO - write status
	// TODO - check if w is nil
}
