package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// _LOGIC_TYPE_QUERY is the logic type string for query logic.
const _LOGIC_TYPE_QUERY = "query"

// inputLogic represents a formal logic specification in JSON.
type inputLogic struct {
	Type           string `json:"type,omitempty"`
	Description    string `json:"description"`
	Target         string `json:"target,omitempty"`
	TargetTypeSpec string `json:"target_type_spec,omitempty"`
	Notation       string `json:"notation,omitempty"`
	Specification  string `json:"specification,omitempty"`
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
		).WithHint("ensure file contains valid JSON syntax")
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrLogicInvalidJSON,
			"failed to parse logic JSON for schema validation: "+err.Error(),
			filename,
		).WithHint("ensure file contains valid JSON syntax")
	}
	if err := logicSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrLogicSchemaViolation,
			"logic JSON does not match schema: "+err.Error(),
			filename,
		).WithHint("run: req_check --schema logic")
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
		).WithField("description").WithHint("add a non-empty \"description\" field")
	}

	// Description cannot be only whitespace
	if strings.TrimSpace(logic.Description) == "" {
		return NewParseError(
			ErrLogicDescriptionEmpty,
			"logic description cannot be empty or whitespace only, got '"+logic.Description+"'",
			filename,
		).WithField("description").WithHint("add a non-empty \"description\" field")
	}

	// Target validation based on logic type (when type is specified).
	switch logic.Type {
	case "state_change", _LOGIC_TYPE_QUERY, "let":
		if logic.Target == "" {
			return NewParseError(
				ErrLogicTargetRequired,
				"logic of type '"+logic.Type+"' requires a non-empty 'target' field — for state_change this is the attribute SubKey being set, for query this is the output identifier name, for let this is the local variable name",
				filename,
			).WithField("target").WithHint("state_change/query/let types require a non-empty \"target\" field")
		}
		if (logic.Type == _LOGIC_TYPE_QUERY || logic.Type == "let") && strings.HasPrefix(logic.Target, "_") {
			return NewParseError(
				ErrLogicTargetNoLeadUnderscore,
				logic.Type+" logic target '"+logic.Target+"' cannot start with '_' — use a plain identifier name",
				filename,
			).WithField("target").WithHint("query/let target names cannot start with underscore")
		}
	case "assessment", "safety_rule", "value":
		if logic.Target != "" {
			return NewParseError(
				ErrLogicTargetNotAllowed,
				"logic of type '"+logic.Type+"' must not have a 'target' field, got '"+logic.Target+"' — only state_change, query, and let types use target",
				filename,
			).WithField("target").WithHint("assessment/safety_rule/value types must not have a \"target\" field")
		}
	}

	return nil
}
