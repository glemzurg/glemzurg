package generate

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// DiagramOutputMode selects how Mermaid source is embedded in generated markdown.
type DiagramOutputMode int

const (
	// DiagramInlineMermaid keeps ```mermaid fences for browser-side rendering.
	DiagramInlineMermaid DiagramOutputMode = iota
	// DiagramLinkedSVG writes a sibling SVG and embeds a markdown link to it.
	DiagramLinkedSVG
)

// ContentWriter is an interface for writing generated content.
type ContentWriter interface {
	WriteMarkdown(filename string, content []byte) error
	WriteSVG(filename string, content []byte) error
	WriteCSS(content []byte) error
	DiagramMode() DiagramOutputMode
}

// embedDiagram turns Mermaid source into markdown for the writer's diagram mode.
// LinkedSVG mode renders the source, writes svgFilename via writer, and returns a link.
func embedDiagram(writer ContentWriter, source, svgFilename, label string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", nil
	}

	switch writer.DiagramMode() {
	case DiagramInlineMermaid:
		return "```mermaid\n" + source + "\n```", nil
	case DiagramLinkedSVG:
		svg, err := renderMermaidToSVG(source)
		if err != nil {
			return "", err
		}
		if err := writer.WriteSVG(svgFilename, svg); err != nil {
			return "", err
		}
		return fmt.Sprintf("[![%s](%s)](%s)", label, svgFilename, svgFilename), nil
	default:
		return "", errors.Errorf("unknown diagram output mode: %d", writer.DiagramMode())
	}
}
