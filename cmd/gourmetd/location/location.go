package location

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/upstream"
)

func New(bl balancer.Balancer) *Location {
	s := Location{bl}
	return &s
}

type Location struct {
	balancer balancer.Balancer
}

func (l *Location) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := context.Background()
	c, cancel := context.WithTimeout(c, 200*time.Millisecond)
	defer cancel()

	h := l.balancer.NextServer()
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
	// TODO - append original req path to uri
	req, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Forwarded-From", "XXX")
	return req, nil
}

func writeBack(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()
	// TODO - write some headers also
	io.Copy(w, resp.Body)
	// TODO - write status
	// TODO - check if w is nil
}
