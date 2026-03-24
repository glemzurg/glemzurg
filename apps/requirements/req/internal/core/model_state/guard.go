package model_state

import (
	"fmt"

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

func NewGuard(key identity.Key, name string, logic model_logic.Logic) Guard {
	return Guard{
		Key:   key,
		Name:  name,
		Logic: logic,
	}
}

// Validate validates the Guard struct.
func (g *Guard) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := g.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.GuardKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if g.Key.KeyType != identity.KEY_TYPE_GUARD {
		return coreerr.NewWithValues(ctx, coreerr.GuardKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for guard", g.Key.KeyType), "Key", g.Key.KeyType, identity.KEY_TYPE_GUARD)
	}

	if g.Name == "" {
		return coreerr.New(ctx, coreerr.GuardNameRequired, "Name is required", "Name")
	}

	if err := g.Logic.Validate(ctx); err != nil {
		return coreerr.New(ctx, coreerr.GuardLogicInvalid, fmt.Sprintf("logic: %s", err.Error()), "Logic")
	}
	if g.Logic.Type != model_logic.LogicTypeAssessment {
		return coreerr.NewWithValues(ctx, coreerr.GuardLogicTypeInvalid, fmt.Sprintf("logic kind must be '%s', got '%s'", model_logic.LogicTypeAssessment, g.Logic.Type), "Logic.Type", g.Logic.Type, model_logic.LogicTypeAssessment)
	}

	return nil
}

// ValidateWithParent validates the Guard, its key's parent relationship, and all children.
// The parent must be a Class.
func (g *Guard) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := g.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := g.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Guard's logic must use the guard's exact key.
	if g.Logic.Key != g.Key {
		return coreerr.NewWithValues(ctx, coreerr.GuardLogicKeyMismatch, fmt.Sprintf("logic key '%s' does not match guard key '%s'", g.Logic.Key.String(), g.Key.String()), "Logic.Key", g.Logic.Key.String(), g.Key.String())
	}
	// Validate the logic's key parent relationship.
	if err := g.Logic.ValidateWithParent(ctx, parent); err != nil {
		return err
	}
	return nil
}
