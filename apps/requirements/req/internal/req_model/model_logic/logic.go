package model_logic

import "github.com/go-playground/validator/v10"

// NotationTLAPlus is the only supported notation for logic specifications.
const NotationTLAPlus = "TLA+"

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// Logic represents a formal logic specification attached to a model element.
type Logic struct {
	Id            string `validate:"required"`
	Description   string `validate:"required"`
	Notation      string `validate:"required,oneof=TLA+"`
	Specification string // Optional logic specification body.
}

// NewLogic creates a new Logic and validates it.
func NewLogic(id, description, notation, specification string) (logic Logic, err error) {
	logic = Logic{
		Id:            id,
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
	return _validate.Struct(l)
}
