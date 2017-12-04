package balancer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/upstream"
)

func TestRoundRobin(t *testing.T) {
	cases := map[string]struct {
		servers func() ([]*upstream.Server, []*upstream.Server)
		n       int
		wantErr error
	}{
		"weightless": {
			n: 5,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, false)
				return s, []*upstream.Server{s[0], s[1], s[2], s[3], s[4]}
			},
		},
		"weightless health fail": {
			n: 5,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, false)
				failServer(t, s[1])
				failServer(t, s[2])
				return s, []*upstream.Server{s[0], s[3], s[4], s[0], s[3]}
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
		"weighted overflow health fail": {
			n: 11,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(5, true)
				failServer(t, s[1])
				failServer(t, s[4])
				return s, []*upstream.Server{s[0], s[2], s[2], s[3], s[3], s[3], s[0], s[2], s[2], s[3], s[3]}
			},
		},
		"all unhealthy": {
			n: 11,
			servers: func() ([]*upstream.Server, []*upstream.Server) {
				s := dummyServers(2, true)
				failServer(t, s[0])
				failServer(t, s[1])
				return s, nil
			},
			wantErr: balancer.ErrUpstreamUnavailable,
		},
	}

	// TODO - Test all servers unhealthy
	// TODO Test concurrently

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sv, eseq := c.servers()
			bl := balancer.NewRoundRobin(sv)
			var seq []*upstream.Server
			for i := 0; i < c.n; i++ {
				srv, err := bl.NextServer()
				assert.Equal(t, c.wantErr, err)
				if srv != nil {
					seq = append(seq, srv)
				}
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
		srv := upstream.NewServer(
			"http://host1.com",
			upstream.WithWeight(wg),
			upstream.WithFailTimeout(10*time.Millisecond),
			upstream.WithMaxFail(1),
			upstream.WithQueueSize(5),
		)
		c := make(chan struct{})
		go func() { srv.Run(c) }()
		s = append(
			s,
			srv,
		)
	}
	return s
}

func failServer(t *testing.T, s *upstream.Server) {
	assert.True(t, s.Available())
	done := make(chan error)
	s.Work <- upstream.Request{
		Done: done,
		F: func(c context.Context, uri string) error {
			return fmt.Errorf("foo error")
		},
	}
	<-done
	assert.False(t, s.Available())
}
