package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputUseCaseShared represents how a mud-level use case relates to a sea-level use case.
// The outer map key is the mud use case key, the inner map key is the sea use case key.
type inputUseCaseShared struct {
	ShareType  string `json:"share_type"`
	UmlComment string `json:"uml_comment,omitempty"`
}

// useCaseSharedSchema is the compiled JSON schema for use case shared files.
var useCaseSharedSchema *jsonschema.Schema

// useCaseSharedSchemaContent is the raw JSON schema content for error reporting.
var useCaseSharedSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("use_case_shared.schema.json")
	if err != nil {
		panic("failed to read use_case_shared.schema.json: " + err.Error())
	}
	useCaseSharedSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("use_case_shared.schema.json", strings.NewReader(useCaseSharedSchemaContent)); err != nil {
		panic("failed to add use case shared schema resource: " + err.Error())
	}
	useCaseSharedSchema, err = compiler.Compile("use_case_shared.schema.json")
	if err != nil {
		panic("failed to compile use_case_shared.schema.json: " + err.Error())
	}
}

// parseUseCaseShared parses a use case shared JSON file content into an inputUseCaseShared struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the use case shared schema and returns detailed errors if validation fails.
func parseUseCaseShared(content []byte, filename string) (*inputUseCaseShared, error) {
	var shared inputUseCaseShared

	// Parse JSON
	if err := json.Unmarshal(content, &shared); err != nil {
		return nil, NewParseError(
			ErrUseCaseSharedInvalidJSON,
			"failed to parse use case shared JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseSharedInvalidJSON,
			"failed to parse use case shared JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := useCaseSharedSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrUseCaseSharedSchemaViolation,
			"use case shared JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(useCaseSharedSchemaContent)
	}

	// Validate required fields
	if err := validateUseCaseShared(&shared, filename); err != nil {
		return nil, err
	}

	return &shared, nil
}

// validateUseCaseShared validates an inputUseCaseShared struct.
// The filename parameter is the path to the JSON file being parsed.
func validateUseCaseShared(shared *inputUseCaseShared, filename string) error {
	// ShareType is required
	if shared.ShareType == "" {
		return NewParseError(
			ErrUseCaseSharedShareTypeRequired,
			"use case shared share_type is required, got ''",
			filename,
		).WithField("share_type")
	}

	// ShareType cannot be only whitespace
	if strings.TrimSpace(shared.ShareType) == "" {
		return NewParseError(
			ErrUseCaseSharedShareTypeEmpty,
			"use case shared share_type cannot be empty or whitespace only, got '"+shared.ShareType+"'",
			filename,
		).WithField("share_type")
	}

	return nil
}
