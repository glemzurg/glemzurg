package model_state

import (
	"errors"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Parameter is a typed parameter for actions and queries.
type Parameter struct {
	Key           identity.Key
	Name          string
	DataTypeRules string                    // What are the bounds of this data type.
	Nullable      bool                      // Whether absent (NULL) is a valid value.
	DataType      *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

// NewParameter constructs a Parameter whose identity.Key is parented by the owning
// action or query. The data type rules are parsed into a DataType if they
// can be parsed; parse failures are tolerated (DataType stays nil) but other errors
// are propagated. Name is preserved verbatim; the key's subKey uses the normalized
// form so it satisfies the identifier pattern (same convention as Attribute).
func NewParameter(parentKey identity.Key, name, dataTypeRules string, nullable bool) (param Parameter, err error) {
	paramKey, err := identity.NewParameterKey(parentKey, identity.NormalizeSubKey(name))
	if err != nil {
		return Parameter{}, err
	}

	param = Parameter{
		Key:           paramKey,
		Name:          name,
		DataTypeRules: dataTypeRules,
		Nullable:      nullable,
	}

	if param.DataTypeRules != "" {
		dataTypeKey, err := identity.NewDataTypeKey(paramKey, "")
		if err != nil {
			return Parameter{}, err
		}
		parsedDataType, parseErr := model_data_type.New(dataTypeKey, param.DataTypeRules, nil)
		var cannotParseError *model_data_type.CannotParseError
		if parseErr != nil && !isCannotParseError(parseErr, &cannotParseError) {
			return Parameter{}, parseErr
		}
		param.DataType = parsedDataType
	}

	return param, nil
}

// isCannotParseError checks if the error is a CannotParseError using errors.As.
func isCannotParseError(err error, target **model_data_type.CannotParseError) bool {
	return errors.As(err, target)
}

// Validate validates the Parameter struct.
func (p *Parameter) Validate(ctx *coreerr.ValidationContext) error {
	if err := p.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.ParamKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if p.Key.KeyType != identity.KEY_TYPE_PARAMETER {
		return coreerr.NewWithValues(ctx, coreerr.ParamKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for parameter", p.Key.KeyType), "Key", p.Key.KeyType, identity.KEY_TYPE_PARAMETER)
	}
	if p.Name == "" {
		return coreerr.New(ctx, coreerr.ParamNameRequired, "Name is required", "Name")
	}
	if p.DataTypeRules == "" {
		return coreerr.New(ctx, coreerr.ParamDatatypesRequired, "DataTypeRules is required", "DataTypeRules")
	}
	// The DataType, if present, must be a typed datatype key parented by this parameter.
	if p.DataType != nil {
		if p.DataType.Key.KeyType != identity.KEY_TYPE_DATA_TYPE {
			return coreerr.NewWithValues(
				ctx,
				coreerr.ParamDatatypeKeyMismatch,
				fmt.Sprintf("DataType.Key has wrong KeyType '%s', want '%s'", p.DataType.Key.KeyType, identity.KEY_TYPE_DATA_TYPE),
				"DataType.Key",
				p.DataType.Key.KeyType,
				identity.KEY_TYPE_DATA_TYPE,
			)
		}
		if p.DataType.Key.ParentKey != p.Key.String() {
			return coreerr.NewWithValues(
				ctx,
				coreerr.ParamDatatypeKeyMismatch,
				fmt.Sprintf("DataType.Key parent '%s' does not match Parameter.Key '%s'", p.DataType.Key.ParentKey, p.Key.String()),
				"DataType.Key.ParentKey",
				p.DataType.Key.ParentKey,
				p.Key.String(),
			)
		}
	}
	return nil
}

// ValidateWithParent validates the Parameter and verifies its key is parented by
// the given action or query key.
func (p *Parameter) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	if err := p.Validate(ctx); err != nil {
		return err
	}
	if err := p.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	return nil
}
