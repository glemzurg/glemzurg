package model_class

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
)

// Attribute is a member of a class.
type Attribute struct {
	Key              identity.Key
	Name             string             `validate:"required"`
	Details          string             // Markdown.
	DataTypeRules    string             // What are the bounds of this data type.
	DerivationPolicy *model_logic.Logic `validate:"-"` // If this is a derived attribute, the logic for how it is derived.
	Nullable         bool               // Is this attribute optional.
	UmlComment       string
	// Children
	IndexNums  []uint                    // The indexes this attribute is part of.
	DataType   *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
	Invariants []model_logic.Logic       // Invariants that must hold for this attribute's value.
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
		parsedDataType, err := model_data_type.New(dataTypeKey, attribute.DataTypeRules, nil)

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
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return err
	}
	if a.Key.KeyType != identity.KEY_TYPE_ATTRIBUTE {
		return errors.Errorf("Key: invalid key type '%s' for attribute.", a.Key.KeyType)
	}

	// Validate struct tags (Name required).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	// Validate the derivation policy logic if present.
	if a.DerivationPolicy != nil {
		if err := a.DerivationPolicy.Validate(); err != nil {
			return errors.Wrapf(err, "attribute %s: DerivationPolicy", a.Name)
		}
		if a.DerivationPolicy.Type != model_logic.LogicTypeValue {
			return errors.Errorf("attribute %s: DerivationPolicy logic kind must be '%s', got '%s'", a.Name, model_logic.LogicTypeValue, a.DerivationPolicy.Type)
		}
	}

	// Validate invariants.
	for i, inv := range a.Invariants {
		if err := inv.Validate(); err != nil {
			return errors.Wrapf(err, "attribute invariant %d", i)
		}
		if inv.Type != model_logic.LogicTypeAssessment {
			return errors.Errorf("attribute invariant %d: logic kind must be '%s', got '%s'", i, model_logic.LogicTypeAssessment, inv.Type)
		}
	}

	return nil
}

// SetInvariants sets the invariants for this attribute.
func (a *Attribute) SetInvariants(invariants []model_logic.Logic) {
	a.Invariants = invariants
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
	// Validate derivation policy with attribute as parent.
	if a.DerivationPolicy != nil {
		if err := a.DerivationPolicy.ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "attribute %s: DerivationPolicy", a.Name)
		}
	}
	// Validate invariants with attribute as parent.
	for i, inv := range a.Invariants {
		if err := inv.ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "attribute invariant %d", i)
		}
	}
	return nil
}
