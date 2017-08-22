// Package upstream provides upstream request handling
package upstream

// Request represents upstream request
type Request struct {
	F    func(string) error
	Done chan error
}

// NewServer creates new upstream server instance
// and starts queue handler
func NewServer(uri string, w int) *Server {
	h := Server{
		Enqueue: make(chan *Request, 100),
		uri:     uri,
		weight:  w,
	}
	go h.loop()
	return &h
}

// Server represents upstream server abstraction
// It holds server properties and maintains a request queue
type Server struct {
	Enqueue chan *Request
	uri     string
	weight  int
}

func (h *Server) loop() {
	for r := range h.Enqueue {
		r.Done <- r.F(h.uri)
	}
}
