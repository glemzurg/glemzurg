package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/database"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
)

func main() {

	// Example call: $GOBIN/reqmodel -rootsource example/models -rootoutput example/output/models -model model_a -plantuml /usr/bin/plantuml -debug

	var rootSourcePath, rootOutputPath, model string
	var debug bool
	flag.StringVar(&rootSourcePath, "rootsource", "", "the path to the source models")
	flag.StringVar(&rootOutputPath, "rootoutput", "", "the path to output files")
	flag.StringVar(&model, "model", "", "the model to generate md from")
	flag.BoolVar(&debug, "debug", false, "Enable the debug level of logging")
	flag.Parse()

	// Set the appropriate logging level.
	_ = slog.SetLogLoggerLevel(slog.LevelInfo)
	if debug {
		_ = slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// The files.
	fmt.Printf("\nroot source path: %s\n", rootSourcePath)
	fmt.Printf("root output path: %s\n", rootOutputPath)
	fmt.Printf("model: %s\n", model)

	// Write the requirements to the database to ensure the data is well-formed.
	db, err := database.NewDb()
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Do all the work of updating the markdown from the source.
	err = generate.GenerateMd(debug, db, rootSourcePath, rootOutputPath, model)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	//fmt.Println(helper.JsonPretty(reqs))

	// Everything good.
	os.Exit(0)
}
