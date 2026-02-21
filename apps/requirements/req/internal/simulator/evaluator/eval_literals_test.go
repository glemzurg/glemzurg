package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestLiteralsSuite(t *testing.T) {
	suite.Run(t, new(LiteralsSuite))
}

type LiteralsSuite struct {
	suite.Suite
}

// === String Literals ===

func (s *LiteralsSuite) TestStringLiteral_Simple() {
	node := &ast.StringLiteral{Value: "hello"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello", str.Value())
}

func (s *LiteralsSuite) TestStringLiteral_Empty() {
	node := &ast.StringLiteral{Value: ""}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("", str.Value())
}

func (s *LiteralsSuite) TestStringLiteral_Unicode() {
	node := &ast.StringLiteral{Value: "hello 世界"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello 世界", str.Value())
}

// === Boolean Literals ===

func (s *LiteralsSuite) TestBooleanLiteral_True() {
	node := &ast.BooleanLiteral{Value: true}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LiteralsSuite) TestBooleanLiteral_False() {
	node := &ast.BooleanLiteral{Value: false}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Tuple Literals ===

func (s *LiteralsSuite) TestTupleLiteral_Empty() {
	node := &ast.TupleLiteral{Elements: []ast.Expression{}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(0, tuple.Len())
}

func (s *LiteralsSuite) TestTupleLiteral_SingleElement() {
	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(42),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(1, tuple.Len())
	elem := tuple.At(1).(*object.Number)
	s.Equal("42", elem.Inspect())
}

func (s *LiteralsSuite) TestTupleLiteral_MultipleElements() {
	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			&ast.StringLiteral{Value: "two"},
			// Using natural literal since BooleanLiteral implements Logic, not Expression
			ast.NewIntLiteral(3),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())

	// TLA+ uses 1-based indexing
	num := tuple.At(1).(*object.Number)
	s.Equal("1", num.Inspect())

	str := tuple.At(2).(*object.String)
	s.Equal("two", str.Value())

	num3 := tuple.At(3).(*object.Number)
	s.Equal("3", num3.Inspect())
}

// === Set Literals ===

func (s *LiteralsSuite) TestSetLiteralInt_Empty() {
	node := &ast.SetLiteralInt{Values: []int{}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(0, set.Size())
}

func (s *LiteralsSuite) TestSetLiteralInt_SingleElement() {
	node := &ast.SetLiteralInt{Values: []int{42}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(1, set.Size())
	s.True(set.Contains(object.NewNatural(42)))
}

func (s *LiteralsSuite) TestSetLiteralInt_MultipleElements() {
	node := &ast.SetLiteralInt{Values: []int{1, 2, 3}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
	s.True(set.Contains(object.NewNatural(1)))
	s.True(set.Contains(object.NewNatural(2)))
	s.True(set.Contains(object.NewNatural(3)))
}

func (s *LiteralsSuite) TestSetLiteralInt_Duplicates() {
	// Sets should deduplicate
	node := &ast.SetLiteralInt{Values: []int{1, 1, 2, 2, 3}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
}

func (s *LiteralsSuite) TestSetLiteralEnum_Simple() {
	node := &ast.SetLiteralEnum{Values: []string{"red", "green", "blue"}}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
	s.True(set.Contains(object.NewString("red")))
	s.True(set.Contains(object.NewString("green")))
	s.True(set.Contains(object.NewString("blue")))
}

// === Set Range ===

func (s *LiteralsSuite) TestSetRange_Simple() {
	// 1..5 = {1, 2, 3, 4, 5}
	node := &ast.SetRange{Start: 1, End: 5}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(5, set.Size())
	for i := 1; i <= 5; i++ {
		s.True(set.Contains(object.NewInteger(int64(i))))
	}
}

func (s *LiteralsSuite) TestSetRange_SingleElement() {
	// 5..5 = {5}
	node := &ast.SetRange{Start: 5, End: 5}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(1, set.Size())
	s.True(set.Contains(object.NewInteger(5)))
}

func (s *LiteralsSuite) TestSetRange_Negative() {
	// -2..2 = {-2, -1, 0, 1, 2}
	node := &ast.SetRange{Start: -2, End: 2}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(5, set.Size())
	s.True(set.Contains(object.NewInteger(-2)))
	s.True(set.Contains(object.NewInteger(0)))
	s.True(set.Contains(object.NewInteger(2)))
}

// === Set Constants ===

func (s *LiteralsSuite) TestSetConstant_BOOLEAN() {
	node := &ast.SetConstant{Value: "BOOLEAN"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
	s.True(set.Contains(object.NewBoolean(true)))
	s.True(set.Contains(object.NewBoolean(false)))
}

func (s *LiteralsSuite) TestSetConstant_Nat_Error() {
	// Infinite sets cannot be enumerated
	node := &ast.SetConstant{Value: "Nat"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "infinite set")
}

// === Record Instance ===

func (s *LiteralsSuite) TestRecordInstance_Simple() {
	node := &ast.RecordInstance{
		Bindings: []*ast.FieldBinding{
			{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(10)},
			{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(20)},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)

	x := record.Get("x")
	s.NotNil(x)
	s.Equal("10", x.Inspect())

	y := record.Get("y")
	s.NotNil(y)
	s.Equal("20", y.Inspect())
}

func (s *LiteralsSuite) TestRecordInstance_MixedTypes() {
	node := &ast.RecordInstance{
		Bindings: []*ast.FieldBinding{
			{Field: &ast.Identifier{Value: "name"}, Expression: &ast.StringLiteral{Value: "Alice"}},
			{Field: &ast.Identifier{Value: "age"}, Expression: ast.NewIntLiteral(30)},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)

	name := record.Get("name")
	s.Equal("Alice", name.(*object.String).Value())

	age := record.Get("age")
	s.Equal("30", age.Inspect())
}
