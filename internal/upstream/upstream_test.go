package upstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerHealth(t *testing.T) {
	cases := map[string]struct {
		maxFail              int
		numReqs              int
		failTimeout          time.Duration
		wait                 time.Duration
		weight               int
		buffSz               int
		expectedAvailability bool
		expectedErr          error
		req                  func(int) *Request
	}{
		"available": {
			maxFail:              10,
			numReqs:              50,
			failTimeout:          1 * time.Second,
			weight:               3,
			buffSz:               10,
			expectedAvailability: true,
			wait:                 100 * time.Millisecond,
			req: func(it int) *Request {
				return &Request{
					F: func(context.Context, string) error {
						return nil
					},
					Done: make(chan error, 1),
				}
			},
		},
		"available with some errs": {
			maxFail:              10,
			numReqs:              50,
			failTimeout:          1 * time.Second,
			weight:               3,
			buffSz:               10,
			expectedAvailability: true,
			wait:                 1 * time.Second,
			req: func(it int) *Request {
				return &Request{
					F: func(context.Context, string) error {
						if it < 9 {
							return fmt.Errorf("some upstream err")
						}
						return nil
					},
					Done: make(chan error, 1),
				}
			},
		},
		"available with some timeouts": {
			maxFail:              3,
			numReqs:              10,
			failTimeout:          100 * time.Millisecond,
			weight:               3,
			buffSz:               10,
			expectedAvailability: true,
			wait:                 200 * time.Millisecond,
			req: func(it int) *Request {
				return &Request{
					F: func(c context.Context, u string) error {
						if it < 2 {
							time.Sleep(300 * time.Millisecond)
							return nil
						}
						if c.Err() != nil {
							return ErrPassiveHealthCheck
						}
						return nil
					},
					Done: make(chan error, 1),
				}
			},
		},
		"unavailable due to errs": {
			maxFail:              10,
			numReqs:              50,
			failTimeout:          1 * time.Second,
			weight:               3,
			buffSz:               10,
			expectedAvailability: false,
			wait:                 1 * time.Second,
			req: func(it int) *Request {
				return &Request{
					F: func(context.Context, string) error {
						return fmt.Errorf("some upstream err")
					},
					Done: make(chan error, 1),
				}
			},
		},
		"regain health": {
			maxFail:              3,
			numReqs:              1000,
			failTimeout:          1 * time.Second,
			weight:               3,
			buffSz:               10,
			expectedAvailability: true,
			wait:                 2200 * time.Millisecond,
			req: func(it int) *Request {
				t := time.Now()
				return &Request{
					F: func(c context.Context, u string) error {
						if time.Since(t) > (1 * time.Second) {
							return nil
						}
						return fmt.Errorf("some upstream err")
					},
					Done: make(chan error, 1),
				}
			},
		},
		// test panics on a higher level ??
		// test panics recover (don't consider this as unhealthy server - test)
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			srv := NewServer(
				"foo.com",
				WithWeight(c.weight),
				WithFailTimeout(c.failTimeout),
				WithMaxFail(c.maxFail),
				WithQueueSize(c.buffSz),
			)

			cc := make(chan struct{})
			go srv.Run(cc)

			for i := 0; i < c.numReqs; i++ {
				req := *c.req(i)
				srv.Work <- req
				e := <-req.Done
				if c.expectedErr != nil {
					assert.Equal(t, c.expectedErr, e)
				}
			}
			time.Sleep(c.wait)
			assert.Equal(t, c.expectedAvailability, srv.Available())
			assert.Equal(t, c.weight, srv.Weight())

			cc <- struct{}{}
		})
	}
}
