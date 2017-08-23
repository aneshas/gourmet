// Package balancer provides different balancing algorithms
package balancer

import "github.com/tonto/gourmet/internal/upstream"

// Balancer represents balancing algorithm interface
type Balancer interface {
	NextServer() *upstream.Server
}
