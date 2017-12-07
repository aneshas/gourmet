package balancer

import (
	"math/rand"
	"time"

	"github.com/tonto/gourmet/internal/upstream"
)

// NewRandom creates new Random balancer instance
func NewRandom(s []*upstream.Server) *Random {
	r := Random{
		servers: s,
	}

	return &r
}

// Random represents round robin load balancer
type Random struct {
	servers []*upstream.Server
}

// NextServer returns next available upstream server to receive traffic
func (r *Random) NextServer() (*upstream.Server, error) {
	t := time.Now()
	s := r.nextServer()
	for !s.Available() {
		if time.Since(t) > selectTiemout {
			return nil, ErrUpstreamUnavailable
		}
		s = r.nextServer()
	}
	return s, nil
}

func (r *Random) nextServer() *upstream.Server {
	return r.servers[rand.Intn(len(r.servers))]
}
