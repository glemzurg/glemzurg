package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputParameter represents a typed parameter in JSON.
type inputParameter struct {
	Name          string `json:"name"`
	DataTypeRules string `json:"data_type_rules,omitempty"`
}

// parameterSchema is the compiled JSON schema for parameter objects.
var parameterSchema *jsonschema.Schema

// parameterSchemaContent is the raw JSON schema content for error reporting.
var parameterSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("parameter.schema.json")
	if err != nil {
		panic("failed to read parameter.schema.json: " + err.Error())
	}
	parameterSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("parameter.schema.json", strings.NewReader(parameterSchemaContent)); err != nil {
		panic("failed to add parameter schema resource: " + err.Error())
	}
	parameterSchema, err = compiler.Compile("parameter.schema.json")
	if err != nil {
		panic("failed to compile parameter.schema.json: " + err.Error())
	}
}

// parseParameter parses a parameter JSON object content into an inputParameter struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the parameter schema and returns detailed errors if validation fails.
func parseParameter(content []byte, filename string) (*inputParameter, error) {
	var param inputParameter

	// Parse JSON
	if err := json.Unmarshal(content, &param); err != nil {
		return nil, NewParseError(
			ErrParamInvalidJSON,
			"failed to parse parameter JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrParamInvalidJSON,
			"failed to parse parameter JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := parameterSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrParamSchemaViolation,
			"parameter JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(parameterSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateParameter(&param, filename); err != nil {
		return nil, err
	}

	return &param, nil
}

// validateParameter validates an inputParameter struct.
// The filename parameter is the path to the JSON file being parsed.
func validateParameter(param *inputParameter, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if param.Name == "" {
		return NewParseError(
			ErrParamNameRequired,
			"parameter name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(param.Name) == "" {
		return NewParseError(
			ErrParamNameEmpty,
			"parameter name cannot be empty or whitespace only, got '"+param.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
