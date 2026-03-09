package model_state

import (
	"errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
)

// Parameter is a typed parameter for actions and queries.
type Parameter struct {
	Name          string
	DataTypeRules string                    // What are the bounds of this data type.
	DataType      *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

func NewParameter(name, dataTypeRules string) (param Parameter, err error) {
	param = Parameter{
		Name:          name,
		DataTypeRules: dataTypeRules,
	}

	// Parse the data type rules into a DataType object if possible.
	if param.DataTypeRules != "" {
		// Use the parameter name as the key of this data type.
		parsedDataType, parseErr := model_data_type.New(name, param.DataTypeRules, nil)

		// Only an error if it is not a parse error.
		var cannotParseError *model_data_type.CannotParseError
		if parseErr != nil && !isCannotParseError(parseErr, &cannotParseError) {
			return Parameter{}, parseErr
		}

		// If successfully parsed, save the datatype.
		param.DataType = parsedDataType
	}

	return param, nil
}

// isCannotParseError checks if the error is a CannotParseError using errors.As.
func isCannotParseError(err error, target **model_data_type.CannotParseError) bool {
	return errors.As(err, target)
}

// Validate validates the Parameter struct.
func (p *Parameter) Validate() error {
	if p.Name == "" {
		return coreerr.New(coreerr.ParamNameRequired, "Name is required", "Name")
	}
	if p.DataTypeRules == "" {
		return coreerr.New(coreerr.ParamDatatypesRequired, "DataTypeRules is required", "DataTypeRules")
	}
	return nil
}

// ValidateWithParent validates the Parameter.
// Parameter has no key, so parent validation is not applicable.
func (p *Parameter) ValidateWithParent() error {
	return p.Validate()
}
