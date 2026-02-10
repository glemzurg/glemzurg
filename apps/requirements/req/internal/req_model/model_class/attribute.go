package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Attribute is a member of a class.
type Attribute struct {
	Key              identity.Key
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy *model_logic.Logic // If this is a derived attribute, the logic for how it is derived.
	Nullable         bool               // Is this attribute optional.
	UmlComment       string
	// Children
	IndexNums []uint                    // The indexes this attribute is part of.
	DataType  *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

func NewAttribute(key identity.Key, name, details, dataTypeRules string, derivationPolicy *model_logic.Logic, nullable bool, umlComment string, indexNums []uint) (attribute Attribute, err error) {

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

	if err = attribute.Validate(); err != nil {
		return Attribute{}, err
	}

	return attribute, nil
}

// Validate validates the Attribute struct.
func (a *Attribute) Validate() error {
	if err := validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ATTRIBUTE {
				return errors.Errorf("invalid key type '%s' for attribute", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.Name, validation.Required),
	); err != nil {
		return err
	}

	// Validate the derivation policy logic if present.
	if a.DerivationPolicy != nil {
		if err := a.DerivationPolicy.Validate(); err != nil {
			return errors.Wrapf(err, "attribute %s: DerivationPolicy", a.Name)
		}
	}

	return nil
}

// ValidateWithParent validates the Attribute, its key's parent relationship, and all children.
// The parent must be a Class.
func (a *Attribute) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Attribute has no children with keys that need validation.
	return nil
}
