// Package loader provides JSON serialization and deserialization for req_model.Model.
package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/pkg/errors"
)

// LoadModel reads a JSON file and returns a req_model.Model.
// After deserialization, it re-parses DataTypeRules for any attributes
// where DataType is nil and DataTypeRules is non-empty, then validates
// the model.
func LoadModel(path string) (*req_model.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading model file: %w", err)
	}

	var model req_model.Model
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("parsing model JSON: %w", err)
	}

	// Re-parse DataTypeRules for attributes that lost their DataType during serialization.
	if err := reParseDataTypes(&model); err != nil {
		return nil, fmt.Errorf("re-parsing data types: %w", err)
	}

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("validating model: %w", err)
	}

	return &model, nil
}

// SaveModel writes a req_model.Model to a JSON file.
func SaveModel(model *req_model.Model, path string) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling model: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing model file: %w", err)
	}

	return nil
}

// reParseDataTypes walks all classes and re-parses DataTypeRules for any
// attributes where DataType is nil but DataTypeRules is non-empty.
func reParseDataTypes(model *req_model.Model) error {
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				for attrKey, attr := range class.Attributes {
					if attr.DataType == nil && attr.DataTypeRules != "" {
						dataTypeKey := attr.Key.String()
						parsedDataType, err := model_data_type.New(dataTypeKey, attr.DataTypeRules, nil)

						// Only an error if it is not a parse error.
						var parseError *model_data_type.CannotParseError
						if err != nil && !errors.As(err, &parseError) {
							return fmt.Errorf("attribute %s: %w", attr.Name, err)
						}

						attr.DataType = parsedDataType
						class.Attributes[attrKey] = attr
					}
				}
				subdomain.Classes[classKey] = class
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}
	return nil
}
