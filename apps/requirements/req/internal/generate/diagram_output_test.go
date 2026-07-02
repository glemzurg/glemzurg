package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stubSVG = `<svg xmlns="http://www.w3.org/2000/svg"></svg>`

// linkedCollectWriter records markdown and SVG output for linked-diagram tests.
type linkedCollectWriter struct {
	md  map[string][]byte
	svg map[string][]byte
}

func newLinkedCollectWriter() *linkedCollectWriter {
	return &linkedCollectWriter{
		md:  map[string][]byte{},
		svg: map[string][]byte{},
	}
}

func (c *linkedCollectWriter) WriteMarkdown(filename string, content []byte) error {
	c.md[filename] = content
	return nil
}

func (c *linkedCollectWriter) WriteSVG(filename string, content []byte) error {
	c.svg[filename] = content
	return nil
}

func (c *linkedCollectWriter) WriteCSS([]byte) error { return nil }

func (c *linkedCollectWriter) DiagramMode() DiagramOutputMode { return DiagramLinkedSVG }

func TestEmbedDiagramInlineMermaid(t *testing.T) {
	writer := newCollectWriter()
	embed, err := embedDiagram(writer, "classDiagram\n  class Foo", "model-domains.svg", "Domains")
	require.NoError(t, err)
	assert.Equal(t, "```mermaid\nclassDiagram\n  class Foo\n```", embed)
}

func TestEmbedDiagramLinkedSVG(t *testing.T) {
	t.Cleanup(func() { SetMermaidRenderHook(nil) })
	SetMermaidRenderHook(func(string) ([]byte, error) { return []byte(stubSVG), nil })

	writer := newLinkedCollectWriter()
	embed, err := embedDiagram(writer, "classDiagram\n  class Foo", "model-domains.svg", "Domains")
	require.NoError(t, err)
	assert.Equal(t, "[![Domains](model-domains.svg)](model-domains.svg)", embed)
	require.Contains(t, writer.svg, "model-domains.svg")
	assert.Equal(t, stubSVG, string(writer.svg["model-domains.svg"]))
}

func TestGenerateMdToWriterLinkedSVGHasNoMermaidFences(t *testing.T) {
	t.Cleanup(func() { SetMermaidRenderHook(nil) })
	SetMermaidRenderHook(func(string) ([]byte, error) { return []byte(stubSVG), nil })

	model := test_helper.GetTestModel()
	writer := newLinkedCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	for filename, content := range writer.md {
		assert.NotContains(t, string(content), "```mermaid", "file %s should not contain mermaid fences", filename)
	}
	assert.NotEmpty(t, writer.svg, "linked mode should write diagram SVGs")
}

func TestGenerateMdToWriterInlineMermaidPreservesFences(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	var withMermaid int
	for _, content := range writer.md {
		if strings.Contains(string(content), "```mermaid") {
			withMermaid++
		}
	}
	assert.Positive(t, withMermaid, "inline mode should keep mermaid fences on diagram pages")
}
