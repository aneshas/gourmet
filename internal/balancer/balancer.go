package balancer

import "github.com/tonto/gourmet/internal/upstream"

type Balancer interface {
	NextServer() *upstream.Server
}
