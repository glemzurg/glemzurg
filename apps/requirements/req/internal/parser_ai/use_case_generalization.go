package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputUseCaseGeneralization represents a use case generalization JSON file.
// Use case generalizations define super/sub-type hierarchies between use cases.
type inputUseCaseGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UMLComment    string   `json:"uml_comment,omitempty"`
}

// useCaseGeneralizationSchema is the compiled JSON schema for use case generalization files.
var useCaseGeneralizationSchema *jsonschema.Schema

// useCaseGeneralizationSchemaContent is the raw JSON schema content for error reporting.
var useCaseGeneralizationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("use_case_generalization.schema.json")
	if err != nil {
		panic("failed to read use_case_generalization.schema.json: " + err.Error())
	}
	useCaseGeneralizationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("use_case_generalization.schema.json", strings.NewReader(useCaseGeneralizationSchemaContent)); err != nil {
		panic("failed to add use case generalization schema resource: " + err.Error())
	}
	useCaseGeneralizationSchema, err = compiler.Compile("use_case_generalization.schema.json")
	if err != nil {
		panic("failed to compile use_case_generalization.schema.json: " + err.Error())
	}
}

// parseUseCaseGeneralization parses a use case generalization JSON file content into an inputUseCaseGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the use case generalization schema and returns detailed errors if validation fails.
func parseUseCaseGeneralization(content []byte, filename string) (*inputUseCaseGeneralization, error) {
	var gen inputUseCaseGeneralization

	// Parse JSON
	if err := json.Unmarshal(content, &gen); err != nil {
		return nil, NewParseError(
			ErrUseCaseGenInvalidJSON,
			"failed to parse use case generalization JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseGenInvalidJSON,
			"failed to parse use case generalization JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := useCaseGeneralizationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseGenSchemaViolation,
			"use case generalization JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(useCaseGeneralizationSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateUseCaseGeneralization(&gen, filename); err != nil {
		return nil, err
	}

	return &gen, nil
}

// validateUseCaseGeneralization validates an inputUseCaseGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
func validateUseCaseGeneralization(gen *inputUseCaseGeneralization, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if gen.Name == "" {
		return NewParseError(
			ErrUseCaseGenNameRequired,
			"use case generalization name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(gen.Name) == "" {
		return NewParseError(
			ErrUseCaseGenNameEmpty,
			"use case generalization name cannot be empty or whitespace only, got '"+gen.Name+"'",
			filename,
		).WithField("name")
	}

	// Superclass key is required (schema enforces this, but we provide a clearer error)
	if gen.SuperclassKey == "" {
		return NewParseError(
			ErrUseCaseGenSuperclassRequired,
			"use case generalization superclass_key is required, got ''",
			filename,
		).WithField("superclass_key")
	}

	// Superclass key cannot be only whitespace
	if strings.TrimSpace(gen.SuperclassKey) == "" {
		return NewParseError(
			ErrUseCaseGenSuperclassRequired,
			"use case generalization superclass_key cannot be empty or whitespace only, got '"+gen.SuperclassKey+"'",
			filename,
		).WithField("superclass_key")
	}

	// Subclass keys is required and must have at least one entry (schema enforces this)
	if len(gen.SubclassKeys) == 0 {
		return NewParseError(
			ErrUseCaseGenSubclassesRequired,
			"use case generalization subclass_keys is required and must have at least one entry",
			filename,
		).WithField("subclass_keys")
	}

	// Each subclass key must be non-empty and non-whitespace
	for i, key := range gen.SubclassKeys {
		if key == "" {
			return NewParseError(
				ErrUseCaseGenSubclassesEmpty,
				fmt.Sprintf("use case generalization subclass_keys[%d] cannot be empty", i),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
		if strings.TrimSpace(key) == "" {
			return NewParseError(
				ErrUseCaseGenSubclassesEmpty,
				fmt.Sprintf("use case generalization subclass_keys[%d] cannot be whitespace only, got '%s'", i, key),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
	}

	return nil
}
