package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
)

// Guard is a constraint on an event in a state machine.
type Guard struct {
	Key   identity.Key
	Name  string             // A simple unique name for a guard, for internal use.
	Logic model_logic.Logic  // The formal logic specification for this guard condition.
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
	if err := validation.ValidateStruct(g,
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
	); err != nil {
		return err
	}

	if err := g.Logic.Validate(); err != nil {
		return errors.Wrap(err, "logic")
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
	// Guard has no children with keys that need validation.
	return nil
}
