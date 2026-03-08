package model_spec

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression_type"
)

// TypeParseFunc parses a type specification string and returns the parsed expression type
// and a normalized specification string. Returns (nil, "") if parsing fails — this is
// NOT an error condition, it simply means the specification could not be parsed.
type TypeParseFunc func(specification string) (model_expression_type.ExpressionType, string)

// TypeSpec groups a type declaration's notation, text, and parsed type tree.
// A TypeSpec can be in one of three states:
//  1. Notation-only: Specification is empty, ExpressionType is nil.
//  2. Unparsed: Specification is set, ExpressionType is nil (parse not attempted or failed).
//  3. Fully parsed: Specification is set, ExpressionType is non-nil. ParseOk() returns true.
type TypeSpec struct {
	Notation       string                               // Notation system (currently only TLA+).
	Specification  string                               // Optional specification body text.
	ExpressionType model_expression_type.ExpressionType // Optional parsed expression type tree (nil = not yet parsed).
}

// NewTypeSpec creates a new TypeSpec and validates it.
// If parseFunc is non-nil and specification is non-empty, the parse function is called
// to attempt parsing the specification into an expression type tree. If parsing succeeds,
// the normalized specification replaces the original.
func NewTypeSpec(notation, specification string, parseFunc TypeParseFunc) (spec TypeSpec, err error) {
	spec = TypeSpec{
		Notation:      notation,
		Specification: specification,
	}
	// Attempt parsing if we have a specification and a parse function.
	if specification != "" && parseFunc != nil {
		exprType, normalized := parseFunc(specification)
		if exprType != nil {
			spec.ExpressionType = exprType
			if normalized != "" {
				spec.Specification = normalized
			}
		}
	}
	if err = spec.Validate(); err != nil {
		return TypeSpec{}, err
	}
	return spec, nil
}

// ParseOk returns true if the specification was successfully parsed into an expression type tree.
func (s *TypeSpec) ParseOk() bool {
	return s.ExpressionType != nil
}

// Validate validates the TypeSpec.
func (s *TypeSpec) Validate() error {
	if s.Notation == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.TypespecNotationRequired,
			Message: "Notation is required",
			Field:   "Notation",
			Want:    "one of: tla_plus",
		}
	}
	if s.Notation != "tla_plus" {
		return &coreerr.ValidationError{
			Code:    coreerr.TypespecNotationInvalid,
			Message: fmt.Sprintf("Notation '%s' is not valid", s.Notation),
			Field:   "Notation",
			Got:     s.Notation,
			Want:    "one of: tla_plus",
		}
	}
	if s.ExpressionType != nil {
		if err := s.ExpressionType.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.TypespecExprtypeInvalid,
				Message: fmt.Sprintf("TypeSpec.ExpressionType: %s", err.Error()),
				Field:   "ExpressionType",
			}
		}
	}
	return nil
}
