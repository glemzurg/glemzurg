package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputSubdomainAssociation represents a subdomain dependency JSON file at domain level.
type inputSubdomainAssociation struct {
	ProblemSubdomainKey  string `json:"problem_subdomain_key"`
	SolutionSubdomainKey string `json:"solution_subdomain_key"`
	UmlComment           string `json:"uml_comment,omitempty"`
}

var subdomainAssociationSchema *jsonschema.Schema
var subdomainAssociationSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("subdomain_association.schema.json")
	if err != nil {
		panic("failed to read subdomain_association.schema.json: " + err.Error())
	}
	subdomainAssociationSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("subdomain_association.schema.json", strings.NewReader(subdomainAssociationSchemaContent)); err != nil {
		panic("failed to add subdomain association schema resource: " + err.Error())
	}
	subdomainAssociationSchema, err = compiler.Compile("subdomain_association.schema.json")
	if err != nil {
		panic("failed to compile subdomain_association.schema.json: " + err.Error())
	}
}

func parseSubdomainAssociation(content []byte, filename string) (*inputSubdomainAssociation, error) {
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrSubdomainAssocInvalidJSON,
			"failed to parse subdomain association JSON: "+err.Error(),
			filename,
		).WithHint("ensure file contains valid JSON syntax")
	}
	if err := subdomainAssociationSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrSubdomainAssocSchemaViolation,
			"subdomain association JSON does not match schema: "+err.Error(),
			filename,
		).WithHint("run: req_check --schema subdomain_association")
	}

	var assoc inputSubdomainAssociation
	if err := json.Unmarshal(content, &assoc); err != nil {
		return nil, NewParseError(
			ErrSubdomainAssocInvalidJSON,
			"failed to parse subdomain association JSON: "+err.Error(),
			filename,
		).WithHint("ensure file contains valid JSON syntax")
	}

	if err := validateSubdomainAssoc(&assoc, filename); err != nil {
		return nil, err
	}

	return &assoc, nil
}

func validateSubdomainAssoc(assoc *inputSubdomainAssociation, filename string) error {
	if assoc.ProblemSubdomainKey == "" {
		return NewParseError(
			ErrSubdomainAssocProblemKeyRequired,
			"subdomain association problem_subdomain_key is required, got ''",
			filename,
		).WithField("problem_subdomain_key").WithHint("add a non-empty \"problem_subdomain_key\" referencing a defined subdomain")
	}
	if strings.TrimSpace(assoc.ProblemSubdomainKey) == "" {
		return NewParseError(
			ErrSubdomainAssocProblemKeyEmpty,
			"subdomain association problem_subdomain_key cannot be empty or whitespace only, got '"+assoc.ProblemSubdomainKey+"'",
			filename,
		).WithField("problem_subdomain_key").WithHint("add a non-empty \"problem_subdomain_key\" referencing a defined subdomain")
	}
	if assoc.SolutionSubdomainKey == "" {
		return NewParseError(
			ErrSubdomainAssocSolutionKeyRequired,
			"subdomain association solution_subdomain_key is required, got ''",
			filename,
		).WithField("solution_subdomain_key").WithHint("add a non-empty \"solution_subdomain_key\" referencing a defined subdomain")
	}
	if strings.TrimSpace(assoc.SolutionSubdomainKey) == "" {
		return NewParseError(
			ErrSubdomainAssocSolutionKeyEmpty,
			"subdomain association solution_subdomain_key cannot be empty or whitespace only, got '"+assoc.SolutionSubdomainKey+"'",
			filename,
		).WithField("solution_subdomain_key").WithHint("add a non-empty \"solution_subdomain_key\" referencing a defined subdomain")
	}
	if assoc.ProblemSubdomainKey == assoc.SolutionSubdomainKey {
		return NewParseError(
			ErrSubdomainAssocSameSubdomains,
			"subdomain association problem_subdomain_key and solution_subdomain_key cannot be the same, got '"+assoc.ProblemSubdomainKey+"'",
			filename,
		).WithField("problem_subdomain_key").WithHint("choose different problem and solution subdomain keys")
	}
	return nil
}
