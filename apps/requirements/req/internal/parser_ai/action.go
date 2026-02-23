package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputAction represents an action JSON file.
type inputAction struct {
	Name        string           `json:"name"`
	Details     string           `json:"details,omitempty"`
	Parameters  []inputParameter `json:"parameters,omitempty"`
	Requires    []inputLogic     `json:"requires,omitempty"`
	Guarantees  []inputLogic     `json:"guarantees,omitempty"`
	SafetyRules []inputLogic     `json:"safety_rules,omitempty"`
}

// actionSchema is the compiled JSON schema for action files.
var actionSchema *jsonschema.Schema

// actionSchemaContent is the raw JSON schema content for error reporting.
var actionSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("action.schema.json")
	if err != nil {
		panic("failed to read action.schema.json: " + err.Error())
	}
	actionSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("action.schema.json", strings.NewReader(actionSchemaContent)); err != nil {
		panic("failed to add action schema resource: " + err.Error())
	}
	actionSchema, err = compiler.Compile("action.schema.json")
	if err != nil {
		panic("failed to compile action.schema.json: " + err.Error())
	}
}

// parseAction parses an action JSON file content into an inputAction struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the action schema and returns detailed errors if validation fails.
func parseAction(content []byte, filename string) (*inputAction, error) {
	var action inputAction

	// Parse JSON
	if err := json.Unmarshal(content, &action); err != nil {
		return nil, NewParseError(
			ErrActionInvalidJSON,
			"failed to parse action JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrActionInvalidJSON,
			"failed to parse action JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := actionSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrActionSchemaViolation,
			"action JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(actionSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateAction(&action, filename); err != nil {
		return nil, err
	}

	return &action, nil
}

// validateAction validates an inputAction struct.
// The filename parameter is the path to the JSON file being parsed.
func validateAction(action *inputAction, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if action.Name == "" {
		return NewParseError(
			ErrActionNameRequired,
			"action name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(action.Name) == "" {
		return NewParseError(
			ErrActionNameEmpty,
			"action name cannot be empty or whitespace only, got '"+action.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
