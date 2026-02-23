package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputActor represents an actor JSON file (e.g., actors/customer.actor.json).
type inputActor struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Details         string  `json:"details,omitempty"`
	SuperclassOfKey *string `json:"superclass_of_key,omitempty"`
	SubclassOfKey   *string `json:"subclass_of_key,omitempty"`
	UMLComment      string  `json:"uml_comment,omitempty"`
}

// actorSchema is the compiled JSON schema for actor files.
var actorSchema *jsonschema.Schema

// actorSchemaContent is the raw JSON schema content for error reporting.
var actorSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("actor.schema.json")
	if err != nil {
		panic("failed to read actor.schema.json: " + err.Error())
	}
	actorSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("actor.schema.json", strings.NewReader(actorSchemaContent)); err != nil {
		panic("failed to add actor schema resource: " + err.Error())
	}
	actorSchema, err = compiler.Compile("actor.schema.json")
	if err != nil {
		panic("failed to compile actor.schema.json: " + err.Error())
	}
}

// parseActor parses an actor JSON file content into an inputActor struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the actor schema and returns detailed errors if validation fails.
func parseActor(content []byte, filename string) (*inputActor, error) {
	var actor inputActor

	// Parse JSON
	if err := json.Unmarshal(content, &actor); err != nil {
		return nil, NewParseError(
			ErrActorInvalidJSON,
			"failed to parse actor JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrActorInvalidJSON,
			"failed to parse actor JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := actorSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrActorSchemaViolation,
			"actor JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(actorSchemaContent)
	}

	// Validate required fields
	if err := validateActor(&actor, filename); err != nil {
		return nil, err
	}

	return &actor, nil
}

// validateActor validates an inputActor struct.
// The filename parameter is the path to the JSON file being parsed.
func validateActor(actor *inputActor, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if actor.Name == "" {
		return NewParseError(
			ErrActorNameRequired,
			"actor name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(actor.Name) == "" {
		return NewParseError(
			ErrActorNameEmpty,
			"actor name cannot be empty or whitespace only, got '"+actor.Name+"'",
			filename,
		).WithField("name")
	}

	// Type is required (schema enforces this, but we provide a clearer error)
	if actor.Type == "" {
		return NewParseError(
			ErrActorTypeRequired,
			"actor type is required, got ''",
			filename,
		).WithField("type")
	}

	// Type cannot be only whitespace
	if strings.TrimSpace(actor.Type) == "" {
		return NewParseError(
			ErrActorTypeInvalid,
			"actor type cannot be empty or whitespace only, got '"+actor.Type+"'",
			filename,
		).WithField("type")
	}

	return nil
}
