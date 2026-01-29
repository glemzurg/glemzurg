package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputGeneralization represents a generalization JSON file.
type inputGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UMLComment    string   `json:"uml_comment,omitempty"`
}

// generalizationSchema is the compiled JSON schema for generalization files.
var generalizationSchema *jsonschema.Schema

// generalizationSchemaContent is the raw JSON schema content for error reporting.
var generalizationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("generalization.schema.json")
	if err != nil {
		panic("failed to read generalization.schema.json: " + err.Error())
	}
	generalizationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("generalization.schema.json", strings.NewReader(generalizationSchemaContent)); err != nil {
		panic("failed to add generalization schema resource: " + err.Error())
	}
	generalizationSchema, err = compiler.Compile("generalization.schema.json")
	if err != nil {
		panic("failed to compile generalization.schema.json: " + err.Error())
	}
}

// parseGeneralization parses a generalization JSON file content into an inputGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the generalization schema and returns detailed errors if validation fails.
func parseGeneralization(content []byte, filename string) (*inputGeneralization, error) {
	var gen inputGeneralization

	// Parse JSON
	if err := json.Unmarshal(content, &gen); err != nil {
		return nil, NewParseError(
			ErrGenInvalidJSON,
			"failed to parse generalization JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrGenInvalidJSON,
			"failed to parse generalization JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := generalizationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrGenSchemaViolation,
			"generalization JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(generalizationSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateGeneralization(&gen, filename); err != nil {
		return nil, err
	}

	return &gen, nil
}

// validateGeneralization validates an inputGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
func validateGeneralization(gen *inputGeneralization, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if gen.Name == "" {
		return NewParseError(
			ErrGenNameRequired,
			"generalization name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(gen.Name) == "" {
		return NewParseError(
			ErrGenNameEmpty,
			"generalization name cannot be empty or whitespace only, got '"+gen.Name+"'",
			filename,
		).WithField("name")
	}

	// Superclass key is required (schema enforces this, but we provide a clearer error)
	if gen.SuperclassKey == "" {
		return NewParseError(
			ErrGenSuperclassRequired,
			"generalization superclass_key is required, got ''",
			filename,
		).WithField("superclass_key")
	}

	// Superclass key cannot be only whitespace
	if strings.TrimSpace(gen.SuperclassKey) == "" {
		return NewParseError(
			ErrGenSuperclassRequired,
			"generalization superclass_key cannot be empty or whitespace only, got '"+gen.SuperclassKey+"'",
			filename,
		).WithField("superclass_key")
	}

	// Subclass keys is required and must have at least one entry (schema enforces this)
	if len(gen.SubclassKeys) == 0 {
		return NewParseError(
			ErrGenSubclassesRequired,
			"generalization subclass_keys is required and must have at least one entry",
			filename,
		).WithField("subclass_keys")
	}

	// Each subclass key must be non-empty and non-whitespace
	for i, key := range gen.SubclassKeys {
		if key == "" {
			return NewParseError(
				ErrGenSubclassesEmpty,
				fmt.Sprintf("generalization subclass_keys[%d] cannot be empty", i),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
		if strings.TrimSpace(key) == "" {
			return NewParseError(
				ErrGenSubclassesEmpty,
				fmt.Sprintf("generalization subclass_keys[%d] cannot be whitespace only, got '%s'", i, key),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
	}

	return nil
}
