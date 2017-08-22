package ingres

import (
	"net/http"
	"regexp"
)

type Router struct {
	routes []*entry
}

type entry struct {
	route   *route
	handler http.Handler
}

func New() *Router {
	return &Router{}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, raw *http.Request) {
	h := r.match(raw)
	if h == nil {
		http.NotFound(w, raw)
		return
	}
	h.ServeHTTP(w, raw)
}

func (r *Router) match(req *http.Request) http.Handler {
	for _, e := range r.routes {
		if e.route.match(req.URL.Path) {
			return e.handler
		}
	}
	return nil
}

func (router *Router) AddLocation(pattern string, h http.Handler) {
	router.routes = append(router.routes, &entry{route: &route{regexp.MustCompile(pattern)}, handler: h})
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
