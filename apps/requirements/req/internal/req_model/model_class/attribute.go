package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Attribute is a member of a class.
type Attribute struct {
	Key              identity.Key
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy string // If this is a derived attribute, how is it derived.
	Nullable         bool   // Is this attribute optional.
	UmlComment       string
	// Part of the data in a parsed file.
	IndexNums []uint                    // The indexes this attribute is part of.
	DataType  *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

func NewAttribute(key identity.Key, name, details, dataTypeRules, derivationPolicy string, nullable bool, umlComment string, indexNums []uint) (attribute Attribute, err error) {

	attribute = Attribute{
		Key:              key,
		Name:             name,
		Details:          details,
		DataTypeRules:    dataTypeRules,
		DerivationPolicy: derivationPolicy,
		Nullable:         nullable,
		UmlComment:       umlComment,
		IndexNums:        indexNums,
	}

	// Parse the data type rules into a DataType object if possible.
	if attribute.DataTypeRules != "" {

		// Use the attribute key as the key of this data type.
		dataTypeKey := attribute.Key.String()
		parsedDataType, err := model_data_type.New(dataTypeKey, attribute.DataTypeRules)

		// Only an error if it is not a parse error.
		var parseError *model_data_type.CannotParseError // Use a pointer for type checking.
		if err != nil && !errors.As(err, &parseError) {
			return Attribute{}, err
		}

		// If parse error then save the dataype.
		attribute.DataType = parsedDataType
	}

	err = validation.ValidateStruct(&attribute,
		validation.Field(&attribute.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ATTRIBUTE {
				return errors.Errorf("invalid key type '%s' for attribute", k.KeyType())
			}
			return nil
		})),
		validation.Field(&attribute.Name, validation.Required),
	)
	if err != nil {
		return Attribute{}, errors.WithStack(err)
	}

	return attribute, nil
}
