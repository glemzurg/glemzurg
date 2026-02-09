package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestRecordsControlSuite(t *testing.T) {
	suite.Run(t, new(RecordsControlSuite))
}

type RecordsControlSuite struct {
	suite.Suite
}

// === IF-THEN-ELSE ===

func (s *RecordsControlSuite) TestIfElse_TrueCondition() {
	// IF TRUE THEN 1 ELSE 2 = 1
	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewIntLiteral(1),
		Else:      ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *RecordsControlSuite) TestIfElse_FalseCondition() {
	// IF FALSE THEN 1 ELSE 2 = 2
	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: false},
		Then:      ast.NewIntLiteral(1),
		Else:      ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *RecordsControlSuite) TestIfElse_NestedCondition() {
	// IF (3 < 5) THEN "yes" ELSE "no" = "yes"
	node := &ast.ExpressionIfElse{
		Condition: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "<",
			Right:    ast.NewIntLiteral(5),
		},
		Then: &ast.StringLiteral{Value: "yes"},
		Else: &ast.StringLiteral{Value: "no"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("yes", str.Value())
}

func (s *RecordsControlSuite) TestIfElse_NestedIfElse() {
	// IF FALSE THEN 1 ELSE IF TRUE THEN 2 ELSE 3 = 2
	inner := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewIntLiteral(2),
		Else:      ast.NewIntLiteral(3),
	}
	outer := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: false},
		Then:      ast.NewIntLiteral(1),
		Else:      inner,
	}
	bindings := NewBindings()

	result := Eval(outer, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

// === Record EXCEPT ===

func (s *RecordsControlSuite) TestRecordExcept_SimpleUpdate() {
	// [x EXCEPT !.value = 20] where x = [value ↦ 10]
	record := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(10),
	})
	bindings := NewBindings()
	bindings.Set("x", record, NamespaceGlobal)

	node := &ast.RecordAltered{
		Identifier: &ast.Identifier{Value: "x"},
		Alterations: []*ast.FieldAlteration{
			{
				Field:      &ast.FieldIdentifier{Identifier: nil, Member: "value"},
				Expression: ast.NewIntLiteral(20),
			},
		},
	}

	result := Eval(node, bindings)

	s.False(result.IsError())
	newRecord := result.Value.(*object.Record)
	value := newRecord.Get("value").(*object.Number)
	s.Equal("20", value.Inspect())

	// Original record should be unchanged (immutability)
	origValue := record.Get("value").(*object.Number)
	s.Equal("10", origValue.Inspect())
}

func (s *RecordsControlSuite) TestRecordExcept_MultipleUpdates() {
	// [x EXCEPT !.a = 100, !.b = 200] where x = [a ↦ 1, b ↦ 2]
	record := object.NewRecordFromFields(map[string]object.Object{
		"a": object.NewNatural(1),
		"b": object.NewNatural(2),
	})
	bindings := NewBindings()
	bindings.Set("x", record, NamespaceGlobal)

	node := &ast.RecordAltered{
		Identifier: &ast.Identifier{Value: "x"},
		Alterations: []*ast.FieldAlteration{
			{
				Field:      &ast.FieldIdentifier{Identifier: nil, Member: "a"},
				Expression: ast.NewIntLiteral(100),
			},
			{
				Field:      &ast.FieldIdentifier{Identifier: nil, Member: "b"},
				Expression: ast.NewIntLiteral(200),
			},
		},
	}

	result := Eval(node, bindings)

	s.False(result.IsError())
	newRecord := result.Value.(*object.Record)
	s.Equal("100", newRecord.Get("a").Inspect())
	s.Equal("200", newRecord.Get("b").Inspect())
}

// Note: Testing EXCEPT with @ (ExistingValue) would require ExistingValue to implement
// the Real interface, which it doesn't in this type-safe AST design. The @ reference
// works via the evaluator setting existingValue in bindings, but can only be used
// in Expression contexts (not Real).

// === Assignment (Priming) ===

func (s *RecordsControlSuite) TestAssignment_Simple() {
	// x' = 42
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value:  ast.NewIntLiteral(42),
	}
	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(10), NamespaceGlobal)

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.HasPrimedBindings())
	s.Equal("42", result.PrimedBindings["x"].Inspect())

	// The bindings should also reflect the primed value
	s.True(bindings.IsPrimed("x"))
}

// Note: Testing x' = x + 5 would require Identifier to implement Real interface,
// which it doesn't in this type-safe AST. Arithmetic expressions can only use
// literal numbers or other Real-typed nodes.

func (s *RecordsControlSuite) TestAssignment_NewVariable() {
	// y' = 100 (y doesn't exist yet)
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "y"},
		Value:  ast.NewIntLiteral(100),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.Equal("100", result.PrimedBindings["y"].Inspect())
}

// === Combined Record and Control Flow ===

func (s *RecordsControlSuite) TestIfElse_WithRecords() {
	// IF TRUE THEN [val ↦ 1] ELSE [val ↦ 2]
	bindings := NewBindings()

	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true}, // Use literal since Identifier doesn't implement Logic
		Then: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "val"}, Expression: ast.NewIntLiteral(1)},
			},
		},
		Else: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "val"}, Expression: ast.NewIntLiteral(2)},
			},
		},
	}

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal("1", record.Get("val").Inspect())
}

// === String Operations ===

func (s *RecordsControlSuite) TestStringIndex_Simple() {
	// "hello"[1] = "h"
	node := &ast.StringIndex{
		Str:   &ast.StringLiteral{Value: "hello"},
		Index: ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("h", str.Value())
}

func (s *RecordsControlSuite) TestStringIndex_Middle() {
	// "hello"[3] = "l"
	node := &ast.StringIndex{
		Str:   &ast.StringLiteral{Value: "hello"},
		Index: ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("l", str.Value())
}

func (s *RecordsControlSuite) TestStringIndex_OutOfBounds() {
	// "hi"[5] = error
	node := &ast.StringIndex{
		Str:   &ast.StringLiteral{Value: "hi"},
		Index: ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "out of bounds")
}

func (s *RecordsControlSuite) TestStringConcat_Simple() {
	// "hello" ∘ " " ∘ "world" = "hello world"
	node := &ast.StringInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.StringLiteral{Value: "hello"},
			&ast.StringLiteral{Value: " "},
			&ast.StringLiteral{Value: "world"},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello world", str.Value())
}
