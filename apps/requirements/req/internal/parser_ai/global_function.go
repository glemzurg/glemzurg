package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputGlobalFunction represents a global function/definition in JSON.
// Global functions are referenced from expressions throughout the model.
// Names must start with underscore (e.g., _Max, _SetOfValues).
type inputGlobalFunction struct {
	Name       string     `json:"name"`
	Parameters []string   `json:"parameters,omitempty"`
	Logic      inputLogic `json:"logic"`
}

// globalFunctionSchema is the compiled JSON schema for global function objects.
var globalFunctionSchema *jsonschema.Schema

// globalFunctionSchemaContent is the raw JSON schema content for error reporting.
var globalFunctionSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("global_function.schema.json")
	if err != nil {
		panic("failed to read global_function.schema.json: " + err.Error())
	}
	globalFunctionSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("global_function.schema.json", strings.NewReader(globalFunctionSchemaContent)); err != nil {
		panic("failed to add global function schema resource: " + err.Error())
	}
	globalFunctionSchema, err = compiler.Compile("global_function.schema.json")
	if err != nil {
		panic("failed to compile global_function.schema.json: " + err.Error())
	}
}

// parseGlobalFunction parses a global function JSON object content into an inputGlobalFunction struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the global function schema and returns detailed errors if validation fails.
func parseGlobalFunction(content []byte, filename string) (*inputGlobalFunction, error) {
	var gf inputGlobalFunction

	// Parse JSON
	if err := json.Unmarshal(content, &gf); err != nil {
		return nil, NewParseError(
			ErrGlobalFuncInvalidJSON,
			"failed to parse global function JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrGlobalFuncInvalidJSON,
			"failed to parse global function JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := globalFunctionSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrGlobalFuncSchemaViolation,
			"global function JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(globalFunctionSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateGlobalFunction(&gf, filename); err != nil {
		return nil, err
	}

	return &gf, nil
}

// validateGlobalFunction validates an inputGlobalFunction struct.
// The filename parameter is the path to the JSON file being parsed.
func validateGlobalFunction(gf *inputGlobalFunction, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if gf.Name == "" {
		return NewParseError(
			ErrGlobalFuncNameRequired,
			"global function name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(gf.Name) == "" {
		return NewParseError(
			ErrGlobalFuncNameEmpty,
			"global function name cannot be empty or whitespace only, got '"+gf.Name+"'",
			filename,
		).WithField("name")
	}

	// Name must start with underscore
	if !strings.HasPrefix(gf.Name, "_") {
		return NewParseError(
			ErrGlobalFuncNameNoUnderscore,
			"global function name must start with underscore, got '"+gf.Name+"'",
			filename,
		).WithField("name")
	}

	// Each parameter must be non-empty and non-whitespace
	for i, param := range gf.Parameters {
		if param == "" {
			return NewParseError(
				ErrGlobalFuncParamEmpty,
				fmt.Sprintf("global function parameters[%d] cannot be empty", i),
				filename,
			).WithField(fmt.Sprintf("parameters[%d]", i))
		}
		if strings.TrimSpace(param) == "" {
			return NewParseError(
				ErrGlobalFuncParamEmpty,
				fmt.Sprintf("global function parameters[%d] cannot be whitespace only, got '%s'", i, param),
				filename,
			).WithField(fmt.Sprintf("parameters[%d]", i))
		}
	}

	// Logic description is required (schema enforces this, but we provide a clearer error)
	if gf.Logic.Description == "" {
		return NewParseError(
			ErrGlobalFuncLogicRequired,
			"global function logic description is required, got ''",
			filename,
		).WithField("logic.description")
	}

	// Logic description cannot be only whitespace
	if strings.TrimSpace(gf.Logic.Description) == "" {
		return NewParseError(
			ErrGlobalFuncLogicRequired,
			"global function logic description cannot be empty or whitespace only, got '"+gf.Logic.Description+"'",
			filename,
		).WithField("logic.description")
	}

	return nil
}
