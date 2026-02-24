package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

func main() {

	// Example call:
	//   $GOBIN/req_check /path/to/ai_models/model_a

	var modelPath string
	flag.StringVar(&modelPath, "path", "", "the path to the model (last folder is model name)")
	flag.Parse()

	// If no -path flag, check for positional argument
	if modelPath == "" && flag.NArg() > 0 {
		modelPath = flag.Arg(0)
	}

	// Validate required argument
	if modelPath == "" {
		fmt.Println("Error: model path is required")
		fmt.Println("Usage: req_check <model_path>")
		fmt.Println("       req_check -path <model_path>")
		os.Exit(1)
	}

	// Extract model name from path (last folder)
	model := filepath.Base(modelPath)

	// Always enable debug logging
	_ = slog.SetLogLoggerLevel(slog.LevelDebug)

	// Show configuration
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  model path: %s\n", modelPath)
	fmt.Printf("  model: %s\n", model)
	fmt.Println()

	// Validate the model
	err := validateModel(modelPath, model)
	if err != nil {
		fmt.Printf("Validation failed: %+v\n\n", err)
		os.Exit(1)
	}

	fmt.Println("Validation passed!")
	os.Exit(0)
}

// validateModel reads and validates a model from ai/json format.
func validateModel(modelPath, model string) error {

	// Read the input model into req_model.Model
	var parsedModel *req_model.Model

	fmt.Println("Reading and validating model from ai/json format...")
	inputModel, err := parser_ai.readModelTree(modelPath)
	if err != nil {
		return fmt.Errorf("failed to read/validate ai/json model: %w", err)
	}

	// Convert to req_model.Model
	fmt.Println("Converting to req_model...")
	converted, err := parser_ai.ConvertToModel(inputModel, model)
	if err != nil {
		return fmt.Errorf("failed to convert ai/json to req_model: %w", err)
	}
	parsedModel = converted

	// Validate the req_model
	fmt.Println("Validating req_model...")
	if err := parsedModel.Validate(); err != nil {
		return fmt.Errorf("req_model validation failed: %w", err)
	}

	return nil
}
