package model_state

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
)

// Action is what happens in a transition between states.
type Action struct {
	Key         identity.Key
	Name        string `validate:"required"`
	Details     string
	Requires    []model_logic.Logic // Preconditions to enter this action (must not contain primed variables).
	Guarantees  []model_logic.Logic // Postconditions of this action (primed assignments only, e.g., self.field' = expr).
	SafetyRules []model_logic.Logic // Boolean assertions that must reference primed variables.
	// Children
	Parameters []Parameter // Typed parameters for this action.
}

func NewAction(key identity.Key, name, details string, requires, guarantees, safetyRules []model_logic.Logic, parameters []Parameter) (action Action, err error) {

	action = Action{
		Key:         key,
		Name:        name,
		Details:     details,
		Requires:    requires,
		Guarantees:  guarantees,
		SafetyRules: safetyRules,
		Parameters:  parameters,
	}

	if err = action.Validate(); err != nil {
		return Action{}, err
	}

	return action, nil
}

// Validate validates the Action struct.
func (a *Action) Validate() error {
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return err
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTION {
		return errors.Errorf("Key: invalid key type '%s' for action", a.Key.KeyType)
	}

	// Validate struct tags (Name required).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	for i, req := range a.Requires {
		if err := req.Validate(); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
	}
	for i, guar := range a.Guarantees {
		if err := guar.Validate(); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
	}
	for i, rule := range a.SafetyRules {
		if err := rule.Validate(); err != nil {
			return errors.Wrapf(err, "safety rule %d", i)
		}
	}

	return nil
}

// ValidateWithParent validates the Action, its key's parent relationship, and all children.
// The parent must be a Class.
func (a *Action) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate logic children with action as parent.
	for i := range a.Requires {
		if err := a.Requires[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
	}
	for i := range a.Guarantees {
		if err := a.Guarantees[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
	}
	for i := range a.SafetyRules {
		if err := a.SafetyRules[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "safety rule %d", i)
		}
	}
	// Validate all children.
	for i := range a.Parameters {
		if err := a.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
