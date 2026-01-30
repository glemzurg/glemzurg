package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

// Supported input formats
const (
	InputFormatDataYAML = "data/yaml" // Parser format (YAML files)
	InputFormatAIJSON   = "ai/json"   // AI format (JSON files)
)

func main() {

	// Example calls:
	//   $GOBIN/req_check -rootsource example/models -model model_a
	//   $GOBIN/req_check -input ai/json -rootsource example/ai_models -model model_a

	var rootSourcePath, model string
	var inputFormat string
	var debug bool
	flag.StringVar(&rootSourcePath, "rootsource", "", "the path to the source models")
	flag.StringVar(&model, "model", "", "the model to validate")
	flag.StringVar(&inputFormat, "input", InputFormatDataYAML, "input format: data/yaml or ai/json")
	flag.BoolVar(&debug, "debug", false, "enable the debug level of logging")
	flag.Parse()

	// Validate required arguments
	if rootSourcePath == "" || model == "" {
		fmt.Println("Error: rootsource and model are required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate input format
	inputFormat = strings.ToLower(inputFormat)
	if inputFormat != InputFormatDataYAML && inputFormat != InputFormatAIJSON {
		fmt.Printf("Error: invalid input format '%s'. Valid options: data/yaml, ai/json\n", inputFormat)
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
	fmt.Printf("  model: %s\n", model)
	fmt.Printf("  input format: %s\n", inputFormat)
	fmt.Println()

	// Validate the model
	err := validateModel(rootSourcePath, model, inputFormat)
	if err != nil {
		fmt.Printf("Validation failed: %+v\n\n", err)
		os.Exit(1)
	}

	fmt.Println("Validation passed!")
	os.Exit(0)
}

// validateModel reads and validates a model from the specified format.
func validateModel(rootSourcePath, model, inputFormat string) error {

	sourcePath := filepath.Join(rootSourcePath, model)

	// Read the input model into req_model.Model
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

		// Validate the tree structure
		fmt.Println("Validating ai/json tree structure...")
		if err := parser_ai.ValidateModelTree(inputModel); err != nil {
			return fmt.Errorf("ai/json model tree validation failed: %w", err)
		}

		// Validate completeness
		fmt.Println("Validating ai/json model completeness...")
		if err := parser_ai.ValidateModelCompleteness(inputModel); err != nil {
			return fmt.Errorf("ai/json model completeness validation failed: %w", err)
		}

		// Convert to req_model.Model
		fmt.Println("Converting to req_model...")
		converted, err := parser_ai.ConvertToModel(inputModel, model)
		if err != nil {
			return fmt.Errorf("failed to convert ai/json to req_model: %w", err)
		}
		parsedModel = converted
	}

	// Validate the req_model
	fmt.Println("Validating req_model...")
	if err := parsedModel.Validate(); err != nil {
		return fmt.Errorf("req_model validation failed: %w", err)
	}

	return nil
}
