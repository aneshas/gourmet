package balancer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/upstream"
)

func TestRoundRobin(t *testing.T) {
	cases := map[string]struct {
		servers func() ([]*upstream.Server, []*upstream.Server)
		n       int
	}{
		"weightless": {
			n: 5,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, false)
				return s, []*upstream.Server{s[0], s[1], s[2], s[3], s[4]}
			},
		},
		"weightless overflow": {
			n: 7,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, false)
				return s, []*upstream.Server{s[0], s[1], s[2], s[3], s[4], s[0], s[1]}
			},
		},
		"weighted": {
			n: 5,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, true)
				return s, []*upstream.Server{s[0], s[1], s[2], s[2], s[3]}
			},
		},
		"weighted overflow": {
			n: 11,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(4, true)
				return s, []*upstream.Server{s[0], s[1], s[2], s[2], s[3], s[3], s[3], s[0], s[1], s[2], s[2]}
			},
		},
	}

	// TODO Test concurrently

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sv, eseq := c.servers()
			bl := balancer.NewRoundRobin(sv)
			var seq []*upstream.Server
			for i := 0; i < c.n; i++ {
				seq = append(seq, bl.NextServer())
			}
			assert.Equal(t, eseq, seq)
		})
	}
}

func dummyServers(n int, w bool) []*upstream.Server {
	var s []*upstream.Server
	for i := 0; i < n; i++ {
		wg := 0
		if w {
			wg = i
		}
		s = append(s, upstream.NewServer("http://host1.com", wg))
	}
	return s
}
