package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestBagsSetsCallsSuite(t *testing.T) {
	suite.Run(t, new(BagsSetsCallsSuite))
}

type BagsSetsCallsSuite struct {
	suite.Suite
}

// === SetConditional ===

func (s *BagsSetsCallsSuite) TestSetConditional_FilterAll() {
	// {x ∈ {1, 2, 3} : TRUE} = {1, 2, 3}
	node := &ast.SetConditional{
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
}

func (s *BagsSetsCallsSuite) TestSetConditional_FilterNone() {
	// {x ∈ {1, 2, 3} : FALSE} = {}
	node := &ast.SetConditional{
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(0, set.Size())
}

func (s *BagsSetsCallsSuite) TestSetConditional_EmptySource() {
	// {x ∈ {} : TRUE} = {}
	node := &ast.SetConditional{
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(0, set.Size())
}

func (s *BagsSetsCallsSuite) TestSetConditional_WithComparisonPredicate() {
	// {x ∈ {1, 2, 3, 4, 5} : 2 < 4} = {1, 2, 3, 4, 5} (predicate always true)
	node := &ast.SetConditional{
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3, 4, 5}},
		},
		Predicate: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(2),
			Operator: "<",
			Right:    ast.NewIntLiteral(4),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(5, set.Size())
}

// === Bag Builtins (direct tests) ===

func (s *BagsSetsCallsSuite) TestBagIn_ElementExists() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinBagIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *BagsSetsCallsSuite) TestBagIn_ElementNotExists() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinBagIn([]object.Object{object.NewNatural(5), bag})

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *BagsSetsCallsSuite) TestBagIn_EmptyBag() {
	bag := object.NewBag()

	result := builtinBagIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === LogicInfixBag (Subbag) - direct object tests ===

func (s *BagsSetsCallsSuite) TestSubbag_IsSubbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	// Test directly via object method
	s.True(bag1.IsSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestSubbag_NotSubbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)
	bag1.Add(object.NewNatural(3), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	s.False(bag1.IsSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestSubbag_EqualBags() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	s.True(bag1.IsSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestSubbag_EmptyIsSubbag() {
	bag1 := object.NewBag()

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	s.True(bag1.IsSubBagOf(bag2))
}

// === Proper Subbag (⊏) Tests ===

func (s *BagsSetsCallsSuite) TestProperSubbag_IsProperSubbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	s.True(bag1.IsProperSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestProperSubbag_EqualBagsNotProper() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	// Equal bags are NOT proper subbags
	s.False(bag1.IsProperSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestProperSubbag_EmptyBag() {
	bag1 := object.NewBag() // empty

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)

	// Empty bag is a proper subbag of non-empty bag
	s.True(bag1.IsProperSubBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestProperSubbag_BothEmpty() {
	bag1 := object.NewBag()
	bag2 := object.NewBag()

	// Empty is not a proper subbag of empty (they're equal)
	s.False(bag1.IsProperSubBagOf(bag2))
}

// === Superbag (⊒) Tests ===

func (s *BagsSetsCallsSuite) TestSuperbag_IsSuperbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)

	s.True(bag1.IsSuperBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestSuperbag_NotSuperbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	s.False(bag1.IsSuperBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestSuperbag_EqualBags() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	// Equal bags: bag1 is superbag of bag2
	s.True(bag1.IsSuperBagOf(bag2))
}

// === Proper Superbag (⊐) Tests ===

func (s *BagsSetsCallsSuite) TestProperSuperbag_IsProperSuperbag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)

	s.True(bag1.IsProperSuperBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestProperSuperbag_EqualBagsNotProper() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	// Equal bags are NOT proper superbags
	s.False(bag1.IsProperSuperBagOf(bag2))
}

func (s *BagsSetsCallsSuite) TestProperSuperbag_OfEmptyBag() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)

	bag2 := object.NewBag() // empty

	// Non-empty bag is proper superbag of empty bag
	s.True(bag1.IsProperSuperBagOf(bag2))
}

// === BinaryBagComparison via AST ===

func (s *BagsSetsCallsSuite) TestEvalBagComparison_ProperSubbag() {
	// Test A ⊏ B via evaluator
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagComparison{
		Operator: "⊏",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *BagsSetsCallsSuite) TestEvalBagComparison_SubbagEq() {
	// Test A ⊑ B via evaluator
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagComparison{
		Operator: "⊑",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value()) // Equal bags, so subbag-or-equal is true
}

func (s *BagsSetsCallsSuite) TestEvalBagComparison_ProperSuperbag() {
	// Test A ⊐ B via evaluator
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagComparison{
		Operator: "⊐",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *BagsSetsCallsSuite) TestEvalBagComparison_SuperbagEq() {
	// Test A ⊒ B via evaluator
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(1), 1)
	bag2.Add(object.NewNatural(2), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagComparison{
		Operator: "⊒",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value()) // Equal bags, so superbag-or-equal is true
}

// === CopiesIn ===

func (s *BagsSetsCallsSuite) TestCopiesIn_ElementExists() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinCopiesIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *BagsSetsCallsSuite) TestCopiesIn_ElementNotExists() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinCopiesIn([]object.Object{object.NewNatural(5), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

func (s *BagsSetsCallsSuite) TestCopiesIn_MultipleCopies() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 3) // 3 copies of 1
	bag.Add(object.NewNatural(2), 1)

	result := builtinCopiesIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3", num.Inspect())
}

func (s *BagsSetsCallsSuite) TestCopiesIn_EmptyBag() {
	bag := object.NewBag()

	result := builtinCopiesIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

// === BagInfix (direct object tests) ===

func (s *BagsSetsCallsSuite) TestBagSum_Simple() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(3), 1)

	result := bag1.Sum(bag2)
	s.Equal(3, len(result.Elements()))
}

func (s *BagsSetsCallsSuite) TestBagSum_Overlapping() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(2), 1)
	bag2.Add(object.NewNatural(3), 1)

	result := bag1.Sum(bag2)
	// 3 unique elements, but element 2 has count 2
	s.Equal(3, len(result.Elements()))
	s.Equal(2, result.CopiesIn(object.NewNatural(2)))
}

func (s *BagsSetsCallsSuite) TestBagDifference_Simple() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)
	bag1.Add(object.NewNatural(3), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(2), 1)

	result := bag1.Difference(bag2)
	s.Equal(2, len(result.Elements()))
}

func (s *BagsSetsCallsSuite) TestBagDifference_NoOverlap() {
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(3), 1)
	bag2.Add(object.NewNatural(4), 1)

	result := bag1.Difference(bag2)
	s.Equal(2, len(result.Elements()))
}

// === BinaryBagOperation via AST (⊕ and ⊖) ===

func (s *BagsSetsCallsSuite) TestEvalBagOperation_Sum() {
	// Test A ⊕ B via evaluator using BinaryBagOperation AST node
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 1)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(2), 1)
	bag2.Add(object.NewNatural(3), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagOperation{
		Operator: "⊕",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	resultBag := result.Value.(*object.Bag)
	// {1:1, 2:1} ⊕ {2:1, 3:1} = {1:1, 2:2, 3:1}
	s.Equal(3, len(resultBag.Elements()))
	s.Equal(2, resultBag.CopiesIn(object.NewNatural(2)))
}

func (s *BagsSetsCallsSuite) TestEvalBagOperation_Difference() {
	// Test A ⊖ B via evaluator using BinaryBagOperation AST node
	bag1 := object.NewBag()
	bag1.Add(object.NewNatural(1), 1)
	bag1.Add(object.NewNatural(2), 2)

	bag2 := object.NewBag()
	bag2.Add(object.NewNatural(2), 1)

	bindings := NewBindings()
	bindings.Set("a", bag1, NamespaceGlobal)
	bindings.Set("b", bag2, NamespaceGlobal)

	node := &ast.BinaryBagOperation{
		Operator: "⊖",
		Left:     &ast.Identifier{Value: "a"},
		Right:    &ast.Identifier{Value: "b"},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	resultBag := result.Value.(*object.Bag)
	// {1:1, 2:2} ⊖ {2:1} = {1:1, 2:1}
	s.Equal(2, len(resultBag.Elements()))
	s.Equal(1, resultBag.CopiesIn(object.NewNatural(2)))
}

// === CallExpression (error case) ===

func (s *BagsSetsCallsSuite) TestCallExpression_NotImplemented() {
	// CallExpression should return "not yet implemented" error
	node := &ast.CallExpression{
		FunctionName: &ast.Identifier{Value: "MyFunction"},
		Parameter: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "not yet implemented")
	s.Contains(result.Error.Message, "MyFunction")
}

func (s *BagsSetsCallsSuite) TestCallExpression_WithClass() {
	// Class!FunctionName(record)
	node := &ast.CallExpression{
		Class:        &ast.Identifier{Value: "MyClass"},
		FunctionName: &ast.Identifier{Value: "New"},
		Parameter: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "MyClass!New")
}

func (s *BagsSetsCallsSuite) TestCallExpression_ModelScope() {
	// _FunctionName(record)
	node := &ast.CallExpression{
		ModelScope:   true,
		FunctionName: &ast.Identifier{Value: "Initialize"},
		Parameter: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "_Initialize")
}

func (s *BagsSetsCallsSuite) TestCallExpression_FullyScoped() {
	// Domain!Subdomain!Class!FunctionName(record)
	node := &ast.CallExpression{
		Domain:       &ast.Identifier{Value: "MyDomain"},
		Subdomain:    &ast.Identifier{Value: "MySub"},
		Class:        &ast.Identifier{Value: "MyClass"},
		FunctionName: &ast.Identifier{Value: "DoSomething"},
		Parameter: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "MyDomain!MySub!MyClass!DoSomething")
}
