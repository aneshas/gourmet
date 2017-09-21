package errors_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tonto/gourmet/internal/errors"
)

func TestNewError(t *testing.T) {
	cases := map[string]struct {
		status int
		text   string
		desc   string
		want   string
	}{
		"test 200": {
			status: http.StatusOK,
			text:   http.StatusText(http.StatusOK),
			desc:   "200 foo desc",
			want:   "200 OK <200 foo desc>",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := errors.New(c.status, c.text, c.desc)
			assert.Equal(t, c.want, err.Error())
		})
	}
}
