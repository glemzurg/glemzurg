package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestCollectionsSuite(t *testing.T) {
	suite.Run(t, new(CollectionsSuite))
}

type CollectionsSuite struct {
	suite.Suite
}

// === Set Union (∪) ===

func (s *CollectionsSuite) TestSetUnion_Simple() {
	// {1, 2} ∪ {2, 3} = {1, 2, 3}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "∪",
		Right:    &ast.SetLiteralInt{Values: []int{2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
	s.True(set.Contains(object.NewNatural(1)))
	s.True(set.Contains(object.NewNatural(2)))
	s.True(set.Contains(object.NewNatural(3)))
}

func (s *CollectionsSuite) TestSetUnion_EmptySet() {
	// {} ∪ {1, 2} = {1, 2}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{}},
		Operator: "∪",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
}

// === Set Intersection (∩) ===

func (s *CollectionsSuite) TestSetIntersection_Simple() {
	// {1, 2, 3} ∩ {2, 3, 4} = {2, 3}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Operator: "∩",
		Right:    &ast.SetLiteralInt{Values: []int{2, 3, 4}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
	s.True(set.Contains(object.NewNatural(2)))
	s.True(set.Contains(object.NewNatural(3)))
}

func (s *CollectionsSuite) TestSetIntersection_Disjoint() {
	// {1, 2} ∩ {3, 4} = {}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "∩",
		Right:    &ast.SetLiteralInt{Values: []int{3, 4}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(0, set.Size())
}

// === Set Difference (\) ===

func (s *CollectionsSuite) TestSetDifference_Simple() {
	// {1, 2, 3} \ {2} = {1, 3}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Operator: "\\",
		Right:    &ast.SetLiteralInt{Values: []int{2}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
	s.True(set.Contains(object.NewNatural(1)))
	s.True(set.Contains(object.NewNatural(3)))
	s.False(set.Contains(object.NewNatural(2)))
}

func (s *CollectionsSuite) TestSetDifference_NoOverlap() {
	// {1, 2} \ {3, 4} = {1, 2}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "\\",
		Right:    &ast.SetLiteralInt{Values: []int{3, 4}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
}

// === Nested Set Operations ===

func (s *CollectionsSuite) TestSetNested_UnionIntersection() {
	// ({1, 2} ∪ {2, 3}) ∩ {2, 4} = {1, 2, 3} ∩ {2, 4} = {2}
	union := &ast.SetInfix{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "∪",
		Right:    &ast.SetLiteralInt{Values: []int{2, 3}},
	}
	node := &ast.SetInfix{
		Left:     union,
		Operator: "∩",
		Right:    &ast.SetLiteralInt{Values: []int{2, 4}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(1, set.Size())
	s.True(set.Contains(object.NewNatural(2)))
}

// === Set Membership (∈) ===

func (s *CollectionsSuite) TestMembership_Contains() {
	// 2 ∈ {1, 2, 3} = TRUE
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(2),
		Operator: "∈",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *CollectionsSuite) TestMembership_NotContains() {
	// 5 ∈ {1, 2, 3} = FALSE
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(5),
		Operator: "∈",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *CollectionsSuite) TestMembership_NotIn() {
	// 5 ∉ {1, 2, 3} = TRUE
	node := &ast.LogicMembership{
		Left:     ast.NewIntLiteral(5),
		Operator: "∉",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Set Comparison (⊆, ⊂, etc.) ===

func (s *CollectionsSuite) TestSetSubset() {
	// {1, 2} ⊆ {1, 2, 3} = TRUE
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "⊆",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *CollectionsSuite) TestSetSubsetEqual() {
	// {1, 2, 3} ⊆ {1, 2, 3} = TRUE
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Operator: "⊆",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *CollectionsSuite) TestSetProperSubset() {
	// {1, 2} ⊂ {1, 2, 3} = TRUE
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "⊂",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *CollectionsSuite) TestSetProperSubsetEqual_False() {
	// {1, 2, 3} ⊂ {1, 2, 3} = FALSE (equal sets are not proper subsets)
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Operator: "⊂",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *CollectionsSuite) TestSetEquals() {
	// {1, 2, 3} = {3, 2, 1} = TRUE (sets are unordered)
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Operator: "=",
		Right:    &ast.SetLiteralInt{Values: []int{3, 2, 1}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *CollectionsSuite) TestSetNotEquals() {
	// {1, 2} ≠ {1, 2, 3} = TRUE
	node := &ast.LogicInfixSet{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "≠",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Bag Operations ===

func (s *CollectionsSuite) TestBagUnion() {
	// Bag union uses MAX of counts (not sum)
	// Testing at object level since BagInfix requires ast.Bag types

	leftBag := object.NewBag()
	leftBag.Add(object.NewNatural(1), 2) // 2 copies of 1
	leftBag.Add(object.NewNatural(2), 1) // 1 copy of 2

	rightBag := object.NewBag()
	rightBag.Add(object.NewNatural(1), 1) // 1 copy of 1
	rightBag.Add(object.NewNatural(3), 1) // 1 copy of 3

	// Directly test bag union (uses max of counts)
	result := leftBag.Union(rightBag)
	s.Equal(4, result.Size())                         // max(2,1) + 1 + 1 = 4
	s.Equal(2, result.CopiesIn(object.NewNatural(1))) // max(2, 1) = 2 copies of 1
	s.Equal(1, result.CopiesIn(object.NewNatural(2))) // 1 copy of 2
	s.Equal(1, result.CopiesIn(object.NewNatural(3))) // 1 copy of 3
}

func (s *CollectionsSuite) TestBagDifference() {
	// Bag difference: test at object level
	leftBag := object.NewBag()
	leftBag.Add(object.NewNatural(1), 2) // 2 copies of 1
	leftBag.Add(object.NewNatural(2), 1) // 1 copy of 2

	rightBag := object.NewBag()
	rightBag.Add(object.NewNatural(1), 1) // 1 copy of 1

	result := leftBag.Difference(rightBag)
	s.Equal(2, result.Size())
	s.Equal(1, result.CopiesIn(object.NewNatural(1))) // 2 - 1 = 1 copy
	s.Equal(1, result.CopiesIn(object.NewNatural(2))) // 1 copy unchanged
}

// === Bag Contains ===

func (s *CollectionsSuite) TestBagContains() {
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)

	s.True(bag.Contains(object.NewNatural(1)))
	s.True(bag.Contains(object.NewNatural(2)))
	s.False(bag.Contains(object.NewNatural(3)))
}

// === Set to Bag / Bag to Set ===

func (s *CollectionsSuite) TestBagToSet() {
	// Converting a bag with duplicates to a set removes duplicates
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 2) // 2 copies
	bag.Add(object.NewNatural(2), 1) // 1 copy

	elements := bag.Elements()
	set := object.NewSetFromElements(elements)

	s.Equal(2, set.Size())
	s.True(set.Contains(object.NewNatural(1)))
	s.True(set.Contains(object.NewNatural(2)))
}

// === Set with Strings ===

func (s *CollectionsSuite) TestSetWithStrings() {
	// {"a", "b"} ∪ {"b", "c"} = {"a", "b", "c"}
	node := &ast.SetInfix{
		Left:     &ast.SetLiteralEnum{Values: []string{"a", "b"}},
		Operator: "∪",
		Right:    &ast.SetLiteralEnum{Values: []string{"b", "c"}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
	s.True(set.Contains(object.NewString("a")))
	s.True(set.Contains(object.NewString("b")))
	s.True(set.Contains(object.NewString("c")))
}
