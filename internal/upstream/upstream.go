package upstream

type Algorithm int

type Upstream struct {
	Name    string
	Servers []*Server
}

type Request struct {
	F    func(string) error
	Done chan error
}

func NewServer(uri string, w int) *Server {
	h := Server{
		Enqueue: make(chan *Request, 100),
		uri:     uri,
		weight:  w,
	}
	go h.loop()
	return &h
}

type Server struct {
	Enqueue  chan *Request
	reqCount int32
	uri      string
	weight   int
}

func (h *Server) loop() {
	for r := range h.Enqueue {
		r.Done <- r.F(h.uri)
	}
}
