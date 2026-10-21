// Command postgresql reads a requirements JSON tree and PostgreSQL annotation files.
// The JSON tree is the one req writes. Annotation files use the .annot.json suffix
// and are parsed here, not by the requirements parser.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/load"
	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/walk"
)

func main() {
	if err := Execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// Execute runs the command with args and writes the walk to stdout.
func Execute(args []string, stdout io.Writer) error {
	modelDir, annotDir, err := parseArgs(args)
	if err != nil {
		return err
	}
	result, err := load.Load(modelDir, annotDir)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, walk.Format(result))
	return err
}

func parseArgs(args []string) (string, string, error) {
	fs := flag.NewFlagSet("postgresql", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var modelDir, annotDir string
	fs.StringVar(&modelDir, "model", "", "directory of JSON files written by req")
	fs.StringVar(&annotDir, "annot", "", "directory of PostgreSQL .annot.json files (default: the model directory)")
	if err := fs.Parse(args); err != nil {
		return "", "", err
	}
	if modelDir == "" {
		return "", "", fmt.Errorf("model directory is required")
	}
	if annotDir == "" {
		annotDir = modelDir
	}
	return modelDir, annotDir, nil
}
