package model_spec

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
)

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// ExpressionParseFunc parses a specification string and returns the parsed expression
// and a normalized specification string. Returns (nil, "") if parsing fails — this is
// NOT an error condition, it simply means the specification could not be parsed.
type ExpressionParseFunc func(specification string) (model_expression.Expression, string)

// ExpressionSpec groups a formal specification's notation, text, and parsed expression tree.
// An ExpressionSpec can be in one of three states:
//  1. Notation-only: Specification is empty, Expression is nil.
//  2. Unparsed: Specification is set, Expression is nil (parse not attempted or failed).
//  3. Fully parsed: Specification is set, Expression is non-nil. ParseOk() returns true.
type ExpressionSpec struct {
	Notation      string                       `validate:"required,oneof=tla_plus"` // Notation system (currently only TLA+).
	Specification string                       // Optional specification body text.
	Expression    model_expression.Expression  // Optional parsed expression tree (nil = not yet parsed).
}

// NewExpressionSpec creates a new ExpressionSpec and validates it.
// If parseFunc is non-nil and specification is non-empty, the parse function is called
// to attempt parsing the specification into an expression tree. If parsing succeeds,
// the normalized specification replaces the original.
func NewExpressionSpec(notation, specification string, parseFunc ExpressionParseFunc) (spec ExpressionSpec, err error) {
	spec = ExpressionSpec{
		Notation:      notation,
		Specification: specification,
	}
	// Attempt parsing if we have a specification and a parse function.
	if specification != "" && parseFunc != nil {
		expr, normalized := parseFunc(specification)
		if expr != nil {
			spec.Expression = expr
			if normalized != "" {
				spec.Specification = normalized
			}
		}
	}
	if err = spec.Validate(); err != nil {
		return ExpressionSpec{}, err
	}
	return spec, nil
}

// ParseOk returns true if the specification was successfully parsed into an expression tree.
func (s *ExpressionSpec) ParseOk() bool {
	return s.Expression != nil
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
