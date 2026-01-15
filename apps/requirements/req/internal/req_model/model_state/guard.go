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

	err = validation.ValidateStruct(&guard,
		validation.Field(&guard.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_GUARD {
				return errors.Errorf("invalid key type '%s' for guard", k.KeyType())
			}
			return nil
		})),
		validation.Field(&guard.Name, validation.Required),
		validation.Field(&guard.Details, validation.Required),
	)
	if err != nil {
		return Guard{}, errors.WithStack(err)
	}

	return guard, nil
}
