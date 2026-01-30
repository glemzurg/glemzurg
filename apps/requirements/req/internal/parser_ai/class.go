package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputAttribute represents an attribute within a class.
type inputAttribute struct {
	Name             string `json:"name"`
	DataTypeRules    string `json:"data_type_rules,omitempty"`
	Details          string `json:"details,omitempty"`
	DerivationPolicy string `json:"derivation_policy,omitempty"`
	Nullable         bool   `json:"nullable,omitempty"`
	UMLComment       string `json:"uml_comment,omitempty"`
}

// inputClass represents a class JSON file.
type inputClass struct {
	Name       string                     `json:"name"`
	Details    string                     `json:"details,omitempty"`
	ActorKey   string                     `json:"actor_key,omitempty"`
	UMLComment string                     `json:"uml_comment,omitempty"`
	Attributes map[string]*inputAttribute `json:"attributes,omitempty"`
	Indexes    [][]string                 `json:"indexes,omitempty"`

	// Children (not from JSON, populated during directory traversal)
	StateMachine *inputStateMachine    `json:"-"`
	Actions      map[string]*inputAction `json:"-"`
	Queries      map[string]*inputQuery  `json:"-"`
}

// classSchema is the compiled JSON schema for class files.
var classSchema *jsonschema.Schema

// classSchemaContent is the raw JSON schema content for error reporting.
var classSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("class.schema.json")
	if err != nil {
		panic("failed to read class.schema.json: " + err.Error())
	}
	classSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("class.schema.json", strings.NewReader(classSchemaContent)); err != nil {
		panic("failed to add class schema resource: " + err.Error())
	}
	classSchema, err = compiler.Compile("class.schema.json")
	if err != nil {
		panic("failed to compile class.schema.json: " + err.Error())
	}
}

// parseClass parses a class JSON file content into an inputClass struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the class schema and returns detailed errors if validation fails.
func parseClass(content []byte, filename string) (*inputClass, error) {
	var class inputClass

	// Parse JSON
	if err := json.Unmarshal(content, &class); err != nil {
		return nil, NewParseError(
			ErrClassInvalidJSON,
			"failed to parse class JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrClassInvalidJSON,
			"failed to parse class JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := classSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrClassSchemaViolation,
			"class JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(classSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateClass(&class, filename); err != nil {
		return nil, err
	}

	return &class, nil
}

// validateClass validates an inputClass struct.
// The filename parameter is the path to the JSON file being parsed.
func validateClass(class *inputClass, filename string) error {
	// Name is required (schema enforces this, but we provide a clearer error)
	if class.Name == "" {
		return NewParseError(
			ErrClassNameRequired,
			"class name is required, got ''",
			filename,
		).WithField("name")
	}

	// Name cannot be only whitespace
	if strings.TrimSpace(class.Name) == "" {
		return NewParseError(
			ErrClassNameEmpty,
			"class name cannot be empty or whitespace only, got '"+class.Name+"'",
			filename,
		).WithField("name")
	}

	// Validate attributes if present
	for attrKey, attr := range class.Attributes {
		// Attribute name is required (schema enforces this)
		if attr.Name == "" {
			return NewParseError(
				ErrClassAttributeNameEmpty,
				fmt.Sprintf("attribute '%s' name is required, got ''", attrKey),
				filename,
			).WithField("attributes." + attrKey + ".name")
		}

		// Attribute name cannot be only whitespace
		if strings.TrimSpace(attr.Name) == "" {
			return NewParseError(
				ErrClassAttributeNameEmpty,
				fmt.Sprintf("attribute '%s' name cannot be empty or whitespace only, got '%s'", attrKey, attr.Name),
				filename,
			).WithField("attributes." + attrKey + ".name")
		}
	}

	// Validate indexes if present
	for i, index := range class.Indexes {
		// Each index must have at least one attribute (schema enforces minItems: 1)
		if len(index) == 0 {
			return NewParseError(
				ErrClassIndexInvalid,
				fmt.Sprintf("index[%d] must have at least one attribute key", i),
				filename,
			).WithField(fmt.Sprintf("indexes[%d]", i))
		}

		// Each attribute key in the index must be non-empty
		for j, attrKey := range index {
			if attrKey == "" {
				return NewParseError(
					ErrClassIndexInvalid,
					fmt.Sprintf("index[%d][%d] attribute key cannot be empty", i, j),
					filename,
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j))
			}
			if strings.TrimSpace(attrKey) == "" {
				return NewParseError(
					ErrClassIndexInvalid,
					fmt.Sprintf("index[%d][%d] attribute key cannot be whitespace only, got '%s'", i, j, attrKey),
					filename,
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j))
			}
		}
	}

	return nil
}
