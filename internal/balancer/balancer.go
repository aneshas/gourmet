// Package balancer provides different balancing algorithms
package balancer

import (
	"errors"

	"github.com/tonto/gourmet/internal/upstream"
)

var (
	// ErrUpstreamUnavailable is returnd when all servers in a
	// given upstream are in a unhealthy state, thus no server can be selected
	ErrUpstreamUnavailable = errors.New("all upstream servers are in an unhealthy state")
)

// Balancer represents balancing algorithm interface
type Balancer interface {
	NextServer() (*upstream.Server, error)
}
