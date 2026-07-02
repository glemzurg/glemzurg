package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassesMermaidStereotypeLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		nodeID string
		want   string
	}{
		{name: "actor", nodeID: "class_example", want: "<<actor>> class_example\n"},
		{name: "association", nodeID: "assoc_example", want: "<<association>> assoc_example\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classesMermaidStereotypeLine(tc.name, tc.nodeID))
		})
	}
}
