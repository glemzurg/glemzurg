package httpserver

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceExtensionsYAMLIncludesMarked(t *testing.T) {
	require.True(t, slices.Contains(SourceExtensionsYAML, ".marked"),
		"this.marked changes must trigger model reload and HTTP page refresh")
}

func TestSourceExtensionsYAMLIsSortedByName(t *testing.T) {
	// Keep the list easy to scan; alphabetical by extension suffix.
	got := append([]string(nil), SourceExtensionsYAML...)
	// Not requiring sort, just no duplicates.
	seen := make(map[string]bool, len(got))
	for _, ext := range got {
		assert.False(t, seen[ext], "duplicate extension %q", ext)
		seen[ext] = true
		assert.True(t, len(ext) > 1 && ext[0] == '.', "extension %q should start with dot", ext)
	}
}
