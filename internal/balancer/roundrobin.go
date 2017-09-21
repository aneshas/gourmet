package balancer

import (
	"sync"

	"github.com/tonto/gourmet/internal/upstream"
)

// NewRoundRobin creates new RoundRobin instance
func NewRoundRobin(s []*upstream.Server) *RoundRobin {
	bl := RoundRobin{
		servers: s,
		wmap:    make(map[*upstream.Server]int),
	}

	return &bl
}

// RoundRobin represents round robin load balancer
type RoundRobin struct {
	servers []*upstream.Server
	wmap    map[*upstream.Server]int
	next    int32
	m       sync.Mutex
}

// NextServer returns next available upstream server to receive traffic
func (bl *RoundRobin) NextServer() (*upstream.Server, error) {
	// TODO - Next server has to be selected within 100ms otherwise throw unhealthy err
	s := bl.nextServer()
	for !s.Available() {
		s = bl.nextServer()
	}
	return s, nil
}

func (bl *RoundRobin) nextServer() *upstream.Server {
	bl.m.Lock()
	defer bl.m.Unlock()

	i := bl.next
	next := bl.next + 1

	if i >= int32(len(bl.servers)) {
		next = 1
		i = 0
	}

	cs := bl.servers[i]

	if cs.Weight() > 1 {
		nc := bl.wmap[cs] + 1
		if nc == cs.Weight() {
			bl.wmap[cs] = 0
		}
		if nc < cs.Weight() {
			next = i
			bl.wmap[cs]++
		}
	}

	bl.next = next

	return bl.servers[i]
}
