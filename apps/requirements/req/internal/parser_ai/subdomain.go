package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputSubdomain represents a subdomain.json file.
type inputSubdomain struct {
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	UMLComment string `json:"uml_comment,omitempty"`

	// Children (not from JSON, populated during directory traversal)
	Classes                  map[string]*inputClass                          `json:"-"`
	Generalizations          map[string]*inputClassGeneralization                 `json:"-"`
	Associations             map[string]*inputAssociation                    `json:"-"`
	UseCases                 map[string]*inputUseCase                        `json:"-"`
	UseCaseGeneralizations   map[string]*inputUseCaseGeneralization          `json:"-"`
	UseCaseShares            map[string]map[string]*inputUseCaseShared       `json:"-"`
}

// subdomainSchema is the compiled JSON schema for subdomain files.
var subdomainSchema *jsonschema.Schema

// subdomainSchemaContent is the raw JSON schema content for error reporting.
var subdomainSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("subdomain.schema.json")
	if err != nil {
		panic("failed to read subdomain.schema.json: " + err.Error())
	}
	subdomainSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("subdomain.schema.json", strings.NewReader(subdomainSchemaContent)); err != nil {
		panic("failed to add subdomain schema resource: " + err.Error())
	}
	subdomainSchema, err = compiler.Compile("subdomain.schema.json")
	if err != nil {
		panic("failed to compile subdomain.schema.json: " + err.Error())
	}
}

// parseSubdomain parses a subdomain JSON file content into an inputSubdomain struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the subdomain schema and returns detailed errors if validation fails.
func parseSubdomain(content []byte, filename string) (*inputSubdomain, error) {
	var subdomain inputSubdomain

	// Parse JSON
	if err := json.Unmarshal(content, &subdomain); err != nil {
		return nil, NewParseError(
			ErrSubdomainInvalidJSON,
			"failed to parse subdomain JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrSubdomainInvalidJSON,
			"failed to parse subdomain JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := subdomainSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrSubdomainSchemaViolation,
			"subdomain JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(subdomainSchemaContent)
	}

	// Validate required fields
	if err := validateSubdomain(&subdomain, filename); err != nil {
		return nil, err
	}

	return &subdomain, nil
}

// validateSubdomain validates an inputSubdomain struct.
// The filename parameter is the path to the JSON file being parsed.
func validateSubdomain(subdomain *inputSubdomain, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if subdomain.Name == "" {
		return NewParseError(
			ErrSubdomainNameRequired,
			"subdomain name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(subdomain.Name) == "" {
		return NewParseError(
			ErrSubdomainNameEmpty,
			"subdomain name cannot be empty or whitespace only, got '"+subdomain.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
