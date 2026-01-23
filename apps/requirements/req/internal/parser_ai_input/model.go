package parser_ai_input

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/errors"
)

// ParseModel parses a model.json file content into an InputModel struct.
// It validates the input against the model schema and returns detailed errors if validation fails.
func ParseModel(content []byte) (*InputModel, error) {
	var model InputModel

	// Parse JSON
	if err := json.Unmarshal(content, &model); err != nil {
		return nil, errors.NewParseError(
			errors.ErrModelInvalidJSON,
			"failed to parse model JSON: "+err.Error(),
			"Ensure the model.json file contains valid JSON. Check for missing commas, unquoted strings, or trailing commas.",
		)
	}

	// Validate required fields
	if err := validateModel(&model); err != nil {
		return nil, err
	}

	return &model, nil
}

// validateModel validates an InputModel struct.
func validateModel(model *InputModel) error {
	// Name is required
	if model.Name == "" {
		return errors.NewParseError(
			errors.ErrModelNameRequired,
			"model name is required",
			`Add a "name" field to your model.json file. Example: {"name": "My Model", "details": "Optional description"}`,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(model.Name) == "" {
		return errors.NewParseError(
			errors.ErrModelNameEmpty,
			"model name cannot be empty or whitespace only",
			`The "name" field must contain actual text, not just spaces. Example: {"name": "My Model"}`,
		).WithField("name")
	}

	return nil
}
