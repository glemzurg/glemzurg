package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputLogic represents a formal logic specification in JSON.
type inputLogic struct {
	Type          string `json:"type,omitempty"`
	Description   string `json:"description"`
	Notation      string `json:"notation,omitempty"`
	Specification string `json:"specification,omitempty"`
}

// logicSchema is the compiled JSON schema for logic objects.
var logicSchema *jsonschema.Schema

// logicSchemaContent is the raw JSON schema content for error reporting.
var logicSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("logic.schema.json")
	if err != nil {
		panic("failed to read logic.schema.json: " + err.Error())
	}
	logicSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("logic.schema.json", strings.NewReader(logicSchemaContent)); err != nil {
		panic("failed to add logic schema resource: " + err.Error())
	}
	logicSchema, err = compiler.Compile("logic.schema.json")
	if err != nil {
		panic("failed to compile logic.schema.json: " + err.Error())
	}
}

// parseLogic parses a logic JSON object content into an inputLogic struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the logic schema and returns detailed errors if validation fails.
func parseLogic(content []byte, filename string) (*inputLogic, error) {
	var logic inputLogic

	// Parse JSON
	if err := json.Unmarshal(content, &logic); err != nil {
		return nil, NewParseError(
			ErrLogicInvalidJSON,
			"failed to parse logic JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrLogicInvalidJSON,
			"failed to parse logic JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := logicSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrLogicSchemaViolation,
			"logic JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(logicSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateLogic(&logic, filename); err != nil {
		return nil, err
	}

	return &logic, nil
}

// validateLogic validates an inputLogic struct.
// The filename parameter is the path to the JSON file being parsed.
func validateLogic(logic *inputLogic, filename string) error {
	// Description is required (schema enforces this, but we provide a clearer error)
	if logic.Description == "" {
		return NewParseError(
			ErrLogicDescriptionRequired,
			"logic description is required, got ''",
			filename,
		).WithField("description")
	}

	// Description cannot be only whitespace
	if strings.TrimSpace(logic.Description) == "" {
		return NewParseError(
			ErrLogicDescriptionEmpty,
			"logic description cannot be empty or whitespace only, got '"+logic.Description+"'",
			filename,
		).WithField("description")
	}

	return nil
}
