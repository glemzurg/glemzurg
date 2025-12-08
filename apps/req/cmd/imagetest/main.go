package main

import (
	"bytes"
	"context"
	"embed"
	"io/fs"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/pkg/errors"
)

//go:embed images
var imagesFolder embed.FS

func init() {
	// Create a sub-FS rooted at "images" so the root of the virtual FS is the contents of images/.
	subFS, err := fs.Sub(imagesFolder, "images")
	if err != nil {
		log.Fatalf("failed to create sub-FS: %v", err)
	}
	graphviz.SetFileSystem(subFS)
}

func main() {
	ctx := context.Background()

	// Create the Graphviz instance.
	gv, err := graphviz.New(ctx)
	if err != nil {
		log.Fatalf("failed to create graphviz: %v", err)
	}
	defer gv.Close()

	// Define the DOT as a literal string.
	dot := `digraph G {
        n [shape=none, label=<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0">
            <TR><TD><IMG SRC="system.svg"/></TD></TR>
        </TABLE>>];
    }`

	// Parse the DOT into a graph (as per the provided logic).
	parsedGraph, err := graphviz.ParseBytes([]byte(dot))
	if err != nil {
		log.Fatalf("failed to parse DOT: %v", errors.WithStack(err))
	}
	defer parsedGraph.Close()

	// Render the SVG as a string using the provided logic.
	var buf bytes.Buffer
	if err := gv.Render(ctx, parsedGraph, graphviz.SVG, &buf); err != nil {
		log.Fatalf("failed to render: %v", errors.WithStack(err))
	}

	log.Println(buf.String()) // Or write to file, etc.
}
