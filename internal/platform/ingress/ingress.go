// Package ingress provides ingress routers
package ingress

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/tonto/gourmet/internal/errors"
)

// Ingress represents net/http ingress implementation
type Ingress struct {
	routes []*entry
	logger *log.Logger
}

type entry struct {
	route   *route
	handler ProtocolHandler
}

// ProtocolHandler represents an interface for protocol handlers
type ProtocolHandler interface {
	// ServerRequest implementations should return an upstream
	// http response ready to send back to clint and a nil err
	// Maybe it should receive custom Request that holds
	// headers, body, path - only stuff relevant (so this ingress can be reused with other protocols)
	ServeRequest(*http.Request) (*http.Response, error)
}

// New creates new http ingress instance
func New(l *log.Logger) *Ingress {
	return &Ingress{
		logger: l,
	}
}

// ServeHTTP implements http.Handler
func (igr *Ingress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	igr.logger.Printf("%s %s IP: %s", r.Method, r.URL.Path, r.RemoteAddr)

	ph, err := igr.match(r)
	if err != nil {
		igr.writeRouteErr(w, r)
		return
	}

	igr.handleReq(w, r, ph)
}

func (igr *Ingress) match(r *http.Request) (ProtocolHandler, error) {
	for _, e := range igr.routes {
		str := r.Host + r.URL.Path
		if path, ok := e.route.match(str); ok {
			r.URL.Path = "/" + path
			return e.handler, nil
		}
	}
	return nil, fmt.Errorf("no matching route")
}

func (igr *Ingress) writeRouteErr(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error":"not found"}`)
		return
	}

	w.WriteHeader(http.StatusNotFound)

	err := writeErrTpl(
		w,
		errors.New(
			http.StatusNotFound,
			http.StatusText(http.StatusNotFound),
			"the path "+r.URL.Path+" could not be found on the server.",
		),
	)

	if err != nil {
		http.NotFound(w, r)
	}
}

func (igr *Ingress) handleReq(w http.ResponseWriter, r *http.Request, ph ProtocolHandler) {
	resp, err := ph.ServeRequest(r)
	select {
	case <-r.Context().Done():
		return
	default:
		if err != nil {
			switch r.Header.Get("Accept") {
			case "application/json":
				igr.writerJSONErr(w, err)
			default:
				igr.writerTextErr(w, err)
			}
			return
		}
		// TODO - write some headers also
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	}
}

func (igr *Ingress) writerJSONErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")

	e := interface{}(err)
	ge, ok := e.(*errors.Error)
	if !ok {
		igr.writeInternalErr(w)
		return
	}

	data, err := json.Marshal(ge)
	if err != nil {
		igr.writeInternalErr(w)
		return
	}

	w.WriteHeader(ge.Status)
	w.Write(data)
}

func (igr *Ingress) writerTextErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/html")

	e := interface{}(err)
	gerr, ok := e.(*errors.Error)
	if !ok {
		igr.writeInternalErr(w)
		return
	}

	w.WriteHeader(gerr.Status)

	err = writeErrTpl(w, gerr)
	if err != nil {
		igr.writeInternalErr(w)
	}
}

func (igr *Ingress) writeInternalErr(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"status":500,"status_text":"internal server error"}`)
}

// RegisterLocHandler registers location regex path with a location protocol handler
func (igr *Ingress) RegisterLocHandler(pattern string, ph ProtocolHandler) {
	igr.routes = append(igr.routes, &entry{route: &route{regexp.MustCompile(pattern)}, handler: ph})
}

type route struct {
	*regexp.Regexp
}

func (r *route) match(target string) (string, bool) {
	if m := r.FindStringSubmatch(target); m != nil {
		return m[len(m)-1], true
	}
	return "", false
}
