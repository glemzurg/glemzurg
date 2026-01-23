package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Guard is a constraint on an event in a state machine.
type Guard struct {
	Key     identity.Key
	Name    string // A simple unique name for a guard, for internal use.
	Details string // How the details of the guard are represented, what shows in the uml.
}

func NewGuard(key identity.Key, name, details string) (guard Guard, err error) {

	guard = Guard{
		Key:     key,
		Name:    name,
		Details: details,
	}

	if err = guard.Validate(); err != nil {
		return Guard{}, err
	}

	return guard, nil
}

// Validate validates the Guard struct.
func (g *Guard) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_GUARD {
				return errors.Errorf("invalid key type '%s' for guard", k.KeyType())
			}
			return nil
		})),
		validation.Field(&g.Name, validation.Required),
		validation.Field(&g.Details, validation.Required),
	)
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
	// Guard has no children with keys that need validation.
	return nil
}
