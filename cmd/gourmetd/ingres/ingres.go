// Package ingres provides ingres routers
package ingres

import (
	"net/http"
	"regexp"
)

// RegExRouter represents regular expression router
type RegExRouter struct {
	routes []*entry
}

type entry struct {
	route   *route
	handler http.Handler
}

// NewRegEx creates new router instance
func NewRegEx() *RegExRouter {
	return &RegExRouter{}
}

// ServeHTTP implements http.Handler
func (r *RegExRouter) ServeHTTP(w http.ResponseWriter, raw *http.Request) {
	h := r.match(raw)
	if h == nil {
		http.NotFound(w, raw)
		return
	}
	h.ServeHTTP(w, raw)
}

func (r *RegExRouter) match(req *http.Request) http.Handler {
	for _, e := range r.routes {
		if e.route.match(req.URL.Path) {
			return e.handler
		}
	}
	return nil
}

// AddLocation registers location path regex with a location handler
func (r *RegExRouter) AddLocation(pattern string, h http.Handler) {
	r.routes = append(r.routes, &entry{route: &route{regexp.MustCompile(pattern)}, handler: h})
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
