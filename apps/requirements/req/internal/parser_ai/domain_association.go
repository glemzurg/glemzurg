package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputDomainAssociation represents a domain-level association JSON file.
// Domain associations describe constraint relationships between domains
// (problem domain enforces requirements on solution domain).
type inputDomainAssociation struct {
	ProblemDomainKey  string `json:"problem_domain_key"`
	SolutionDomainKey string `json:"solution_domain_key"`
	UmlComment        string `json:"uml_comment,omitempty"`
}

// domainAssociationSchema is the compiled JSON schema for domain association files.
var domainAssociationSchema *jsonschema.Schema

// domainAssociationSchemaContent is the raw JSON schema content for error reporting.
var domainAssociationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("domain_association.schema.json")
	if err != nil {
		panic("failed to read domain_association.schema.json: " + err.Error())
	}
	domainAssociationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("domain_association.schema.json", strings.NewReader(domainAssociationSchemaContent)); err != nil {
		panic("failed to add domain association schema resource: " + err.Error())
	}
	domainAssociationSchema, err = compiler.Compile("domain_association.schema.json")
	if err != nil {
		panic("failed to compile domain_association.schema.json: " + err.Error())
	}
}

// parseDomainAssociation parses a domain association JSON file content into an inputDomainAssociation struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the domain association schema and returns detailed errors if validation fails.
func parseDomainAssociation(content []byte, filename string) (*inputDomainAssociation, error) {
	var assoc inputDomainAssociation

	// Parse JSON
	if err := json.Unmarshal(content, &assoc); err != nil {
		return nil, NewParseError(
			ErrDomainAssocInvalidJSON,
			"failed to parse domain association JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrDomainAssocInvalidJSON,
			"failed to parse domain association JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := domainAssociationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrDomainAssocSchemaViolation,
			"domain association JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(domainAssociationSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateDomainAssoc(&assoc, filename); err != nil {
		return nil, err
	}

	return &assoc, nil
}

// validateDomainAssoc validates an inputDomainAssociation struct.
// The filename parameter is the path to the JSON file being parsed.
func validateDomainAssoc(assoc *inputDomainAssociation, filename string) error {
	// Problem domain key is required
	if assoc.ProblemDomainKey == "" {
		return NewParseError(
			ErrDomainAssocProblemKeyRequired,
			"domain association problem_domain_key is required, got ''",
			filename,
		).WithField("problem_domain_key")
	}

	// Problem domain key cannot be only whitespace
	if strings.TrimSpace(assoc.ProblemDomainKey) == "" {
		return NewParseError(
			ErrDomainAssocProblemKeyEmpty,
			"domain association problem_domain_key cannot be empty or whitespace only, got '"+assoc.ProblemDomainKey+"'",
			filename,
		).WithField("problem_domain_key")
	}

	// Solution domain key is required
	if assoc.SolutionDomainKey == "" {
		return NewParseError(
			ErrDomainAssocSolutionKeyRequired,
			"domain association solution_domain_key is required, got ''",
			filename,
		).WithField("solution_domain_key")
	}

	// Solution domain key cannot be only whitespace
	if strings.TrimSpace(assoc.SolutionDomainKey) == "" {
		return NewParseError(
			ErrDomainAssocSolutionKeyEmpty,
			"domain association solution_domain_key cannot be empty or whitespace only, got '"+assoc.SolutionDomainKey+"'",
			filename,
		).WithField("solution_domain_key")
	}

	return nil
}
