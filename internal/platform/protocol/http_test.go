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

var m sync.Mutex
var chans = make(map[string]chan epreq)

type epreq struct {
	r http.Request
	b string
}

func TestHTTPServeRequest(t *testing.T) {
	cases := map[string]struct {
		opts          []protocol.HTTPOption
		bl            *mockbl
		reqURL        string
		reqMtd        string
		reqBody       []byte
		headers       map[string]string
		wantHeaders   map[string]string
		customHeaders map[string]string
		assert        func(*testing.T, epreq)
		assertResp    func()
		wantErr       bool
	}{
		"test automatic headers": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers",
			assert: func(t *testing.T, req epreq) {
				r := req.r
				assert.Equal(t, "/headers", r.URL.Path)
				assert.Equal(t, "Close", r.Header.Get("Connection"))
				assert.Equal(t, "localhost:8080", r.Header.Get("X-Forwarded-Host"))
				assert.Equal(t, true, testIP("127.0.0.1", r.Header.Get("X-Real-IP")))
				assert.Equal(t, http.NoBody, r.Body)
			},
		},
		"test req headers": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "POST",
			reqURL: "/headers",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"X-Some-Header": "1024",
			},
			wantHeaders: map[string]string{
				"Content-Type":  "application/json",
				"X-Some-Header": "1024",
			},
		},
		"test headers pass": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "POST",
			reqURL: "/headers",
			opts: []protocol.HTTPOption{
				protocol.WithHTTPHeaders(map[string]string{
					"Content-Type":  "application/json",
					"X-Some-Header": "1024",
				}),
			},
			wantHeaders: map[string]string{
				"Content-Type":  "application/json",
				"X-Some-Header": "1024",
			},
		},
		"test query params": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "GET",
			reqURL: "/headers?foo=bar&bar=baz",
			assert: func(t *testing.T, r epreq) {
				assert.Equal(t, "bar", r.r.URL.Query().Get("foo"))
				assert.Equal(t, "baz", r.r.URL.Query().Get("bar"))
			},
		},
		"test request body": {
			bl:      &mockbl{RW: &rw{}},
			reqMtd:  "POST",
			reqURL:  "/headers",
			reqBody: []byte("test body"),
			assert: func(t *testing.T, r epreq) {
				assert.Equal(t, "test body", r.b)
			},
		},
		"test svc unavailable": {
			bl:      &mockbl{RW: &rw{}},
			reqMtd:  "POST",
			reqURL:  "/unavailable",
			reqBody: []byte("test body"),
			wantErr: true,
		},
		"test upstreams unavailable": {
			bl:      &mockbl{RW: &rw{}, Err: true},
			reqMtd:  "POST",
			reqURL:  "/unavailable",
			reqBody: []byte("test body"),
			wantErr: true,
		},
		"test req timeout": {
			bl:     &mockbl{RW: &rw{}},
			reqMtd: "POST",
			reqURL: "/timeout",
			opts: []protocol.HTTPOption{
				protocol.WithHTTPRequestTimeout(time.Millisecond),
			},
			wantErr: true,
		},
	}

	close := spinUpUpstreams()
	defer close()

	for name, c := range cases {
		m.Lock()
		chans[name] = make(chan epreq, 5)
		m.Unlock()

		t.Run(name, func(t *testing.T) {
			c := c

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

			_, err := h.ServeRequest(r)

			if c.wantErr != (err != nil) {
				t.Fatalf("error should be %v got: %v", c.wantErr, err)
			}

			if c.wantErr && (err != nil) {
				return
			}

			m.Lock()
			req := <-chans[name]
			m.Unlock()

			if c.assert != nil {
				c.assert(t, req)
			}

			if c.wantHeaders != nil {
				for h, v := range c.wantHeaders {
					assert.Equal(t, v, req.r.Header.Get(h))
				}
			}
		})
	}
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

func spinUpUpstreams() func() {
	s := http.Server{
		Addr:    "localhost:8081",
		Handler: http.DefaultServeMux,
	}

	http.DefaultServeMux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		sendReq(r)
	})

	http.DefaultServeMux.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		sendReq(r)
	})

	http.DefaultServeMux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		sendReq(r)
	})

	http.DefaultServeMux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "response body")
		sendReq(r)
	})

	http.DefaultServeMux.HandleFunc("/internal_error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "response body")
		sendReq(r)
	})

	// TODO - Add more endpoints with responses
	go func() { s.ListenAndServe() }()
	return func() {
		s.Shutdown(context.Background())
	}
}

func sendReq(r *http.Request) {
	m.Lock()
	b, _ := ioutil.ReadAll(r.Body)
	chans[r.Header.Get("X-Test-Name")] <- epreq{
		r: *r,
		b: string(b),
	}
	m.Unlock()
}

type mockbl struct {
	Next *upstream.Server
	RW   http.ResponseWriter
	Err  bool
}

func (m *mockbl) NextServer() (*upstream.Server, error) {
	if m.Err {
		return nil, fmt.Errorf("upstream unavailable")
	}
	m.Next = newServer()
	return m.Next, nil
}

func newServer() *upstream.Server {
	s := upstream.NewServer(
		"localhost:8081/",
		upstream.WithFailTimeout(10*time.Millisecond),
		upstream.WithMaxFail(10),
		upstream.WithQueueSize(1),
	)

	c := make(chan struct{})
	go func() { s.Run(c) }()

	return s
}

type rw struct{}

func (r *rw) Header() http.Header {
	return make(http.Header)
}

func (r *rw) Write([]byte) (int, error) {
	return 0, nil
}

func (r *rw) WriteHeader(int) {}
