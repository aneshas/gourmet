package ingres

import (
	"net/http"
	"regexp"
)

type RegExRouter struct {
	routes []*entry
}

type entry struct {
	route   *route
	handler http.Handler
}

func NewRegEx() *RegExRouter {
	return &RegExRouter{}
}

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

func (router *RegExRouter) AddLocation(pattern string, h http.Handler) {
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
