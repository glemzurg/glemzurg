package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/glemzurg/go-tlaplus/internal/req_model/model_data_type"
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
		parsedDataType, parseErr := model_data_type.New(name, param.DataTypeRules)

		// Only an error if it is not a parse error.
		var cannotParseError *model_data_type.CannotParseError
		if parseErr != nil && !isCannotParseError(parseErr, &cannotParseError) {
			return Parameter{}, parseErr
		}

		// If successfully parsed, save the datatype.
		param.DataType = parsedDataType
	}

	if err = param.Validate(); err != nil {
		return Parameter{}, err
	}

	return param, nil
}

// isCannotParseError checks if the error is a CannotParseError using type assertion.
func isCannotParseError(err error, target **model_data_type.CannotParseError) bool {
	if err == nil {
		return false
	}
	// Try type assertion
	if e, ok := err.(*model_data_type.CannotParseError); ok {
		*target = e
		return true
	}
	return false
}

// Validate validates the Parameter struct.
func (p *Parameter) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.DataTypeRules, validation.Required),
	)
}

// ValidateWithParent validates the Parameter.
// Parameter has no key, so parent validation is not applicable.
func (p *Parameter) ValidateWithParent() error {
	return p.Validate()
}
