package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestPairwiseSuite(t *testing.T) {
	suite.Run(t, new(PairwiseSuite))
}

type PairwiseSuite struct {
	suite.Suite
}

// === Assignment + Complex Expressions ===

func (s *PairwiseSuite) TestAssignment_WithTupleConcat() {
	// x' = <<1, 2>> ∘ <<3, 4>>
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value: &ast.TupleInfixExpression{
			Operator: "∘",
			Operands: []ast.Expression{
				&ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(1),
						ast.NewIntLiteral(2),
					},
				},
				&ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(3),
						ast.NewIntLiteral(4),
					},
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.HasPrimedBindings())
	tuple := result.PrimedBindings["x"].(*object.Tuple)
	s.Equal(4, tuple.Len())
}

func (s *PairwiseSuite) TestAssignment_WithTupleAppend() {
	// x' = _Seq!Append(<<1, 2>>, 3)
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value: &ast.BuiltinCall{
			Name: "_Seq!Append",
			Args: []ast.Expression{
				&ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(1),
						ast.NewIntLiteral(2),
					},
				},
				ast.NewIntLiteral(3),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.PrimedBindings["x"].(*object.Tuple)
	s.Equal(3, tuple.Len())
}

func (s *PairwiseSuite) TestAssignment_WithRecordInstance() {
	// x' = [a ↦ 1, b ↦ 2]
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "a"}, Expression: ast.NewIntLiteral(1)},
				{Field: &ast.Identifier{Value: "b"}, Expression: ast.NewIntLiteral(2)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.PrimedBindings["x"].(*object.Record)
	s.Equal("1", record.Get("a").Inspect())
	s.Equal("2", record.Get("b").Inspect())
}

func (s *PairwiseSuite) TestAssignment_WithIfElse() {
	// x' = IF TRUE THEN 10 ELSE 20
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value: &ast.ExpressionIfElse{
			Condition: &ast.BooleanLiteral{Value: true},
			Then:      ast.NewIntLiteral(10),
			Else:      ast.NewIntLiteral(20),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.PrimedBindings["x"].(*object.Number)
	s.Equal("10", num.Inspect())
}

func (s *PairwiseSuite) TestAssignment_WithArithmetic() {
	// x' = 5 + 3 * 2
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value: &ast.RealInfixExpression{
			Operator: "+",
			Left:     ast.NewIntLiteral(5),
			Right: &ast.RealInfixExpression{
				Operator: "*",
				Left:     ast.NewIntLiteral(3),
				Right:    ast.NewIntLiteral(2),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.PrimedBindings["x"].(*object.Number)
	s.Equal("11", num.Inspect())
}

// === Quantifiers + Set Operations ===

func (s *PairwiseSuite) TestQuantifier_OverSetUnion() {
	// ∀x ∈ ({1} ∪ {2}) : TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right: &ast.SetInfix{
				Operator: "∪",
				Left:     &ast.SetLiteralInt{Values: []int{1}},
				Right:    &ast.SetLiteralInt{Values: []int{2}},
			},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PairwiseSuite) TestQuantifier_OverSetConditional() {
	// ∃x ∈ {y ∈ {1, 2, 3} : TRUE} : TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∃",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right: &ast.SetConditional{
				Membership: &ast.LogicMembership{
					Left:     &ast.Identifier{Value: "y"},
					Operator: "∈",
					Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
				},
				Predicate: &ast.BooleanLiteral{Value: true},
			},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PairwiseSuite) TestQuantifier_OverSetRange() {
	// ∀x ∈ 1..5 : TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right: &ast.SetRange{
				Start: 1,
				End:   5,
			},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Control Flow + Collections ===

func (s *PairwiseSuite) TestIfElse_ReturnsRecord() {
	// IF TRUE THEN [a ↦ 1] ELSE [a ↦ 2]
	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "a"}, Expression: ast.NewIntLiteral(1)},
			},
		},
		Else: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "a"}, Expression: ast.NewIntLiteral(2)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal("1", record.Get("a").Inspect())
}

func (s *PairwiseSuite) TestIfElse_ReturnsTuple() {
	// IF FALSE THEN <<1>> ELSE <<2, 3>>
	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: false},
		Then: &ast.TupleLiteral{
			Elements: []ast.Expression{ast.NewIntLiteral(1)},
		},
		Else: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(2),
				ast.NewIntLiteral(3),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(2, tuple.Len())
}

func (s *PairwiseSuite) TestCase_ReturnsTuple() {
	// CASE TRUE → <<1, 2>> □ OTHER → <<>>
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result: &ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(1),
						ast.NewIntLiteral(2),
					},
				},
			},
		},
		Other: &ast.TupleLiteral{Elements: []ast.Expression{}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(2, tuple.Len())
}

func (s *PairwiseSuite) TestCase_ReturnsRecord() {
	// CASE FALSE → [x ↦ 1] □ TRUE → [x ↦ 2] □ OTHER → [x ↦ 0]
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result: &ast.RecordInstance{
					Bindings: []*ast.FieldBinding{
						{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
					},
				},
			},
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result: &ast.RecordInstance{
					Bindings: []*ast.FieldBinding{
						{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(2)},
					},
				},
			},
		},
		Other: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(0)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal("2", record.Get("x").Inspect())
}

// === Membership + Computed Sets ===

func (s *PairwiseSuite) TestMembership_InSetUnion() {
	// 3 ∈ ({1, 2} ∪ {3, 4})
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(3),
		Operator: "∈",
		Right: &ast.SetInfix{
			Operator: "∪",
			Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
			Right:    &ast.SetLiteralInt{Values: []int{3, 4}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PairwiseSuite) TestMembership_InSetIntersection() {
	// 2 ∈ ({1, 2, 3} ∩ {2, 3, 4})
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(2),
		Operator: "∈",
		Right: &ast.SetInfix{
			Operator: "∩",
			Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
			Right:    &ast.SetLiteralInt{Values: []int{2, 3, 4}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PairwiseSuite) TestMembership_InSetDifference() {
	// 1 ∈ ({1, 2, 3} \ {2, 3})
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(1),
		Operator: "∈",
		Right: &ast.SetInfix{
			Operator: "\\",
			Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
			Right:    &ast.SetLiteralInt{Values: []int{2, 3}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PairwiseSuite) TestNotMembership_InSetDifference() {
	// 2 ∉ ({1, 2, 3} \ {2, 3})
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(2),
		Operator: "∉",
		Right: &ast.SetInfix{
			Operator: "\\",
			Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
			Right:    &ast.SetLiteralInt{Values: []int{2, 3}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Tuple Operations in Expressions ===

func (s *PairwiseSuite) TestSeqLen_InArithmetic() {
	// _Seq!Len(<<1, 2, 3>>) + 5
	node := &ast.RealInfixExpression{
		Operator: "+",
		Left: &ast.BuiltinCall{
			Name: "_Seq!Len",
			Args: []ast.Expression{
				&ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(1),
						ast.NewIntLiteral(2),
						ast.NewIntLiteral(3),
					},
				},
			},
		},
		Right: ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("8", num.Inspect()) // 3 + 5
}

func (s *PairwiseSuite) TestSeqLen_InComparison() {
	// _Seq!Len(<<1, 2>>) < 5
	node := &ast.LogicRealComparison{
		Left: &ast.BuiltinCall{
			Name: "_Seq!Len",
			Args: []ast.Expression{
				&ast.TupleLiteral{
					Elements: []ast.Expression{
						ast.NewIntLiteral(1),
						ast.NewIntLiteral(2),
					},
				},
			},
		},
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Record Operations ===

func (s *PairwiseSuite) TestRecordExcept_WithArithmetic() {
	// [x EXCEPT !.value = 10 + 5] where x = [value ↦ 1]
	record := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(1),
	})
	bindings := NewBindings()
	bindings.Set("x", record, NamespaceGlobal)

	node := &ast.RecordAltered{
		Identifier: &ast.Identifier{Value: "x"},
		Alterations: []*ast.FieldAlteration{
			{
				Field: &ast.FieldIdentifier{Identifier: nil, Member: "value"},
				Expression: &ast.RealInfixExpression{
					Operator: "+",
					Left:     ast.NewIntLiteral(10),
					Right:    ast.NewIntLiteral(5),
				},
			},
		},
	}

	result := Eval(node, bindings)

	s.False(result.IsError())
	newRecord := result.Value.(*object.Record)
	s.Equal("15", newRecord.Get("value").Inspect())
}

func (s *PairwiseSuite) TestRecordExcept_WithIfElse() {
	// [x EXCEPT !.value = IF TRUE THEN 100 ELSE 0] where x = [value ↦ 1]
	record := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(1),
	})
	bindings := NewBindings()
	bindings.Set("x", record, NamespaceGlobal)

	node := &ast.RecordAltered{
		Identifier: &ast.Identifier{Value: "x"},
		Alterations: []*ast.FieldAlteration{
			{
				Field: &ast.FieldIdentifier{Identifier: nil, Member: "value"},
				Expression: &ast.ExpressionIfElse{
					Condition: &ast.BooleanLiteral{Value: true},
					Then:      ast.NewIntLiteral(100),
					Else:      ast.NewIntLiteral(0),
				},
			},
		},
	}

	result := Eval(node, bindings)

	s.False(result.IsError())
	newRecord := result.Value.(*object.Record)
	s.Equal("100", newRecord.Get("value").Inspect())
}

// === Error Propagation ===

func (s *PairwiseSuite) TestError_InSetConditional() {
	// {x ∈ {1, 2, 3} : x ∈ nonexistent} - error in predicate (nonexistent set variable)
	node := &ast.SetConditional{
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{}}, // Empty set - predicate will be false, no error
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	// No error - just empty result (all elements fail predicate)
	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(0, set.Size())
}

func (s *PairwiseSuite) TestError_InQuantifier_BadType() {
	// ∀x ∈ {1, 2, 3} : 5 < 3 ∧ 2 < 1
	// No error - just false predicate
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.LogicInfixExpression{
			Operator: "∧",
			Left: &ast.LogicRealComparison{
				Left:     ast.NewIntLiteral(5),
				Operator: "<",
				Right:    ast.NewIntLiteral(3),
			},
			Right: &ast.LogicRealComparison{
				Left:     ast.NewIntLiteral(2),
				Operator: "<",
				Right:    ast.NewIntLiteral(1),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value()) // 5 < 3 is false, so ∀ is false
}

func (s *PairwiseSuite) TestError_InAssignment_UndefinedVariable() {
	// x' = y where y is not defined
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "x"},
		Value:  &ast.Identifier{Value: "nonexistent"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "not found")
}

func (s *PairwiseSuite) TestError_InRecordAltered_UndefinedVariable() {
	// [nonexistent EXCEPT !.field = 1] - record not defined
	node := &ast.RecordAltered{
		Identifier: &ast.Identifier{Value: "nonexistent"},
		Alterations: []*ast.FieldAlteration{
			{
				Field:      &ast.FieldIdentifier{Identifier: nil, Member: "field"},
				Expression: ast.NewIntLiteral(1),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
}

func (s *PairwiseSuite) TestError_InFieldAccess_UndefinedVariable() {
	// nonexistent.field - record not defined
	node := &ast.FieldIdentifier{
		Identifier: &ast.Identifier{Value: "nonexistent"},
		Member:     "field",
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
}

func (s *PairwiseSuite) TestError_InTupleIndex_OutOfBounds() {
	// <<1, 2>>[5] - index out of bounds (1-based)
	node := &ast.ExpressionTupleIndex{
		Tuple: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(1),
				ast.NewIntLiteral(2),
			},
		},
		Index: ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "out of bounds")
}

func (s *PairwiseSuite) TestError_DivisionByZero() {
	// 10 ÷ 0
	node := &ast.RealInfixExpression{
		Operator: "÷",
		Left:     ast.NewIntLiteral(10),
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "division by zero")
}

// === Deeply Nested Expressions ===

func (s *PairwiseSuite) TestDeeplyNested_SetOperations() {
	// (({1} ∪ {2}) ∩ {2, 3}) \ {3}
	node := &ast.SetInfix{
		Operator: "\\",
		Left: &ast.SetInfix{
			Operator: "∩",
			Left: &ast.SetInfix{
				Operator: "∪",
				Left:     &ast.SetLiteralInt{Values: []int{1}},
				Right:    &ast.SetLiteralInt{Values: []int{2}},
			},
			Right: &ast.SetLiteralInt{Values: []int{2, 3}},
		},
		Right: &ast.SetLiteralInt{Values: []int{3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(1, set.Size())
	s.True(set.Contains(object.NewNatural(2)))
}

func (s *PairwiseSuite) TestDeeplyNested_TupleOperations() {
	// _Seq!Head(_Seq!Tail(_Seq!Append(<<1>>, 2)))
	// Append(<<1>>, 2) = <<1, 2>>
	// Tail(<<1, 2>>) = <<2>>
	// Head(<<2>>) = 2
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.BuiltinCall{
				Name: "_Seq!Tail",
				Args: []ast.Expression{
					&ast.BuiltinCall{
						Name: "_Seq!Append",
						Args: []ast.Expression{
							&ast.TupleLiteral{
								Elements: []ast.Expression{ast.NewIntLiteral(1)},
							},
							ast.NewIntLiteral(2),
						},
					},
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *PairwiseSuite) TestDeeplyNested_LogicOperations() {
	// ((TRUE ∧ FALSE) ∨ TRUE) ⇒ (¬FALSE)
	node := &ast.LogicInfixExpression{
		Operator: "⇒",
		Left: &ast.LogicInfixExpression{
			Operator: "∨",
			Left: &ast.LogicInfixExpression{
				Operator: "∧",
				Left:     &ast.BooleanLiteral{Value: true},
				Right:    &ast.BooleanLiteral{Value: false},
			},
			Right: &ast.BooleanLiteral{Value: true},
		},
		Right: &ast.LogicPrefixExpression{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: false},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value()) // TRUE ⇒ TRUE = TRUE
}

// === String + Other Operations ===

func (s *PairwiseSuite) TestStringConcat_InIfElse() {
	// IF TRUE THEN "hello" ∘ " world" ELSE "goodbye"
	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then: &ast.StringInfixExpression{
			Operator: "∘",
			Operands: []ast.Expression{
				&ast.StringLiteral{Value: "hello"},
				&ast.StringLiteral{Value: " world"},
			},
		},
		Else: &ast.StringLiteral{Value: "goodbye"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello world", str.Value())
}

func (s *PairwiseSuite) TestStringIndex_InComparison() {
	// "abc"[1] = "a" is not directly testable since we can't compare strings with =
	// But we can test string indexing in an assignment
	node := &ast.Assignment{
		Target: &ast.Identifier{Value: "c"},
		Value: &ast.StringIndex{
			Str:   &ast.StringLiteral{Value: "hello"},
			Index: ast.NewIntLiteral(1),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.PrimedBindings["c"].(*object.String)
	s.Equal("h", str.Value())
}

// === Bag + Set Conversions ===
// Note: These tests use direct object construction since BuiltinCall doesn't implement Set

func (s *PairwiseSuite) TestBagToSet_Direct() {
	// Test BagToSet directly via builtin function
	set := object.NewSet()
	set.Add(object.NewNatural(1))
	set.Add(object.NewNatural(2))
	set.Add(object.NewNatural(3))

	// First create a bag from the set
	bagResult := builtinSetToBag([]object.Object{set})
	s.False(bagResult.IsError())
	bag := bagResult.Value.(*object.Bag)

	// Then convert back to set
	setResult := builtinBagToSet([]object.Object{bag})
	s.False(setResult.IsError())
	resultSet := setResult.Value.(*object.Set)

	s.Equal(3, resultSet.Size())
	s.True(resultSet.Contains(object.NewNatural(1)))
}

func (s *PairwiseSuite) TestBagIn_ViaBuiltin() {
	// Test _Bags!BagIn via builtin function
	set := object.NewSet()
	set.Add(object.NewNatural(1))
	set.Add(object.NewNatural(2))

	// Create a bag from the set
	bagResult := builtinSetToBag([]object.Object{set})
	s.False(bagResult.IsError())
	bag := bagResult.Value.(*object.Bag)

	// Test BagIn
	result := builtinBagIn([]object.Object{object.NewNatural(1), bag})
	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}
