package parser_ai_input

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/errors"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputModel represents the model.json file.
type inputModel struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// modelSchema is the compiled JSON schema for model.json.
var modelSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	schemaContent, err := json_schemas.Schemas.ReadFile("model.schema.json")
	if err != nil {
		panic("failed to read model.schema.json: " + err.Error())
	}
	if err := compiler.AddResource("model.schema.json", strings.NewReader(string(schemaContent))); err != nil {
		panic("failed to add model schema resource: " + err.Error())
	}
	modelSchema, err = compiler.Compile("model.schema.json")
	if err != nil {
		panic("failed to compile model.schema.json: " + err.Error())
	}
}

// parseModel parses a model.json file content into an inputModel struct.
// It validates the input against the model schema and returns detailed errors if validation fails.
func parseModel(content []byte) (*inputModel, error) {
	var model inputModel

	// Parse JSON
	if err := json.Unmarshal(content, &model); err != nil {
		return nil, errors.NewParseError(
			errors.ErrModelInvalidJSON,
			"failed to parse model JSON: "+err.Error(),
			"Ensure the model.json file contains valid JSON. Check for missing commas, unquoted strings, or trailing commas.",
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, errors.NewParseError(
			errors.ErrModelInvalidJSON,
			"failed to parse model JSON for schema validation: "+err.Error(),
			"Ensure the model.json file contains valid JSON.",
		)
	}
	if err := modelSchema.Validate(jsonData); err != nil {
		return nil, errors.NewParseError(
			errors.ErrModelSchemaViolation,
			"model JSON does not match schema: "+err.Error(),
			"Check your model.json against the expected schema. Required fields: name (string). Optional fields: details (string). No additional properties are allowed.",
		)
	}

	// Validate required fields
	if err := validateModel(&model); err != nil {
		return nil, err
	}

	return &model, nil
}

// validateModel validates an inputModel struct.
func validateModel(model *inputModel) error {
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
