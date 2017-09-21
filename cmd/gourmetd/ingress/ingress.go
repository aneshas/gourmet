// Package ingress provides ingress routers
package ingress

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/tonto/gourmet/internal/errors"
)

// HTTP represents http ingress implementation
type HTTP struct {
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
	ServeRequest(*http.Request) (*http.Response, error)
}

// NewHTTP creates new http ingress instance
func NewHTTP(l *log.Logger) *HTTP {
	return &HTTP{
		logger: l,
	}
}

// ServeHTTP implements http.Handler
func (h *HTTP) ServeHTTP(w http.ResponseWriter, raw *http.Request) {
	// TODO - Move logging to NewLoggedIngress
	h.logger.Printf("%s %s IP: %s", raw.Method, raw.URL.Path, raw.RemoteAddr)
	hdlr := h.match(raw)
	if hdlr == nil {
		// TODO - Check if /etc/gourmet/error.tpl exists
		t := template.New("error tpl")
		t, err := t.Parse(defaultErrTpl)
		if err != nil {
			http.NotFound(w, raw)
			return
		}
		t.Execute(w, errors.NewHTTP(
			http.StatusNotFound,
			http.StatusText(http.StatusNotFound),
			"The path "+raw.URL.Path+" could not be found on the server.",
		))
		return
	}

	resp, err := hdlr.ServeRequest(raw)

	select {
	case <-raw.Context().Done():
		return
	default:
		if err != nil {
			// TODO - Write err page (or json)
			switch raw.Header.Get("Accept") {
			case "application/json":
				h.writerJSONErr(w, err)
			default:
				h.writerErr(w, err)
			}
			return
		}
		// TODO - write some headers also
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	}

}

func (h *HTTP) match(req *http.Request) ProtocolHandler {
	for _, e := range h.routes {
		if e.route.match(req.URL.Path) {
			return e.handler
		}
	}
	return nil
}

func (h *HTTP) writerJSONErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")
	e := interface{}(err)
	ge, ok := e.(*errors.HTTPError)
	if !ok {
		fmt.Fprintf(w, `{"status":500,"status_text":"internal server error"}`)
		return
	}
	data, err := json.Marshal(ge)
	if err != nil {
		fmt.Fprintf(w, `{"status":500,"status_text":"internal server error"}`)
		return
	}
	w.Write(data)
}

// TODO - pass err template
func (h *HTTP) writerErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/html")
	e := interface{}(err)
	ge, ok := e.(*errors.HTTPError)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal error occured")
		return
	}
	t := template.New("error tpl")
	t, err = t.Parse(defaultErrTpl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, ge.Error())
		return
	}
	w.WriteHeader(ge.Status)
	t.Execute(w, ge)
}

// AddLocation registers location path regex with a location handler
func (h *HTTP) AddLocation(pattern string, hdlr ProtocolHandler) {
	h.routes = append(h.routes, &entry{route: &route{regexp.MustCompile(pattern)}, handler: hdlr})
}

type route struct {
	*regexp.Regexp
}

func (r *route) match(target string) bool {
	if m := r.FindStringSubmatch(target); m != nil {
		return true
	}
	return false
}
