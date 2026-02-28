package model_spec

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression_type"
)

// TypeSpec groups a type declaration's notation, text, and parsed type tree.
// These three fields always travel together — you never have one without the others being meaningful.
type TypeSpec struct {
	Notation       string                                   `validate:"required,oneof=tla_plus"` // Notation system (currently only TLA+).
	Specification  string                                   // Optional specification body text.
	ExpressionType model_expression_type.ExpressionType     // Optional parsed expression type tree (nil = not yet parsed).
}

// NewTypeSpec creates a new TypeSpec and validates it.
func NewTypeSpec(notation, specification string, expressionType model_expression_type.ExpressionType) (spec TypeSpec, err error) {
	spec = TypeSpec{
		Notation:       notation,
		Specification:  specification,
		ExpressionType: expressionType,
	}
	if err = spec.Validate(); err != nil {
		return TypeSpec{}, err
	}
	return spec, nil
}

// Validate validates the TypeSpec.
func (s *TypeSpec) Validate() error {
	if err := _validate.Struct(s); err != nil {
		return err
	}
	if s.ExpressionType != nil {
		if err := s.ExpressionType.Validate(); err != nil {
			return errors.Wrap(err, "TypeSpec.ExpressionType")
		}
	}
	return nil
}
