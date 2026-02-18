package model_logic

import (
	"fmt"

	"github.com/go-playground/validator/v10"
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
	Name          string   `validate:"required,startswith=_"` // The definition name (e.g., _Max, _SetOfValues). Must start with underscore.
	Comment       string   // Optional human-readable description of this definition.
	Parameters    []string // The parameter names (e.g., ["x", "y"] for _Max(x, y)).
	Specification Logic    `validate:"required"`
}

// NewGlobalFunction creates a new GlobalFunction and validates it.
func NewGlobalFunction(name, comment string, parameters []string, specification Logic) (gf GlobalFunction, err error) {
	gf = GlobalFunction{
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
	// Validate the specification logic explicitly (Key.Validate() is not called by struct tag validation).
	if err := gf.Specification.Validate(); err != nil {
		return fmt.Errorf("specification: %w", err)
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

// ValidateWithParent validates the GlobalFunction and its specification logic's parent relationship.
// Global function specification keys are root-level (invariant keys with nil parent).
func (gf *GlobalFunction) ValidateWithParent() error {
	if err := gf.Validate(); err != nil {
		return err
	}
	if err := gf.Specification.ValidateWithParent(nil); err != nil {
		return fmt.Errorf("specification: %w", err)
	}
	return nil
}
