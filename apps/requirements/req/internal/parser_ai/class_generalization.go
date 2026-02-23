package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputClassGeneralization represents a class generalization JSON file.
type inputClassGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UMLComment    string   `json:"uml_comment,omitempty"`
}

// classGeneralizationSchema is the compiled JSON schema for class generalization files.
var classGeneralizationSchema *jsonschema.Schema

// classGeneralizationSchemaContent is the raw JSON schema content for error reporting.
var classGeneralizationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("class_generalization.schema.json")
	if err != nil {
		panic("failed to read class_generalization.schema.json: " + err.Error())
	}
	classGeneralizationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("class_generalization.schema.json", strings.NewReader(classGeneralizationSchemaContent)); err != nil {
		panic("failed to add class generalization schema resource: " + err.Error())
	}
	classGeneralizationSchema, err = compiler.Compile("class_generalization.schema.json")
	if err != nil {
		panic("failed to compile class_generalization.schema.json: " + err.Error())
	}
}

// parseClassGeneralization parses a class generalization JSON file content into an inputClassGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the class generalization schema and returns detailed errors if validation fails.
func parseClassGeneralization(content []byte, filename string) (*inputClassGeneralization, error) {
	var gen inputClassGeneralization

	// Parse JSON
	if err := json.Unmarshal(content, &gen); err != nil {
		return nil, NewParseError(
			ErrClassGenInvalidJSON,
			"failed to parse class generalization JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrClassGenInvalidJSON,
			"failed to parse class generalization JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := classGeneralizationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrClassGenSchemaViolation,
			"class generalization JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(classGeneralizationSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateClassGeneralization(&gen, filename); err != nil {
		return nil, err
	}

	return &gen, nil
}

// validateClassGeneralization validates an inputClassGeneralization struct.
// The filename parameter is the path to the JSON file being parsed.
func validateClassGeneralization(gen *inputClassGeneralization, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if gen.Name == "" {
		return NewParseError(
			ErrClassGenNameRequired,
			"class generalization name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(gen.Name) == "" {
		return NewParseError(
			ErrClassGenNameEmpty,
			"class generalization name cannot be empty or whitespace only, got '"+gen.Name+"'",
			filename,
		).WithField("name")
	}

	// Superclass key is required (schema enforces this, but we provide a clearer error)
	if gen.SuperclassKey == "" {
		return NewParseError(
			ErrClassGenSuperclassRequired,
			"class generalization superclass_key is required, got ''",
			filename,
		).WithField("superclass_key")
	}

	// Superclass key cannot be only whitespace
	if strings.TrimSpace(gen.SuperclassKey) == "" {
		return NewParseError(
			ErrClassGenSuperclassRequired,
			"class generalization superclass_key cannot be empty or whitespace only, got '"+gen.SuperclassKey+"'",
			filename,
		).WithField("superclass_key")
	}

	// Subclass keys is required and must have at least one entry (schema enforces this)
	if len(gen.SubclassKeys) == 0 {
		return NewParseError(
			ErrClassGenSubclassesRequired,
			"class generalization subclass_keys is required and must have at least one entry",
			filename,
		).WithField("subclass_keys")
	}

	// Each subclass key must be non-empty and non-whitespace
	for i, key := range gen.SubclassKeys {
		if key == "" {
			return NewParseError(
				ErrClassGenSubclassesEmpty,
				fmt.Sprintf("class generalization subclass_keys[%d] cannot be empty", i),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
		if strings.TrimSpace(key) == "" {
			return NewParseError(
				ErrClassGenSubclassesEmpty,
				fmt.Sprintf("class generalization subclass_keys[%d] cannot be whitespace only, got '%s'", i, key),
				filename,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
	}

	return nil
}
