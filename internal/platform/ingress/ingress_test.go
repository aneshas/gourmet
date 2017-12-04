package ingress

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/errors"
)

func TestIngress(t *testing.T) {
	cases := []struct {
		name        string
		url         string
		body        string
		json        bool
		method      string
		want        string
		wantCode    int
		wantContent string
	}{
		// Test r context timeout
		// Test not found text resp
		// ...
		{
			name:     "test get route",
			method:   "GET",
			url:      "/api/foo",
			want:     "/foo",
			wantCode: http.StatusOK,
		},
		{
			name:     "test get route full path",
			method:   "GET",
			url:      "/static/bar/baz",
			want:     "/static/bar/baz",
			wantCode: http.StatusOK,
		},
		{
			name:     "test post route",
			method:   "POST",
			url:      "/api/foo",
			body:     "post body",
			want:     "post body",
			wantCode: http.StatusOK,
		},
		{
			name:     "test json 404",
			method:   "GET",
			url:      "/404/foo",
			want:     `{"error":"not found"}`,
			json:     true,
			wantCode: http.StatusOK,
		},
		{
			name:     "test json error",
			method:   "GET",
			url:      "/api/jsonerr",
			want:     `{"status":400,"status_text":"Bad Request","description":"/jsonerr"}`,
			json:     true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "test internal error",
			method:   "GET",
			url:      "/api/internalerr",
			want:     `{"status":500,"status_text":"internal server error"}`,
			json:     true,
			wantCode: http.StatusInternalServerError,
		},
	}
	igr := makeigr()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var b io.Reader
			w := httptest.NewRecorder()
			if c.method != "GET" {
				b = strings.NewReader(c.body)
			}
			r, err := http.NewRequest(c.method, c.url, b)
			if err != nil {
				log.Fatal(err)
			}
			if c.json {
				r.Header.Add("Accept", "application/json")
			}
			igr.ServeHTTP(w, r)
			body, _ := ioutil.ReadAll(w.Body)
			assert.Equal(t, []byte(c.want), body)
			assert.Equal(t, c.wantCode, w.Code)
		})
	}
}

func makeigr() *Ingress {
	igr := New(log.New(os.Stdout, "ingress test => ", log.Ldate|log.Ltime|log.Lshortfile))
	igr.RegisterLocHandler("api/(.+)/?", phandler{})
	igr.RegisterLocHandler("(static/bar/.+/?)", phandler{})
	return igr
}

type phandler struct{}

func (ph phandler) ServeRequest(r *http.Request) (*http.Response, error) {
	// TODO - Return err based on url path
	switch r.URL.Path {
	case "/jsonerr":
		return nil, errors.New(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), r.URL.Path)
	case "/internalerr":
		return nil, fmt.Errorf("internal err")
	}

	resp := http.Response{
		StatusCode: http.StatusOK,
	}

	if r.Method == "GET" {
		resp.Body = makeBody(r.URL.Path)
	} else {
		resp.Body = r.Body
	}

	return &resp, nil
}

func makeBody(str string) io.ReadCloser {
	return rbody{
		body: strings.NewReader(str),
	}
}

type rbody struct {
	body io.Reader
}

func (rb rbody) Read(p []byte) (n int, err error) {
	return rb.body.Read(p)
}

func (rb rbody) Close() error { return nil }
