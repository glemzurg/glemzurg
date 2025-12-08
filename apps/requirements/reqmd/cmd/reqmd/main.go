package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/reqmd/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/reqmd/internal/requirements"
)

func main() {

	// Example call: $GOBIN/reqmd -config design/config.json -path design/requirements

	var configFilename string
	var path string
	flag.StringVar(&configFilename, "config", "", "the path to the configuration file")
	flag.StringVar(&path, "path", "", "the path to the requirements file tree")
	flag.Parse()
	fmt.Printf("\nconfig: %s\n", configFilename)

	// The config.
	config, err := requirements.ParseConfig(configFilename)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}
	fmt.Println(helper.JsonPretty(config))

	// The files.
	fmt.Printf("\npath: %s\n\n", path)

	// Create the new requirements.
	req, err := requirements.New(path)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Number unnumbered requirements.
	err = req.NumberAll()
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Add back links between requirements.
	err = req.GenerateReferencedInLists()
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Generate any incompletes.
	err = req.GenerateIncompletes()
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Update files.
	fmt.Printf("\nupdating: \n\n")
	err = req.UpdateFiles(config)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Write summary file (which includes incompletes)
	fmt.Printf("\nsummary: \n\n")
	err = req.WriteSummaryFile(path)
	if err != nil {
		fmt.Printf("%+v\n\n", err)
		os.Exit(1)
	}

	// Have a print line at the end of output.
	fmt.Println()

	// Everything good.
	os.Exit(0)
}
