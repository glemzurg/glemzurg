package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputUseCaseActor represents an actor reference within a use case.
// The map key is the class key of the actor class.
type inputUseCaseActor struct {
	UmlComment string `json:"uml_comment,omitempty"`
}

// inputUseCase represents a use case JSON file.
// Use cases are user stories for the system at various levels (sky/sea/mud).
type inputUseCase struct {
	Name       string                        `json:"name"`
	Details    string                        `json:"details,omitempty"`
	Level      string                        `json:"level"`
	ReadOnly   bool                          `json:"read_only,omitempty"`
	UMLComment string                        `json:"uml_comment,omitempty"`
	Actors     map[string]*inputUseCaseActor `json:"actors,omitempty"`

	// Children (not from JSON, populated during directory traversal)
	Scenarios map[string]*inputScenario `json:"-"`
}

// useCaseSchema is the compiled JSON schema for use case files.
var useCaseSchema *jsonschema.Schema

// useCaseSchemaContent is the raw JSON schema content for error reporting.
var useCaseSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("use_case.schema.json")
	if err != nil {
		panic("failed to read use_case.schema.json: " + err.Error())
	}
	useCaseSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("use_case.schema.json", strings.NewReader(useCaseSchemaContent)); err != nil {
		panic("failed to add use case schema resource: " + err.Error())
	}
	useCaseSchema, err = compiler.Compile("use_case.schema.json")
	if err != nil {
		panic("failed to compile use_case.schema.json: " + err.Error())
	}
}

// parseUseCase parses a use case JSON file content into an inputUseCase struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the use case schema and returns detailed errors if validation fails.
func parseUseCase(content []byte, filename string) (*inputUseCase, error) {
	var uc inputUseCase

	// Parse JSON
	if err := json.Unmarshal(content, &uc); err != nil {
		return nil, NewParseError(
			ErrUseCaseInvalidJSON,
			"failed to parse use case JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseInvalidJSON,
			"failed to parse use case JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := useCaseSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseSchemaViolation,
			"use case JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(useCaseSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateUseCase(&uc, filename); err != nil {
		return nil, err
	}

	return &uc, nil
}

// validateUseCase validates an inputUseCase struct.
// The filename parameter is the path to the JSON file being parsed.
func validateUseCase(uc *inputUseCase, filename string) error {
	// Name is required
	if uc.Name == "" {
		return NewParseError(
			ErrUseCaseNameRequired,
			"use case name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(uc.Name) == "" {
		return NewParseError(
			ErrUseCaseNameEmpty,
			"use case name cannot be empty or whitespace only, got '"+uc.Name+"'",
			filename,
		).WithField("name")
	}

	// Level is required
	if uc.Level == "" {
		return NewParseError(
			ErrUseCaseLevelRequired,
			"use case level is required, got ''",
			filename,
		).WithField("level")
	}

	// Level must be one of the valid values
	switch uc.Level {
	case "sky", "sea", "mud":
		// valid
	default:
		return NewParseError(
			ErrUseCaseLevelInvalid,
			"use case level must be 'sky', 'sea', or 'mud', got '"+uc.Level+"'",
			filename,
		).WithField("level")
	}

	return nil
}
