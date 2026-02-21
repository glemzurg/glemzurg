package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestTuplesSuite(t *testing.T) {
	suite.Run(t, new(TuplesSuite))
}

type TuplesSuite struct {
	suite.Suite
}

// === Tuple Indexing ===

func (s *TuplesSuite) TestTupleIndex_First() {
	// <<1, 2, 3>>[1] = 1
	node := &ast.ExpressionTupleIndex{
		Tuple: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(1),
				ast.NewIntLiteral(2),
				ast.NewIntLiteral(3),
			},
		},
		Index: ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *TuplesSuite) TestTupleIndex_Middle() {
	// <<1, 2, 3>>[2] = 2
	node := &ast.ExpressionTupleIndex{
		Tuple: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(1),
				ast.NewIntLiteral(2),
				ast.NewIntLiteral(3),
			},
		},
		Index: ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *TuplesSuite) TestTupleIndex_OutOfBounds() {
	// <<1, 2, 3>>[5] = error
	node := &ast.ExpressionTupleIndex{
		Tuple: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(1),
				ast.NewIntLiteral(2),
				ast.NewIntLiteral(3),
			},
		},
		Index: ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "out of bounds")
}

// === Sequence Head ===

func (s *TuplesSuite) TestSeqHead_Simple() {
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

func (s *TuplesSuite) TestSeqHead_Empty() {
	// _Seq!Head(<<>>) = error
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "empty")
}

// === Sequence Tail ===

func (s *TuplesSuite) TestSeqTail_Simple() {
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
	s.Equal("2", tuple.At(1).Inspect())
	s.Equal("3", tuple.At(2).Inspect())
}

func (s *TuplesSuite) TestSeqTail_SingleElement() {
	// _Seq!Tail(<<1>>) = <<>>
	node := &ast.BuiltinCall{
		Name: "_Seq!Tail",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
				},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(0, tuple.Len())
}

// === Sequence Append ===

func (s *TuplesSuite) TestSeqAppend_Simple() {
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
	s.Equal("1", tuple.At(1).Inspect())
	s.Equal("2", tuple.At(2).Inspect())
	s.Equal("3", tuple.At(3).Inspect())
}

func (s *TuplesSuite) TestSeqAppend_Empty() {
	// _Seq!Append(<<>>, 1) = <<1>>
	node := &ast.BuiltinCall{
		Name: "_Seq!Append",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{},
			},
			ast.NewIntLiteral(1),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(1, tuple.Len())
	s.Equal("1", tuple.At(1).Inspect())
}

// === Sequence Length ===

func (s *TuplesSuite) TestSeqLen_Simple() {
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

func (s *TuplesSuite) TestSeqLen_Empty() {
	// _Seq!Len(<<>>) = 0
	node := &ast.BuiltinCall{
		Name: "_Seq!Len",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

// === Tuple Concatenation ===

func (s *TuplesSuite) TestTupleConcat_Simple() {
	// <<1, 2>> ∘ <<3, 4>> = <<1, 2, 3, 4>>
	node := &ast.TupleInfixExpression{
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
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(4, tuple.Len())
	s.Equal("1", tuple.At(1).Inspect())
	s.Equal("4", tuple.At(4).Inspect())
}

func (s *TuplesSuite) TestTupleConcat_Empty() {
	// <<>> ∘ <<1, 2>> = <<1, 2>>
	node := &ast.TupleInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{}},
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
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

func (s *TuplesSuite) TestTupleConcat_ThreeOperands() {
	// <<1>> ∘ <<2>> ∘ <<3>> via TupleConcat AST node (parser output)
	node := &ast.TupleConcat{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{ast.NewIntLiteral(1)},
			},
			&ast.TupleLiteral{
				Elements: []ast.Expression{ast.NewIntLiteral(2)},
			},
			&ast.TupleLiteral{
				Elements: []ast.Expression{ast.NewIntLiteral(3)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)
	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
	s.Equal("1", tuple.At(1).Inspect())
	s.Equal("2", tuple.At(2).Inspect())
	s.Equal("3", tuple.At(3).Inspect())
}

func (s *TuplesSuite) TestTupleConcat_WithVariables() {
	// a ∘ b where a and b are tuple variables
	tuple1 := object.NewTupleFromElements([]object.Object{
		object.NewNatural(1),
		object.NewNatural(2),
	})
	tuple2 := object.NewTupleFromElements([]object.Object{
		object.NewNatural(3),
	})

	bindings := NewBindings()
	bindings.Set("a", tuple1, NamespaceGlobal)
	bindings.Set("b", tuple2, NamespaceGlobal)

	node := &ast.TupleConcat{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.Identifier{Value: "a"},
			&ast.Identifier{Value: "b"},
		},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
}

// === String Concatenation ===

func (s *TuplesSuite) TestStringConcat_Simple() {
	// "hello" ∘ "world" = "helloworld"
	node := &ast.StringInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.StringLiteral{Value: "hello"},
			&ast.StringLiteral{Value: "world"},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)
	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("helloworld", str.Value())
}

func (s *TuplesSuite) TestStringConcat_ThreeStrings() {
	// "a" ∘ "b" ∘ "c" = "abc"
	node := &ast.StringInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.StringLiteral{Value: "a"},
			&ast.StringLiteral{Value: "b"},
			&ast.StringLiteral{Value: "c"},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)
	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("abc", str.Value())
}

func (s *TuplesSuite) TestStringConcat_WithVariables() {
	// a ∘ b where a and b are string variables
	bindings := NewBindings()
	bindings.Set("a", object.NewString("foo"), NamespaceGlobal)
	bindings.Set("b", object.NewString("bar"), NamespaceGlobal)

	node := &ast.StringInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.Identifier{Value: "a"},
			&ast.Identifier{Value: "b"},
		},
	}

	result := Eval(node, bindings)
	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("foobar", str.Value())
}

func (s *TuplesSuite) TestStringConcat_EmptyString() {
	// "" ∘ "hello" = "hello"
	node := &ast.StringInfixExpression{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.StringLiteral{Value: ""},
			&ast.StringLiteral{Value: "hello"},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)
	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello", str.Value())
}

// === Stack Operations ===

func (s *TuplesSuite) TestStackPush_Simple() {
	// _Stack!Push(<<2, 3>>, 1) = <<1, 2, 3>>
	node := &ast.BuiltinCall{
		Name: "_Stack!Push",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(2),
					ast.NewIntLiteral(3),
				},
			},
			ast.NewIntLiteral(1),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
	s.Equal("1", tuple.At(1).Inspect()) // Pushed element is at front
	s.Equal("2", tuple.At(2).Inspect())
	s.Equal("3", tuple.At(3).Inspect())
}

func (s *TuplesSuite) TestStackPop_Simple() {
	// _Stack!Pop(<<1, 2, 3>>) = 1 (returns the head)
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

// === Queue Operations ===

func (s *TuplesSuite) TestQueueEnqueue_Simple() {
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
	s.Equal("3", tuple.At(3).Inspect()) // Enqueued at end
}

func (s *TuplesSuite) TestQueueDequeue_Simple() {
	// _Queue!Dequeue(<<1, 2, 3>>) = 1 (returns the head)
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

// === Nested Tuple Operations ===

func (s *TuplesSuite) TestNested_HeadOfTail() {
	// _Seq!Head(_Seq!Tail(<<1, 2, 3>>)) = 2
	tail := &ast.BuiltinCall{
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
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{tail},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}
