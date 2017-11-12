package ingress_test

import (
	"testing"

	"github.com/tonto/gourmet/internal/platform/ingress"
)

func TestIngress(t *testing.T) {
	cases := []struct {
		name        string
		ph          ingress.ProtocolHandler
		url         string
		want        []byte
		wantCode    int
		wantContent string
	}{
	// Test resp
	// Test r timeout
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

		})
	}
}
