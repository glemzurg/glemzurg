package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventStreamModelKey(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "model only", path: "evenplay", want: "evenplay"},
		{name: "legacy per file", path: "evenplay/domain-domain.finance.md", want: "evenplay"},
		{name: "trailing slash", path: "evenplay/", want: "evenplay"},
		{name: "empty", path: "", want: ""},
		{name: "slashes only", path: "/", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, eventStreamModelKey(tc.path))
		})
	}
}
