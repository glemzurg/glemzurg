package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputObject represents an object participating in a scenario.
type inputObject struct {
	ObjectNumber uint   `json:"object_number"`
	Name         string `json:"name,omitempty"`
	NameStyle    string `json:"name_style"`
	ClassKey     string `json:"class_key"`
	Multi        bool   `json:"multi,omitempty"`
	UmlComment   string `json:"uml_comment,omitempty"`
}

// inputStep represents a step in a scenario's step tree.
// Steps form a recursive tree structure (AST) representing the scenario flow.
type inputStep struct {
	StepType      string      `json:"step_type"`
	LeafType      *string     `json:"leaf_type,omitempty"`
	Statements    []inputStep `json:"statements,omitempty"`
	Condition     string      `json:"condition,omitempty"`
	Description   string      `json:"description,omitempty"`
	FromObjectKey *string     `json:"from_object_key,omitempty"`
	ToObjectKey   *string     `json:"to_object_key,omitempty"`
	EventKey      *string     `json:"event_key,omitempty"`
	QueryKey      *string     `json:"query_key,omitempty"`
	ScenarioKey   *string     `json:"scenario_key,omitempty"`
}

// inputScenario represents a scenario JSON file.
// Scenarios document specific flows through a use case (e.g., sequence diagrams).
type inputScenario struct {
	Name    string                  `json:"name"`
	Details string                  `json:"details,omitempty"`
	Objects map[string]*inputObject `json:"objects,omitempty"`
	Steps   *inputStep              `json:"steps,omitempty"`
}

// scenarioSchema is the compiled JSON schema for scenario files.
var scenarioSchema *jsonschema.Schema

// scenarioSchemaContent is the raw JSON schema content for error reporting.
var scenarioSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("scenario.schema.json")
	if err != nil {
		panic("failed to read scenario.schema.json: " + err.Error())
	}
	scenarioSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("scenario.schema.json", strings.NewReader(scenarioSchemaContent)); err != nil {
		panic("failed to add scenario schema resource: " + err.Error())
	}
	scenarioSchema, err = compiler.Compile("scenario.schema.json")
	if err != nil {
		panic("failed to compile scenario.schema.json: " + err.Error())
	}
}

// parseScenario parses a scenario JSON file content into an inputScenario struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the scenario schema and returns detailed errors if validation fails.
func parseScenario(content []byte, filename string) (*inputScenario, error) {
	var scenario inputScenario

	// Parse JSON
	if err := json.Unmarshal(content, &scenario); err != nil {
		return nil, NewParseError(
			ErrScenarioInvalidJSON,
			"failed to parse scenario JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrScenarioInvalidJSON,
			"failed to parse scenario JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := scenarioSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrScenarioSchemaViolation,
			"scenario JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(scenarioSchemaContent)
	}

	// Validate required fields
	if err := validateScenario(&scenario, filename); err != nil {
		return nil, err
	}

	return &scenario, nil
}

// validateScenario validates an inputScenario struct.
// The filename parameter is the path to the JSON file being parsed.
func validateScenario(scenario *inputScenario, filename string) error {
	// Name is required
	if scenario.Name == "" {
		return NewParseError(
			ErrScenarioNameRequired,
			"scenario name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(scenario.Name) == "" {
		return NewParseError(
			ErrScenarioNameEmpty,
			"scenario name cannot be empty or whitespace only, got '"+scenario.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
