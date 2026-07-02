package generate

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMermaidToSVGWithMMDC(t *testing.T) {
	if _, err := exec.LookPath("mmdc"); err != nil {
		t.Skip("mmdc not installed")
	}

	t.Cleanup(func() { SetMermaidRenderHook(nil) })

	svg, err := renderMermaidToSVG("classDiagram\n  class Example")
	if err != nil && mmdcBrowserUnavailable(err) {
		t.Skip("mmdc installed but headless browser unavailable in this environment")
	}
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(svg), "<svg"), "mmdc should return SVG markup")
}

func mmdcBrowserUnavailable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Could not find Chrome") ||
		strings.Contains(msg, "Failed to launch the browser process")
}
