package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	cases := map[string]struct {
		expectedCfg *Config
		expectedErr error
	}{
		"invalid_toml_format":      {expectedErr: errInvalidTOML},
		"upstream_err":             {expectedErr: errNoUpstreams},
		"static_provider_err":      {expectedErr: errNoServers},
		"static_provider_path_err": {expectedErr: errNoServerPath},
		"server_err":               {expectedErr: errNoServer},
		"server_locations_err":     {expectedErr: errNoServerLocations},
		"upstream_mismatch":        {expectedErr: errUpstreamMismatch},
		"defaults": {
			expectedCfg: &Config{
				Upstreams: map[string]*Upstream{
					"front":   &Upstream{Balancer: "round_robin", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo.com", Weight: 5, MaxFail: 10, FailTimeout: 1}}},
					"backend": &Upstream{Balancer: "round_robin", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo1.com", Weight: 5, MaxFail: 10, FailTimeout: 1}, &UpstreamServer{Path: "http://api.foo2.com", Weight: 0, MaxFail: 10, FailTimeout: 1}}}},
				Server: &Server{Port: 8080, Locations: []ServerLocation{ServerLocation{Path: "/api", HTTPPass: "backend"}, ServerLocation{Path: "/", HTTPPass: "front"}}},
			},
		},
		"valid": {
			expectedCfg: &Config{
				Upstreams: map[string]*Upstream{
					"front":   &Upstream{Balancer: "round_robin", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo.com", Weight: 5, MaxFail: 15, FailTimeout: 5}}},
					"backend": &Upstream{Balancer: "round_robin", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo1.com", Weight: 5, MaxFail: 10, FailTimeout: 1}, &UpstreamServer{Path: "http://api.foo2.com", Weight: 0, MaxFail: 10, FailTimeout: 1}}},
				},
				Server: &Server{Port: 80, Locations: []ServerLocation{ServerLocation{Path: "/api", HTTPPass: "backend"}, ServerLocation{Path: "/", HTTPPass: "front"}}},
			},
		},
		"valid_random": {
			expectedCfg: &Config{
				Upstreams: map[string]*Upstream{
					"front":   &Upstream{Balancer: "random", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo.com", Weight: 5, MaxFail: 15, FailTimeout: 5}}},
					"backend": &Upstream{Balancer: "round_robin", Provider: "static", Servers: []*UpstreamServer{&UpstreamServer{Path: "http://api.foo1.com", Weight: 5, MaxFail: 10, FailTimeout: 1}, &UpstreamServer{Path: "http://api.foo2.com", Weight: 0, MaxFail: 10, FailTimeout: 1}}},
				},
				Server: &Server{Port: 80, Locations: []ServerLocation{ServerLocation{Path: "/api", HTTPPass: "backend"}, ServerLocation{Path: "/", HTTPPass: "front"}}},
			},
		},
		// TODO
		// Add tests for misspelled options eg. round_rob
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cfg, err := Parse(mustOpenConfigF(t, name))
			assert.Equal(t, c.expectedCfg, cfg)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func mustOpenConfigF(t *testing.T, fname string) io.Reader {
	f, err := os.Open(filepath.Join("testdata", fname+".toml"))
	if err != nil {
		t.Fatal(err)
	}
	return f
}
