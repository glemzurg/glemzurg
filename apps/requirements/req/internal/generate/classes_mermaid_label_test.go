package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassesMermaidStereotypeAnnotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "actor", want: "<<actor>>"},
		{name: "association", want: "<<association>>"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classesMermaidStereotypeAnnotation(tc.name))
		})
	}
}
