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
func (h *Ingress) ServeHTTP(w http.ResponseWriter, raw *http.Request) {
	h.logger.Printf("%s %s IP: %s", raw.Method, raw.URL.Path, raw.RemoteAddr)
	hdlr := h.match(raw)
	if hdlr == nil {
		h.logger.Printf("Response: %d %s", http.StatusNotFound, http.StatusText(http.StatusNotFound))
		// TODO - Check if /etc/gourmet/error.tpl exists
		t := template.New("error tpl")
		t, err := t.Parse(defaultErrTpl)
		if err != nil {
			http.NotFound(w, raw)
			return
		}
		t.Execute(w, errors.New(
			http.StatusNotFound,
			http.StatusText(http.StatusNotFound),
			"The path "+raw.URL.Path+" could not be found on the server.",
		))
		return
	}
	h.handleReq(w, raw, hdlr)
}

func (h *Ingress) match(r *http.Request) ProtocolHandler {
	for _, e := range h.routes {
		if path, ok := e.route.match(r.URL.Path); ok {
			r.URL.Path = "/" + path
			return e.handler
		}
	}
	return nil
}

func (h *Ingress) handleReq(w http.ResponseWriter, r *http.Request, ph ProtocolHandler) {
	resp, err := ph.ServeRequest(r)
	select {
	case <-r.Context().Done():
		return
	default:
		if err != nil {
			switch r.Header.Get("Accept") {
			case "application/json":
				h.writerJSONErr(w, err)
			default:
				h.writerErr(w, err)
			}
			return
		}
		// TODO - write some headers also
		defer resp.Body.Close()
		h.logger.Printf("Response: %d %s", resp.StatusCode, resp.Status)
		io.Copy(w, resp.Body)
	}
}

func (h *Ingress) writerJSONErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")
	e := interface{}(err)
	ge, ok := e.(*errors.Error)
	if !ok {
		h.writeInternalErr(w)
		return
	}
	h.logger.Printf("Response: %d %s", ge.Status, ge.StatusText)
	data, err := json.Marshal(ge)
	if err != nil {
		h.writeInternalErr(w)
		return
	}
	w.Write(data)
}

func (h *Ingress) writeInternalErr(w http.ResponseWriter) {
	h.logger.Printf("Response: %d %s", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"status":500,"status_text":"internal server error"}`)
}

// TODO - pass err template
func (h *Ingress) writerErr(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/html")
	e := interface{}(err)
	ge, ok := e.(*errors.Error)
	if !ok {
		// TODO - Move this to writeInternalErr
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Printf("Response: %d %s", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		fmt.Fprintf(w, "Internal error occured")
		return
	}
	h.logger.Printf("Response: %d %s", ge.Status, ge.StatusText)
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

// RegisterLocProto registers location regex path with a location protocol handler
func (h *Ingress) RegisterLocProto(pattern string, hdlr ProtocolHandler) {
	h.routes = append(h.routes, &entry{route: &route{regexp.MustCompile(pattern)}, handler: hdlr})
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
