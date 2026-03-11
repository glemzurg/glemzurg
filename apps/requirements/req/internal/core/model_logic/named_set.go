package model_logic

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// NamedSet represents a reusable set definition at the model level.
// Named sets define well-known collections that can be referenced
// from behavioral logic (requires, guarantees, etc.) via NamedSetRef expressions.
type NamedSet struct {
	Key         identity.Key              // Unique key of type "nset".
	Name        string                    // Required: display name.
	Description string                    // Optional description.
	Spec        logic_spec.ExpressionSpec // Notation + Specification + Expression for the set definition.
	TypeSpec    *logic_spec.TypeSpec      // Optional precise type specification.
}

// NewNamedSet creates a new NamedSet.
func NewNamedSet(key identity.Key, name, description string, spec logic_spec.ExpressionSpec, typeSpec *logic_spec.TypeSpec) NamedSet {
	return NamedSet{
		Key:         key,
		Name:        name,
		Description: description,
		Spec:        spec,
		TypeSpec:    typeSpec,
	}
}

// Validate validates the NamedSet struct.
func (ns *NamedSet) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := ns.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.NsetKeyInvalid, fmt.Sprintf("NamedSet key failed validation: %s", err.Error()), "Key")
	}
	if ns.Key.KeyType != identity.KEY_TYPE_NAMED_SET {
		return coreerr.NewWithValues(ctx, coreerr.NsetKeyTypeInvalid, fmt.Sprintf("invalid key type '%s' for named set", ns.Key.KeyType), "Key", ns.Key.KeyType, identity.KEY_TYPE_NAMED_SET)
	}

	// Validate Name is not empty.
	if ns.Name == "" {
		return coreerr.New(ctx, coreerr.NsetNameRequired, "Name is required", "Name")
	}

	// Validate the ExpressionSpec.
	if err := ns.Spec.Validate(ctx); err != nil {
		return coreerr.New(ctx, coreerr.NsetSpecInvalid, fmt.Sprintf("named set '%s' Spec failed validation: %s", ns.Key.String(), err.Error()), "Spec")
	}

	// Validate TypeSpec if present.
	if ns.TypeSpec != nil {
		if err := ns.TypeSpec.Validate(ctx); err != nil {
			return coreerr.New(ctx, coreerr.NsetTypespecInvalid, fmt.Sprintf("named set '%s' TypeSpec failed validation: %s", ns.Key.String(), err.Error()), "TypeSpec")
		}
	}

	return nil
}

// ValidateWithParent validates the NamedSet and its key's parent relationship.
// Named set keys are root-level (nil parent).
func (ns *NamedSet) ValidateWithParent(ctx *coreerr.ValidationContext) error {
	if err := ns.Validate(ctx); err != nil {
		return err
	}
	if err := ns.Key.ValidateParentWithContext(ctx, nil); err != nil {
		return err
	}
	return nil
}
