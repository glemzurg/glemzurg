package model_logic

import (
	"github.com/go-playground/validator/v10"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// NotationTLAPlus is the only supported notation for logic specifications.
const NotationTLAPlus = "tla_plus"

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// Logic represents a formal logic specification attached to a model element.
type Logic struct {
	Key           identity.Key // The key is unique in the whole model, and built on the key of the containing object.
	Description   string       `validate:"required"`
	Notation      string       `validate:"required,oneof=tla_plus"`
	Specification string       // Optional logic specification body.
}

// NewLogic creates a new Logic and validates it.
func NewLogic(key identity.Key, description, notation, specification string) (logic Logic, err error) {
	logic = Logic{
		Key:           key,
		Description:   description,
		Notation:      notation,
		Specification: specification,
	}

	if err = logic.Validate(); err != nil {
		return Logic{}, err
	}

	return logic, nil
}

// Validate validates the Logic struct.
func (l *Logic) Validate() error {
	if err := l.Key.Validate(); err != nil {
		return err
	}
	return _validate.Struct(l)
}

// ValidateWithParent validates the Logic and its key's parent relationship.
func (l *Logic) ValidateWithParent(parent *identity.Key) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if err := l.Key.ValidateParent(parent); err != nil {
		return err
	}
	return nil
}
