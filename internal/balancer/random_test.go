package balancer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/upstream"
)

func TestRandom(t *testing.T) {
	cases := map[string]struct {
		servers func() []*upstream.Server
		n       int
		wantErr error
	}{
		"single server": {
			n: 5,
			servers: func() []*upstream.Server {
				return dummyServers(1, false)
			},
		},
		"multiple servers fail": {
			n: 5,
			servers: func() []*upstream.Server {
				s := dummyServers(2, false)
				failServer(t, s[0])
				failServer(t, s[1])
				return s
			},
			wantErr: balancer.ErrUpstreamUnavailable,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := c.servers()
			time.Sleep(240 * time.Millisecond)
			bl := balancer.NewRandom(s)
			for i := 0; i < c.n; i++ {
				_, err := bl.NextServer()
				assert.Equal(t, c.wantErr, err)
			}
		})
	}
}
