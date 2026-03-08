package model_state

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Guard is a constraint on an event in a state machine.
type Guard struct {
	Key   identity.Key
	Name  string            // A simple unique name for a guard, for internal use.
	Logic model_logic.Logic // The formal logic specification for this guard condition.
}

func NewGuard(key identity.Key, name string, logic model_logic.Logic) (guard Guard, err error) {
	guard = Guard{
		Key:   key,
		Name:  name,
		Logic: logic,
	}

	if err = guard.Validate(); err != nil {
		return Guard{}, err
	}

	return guard, nil
}

// Validate validates the Guard struct.
func (g *Guard) Validate() error {
	// Validate the key.
	if err := g.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if g.Key.KeyType != identity.KEY_TYPE_GUARD {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardKeyTypeInvalid,
			Message: fmt.Sprintf("Key: invalid key type '%s' for guard", g.Key.KeyType),
			Field:   "Key",
			Got:     g.Key.KeyType,
			Want:    identity.KEY_TYPE_GUARD,
		}
	}

	if g.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}

	if err := g.Logic.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardLogicInvalid,
			Message: fmt.Sprintf("logic: %s", err.Error()),
			Field:   "Logic",
		}
	}
	if g.Logic.Type != model_logic.LogicTypeAssessment {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardLogicTypeInvalid,
			Message: fmt.Sprintf("logic kind must be '%s', got '%s'", model_logic.LogicTypeAssessment, g.Logic.Type),
			Field:   "Logic.Type",
			Got:     g.Logic.Type,
			Want:    model_logic.LogicTypeAssessment,
		}
	}

	return nil
}

// ValidateWithParent validates the Guard, its key's parent relationship, and all children.
// The parent must be a Class.
func (g *Guard) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := g.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := g.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Guard's logic must use the guard's exact key.
	if g.Logic.Key != g.Key {
		return &coreerr.ValidationError{
			Code:    coreerr.GuardLogicKeyMismatch,
			Message: fmt.Sprintf("logic key '%s' does not match guard key '%s'", g.Logic.Key.String(), g.Key.String()),
			Field:   "Logic.Key",
			Got:     g.Logic.Key.String(),
			Want:    g.Key.String(),
		}
	}
	// Validate the logic's key parent relationship.
	if err := g.Logic.ValidateWithParent(parent); err != nil {
		return errors.Wrap(err, "logic")
	}
	return nil
}
