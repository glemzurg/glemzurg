package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/database"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/httpserver"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
)

// Supported input formats
const (
	InputFormatDataYAML = "data/yaml" // Parser format (YAML files)
	InputFormatAIJSON   = "ai/json"   // AI format (JSON files)
)

// Supported output formats
const (
	OutputFormatDataYAML = "data/yaml" // Parser format (YAML files)
	OutputFormatMD       = "md"        // Markdown documentation
	OutputFormatAIJSON   = "ai/json"   // AI format (JSON files)
)

func main() {

	// Example calls:
	// Default: data/yaml to md
	//   $GOBIN/req -rootsource example/models -rootoutput example/output/models -model model_a
	//
	// Convert ai/json to md
	//   $GOBIN/req -input ai/json -output md -rootsource example/ai_models -rootoutput example/output/models -model model_a
	//
	// Convert data/yaml to ai/json
	//   $GOBIN/req -input data/yaml -output ai/json -rootsource example/models -rootoutput example/ai_models -model model_a
	//
	// Convert ai/json to data/yaml
	//   $GOBIN/req -input ai/json -output data/yaml -rootsource example/ai_models -rootoutput example/models -model model_a
	//
	// HTTP server mode (serves in-memory generated content for a single model):
	//   $GOBIN/req -http -port 8080 -rootsource example/models -model model_a

	var rootSourcePath, rootOutputPath, model string
	var inputFormat, outputFormat string
	var debug, skipDB bool
	var httpMode bool
	var port string
	flag.StringVar(&rootSourcePath, "rootsource", "", "the path to the source models")
	flag.StringVar(&rootOutputPath, "rootoutput", "", "the path to output files")
	flag.StringVar(&model, "model", "", "the model to process")
	flag.StringVar(&inputFormat, "input", InputFormatDataYAML, "input format: data/yaml or ai/json")
	flag.StringVar(&outputFormat, "output", OutputFormatMD, "output format: data/yaml, md, or ai/json")
	flag.BoolVar(&debug, "debug", false, "enable the debug level of logging")
	flag.BoolVar(&skipDB, "skipdb", false, "skip database validation step")
	flag.BoolVar(&httpMode, "http", false, "start HTTP server mode")
	flag.StringVar(&port, "port", "8080", "port for HTTP server (only used with -http)")
	flag.Parse()

	// Validate input format
	inputFormat = strings.ToLower(inputFormat)
	if inputFormat != InputFormatDataYAML && inputFormat != InputFormatAIJSON {
		log.Printf("Error: invalid input format '%s'. Valid options: data/yaml, ai/json", inputFormat)
		os.Exit(1)
	}

	// Validate output format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != OutputFormatDataYAML && outputFormat != OutputFormatMD && outputFormat != OutputFormatAIJSON {
		log.Printf("Error: invalid output format '%s'. Valid options: data/yaml, md, ai/json", outputFormat)
		os.Exit(1)
	}

	// Set the appropriate logging level.
	_ = slog.SetLogLoggerLevel(slog.LevelInfo)
	if debug {
		_ = slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// HTTP server mode
	if httpMode {
		if rootSourcePath == "" || model == "" {
			log.Println("Error: rootsource and model are required for HTTP server mode")
			flag.Usage()
			os.Exit(1)
		}

		// Show configuration
		log.Printf("Configuration:")
		log.Printf("  root source path: %s", rootSourcePath)
		log.Printf("  model: %s", model)
		log.Printf("  input format: %s", inputFormat)
		log.Printf("  port: %s", port)
		log.Println()

		runHTTPServer(rootSourcePath, model, port, inputFormat)
		return
	}

	// Validate required arguments for conversion mode
	if rootSourcePath == "" || rootOutputPath == "" || model == "" {
		log.Println("Error: rootsource, rootoutput, and model are required")
		flag.Usage()
		os.Exit(1)
	}

	// Show configuration
	log.Printf("Configuration:")
	log.Printf("  root source path: %s", rootSourcePath)
	log.Printf("  root output path: %s", rootOutputPath)
	log.Printf("  model: %s", model)
	log.Printf("  input format: %s", inputFormat)
	log.Printf("  output format: %s", outputFormat)
	log.Println()

	// Process the conversion
	err := processConversion(debug, skipDB, rootSourcePath, rootOutputPath, model, inputFormat, outputFormat)
	if err != nil {
		log.Printf("Error: %+v", err)
		os.Exit(1)
	}

	// Everything good.
	os.Exit(0)
}

// processConversion handles the input/output conversion based on formats
func processConversion(debug, skipDB bool, rootSourcePath, rootOutputPath, model, inputFormat, outputFormat string) error {

	sourcePath := filepath.Join(rootSourcePath, model)
	outputPath := filepath.Join(rootOutputPath, model)

	// Step 1: Read the input model into core.Model
	var parsedModel *core.Model

	switch inputFormat {
	case InputFormatDataYAML:
		log.Println("Reading model from data/yaml format...")
		m, err := parser_human.Parse(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to parse data/yaml model: %w", err)
		}
		parsedModel = &m

	case InputFormatAIJSON:
		log.Println("Reading model from ai/json format...")
		m, err := parser_ai.ReadModel(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read ai/json model: %w", err)
		}
		parsedModel = &m
	}

	// Step 2: Optionally validate through database
	if !skipDB && outputFormat == OutputFormatMD {
		db, err := database.NewDb()
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Println("Exercising data model through database...")
		err = database.WriteModel(db, *parsedModel)
		if err != nil {
			return fmt.Errorf("failed to write model to database: %w", err)
		}
		m, err := database.ReadModel(db, parsedModel.Key)
		if err != nil {
			return fmt.Errorf("failed to read model from database: %w", err)
		}
		parsedModel = &m
	}

	// Step 3: Write the output in the desired format
	switch outputFormat {
	case OutputFormatMD:
		log.Println("Generating markdown output...")
		// Use the already-parsed model to generate markdown
		err := generate.GenerateMdFromModel(debug, outputPath, *parsedModel)
		if err != nil {
			return fmt.Errorf("failed to generate markdown: %w", err)
		}

	case OutputFormatAIJSON:
		log.Println("Converting to ai/json format...")
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := parser_ai.WriteModel(*parsedModel, outputPath); err != nil {
			return fmt.Errorf("failed to write ai/json model: %w", err)
		}
		log.Printf("Model written to: %s", outputPath)

	case OutputFormatDataYAML:
		log.Println("Converting to data/yaml format...")
		if err := parser_human.Write(*parsedModel, outputPath); err != nil {
			return fmt.Errorf("failed to write data/yaml model: %w", err)
		}
		log.Printf("Model written to: %s", outputPath)
	}

	log.Println("Done!")
	return nil
}

// runHTTPServer starts the HTTP server in watch mode, serving in-memory generated content for a single model.
func runHTTPServer(rootSourcePath, model, port, inputFormat string) {
	modelPath := filepath.Join(rootSourcePath, model)
	log.Printf("Starting HTTP server on port :%s for model %s (format: %s)", port, model, inputFormat)

	// Create the model store and server
	store := httpserver.NewModelStore()
	server := httpserver.NewServer(store)

	// Create and start the source watcher for the specific model
	watcher, err := httpserver.NewSourceWatcher(modelPath, inputFormat, store, server)
	if err != nil {
		log.Fatalf("Failed to create source watcher: %v", err)
	}

	// Load the model
	log.Printf("Loading model %s...", model)
	if err := watcher.LoadModel(); err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	// Start watching for changes
	if err := watcher.Start(); err != nil {
		log.Fatalf("Failed to start source watcher: %v", err)
	}
	defer func() { _ = watcher.Close() }()

	// Start the HTTP server
	log.Printf("Server ready at http://localhost:%s/%s/model.md", port, model)
	log.Fatal(http.ListenAndServe(":"+port, server.Handler()))
}
