package protocol_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/platform/protocol"
	"github.com/tonto/gourmet/internal/upstream"
)

// https://www.digitalocean.com/community/tutorials/understanding-nginx-http-proxying-load-balancing-buffering-and-caching

const (
	balancerPath = "http://localhost:8080"
)

func TestHTTPUpstreamRequest(t *testing.T) {
	cases := map[string]struct {
		opts          []protocol.HTTPOption
		bl            *mockbl
		reqURL        string
		reqMtd        string
		reqBody       []byte
		headers       map[string]string
		customHeaders map[string]string
		// assert tests upstream request
		assert func(*testing.T, http.Request)
		// TODO - Add expected response
		assertResp func()
	}{
		"test automatic headers": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers",

			assert: func(t *testing.T, r http.Request) {
				assert.Equal(t, "/headers", r.URL.Path)
				assert.Equal(t, "Close", r.Header.Get("Connection"))
				assert.Equal(t, "localhost:8080", r.Header.Get("X-Forwarded-Host"))
				assert.Equal(t, true, testIP("127.0.0.1", r.Header.Get("X-Real-IP")))
				assert.Equal(t, http.NoBody, r.Body)
			},
		},
		"test headers pass": {
			bl:      &mockbl{RW: &rw{}},
			reqMtd:  "POST",
			reqURL:  "/headers",
			reqBody: []byte("test body"),
			headers: map[string]string{
				"Content-Type":  "application/json",
				"X-Some-Header": "1024",
			},
		},
		"test query params": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers?foo=bar&bar=baz",
			assert: func(t *testing.T, r http.Request) {
				assert.Equal(t, "bar", r.URL.Query().Get("foo"))
				assert.Equal(t, "baz", r.URL.Query().Get("bar"))
			},
		},
		"test request body": {
			bl:      &mockbl{RW: &rw{}},
			reqMtd:  "POST",
			reqURL:  "/headers",
			reqBody: []byte("test body"),
			assert: func(t *testing.T, r http.Request) {
				b, _ := ioutil.ReadAll(r.Body)
				assert.Equal(t, []byte("test body"), b)
			},
		},

		// TODO - Test error paths here and below
	}

	uc, close := spinUpUpstreams()
	defer close()
	m := sync.Mutex{}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := protocol.NewHTTP(c.bl, c.opts...)

			var body io.Reader
			if c.reqBody != nil {
				rb := c.reqBody
				body = bytes.NewReader(rb)
			}

			r := httptest.NewRequest(c.reqMtd, balancerPath+c.reqURL, body)
			r.Header.Set("X-Test-Name", name)
			r.RemoteAddr = "127.0.0.1"

			if c.headers != nil {
				for h, v := range c.headers {
					r.Header.Add(h, v)
				}
			}

			m.Lock()
			uc[name] = make(chan http.Request)
			m.Unlock()

			go func() { h.ServeRequest(r) }()

			time.Sleep(10 * time.Millisecond)

			req := <-uc[name]

			if c.assert != nil {
				c.assert(t, req)
			}

			if c.headers != nil {
				for h, v := range c.headers {
					assert.Equal(t, v, req.Header.Get(h))
				}
			}
		})
	}
}

// TODO - Move this to ingress_test.go
func TestHTTPUpstreamResponse(t *testing.T) {
	/*
		"test request timeout": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers",
			requestTimeout: 1*time.Nanosecond
			// TODO - Add timed out page
		},
	*/

	// test error pages in cmd when doing functional(
	// also test assemble there (input config yaml and test the whole integration)
	// Test timeout and error response codes

	// TODO - Test response body
	// test response headers
	// test response code
}

func testIP(expected, ip string) bool {
	log.Println(ip)
	ips := strings.Split(ip, ":")
	if ips[0] == expected {
		return true
	}
	return false
}

func doReq(r *http.Request) {
	http.DefaultClient.Do(r)
}

func spinUpUpstreams() (map[string]chan http.Request, func()) {
	c := make(map[string]chan http.Request)
	s := http.Server{
		Addr:    "localhost:8081",
		Handler: http.DefaultServeMux,
	}
	http.DefaultServeMux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		c[r.Header.Get("X-Test-Name")] <- *r
	})
	http.DefaultServeMux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "response body")
		c[r.Header.Get("X-Test-Name")] <- *r
	})
	http.DefaultServeMux.HandleFunc("/internal_error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "response body")
		c[r.Header.Get("X-Test-Name")] <- *r
	})
	// TODO - Add more endpoints with responses
	go func() { s.ListenAndServe() }()
	return c, func() {
		s.Shutdown(context.Background())
	}
}

type mockbl struct {
	Next *upstream.Server
	RW   http.ResponseWriter
}

type rw struct{}

func (m *mockbl) NextServer() (*upstream.Server, error) {
	m.Next = newServer()
	return m.Next, nil
}

func newServer() *upstream.Server {
	s := upstream.NewServer(
		"localhost:8081/",
		upstream.WithFailTimeout(5*time.Second),
		upstream.WithMaxFail(10),
		upstream.WithQueueSize(1),
	)

	return s
}

func (r *rw) Header() http.Header {
	return make(http.Header)
}

func (r *rw) Write([]byte) (int, error) {
	return 0, nil
}

func (r *rw) WriteHeader(int) {

}
