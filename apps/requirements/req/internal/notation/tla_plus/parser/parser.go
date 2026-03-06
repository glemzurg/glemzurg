// Package parser provides a minimal TLA+ expression parser.
// It uses a Pigeon-generated PEG parser to produce ast.Expression nodes.
package parser

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

// ParseExpression parses a TLA+ expression string and returns an ast.Expression.
// Returns an error if the input is not a valid TLA+ expression.
func ParseExpression(input string) (ast.Expression, error) {
	result, err := Parse("", []byte(input))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	expr, ok := result.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("unexpected parse result type: %T", result)
	}

	return expr, nil
}

// ParseExpressionList parses multiple TLA+ expression strings.
// Returns the list of parsed expressions, or an error if any expression is invalid.
func ParseExpressionList(inputs []string) ([]ast.Expression, error) {
	expressions := make([]ast.Expression, len(inputs))

	for i, input := range inputs {
		expr, err := ParseExpression(input)
		if err != nil {
			return nil, fmt.Errorf("expression %d: %w", i, err)
		}
		expressions[i] = expr
	}

	return expressions, nil
}

// MustParseExpression is like ParseExpression but panics on error.
// Useful for tests and static initialization.
func MustParseExpression(input string) ast.Expression {
	expr, err := ParseExpression(input)
	if err != nil {
		panic(err)
	}
	return expr
}
