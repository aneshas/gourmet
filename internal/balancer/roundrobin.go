// Package balancer provides different balancing algorithms
package balancer

import (
	"sync"

	"github.com/tonto/gourmet/internal/upstream"
)

// NewRoundRobin creates new RoundRobin instance
func NewRoundRobin(s []*upstream.Server) *RoundRobin {
	rr := RoundRobin{
		servers: s,
	}
	return &rr
}

// RoundRobin presents round robin load balancing algorithm
type RoundRobin struct {
	servers []*upstream.Server
	next    int32
	m       sync.Mutex
}

// NextHost returns next available host to receive traffic
func (rr *RoundRobin) NextServer() *upstream.Server {
	rr.m.Lock()
	defer rr.m.Unlock()

	i := rr.next

	if i >= int32(len(rr.servers)) {
		rr.next = 1
		return rr.servers[0]
	}

	rr.next += 1
	return rr.servers[i]
}
