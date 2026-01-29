package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputDomain represents a domain.json file.
type inputDomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	Realized   bool   `json:"realized,omitempty"`
	UMLComment string `json:"uml_comment,omitempty"`
}

// domainSchema is the compiled JSON schema for domain files.
var domainSchema *jsonschema.Schema

// domainSchemaContent is the raw JSON schema content for error reporting.
var domainSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("domain.schema.json")
	if err != nil {
		panic("failed to read domain.schema.json: " + err.Error())
	}
	domainSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("domain.schema.json", strings.NewReader(domainSchemaContent)); err != nil {
		panic("failed to add domain schema resource: " + err.Error())
	}
	domainSchema, err = compiler.Compile("domain.schema.json")
	if err != nil {
		panic("failed to compile domain.schema.json: " + err.Error())
	}
}

// parseDomain parses a domain JSON file content into an inputDomain struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the domain schema and returns detailed errors if validation fails.
func parseDomain(content []byte, filename string) (*inputDomain, error) {
	var domain inputDomain

	// Parse JSON
	if err := json.Unmarshal(content, &domain); err != nil {
		return nil, NewParseError(
			ErrDomainInvalidJSON,
			"failed to parse domain JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrDomainInvalidJSON,
			"failed to parse domain JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := domainSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrDomainSchemaViolation,
			"domain JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(domainSchemaContent)
	}

	// Validate required fields
	if err := validateDomain(&domain, filename); err != nil {
		return nil, err
	}

	return &domain, nil
}

// validateDomain validates an inputDomain struct.
// The filename parameter is the path to the JSON file being parsed.
func validateDomain(domain *inputDomain, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if domain.Name == "" {
		return NewParseError(
			ErrDomainNameRequired,
			"domain name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(domain.Name) == "" {
		return NewParseError(
			ErrDomainNameEmpty,
			"domain name cannot be empty or whitespace only, got '"+domain.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
