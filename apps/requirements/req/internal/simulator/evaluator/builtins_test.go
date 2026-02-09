package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestBuiltinsSuite(t *testing.T) {
	suite.Run(t, new(BuiltinsSuite))
}

type BuiltinsSuite struct {
	suite.Suite
}

// === Sequence Tests ===

func (s *BuiltinsSuite) TestSeqHead_Simple() {
	// _Seq!Head(<<1, 2, 3>>) = 1
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
					ast.NewIntLiteral(3),
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *BuiltinsSuite) TestSeqHead_Empty() {
	// _Seq!Head(<<>>) = error
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "empty")
}

func (s *BuiltinsSuite) TestSeqTail_Simple() {
	// _Seq!Tail(<<1, 2, 3>>) = <<2, 3>>
	node := &ast.BuiltinCall{
		Name: "_Seq!Tail",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
					ast.NewIntLiteral(3),
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(2, tuple.Len())
}

func (s *BuiltinsSuite) TestSeqAppend_Simple() {
	// _Seq!Append(<<1, 2>>, 3) = <<1, 2, 3>>
	node := &ast.BuiltinCall{
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
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
}

func (s *BuiltinsSuite) TestSeqLen_Simple() {
	// _Seq!Len(<<1, 2, 3>>) = 3
	node := &ast.BuiltinCall{
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
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3", num.Inspect())
}

func (s *BuiltinsSuite) TestSeqLen_Empty() {
	// _Seq!Len(<<>>) = 0
	node := &ast.BuiltinCall{
		Name: "_Seq!Len",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

// === Stack Tests (LIFO) ===

func (s *BuiltinsSuite) TestStackPush_Simple() {
	// _Stack!Push(<<1, 2>>, 0) = <<0, 1, 2>>
	node := &ast.BuiltinCall{
		Name: "_Stack!Push",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
				},
			},
			ast.NewIntLiteral(0),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
	// First element should be 0 (pushed to front)
	s.Equal("0", tuple.At(1).Inspect())
}

func (s *BuiltinsSuite) TestStackPop_Simple() {
	// _Stack!Pop(<<1, 2, 3>>) = 1
	node := &ast.BuiltinCall{
		Name: "_Stack!Pop",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
					ast.NewIntLiteral(3),
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *BuiltinsSuite) TestStackPop_Empty() {
	// _Stack!Pop(<<>>) = error
	node := &ast.BuiltinCall{
		Name: "_Stack!Pop",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "empty")
}

// === Queue Tests (FIFO) ===

func (s *BuiltinsSuite) TestQueueEnqueue_Simple() {
	// _Queue!Enqueue(<<1, 2>>, 3) = <<1, 2, 3>>
	node := &ast.BuiltinCall{
		Name: "_Queue!Enqueue",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
				},
			},
			ast.NewIntLiteral(3),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
	// Last element should be 3
	s.Equal("3", tuple.At(3).Inspect())
}

func (s *BuiltinsSuite) TestQueueDequeue_Simple() {
	// _Queue!Dequeue(<<1, 2, 3>>) = 1
	node := &ast.BuiltinCall{
		Name: "_Queue!Dequeue",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
					ast.NewIntLiteral(3),
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *BuiltinsSuite) TestQueueDequeue_Empty() {
	// _Queue!Dequeue(<<>>) = error
	node := &ast.BuiltinCall{
		Name: "_Queue!Dequeue",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{}},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "empty")
}

// === Bag Tests ===
// Note: Bag tests use direct object creation since SetLiteralInt doesn't implement Expression

func (s *BuiltinsSuite) TestSetToBag_Simple() {
	// Create a set directly and test SetToBag builtin
	set := object.NewSet()
	set.Add(object.NewNatural(1))
	set.Add(object.NewNatural(2))
	set.Add(object.NewNatural(3))

	result := builtinSetToBag([]object.Object{set})

	s.False(result.IsError())
	bag := result.Value.(*object.Bag)
	s.Equal(3, len(bag.Elements()))
}

func (s *BuiltinsSuite) TestBagToSet_Simple() {
	// Create a bag directly and test BagToSet builtin
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinBagToSet([]object.Object{bag})

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
}

func (s *BuiltinsSuite) TestCopiesIn_Found() {
	// Create a bag and test CopiesIn
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinCopiesIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *BuiltinsSuite) TestCopiesIn_NotFound() {
	// Create a bag and test CopiesIn for element not in bag
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinCopiesIn([]object.Object{object.NewNatural(99), bag})

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

func (s *BuiltinsSuite) TestBagIn_Found() {
	// Create a bag and test BagIn
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinBagIn([]object.Object{object.NewNatural(1), bag})

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *BuiltinsSuite) TestBagIn_NotFound() {
	// Create a bag and test BagIn for element not in bag
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)
	bag.Add(object.NewNatural(3), 1)

	result := builtinBagIn([]object.Object{object.NewNatural(99), bag})

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Error Tests ===

func (s *BuiltinsSuite) TestUnknownBuiltin() {
	// _Unknown!Function() = error
	node := &ast.BuiltinCall{
		Name: "_Unknown!Function",
		Args: []ast.Expression{},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "unknown builtin")
}

func (s *BuiltinsSuite) TestWrongArgCount() {
	// _Seq!Head() without args = error
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "requires 1 argument")
}

func (s *BuiltinsSuite) TestWrongArgType() {
	// _Seq!Head(5) with non-tuple = error
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "requires Tuple")
}
