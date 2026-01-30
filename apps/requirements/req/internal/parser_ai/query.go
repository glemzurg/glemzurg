package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputQuery represents a query JSON file.
type inputQuery struct {
	Name       string   `json:"name"`
	Details    string   `json:"details,omitempty"`
	Requires   []string `json:"requires,omitempty"`
	Guarantees []string `json:"guarantees,omitempty"`
}

var querySchema *jsonschema.Schema
var querySchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("query.schema.json")
	if err != nil {
		panic("failed to read query.schema.json: " + err.Error())
	}
	querySchemaContent = string(schemaBytes)
	if err := compiler.AddResource("query.schema.json", strings.NewReader(querySchemaContent)); err != nil {
		panic("failed to add query schema resource: " + err.Error())
	}
	querySchema, err = compiler.Compile("query.schema.json")
	if err != nil {
		panic("failed to compile query.schema.json: " + err.Error())
	}
}

// parseQuery parses and validates a query JSON file.
func parseQuery(content []byte, filename string) (*inputQuery, error) {
	var query inputQuery
	if err := json.Unmarshal(content, &query); err != nil {
		return nil, NewParseError(ErrQueryInvalidJSON, "failed to parse query JSON: "+err.Error(), filename)
	}

	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(ErrQueryInvalidJSON, "failed to parse query JSON for schema validation: "+err.Error(), filename)
	}

	if err := querySchema.Validate(jsonData); err != nil {
		return nil, NewParseError(ErrQuerySchemaViolation, "query JSON does not match schema: "+err.Error(), filename).WithSchema(querySchemaContent)
	}

	if err := validateQuery(&query, filename); err != nil {
		return nil, err
	}

	return &query, nil
}

// validateQuery performs custom validation beyond JSON schema.
func validateQuery(query *inputQuery, filename string) error {
	if query.Name == "" {
		return NewParseError(ErrQueryNameRequired, "query name is required, got ''", filename).WithField("name")
	}
	if strings.TrimSpace(query.Name) == "" {
		return NewParseError(ErrQueryNameEmpty, "query name cannot be empty or whitespace only, got '"+query.Name+"'", filename).WithField("name")
	}
	return nil
}
