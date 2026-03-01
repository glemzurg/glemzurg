package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/assert"
)

func TestNumberLiterals(t *testing.T) {
	tests := []struct {
		name    string
		node    ast.Node
		kind    object.NumberKind
		inspect string
	}{
		// Natural numbers (non-negative integers)
		{"zero", ast.NewIntLiteral(0), object.KindNatural, "0"},
		{"positive small", ast.NewIntLiteral(42), object.KindNatural, "42"},
		{"positive large", ast.NewIntLiteral(1000000), object.KindNatural, "1000000"},

		// Integers (negative)
		{"negative", ast.NewIntLiteral(-42), object.KindInteger, "-42"},
		{"negative large", ast.NewIntLiteral(-1000000), object.KindInteger, "-1000000"},

		// Rationals (fractions that don't simplify to integers)
		{"simple fraction", ast.NewFractionExpr(ast.NewIntLiteral(3), ast.NewIntLiteral(2)), object.KindRational, "3/2"},
		{"negative fraction", ast.NewFractionExpr(ast.NewIntLiteral(-3), ast.NewIntLiteral(4)), object.KindRational, "-3/4"},
		{"fraction simplifies", ast.NewFractionExpr(ast.NewIntLiteral(6), ast.NewIntLiteral(4)), object.KindRational, "3/2"},

		// Fractions that simplify to integers
		{"fraction to natural", ast.NewFractionExpr(ast.NewIntLiteral(4), ast.NewIntLiteral(2)), object.KindNatural, "2"},
		{"fraction to integer", ast.NewFractionExpr(ast.NewIntLiteral(-6), ast.NewIntLiteral(2)), object.KindInteger, "-3"},
		{"fraction to zero", ast.NewFractionExpr(ast.NewIntLiteral(0), ast.NewIntLiteral(5)), object.KindNatural, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindings := NewBindings()
			result := Eval(tt.node, bindings)

			assert.False(t, result.IsError(), "unexpected error: %v", result.Error)
			num := result.Value.(*object.Number)
			assert.Equal(t, tt.kind, num.Kind(), "wrong kind")
			assert.Equal(t, tt.inspect, num.Inspect(), "wrong inspect value")
		})
	}
}
