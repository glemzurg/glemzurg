package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/database"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
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

	var rootSourcePath, rootOutputPath, model string
	var inputFormat, outputFormat string
	var debug, skipDB bool
	flag.StringVar(&rootSourcePath, "rootsource", "", "the path to the source models")
	flag.StringVar(&rootOutputPath, "rootoutput", "", "the path to output files")
	flag.StringVar(&model, "model", "", "the model to process")
	flag.StringVar(&inputFormat, "input", InputFormatDataYAML, "input format: data/yaml or ai/json")
	flag.StringVar(&outputFormat, "output", OutputFormatMD, "output format: data/yaml, md, or ai/json")
	flag.BoolVar(&debug, "debug", false, "enable the debug level of logging")
	flag.BoolVar(&skipDB, "skipdb", false, "skip database validation step")
	flag.Parse()

	// Validate required arguments
	if rootSourcePath == "" || rootOutputPath == "" || model == "" {
		fmt.Println("Error: rootsource, rootoutput, and model are required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate input format
	inputFormat = strings.ToLower(inputFormat)
	if inputFormat != InputFormatDataYAML && inputFormat != InputFormatAIJSON {
		fmt.Printf("Error: invalid input format '%s'. Valid options: data/yaml, ai/json\n", inputFormat)
		os.Exit(1)
	}

	// Validate output format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != OutputFormatDataYAML && outputFormat != OutputFormatMD && outputFormat != OutputFormatAIJSON {
		fmt.Printf("Error: invalid output format '%s'. Valid options: data/yaml, md, ai/json\n", outputFormat)
		os.Exit(1)
	}


	// Set the appropriate logging level.
	_ = slog.SetLogLoggerLevel(slog.LevelInfo)
	if debug {
		_ = slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// Show configuration
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  root source path: %s\n", rootSourcePath)
	fmt.Printf("  root output path: %s\n", rootOutputPath)
	fmt.Printf("  model: %s\n", model)
	fmt.Printf("  input format: %s\n", inputFormat)
	fmt.Printf("  output format: %s\n", outputFormat)
	fmt.Println()

	// Process the conversion
	err := processConversion(debug, skipDB, rootSourcePath, rootOutputPath, model, inputFormat, outputFormat)
	if err != nil {
		fmt.Printf("Error: %+v\n\n", err)
		os.Exit(1)
	}

	// Everything good.
	os.Exit(0)
}

// processConversion handles the input/output conversion based on formats
func processConversion(debug, skipDB bool, rootSourcePath, rootOutputPath, model, inputFormat, outputFormat string) error {

	sourcePath := filepath.Join(rootSourcePath, model)
	outputPath := filepath.Join(rootOutputPath, model)

	// Step 1: Read the input model into req_model.Model
	var parsedModel *req_model.Model

	switch inputFormat {
	case InputFormatDataYAML:
		fmt.Println("Reading model from data/yaml format...")
		m, err := parser.Parse(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to parse data/yaml model: %w", err)
		}
		parsedModel = &m

	case InputFormatAIJSON:
		fmt.Println("Reading model from ai/json format...")
		inputModel, err := parser_ai.ReadModelTree(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read ai/json model: %w", err)
		}

		// Convert to req_model.Model
		converted, err := parser_ai.ConvertToModel(inputModel, model)
		if err != nil {
			return fmt.Errorf("failed to convert ai/json to req_model: %w", err)
		}
		parsedModel = converted
	}

	// Step 2: Optionally validate through database
	if !skipDB && outputFormat == OutputFormatMD {
		db, err := database.NewDb()
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Println("Exercising data model through database...")
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
		fmt.Println("Generating markdown output...")
		// Use the already-parsed model to generate markdown
		err := generate.GenerateMdFromModel(debug, outputPath, *parsedModel)
		if err != nil {
			return fmt.Errorf("failed to generate markdown: %w", err)
		}

	case OutputFormatAIJSON:
		fmt.Println("Converting to ai/json format...")
		// Convert req_model.Model to inputModel
		inputModel, err := parser_ai.ConvertFromModel(parsedModel)
		if err != nil {
			return fmt.Errorf("failed to convert to ai/json format: %w", err)
		}

		// Write to filesystem
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := parser_ai.WriteModelTree(inputModel, outputPath); err != nil {
			return fmt.Errorf("failed to write ai/json model: %w", err)
		}
		fmt.Printf("Model written to: %s\n", outputPath)

	case OutputFormatDataYAML:
		fmt.Println("Converting to data/yaml format...")
		if err := parser.Write(*parsedModel, outputPath); err != nil {
			return fmt.Errorf("failed to write data/yaml model: %w", err)
		}
		fmt.Printf("Model written to: %s\n", outputPath)
	}

	fmt.Println("\nDone!")
	return nil
}
