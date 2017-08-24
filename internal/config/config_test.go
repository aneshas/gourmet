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
		"upstream_err":             {expectedErr: errNoUpstreams},
		"static_provider_err":      {expectedErr: errNoServers},
		"static_provider_path_err": {expectedErr: errNoServerPath},
		"server_err":               {expectedErr: errNoServer},
		"server_locations_err":     {expectedErr: errNoServerLocations},
		"upstream_mismatch":        {expectedErr: errUpstreamMismatch},
		"defaults": {
			expectedCfg: &Config{8080, &config{Upstreams: map[string]*cfgUpstream{"front": &cfgUpstream{Balancer: "round_robin", Provider: "static", Servers: []upstreamServer{upstreamServer{Path: "http://api.foo.com", Weight: 5}}}, "backend": &cfgUpstream{Balancer: "round_robin", Provider: "static", Servers: []upstreamServer{upstreamServer{Path: "http://api.foo1.com", Weight: 5}, upstreamServer{Path: "http://api.foo2.com", Weight: 0}}}}, Server: &server{Port: 8080, Locations: []serverLocation{serverLocation{Path: "/api", HTTPPass: "backend"}, serverLocation{Path: "/", HTTPPass: "front"}}}}},
		},
		"valid": {
			expectedCfg: &Config{80, &config{Upstreams: map[string]*cfgUpstream{"front": &cfgUpstream{Balancer: "round_robin", Provider: "static", Servers: []upstreamServer{upstreamServer{Path: "http://api.foo.com", Weight: 5}}}, "backend": &cfgUpstream{Balancer: "round_robin", Provider: "static", Servers: []upstreamServer{upstreamServer{Path: "http://api.foo1.com", Weight: 5}, upstreamServer{Path: "http://api.foo2.com", Weight: 0}}}}, Server: &server{Port: 80, Locations: []serverLocation{serverLocation{Path: "/api", HTTPPass: "backend"}, serverLocation{Path: "/", HTTPPass: "front"}}}}},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cfg, err := New(mustOpenConfigF(t, name))
			assert.Equal(t, c.expectedCfg, cfg)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

func mustOpenConfigF(t *testing.T, fname string) io.Reader {
	f, err := os.Open(filepath.Join("test_fixture", fname+".toml"))
	if err != nil {
		t.Fatal(err)
	}
	return f
}
