package model_spec

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression"
)

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// ExpressionSpec groups a formal specification's notation, text, and parsed expression tree.
// These three fields always travel together — you never have one without the others being meaningful.
type ExpressionSpec struct {
	Notation      string                       `validate:"required,oneof=tla_plus"` // Notation system (currently only TLA+).
	Specification string                       // Optional specification body text.
	Expression    model_expression.Expression  // Optional parsed expression tree (nil = not yet parsed).
}

// NewExpressionSpec creates a new ExpressionSpec and validates it.
func NewExpressionSpec(notation, specification string, expression model_expression.Expression) (spec ExpressionSpec, err error) {
	spec = ExpressionSpec{
		Notation:      notation,
		Specification: specification,
		Expression:    expression,
	}
	if err = spec.Validate(); err != nil {
		return ExpressionSpec{}, err
	}
	return spec, nil
}

// Validate validates the ExpressionSpec.
func (s *ExpressionSpec) Validate() error {
	if err := _validate.Struct(s); err != nil {
		return err
	}
	if s.Expression != nil {
		if err := s.Expression.Validate(); err != nil {
			return errors.Wrap(err, "ExpressionSpec.Expression")
		}
	}
	return nil
}
