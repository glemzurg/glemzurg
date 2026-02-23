package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputClassAssociation represents an association JSON file.
type inputClassAssociation struct {
	Name                string  `json:"name"`
	Details             string  `json:"details,omitempty"`
	FromClassKey        string  `json:"from_class_key"`
	FromMultiplicity    string  `json:"from_multiplicity"`
	ToClassKey          string  `json:"to_class_key"`
	ToMultiplicity      string  `json:"to_multiplicity"`
	AssociationClassKey *string `json:"association_class_key,omitempty"`
	UmlComment          string  `json:"uml_comment,omitempty"`
}

var associationSchema *jsonschema.Schema
var associationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("association.schema.json")
	if err != nil {
		panic("failed to read association.schema.json: " + err.Error())
	}
	associationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("association.schema.json", strings.NewReader(associationSchemaContent)); err != nil {
		panic("failed to add association schema resource: " + err.Error())
	}
	associationSchema, err = compiler.Compile("association.schema.json")
	if err != nil {
		panic("failed to compile association.schema.json: " + err.Error())
	}
}

// parseAssociation parses and validates an association JSON file.
func parseAssociation(content []byte, filename string) (*inputClassAssociation, error) {
	var assoc inputClassAssociation
	if err := json.Unmarshal(content, &assoc); err != nil {
		return nil, NewParseError(ErrAssocInvalidJSON, "failed to parse association JSON: "+err.Error(), filename)
	}

	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(ErrAssocInvalidJSON, "failed to parse association JSON for schema validation: "+err.Error(), filename)
	}

	if err := associationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(ErrAssocSchemaViolation, "association JSON does not match schema: "+err.Error(), filename).WithSchema(associationSchemaContent)
	}

	if err := validateAssociation(&assoc, filename); err != nil {
		return nil, err
	}

	return &assoc, nil
}

// validateAssociation performs custom validation beyond JSON schema.
func validateAssociation(assoc *inputClassAssociation, filename string) error {
	// Validate name
	if assoc.Name == "" {
		return NewParseError(ErrAssocNameRequired, "association name is required, got ''", filename).WithField("name")
	}
	if strings.TrimSpace(assoc.Name) == "" {
		return NewParseError(ErrAssocNameEmpty, "association name cannot be empty or whitespace only, got '"+assoc.Name+"'", filename).WithField("name")
	}

	// Validate from_class_key
	if assoc.FromClassKey == "" {
		return NewParseError(ErrAssocFromClassRequired, "association from_class_key is required, got ''", filename).WithField("from_class_key")
	}
	if strings.TrimSpace(assoc.FromClassKey) == "" {
		return NewParseError(ErrAssocFromClassRequired, "association from_class_key cannot be empty or whitespace only, got '"+assoc.FromClassKey+"'", filename).WithField("from_class_key")
	}

	// Validate from_multiplicity
	if assoc.FromMultiplicity == "" {
		return NewParseError(ErrAssocFromMultRequired, "association from_multiplicity is required, got ''", filename).WithField("from_multiplicity")
	}
	if strings.TrimSpace(assoc.FromMultiplicity) == "" {
		return NewParseError(ErrAssocFromMultRequired, "association from_multiplicity cannot be empty or whitespace only, got '"+assoc.FromMultiplicity+"'", filename).WithField("from_multiplicity")
	}

	// Validate to_class_key
	if assoc.ToClassKey == "" {
		return NewParseError(ErrAssocToClassRequired, "association to_class_key is required, got ''", filename).WithField("to_class_key")
	}
	if strings.TrimSpace(assoc.ToClassKey) == "" {
		return NewParseError(ErrAssocToClassRequired, "association to_class_key cannot be empty or whitespace only, got '"+assoc.ToClassKey+"'", filename).WithField("to_class_key")
	}

	// Validate to_multiplicity
	if assoc.ToMultiplicity == "" {
		return NewParseError(ErrAssocToMultRequired, "association to_multiplicity is required, got ''", filename).WithField("to_multiplicity")
	}
	if strings.TrimSpace(assoc.ToMultiplicity) == "" {
		return NewParseError(ErrAssocToMultRequired, "association to_multiplicity cannot be empty or whitespace only, got '"+assoc.ToMultiplicity+"'", filename).WithField("to_multiplicity")
	}

	return nil
}
