package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Action is what happens in a transition between states.
type Action struct {
	Key            identity.Key
	Name           string
	Details        string
	Requires       []string // Human-readable preconditions to enter this action.
	Guarantees     []string // Human-readable postconditions of this action.
	TlaRequires    []string // TLA+ expressions for preconditions (must not contain primed variables).
	TlaGuarantees  []string // TLA+ primed assignments only (e.g., self.field' = expr).
	TlaSafetyRules []string // TLA+ boolean assertions that must reference primed variables.
	// Children
	Parameters []Parameter // Typed parameters for this action.
}

func NewAction(key identity.Key, name, details string, requires, guarantees, tlaRequires, tlaGuarantees, tlaSafetyRules []string, parameters []Parameter) (action Action, err error) {

	action = Action{
		Key:            key,
		Name:           name,
		Details:        details,
		Requires:       requires,
		Guarantees:     guarantees,
		TlaRequires:    tlaRequires,
		TlaGuarantees:  tlaGuarantees,
		TlaSafetyRules: tlaSafetyRules,
		Parameters:     parameters,
	}

	if err = action.Validate(); err != nil {
		return Action{}, err
	}

	return action, nil
}

// Validate validates the Action struct.
func (a *Action) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTION {
				return errors.Errorf("invalid key type '%s' for action", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.Name, validation.Required),
	)
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
	// Validate all children.
	for i := range a.Parameters {
		if err := a.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
