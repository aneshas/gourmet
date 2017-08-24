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
		// TODO - This should be configurable
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

// Weight returns weight assigned to upstream server
func (s *Server) Weight() int {
	return s.weight
}

func (s *Server) loop() {
	for r := range s.Enqueue {
		// TODO - handle panic
		r.Done <- r.F(s.uri)
	}
}
