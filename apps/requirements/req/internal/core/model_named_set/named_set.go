package model_named_set

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// NamedSet represents a reusable set definition at the model level.
// Named sets define well-known collections that can be referenced
// from behavioral logic (requires, guarantees, etc.) via NamedSetRef expressions.
type NamedSet struct {
	Key         identity.Key              // Unique key of type "nset".
	Name        string                    // Required: display name.
	Description string                    // Optional description.
	Spec        model_spec.ExpressionSpec // Notation + Specification + Expression for the set definition.
	TypeSpec    *model_spec.TypeSpec      // Optional precise type specification.
}

// NewNamedSet creates a new NamedSet.
func NewNamedSet(key identity.Key, name, description string, spec model_spec.ExpressionSpec, typeSpec *model_spec.TypeSpec) NamedSet {
	return NamedSet{
		Key:         key,
		Name:        name,
		Description: description,
		Spec:        spec,
		TypeSpec:    typeSpec,
	}
}

// Validate validates the NamedSet struct.
func (ns *NamedSet) Validate() error {
	// Validate the key.
	if err := ns.Key.Validate(); err != nil {
		return coreerr.New(coreerr.NsetKeyInvalid, fmt.Sprintf("NamedSet key failed validation: %s", err.Error()), "Key")
	}
	if ns.Key.KeyType != identity.KEY_TYPE_NAMED_SET {
		return coreerr.NewWithValues(coreerr.NsetKeyTypeInvalid, fmt.Sprintf("invalid key type '%s' for named set", ns.Key.KeyType), "Key", ns.Key.KeyType, identity.KEY_TYPE_NAMED_SET)
	}

	// Validate Name is not empty.
	if ns.Name == "" {
		return coreerr.New(coreerr.NsetNameRequired, "Name is required", "Name")
	}

	// Validate the ExpressionSpec.
	if err := ns.Spec.Validate(); err != nil {
		return coreerr.New(coreerr.NsetSpecInvalid, fmt.Sprintf("named set '%s' Spec failed validation: %s", ns.Key.String(), err.Error()), "Spec")
	}

	// Validate TypeSpec if present.
	if ns.TypeSpec != nil {
		if err := ns.TypeSpec.Validate(); err != nil {
			return coreerr.New(coreerr.NsetTypespecInvalid, fmt.Sprintf("named set '%s' TypeSpec failed validation: %s", ns.Key.String(), err.Error()), "TypeSpec")
		}
	}

	return nil
}

// ValidateWithParent validates the NamedSet and its key's parent relationship.
// Named set keys are root-level (nil parent).
func (ns *NamedSet) ValidateWithParent() error {
	if err := ns.Validate(); err != nil {
		return err
	}
	if err := ns.Key.ValidateParent(nil); err != nil {
		return errors.Wrapf(err, "named set '%s' parent", ns.Key.String())
	}
	return nil
}
