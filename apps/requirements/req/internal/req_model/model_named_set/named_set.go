package model_named_set

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
)

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// NamedSet represents a reusable set definition at the model level.
// Named sets define well-known collections that can be referenced
// from behavioral logic (requires, guarantees, etc.) via NamedSetRef expressions.
type NamedSet struct {
	Key         identity.Key          // Unique key of type "nset".
	Name        string                `validate:"required"`
	Description string                // Optional description.
	Spec        model_spec.ExpressionSpec // Notation + Specification + Expression for the set definition.
	TypeSpec    *model_spec.TypeSpec  // Optional precise type specification.
}

// NewNamedSet creates a new NamedSet and validates it.
func NewNamedSet(key identity.Key, name, description string, spec model_spec.ExpressionSpec, typeSpec *model_spec.TypeSpec) (ns NamedSet, err error) {
	ns = NamedSet{
		Key:         key,
		Name:        name,
		Description: description,
		Spec:        spec,
		TypeSpec:    typeSpec,
	}

	if err = ns.Validate(); err != nil {
		return NamedSet{}, err
	}

	return ns, nil
}

// Validate validates the NamedSet struct.
func (ns *NamedSet) Validate() error {
	// Validate the key.
	if err := ns.Key.Validate(); err != nil {
		return err
	}
	if ns.Key.KeyType != identity.KEY_TYPE_NAMED_SET {
		return errors.Errorf("Key: invalid key type '%s' for named set", ns.Key.KeyType)
	}

	// Validate struct tags.
	if err := _validate.Struct(ns); err != nil {
		return err
	}

	// Validate the ExpressionSpec.
	if err := ns.Spec.Validate(); err != nil {
		return errors.Wrapf(err, "named set '%s' spec", ns.Key.String())
	}

	// Validate TypeSpec if present.
	if ns.TypeSpec != nil {
		if err := ns.TypeSpec.Validate(); err != nil {
			return errors.Wrapf(err, "named set '%s' type spec", ns.Key.String())
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
		return err
	}
	return nil
}
