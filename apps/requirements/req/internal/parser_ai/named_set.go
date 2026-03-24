package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputNamedSet represents a named set JSON file.
type inputNamedSet struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Notation      string `json:"notation,omitempty"`
	Specification string `json:"specification,omitempty"`
	TypeSpec      string `json:"type_spec,omitempty"`
}

// namedSetSchema is the compiled JSON schema for named set files.
var namedSetSchema *jsonschema.Schema

// namedSetSchemaContent is the raw JSON schema content for error reporting.
var namedSetSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("named_set.schema.json")
	if err != nil {
		panic("failed to read named_set.schema.json: " + err.Error())
	}
	namedSetSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("named_set.schema.json", strings.NewReader(namedSetSchemaContent)); err != nil {
		panic("failed to add named set schema resource: " + err.Error())
	}
	namedSetSchema, err = compiler.Compile("named_set.schema.json")
	if err != nil {
		panic("failed to compile named_set.schema.json: " + err.Error())
	}
}

// parseNamedSet parses a named set JSON file content into an inputNamedSet struct.
func parseNamedSet(content []byte, filename string) (*inputNamedSet, error) {
	// Validate JSON syntax and schema first (using untyped parse).
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrNamedSetInvalidJSON,
			"failed to parse named set JSON: "+err.Error(),
			filename,
		).WithHint("ensure file contains valid JSON syntax")
	}
	if err := namedSetSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrNamedSetSchemaViolation,
			"named set JSON does not match schema: "+err.Error(),
			filename,
		).WithHint("run: req_check --schema named_set")
	}

	// Unmarshal into typed struct (schema already validated structure).
	var ns inputNamedSet
	if err := json.Unmarshal(content, &ns); err != nil {
		return nil, NewParseError(
			ErrNamedSetInvalidJSON,
			"failed to parse named set JSON: "+err.Error(),
			filename,
		).WithHint("ensure file contains valid JSON syntax")
	}

	if err := validateNamedSet(&ns, filename); err != nil {
		return nil, err
	}

	return &ns, nil
}

// validateNamedSet validates an inputNamedSet struct.
func validateNamedSet(ns *inputNamedSet, filename string) error {
	if ns.Name == "" {
		return NewParseError(
			ErrNamedSetNameRequired,
			"named set name is required, got ''",
			filename,
		).WithField("name").WithHint("add a non-empty \"name\" field starting with underscore")
	}

	if strings.TrimSpace(ns.Name) == "" {
		return NewParseError(
			ErrNamedSetNameEmpty,
			"named set name cannot be empty or whitespace only, got '"+ns.Name+"'",
			filename,
		).WithField("name").WithHint("add a non-empty \"name\" field starting with underscore")
	}

	if !strings.HasPrefix(ns.Name, "_") {
		return NewParseError(
			ErrNamedSetNameNoUnderscore,
			"named set name must start with underscore, got '"+ns.Name+"'",
			filename,
		).WithField("name").WithHint("named set names must start with underscore, e.g. \"_OrderStatuses\"")
	}

	return nil
}
