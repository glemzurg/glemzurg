package model_logic

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// GlobalFunction represents a global definition that can be referenced
// from other expressions throughout the model.
//
// Definitions can be:
//   - A set for membership checks: x \in _SetOfValues
//   - A function for data transformation: _Max(x, y)
//   - A common boolean predicate: _HasAChild(x)
//
// All global definitions must have a leading underscore to distinguish them
// from class-scoped actions.
type GlobalFunction struct {
	Key           identity.Key
	Name          string   `validate:"required,startswith=_"` // The definition name (e.g., _Max, _SetOfValues). Must start with underscore.
	Comment       string   // Optional human-readable description of this definition.
	Parameters    []string // The parameter names (e.g., ["x", "y"] for _Max(x, y)).
	Specification Logic    `validate:"required"`
}

// NewGlobalFunction creates a new GlobalFunction and validates it.
func NewGlobalFunction(key identity.Key, name, comment string, parameters []string, specification Logic) (gf GlobalFunction, err error) {
	gf = GlobalFunction{
		Key:           key,
		Name:          name,
		Comment:       comment,
		Parameters:    parameters,
		Specification: specification,
	}

	if err = gf.Validate(); err != nil {
		return GlobalFunction{}, err
	}

	return gf, nil
}

// Validate validates the GlobalFunction struct.
func (gf *GlobalFunction) Validate() error {
	// Validate the key.
	if err := gf.Key.Validate(); err != nil {
		return err
	}
	if gf.Key.KeyType != identity.KEY_TYPE_GLOBAL_FUNCTION {
		return errors.Errorf("Key: invalid key type '%s' for global function", gf.Key.KeyType)
	}

	// Validate the specification logic.
	if err := gf.Specification.Validate(); err != nil {
		return fmt.Errorf("specification: %w", err)
	}

	// Specification logic must use the global function's exact key.
	if gf.Specification.Key != gf.Key {
		return errors.Errorf("specification key '%s' does not match global function key '%s'", gf.Specification.Key.String(), gf.Key.String())
	}

	if err := _validate.Struct(gf); err != nil {
		// Wrap startswith error with a clearer message for the underscore rule.
		if _, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range err.(validator.ValidationErrors) {
				if fe.Field() == "Name" && fe.Tag() == "startswith" {
					return fmt.Errorf("global function name '%s' must start with underscore", gf.Name)
				}
			}
		}
		return err
	}
	return nil
}

// ValidateWithParent validates the GlobalFunction and its key's parent relationship.
// Global function keys are root-level (nil parent).
func (gf *GlobalFunction) ValidateWithParent() error {
	if err := gf.Validate(); err != nil {
		return err
	}
	if err := gf.Key.ValidateParent(nil); err != nil {
		return err
	}
	return nil
}
