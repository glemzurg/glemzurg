package model_class

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Attribute is a member of a class.
type Attribute struct {
	Key              identity.Key
	Name             string
	Details          string             // Markdown.
	DataTypeRules    string             // What are the bounds of this data type.
	DerivationPolicy *model_logic.Logic // If this is a derived attribute, the logic for how it is derived.
	Nullable         bool               // Is this attribute optional.
	UmlComment       string
	// Children
	IndexNums  []uint                    // The indexes this attribute is part of.
	DataType   *model_data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
	Invariants []model_logic.Logic       // Invariants that must hold for this attribute's value.
}

// AttributeAnnotations holds optional annotation data for an attribute.
type AttributeAnnotations struct {
	UmlComment string
	IndexNums  []uint
}

func NewAttribute(key identity.Key, name, details, dataTypeRules string, derivationPolicy *model_logic.Logic, nullable bool, annotations AttributeAnnotations) (attribute Attribute, err error) {
	attribute = Attribute{
		Key:              key,
		Name:             name,
		Details:          details,
		DataTypeRules:    dataTypeRules,
		DerivationPolicy: derivationPolicy,
		Nullable:         nullable,
		UmlComment:       annotations.UmlComment,
		IndexNums:        annotations.IndexNums,
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

	return attribute, nil
}

// Validate validates the Attribute struct.
func (a *Attribute) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := a.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AttrKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_ATTRIBUTE {
		return coreerr.NewWithValues(ctx, coreerr.AttrKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for attribute", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_ATTRIBUTE)
	}

	// Name is required.
	if a.Name == "" {
		return coreerr.New(ctx, coreerr.AttrNameRequired, "Name is required", "Name")
	}

	// Validate the derivation policy logic if present.
	if a.DerivationPolicy != nil {
		derivCtx := ctx.Child("derivationPolicy", a.Key.String())
		if err := a.DerivationPolicy.Validate(derivCtx); err != nil {
			return coreerr.New(derivCtx, coreerr.AttrDerivationTypeInvalid, fmt.Sprintf("attribute %s: DerivationPolicy: %s", a.Name, err.Error()), "DerivationPolicy")
		}
		if a.DerivationPolicy.Type != model_logic.LogicTypeValue {
			return coreerr.NewWithValues(ctx, coreerr.AttrDerivationTypeInvalid, fmt.Sprintf("attribute %s: DerivationPolicy logic kind must be '%s', got '%s'", a.Name, model_logic.LogicTypeValue, a.DerivationPolicy.Type), "DerivationPolicy", a.DerivationPolicy.Type, string(model_logic.LogicTypeValue))
		}
	}

	// Validate invariants.
	attrInvLetTargets := make(map[string]bool)
	for i, inv := range a.Invariants {
		invCtx := ctx.Child("invariant", fmt.Sprintf("%d", i))
		if err := inv.Validate(invCtx); err != nil {
			return coreerr.New(invCtx, coreerr.AttrInvariantTypeInvalid, fmt.Sprintf("attribute invariant %d: %s", i, err.Error()), "Invariants")
		}
		if inv.Type != model_logic.LogicTypeAssessment && inv.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(invCtx, coreerr.AttrInvariantTypeInvalid, fmt.Sprintf("attribute invariant %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, inv.Type), "Invariants", inv.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if inv.Type == model_logic.LogicTypeLet {
			if attrInvLetTargets[inv.Target] {
				return coreerr.NewWithValues(invCtx, coreerr.AttrInvariantDuplicateLet, fmt.Sprintf("attribute invariant %d: duplicate let target %q", i, inv.Target), "Invariants", inv.Target, "")
			}
			attrInvLetTargets[inv.Target] = true
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
func (a *Attribute) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Validate derivation policy with attribute as parent.
	if a.DerivationPolicy != nil {
		derivCtx := ctx.Child("derivationPolicy", a.Key.String())
		if err := a.DerivationPolicy.ValidateWithParent(derivCtx, &a.Key); err != nil {
			return coreerr.New(derivCtx, coreerr.AttrDerivationTypeInvalid, fmt.Sprintf("attribute %s: DerivationPolicy: %s", a.Name, err.Error()), "DerivationPolicy")
		}
	}
	// Validate invariants with attribute as parent.
	for i, inv := range a.Invariants {
		invCtx := ctx.Child("invariant", fmt.Sprintf("%d", i))
		if err := inv.ValidateWithParent(invCtx, &a.Key); err != nil {
			return coreerr.New(invCtx, coreerr.AttrInvariantTypeInvalid, fmt.Sprintf("attribute invariant %d: %s", i, err.Error()), "Invariants")
		}
	}
	return nil
}
