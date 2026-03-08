package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
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
		log.Println("Error: model path is required")
		log.Println("Usage: req_check <model_path>")
		log.Println("       req_check -path <model_path>")
		os.Exit(1)
	}

	// Extract model name from path (last folder)
	model := filepath.Base(modelPath)

	// Always enable debug logging
	_ = slog.SetLogLoggerLevel(slog.LevelDebug)

	// Show configuration
	log.Printf("Configuration:")
	log.Printf("  model path: %s", modelPath)
	log.Printf("  model: %s", model)
	log.Println()

	// Validate the model
	err := validateModel(modelPath)
	if err != nil {
		log.Printf("Validation failed: %+v", err)
		os.Exit(1)
	}

	log.Println("Validation passed!")
	os.Exit(0)
}

// validateModel reads and validates a model from ai/json format.
func validateModel(modelPath string) error {
	// Read the input model into core.Model
	var parsedModel *core.Model

	log.Println("Reading and validating model from ai/json format...")
	m, err := parser_ai.ReadModel(modelPath)
	if err != nil {
		return fmt.Errorf("failed to read/validate ai/json model: %w", err)
	}
	parsedModel = &m

	// Validate the req_model
	log.Println("Validating core...")
	if err := parsedModel.Validate(); err != nil {
		return fmt.Errorf("req_model validation failed: %w", err)
	}

	return nil
}
