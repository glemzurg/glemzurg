package generate

import (
	"bytes"
	"context"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/pkg/errors"
)

// Create a mutext for adding thread safety.
var _graphvizMutex = &sync.Mutex{}

// Convert a DOT input to an SVG.
func graphvizDotToSvg(dot string) (svg string, err error) {
	ctx := context.Background()

	// Add thread safety to library.
	_graphvizMutex.Lock()
	defer _graphvizMutex.Unlock()

	// Create a new graphviz object for generating images.
	g, err := graphviz.New(ctx)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer g.Close()

	// Create a graph from the input.
	graph, err := graphviz.ParseBytes([]byte(dot))
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer graph.Close()

	// Render the SVG as a string.
	var buf bytes.Buffer
	if err := g.Render(ctx, graph, graphviz.SVG, &buf); err != nil {
		return "", errors.WithStack(err)
	}

	return buf.String(), nil
}
