package protocol_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/platform/protocol"
	"github.com/tonto/gourmet/internal/upstream"
)

// https://www.digitalocean.com/community/tutorials/understanding-nginx-http-proxying-load-balancing-buffering-and-caching
/*
Nginx gets rid of any empty headers. There is no point of passing along empty values to another server; it would only serve to bloat the request.
Nginx, by default, will consider any header that contains underscores as invalid. It will remove these from the proxied request. If you wish to have Nginx interpret these as valid, you can set the underscores_in_headers directive to "on", otherwise your headers will never make it to the backend server.
The "Host" header is re-written to the value defined by the $proxy_host variable. This will be the IP address or name and port number of the upstream, directly as defined by the proxy_pass directive.
The "Connection" header is changed to "close". This header is used to signal information about the particular connection established between two parties. In this instance, Nginx sets this to "close" to indicate to the upstream server that this connection will be closed once the original request is responded to. The upstream should not expect this connection to be persistent.
*/

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
		assert func(*testing.T, *http.Request)
		// TODO - Add expected response
		assertResp func()
	}{
		"test automatic headers": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers",

			assert: func(t *testing.T, r *http.Request) {
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

		// TODO - Test pass custom headers

		"test request body": {
			bl:      &mockbl{RW: &rw{}},
			reqMtd:  "POST",
			reqURL:  "/headers",
			reqBody: []byte("test body"),
			assert: func(t *testing.T, r *http.Request) {
				b, _ := ioutil.ReadAll(r.Body)
				assert.Equal(t, []byte("test body"), b)
			},
		},

		"test request timeout": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers",
			opts:   []protocol.HTTPOption{protocol.WithHTTPRequestTimeout(5 * time.Millisecond)},
			// TODO - Add timed out page
		},
	}

	// TODO
	// Test POST body
	// test error pages

	uc, close := spinUpUpstreams()
	defer close()

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := protocol.NewHTTP(c.bl, c.opts...)

			var body io.Reader
			if c.reqBody != nil {
				body = bytes.NewReader(c.reqBody)
			}

			r := httptest.NewRequest(c.reqMtd, balancerPath+c.reqURL, body)
			r.RemoteAddr = "127.0.0.1"

			if c.headers != nil {
				for h, v := range c.headers {
					r.Header.Add(h, v)
				}
			}

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			req := <-uc

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

func TestHTTPUpstreamResponse(t *testing.T) {

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

func spinUpUpstreams() (chan *http.Request, func()) {
	c := make(chan *http.Request)
	s := http.Server{
		Addr:    "localhost:8081",
		Handler: http.DefaultServeMux,
	}
	http.DefaultServeMux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		c <- r
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

func (m *mockbl) NextServer() *upstream.Server {
	m.Next = newServer()
	return m.Next
}

func newServer() *upstream.Server {
	s := upstream.Server{
		Enqueue: make(chan *upstream.Request, 1),
	}
	go func() {
		for r := range s.Enqueue {
			r.Done <- r.F("localhost:8081/")
		}
	}()
	return &s
}

func (r *rw) Header() http.Header {
	return make(http.Header)
}

func (r *rw) Write([]byte) (int, error) {
	return 0, nil
}

func (r *rw) WriteHeader(int) {

}
