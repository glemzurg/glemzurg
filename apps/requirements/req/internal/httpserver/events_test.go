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
		{name: "model only", path: "test_model", want: "test_model"},
		{name: "legacy per file", path: "test_model/domain-domain.commerce.md", want: "test_model"},
		{name: "trailing slash", path: "test_model/", want: "test_model"},
		{name: "empty", path: "", want: ""},
		{name: "slashes only", path: "/", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, eventStreamModelKey(tc.path))
		})
	}
}
