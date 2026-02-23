package parser_ai

import (
	"encoding/json"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputModel represents the model.json file.
type inputModel struct {
	Name       string       `json:"name"`
	Details    string       `json:"details,omitempty"`
	Invariants []inputLogic `json:"invariants,omitempty"`

	// Children (not from JSON, populated during directory traversal)
	Actors                map[string]*inputActor                `json:"-"`
	ActorGeneralizations  map[string]*inputActorGeneralization  `json:"-"`
	GlobalFunctions       map[string]*inputGlobalFunction       `json:"-"`
	Domains               map[string]*inputDomain               `json:"-"`
	DomainAssociations    map[string]*inputDomainAssociation    `json:"-"`
	ClassAssociations     map[string]*inputClassAssociation          `json:"-"`
}

// modelSchema is the compiled JSON schema for model.json.
var modelSchema *jsonschema.Schema

// modelSchemaContent is the raw JSON schema content for error reporting.
var modelSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("model.schema.json")
	if err != nil {
		panic("failed to read model.schema.json: " + err.Error())
	}
	modelSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("model.schema.json", strings.NewReader(modelSchemaContent)); err != nil {
		panic("failed to add model schema resource: " + err.Error())
	}
	modelSchema, err = compiler.Compile("model.schema.json")
	if err != nil {
		panic("failed to compile model.schema.json: " + err.Error())
	}
}

// parseModel parses a model.json file content into an inputModel struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the model schema and returns detailed errors if validation fails.
func parseModel(content []byte, filename string) (*inputModel, error) {
	var model inputModel

	// Parse JSON
	if err := json.Unmarshal(content, &model); err != nil {
		return nil, NewParseError(
			ErrModelInvalidJSON,
			"failed to parse model JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrModelInvalidJSON,
			"failed to parse model JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := modelSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrModelSchemaViolation,
			"model JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(modelSchemaContent)
	}

	// Validate required fields
	if err := validateModel(&model, filename); err != nil {
		return nil, err
	}

	return &model, nil
}

// validateModel validates an inputModel struct.
// The filename parameter is the path to the JSON file being parsed.
func validateModel(model *inputModel, filename string) error {
	// Name is required
	if model.Name == "" {
		return NewParseError(
			ErrModelNameRequired,
			"model name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(model.Name) == "" {
		return NewParseError(
			ErrModelNameEmpty,
			"model name cannot be empty or whitespace only, got '"+model.Name+"'",
			filename,
		).WithField("name")
	}

	return nil
}
