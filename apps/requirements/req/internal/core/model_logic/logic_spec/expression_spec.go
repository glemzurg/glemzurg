package logic_spec

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

// ExpressionParseFunc parses a specification string and returns the parsed expression
// and a normalized specification string. Returns (nil, "") if parsing fails — this is
// NOT an error condition, it simply means the specification could not be parsed.
type ExpressionParseFunc func(specification string) (logic_expression.Expression, string)

// ExpressionSpec groups a formal specification's notation, text, and parsed expression tree.
// An ExpressionSpec can be in one of three states:
//  1. Notation-only: Specification is empty, Expression is nil.
//  2. Unparsed: Specification is set, Expression is nil (parse not attempted or failed).
//  3. Fully parsed: Specification is set, Expression is non-nil. ParseOk() returns true.
type ExpressionSpec struct {
	Notation      string                      // Notation system (currently only TLA+).
	Specification string                      // Optional specification body text.
	Expression    logic_expression.Expression // Optional parsed expression tree (nil = not yet parsed).
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
	return spec, nil
}

// ParseOk returns true if the specification was successfully parsed into an expression tree.
func (s *ExpressionSpec) ParseOk() bool {
	return s.Expression != nil
}

// Validate validates the ExpressionSpec.
func (s *ExpressionSpec) Validate(ctx *coreerr.ValidationContext) error {
	if s.Notation == "" {
		return coreerr.NewWithValues(ctx, coreerr.ExprspecNotationRequired, "Notation is required", "Notation", "", "one of: tla_plus")
	}
	if s.Notation != "tla_plus" {
		return coreerr.NewWithValues(ctx, coreerr.ExprspecNotationInvalid, fmt.Sprintf("Notation '%s' is not valid", s.Notation), "Notation", s.Notation, "one of: tla_plus")
	}
	if s.Expression != nil {
		if err := s.Expression.Validate(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ExprspecExpressionInvalid, fmt.Sprintf("ExpressionSpec.Expression: %s", err.Error()), "Expression")
		}
	}
	return nil
}
