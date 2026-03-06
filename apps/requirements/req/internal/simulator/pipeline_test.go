package simulator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/suite"
)

func TestPipelineSuite(t *testing.T) {
	suite.Run(t, new(PipelineSuite))
}

type PipelineSuite struct {
	suite.Suite
}

func (s *PipelineSuite) TestEval_NaturalLiteral() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := ast.NewIntLiteral(42)
	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *PipelineSuite) TestEval_BooleanLiteral() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := &ast.BooleanLiteral{Value: true}
	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *PipelineSuite) TestEval_StringLiteral() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := &ast.StringLiteral{Value: "hello"}
	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello", str.Value())
}

func (s *PipelineSuite) TestEval_Arithmetic() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	// 3 + 4 = 7
	node := &ast.RealInfixExpression{
		Operator: "+",
		Left:     ast.NewIntLiteral(3),
		Right:    ast.NewIntLiteral(4),
	}

	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("7", num.Inspect())
}

func (s *PipelineSuite) TestEval_SetLiteral() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	// SetLiteralEnum uses string values
	node := &ast.SetLiteralEnum{
		Values: []string{"a", "b", "c"},
	}

	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(3, set.Size())
}

func (s *PipelineSuite) TestInferType_NaturalLiteral() {
	pipeline := NewPipeline()

	node := ast.NewIntLiteral(42)
	typ, err := pipeline.InferType(node)

	s.NoError(err)
	s.IsType(types.Number{}, typ)
}

func (s *PipelineSuite) TestInferType_BooleanLiteral() {
	pipeline := NewPipeline()

	node := &ast.BooleanLiteral{Value: true}
	typ, err := pipeline.InferType(node)

	s.NoError(err)
	s.IsType(types.Boolean{}, typ)
}

func (s *PipelineSuite) TestInferType_SetLiteral() {
	pipeline := NewPipeline()

	// SetLiteralEnum uses string values
	node := &ast.SetLiteralEnum{
		Values: []string{"a", "b"},
	}

	typ, err := pipeline.InferType(node)

	s.NoError(err)
	setType, ok := typ.(types.Set)
	s.True(ok)
	s.IsType(types.String{}, setType.Element)
}

func (s *PipelineSuite) TestCompile_ReusableExpression() {
	pipeline := NewPipeline()

	node := ast.NewIntLiteral(100)
	compiled, err := pipeline.Compile(node)

	s.NoError(err)
	s.NotNil(compiled)

	// Evaluate multiple times with different bindings
	result1 := compiled.Eval(evaluator.NewBindings())
	result2 := compiled.Eval(evaluator.NewBindings())

	s.False(result1.IsError())
	s.False(result2.IsError())
	s.Equal("100", result1.Value.(*object.Number).Inspect())
	s.Equal("100", result2.Value.(*object.Number).Inspect())
}

func (s *PipelineSuite) TestCompiledExpr_Type() {
	pipeline := NewPipeline()

	node := &ast.BooleanLiteral{Value: false}
	compiled, err := pipeline.Compile(node)

	s.NoError(err)
	s.IsType(types.Boolean{}, compiled.Type())
}

func (s *PipelineSuite) TestRun_ConvenienceFunction() {
	bindings := evaluator.NewBindings()
	node := ast.NewIntLiteral(55)

	result := Run(node, bindings)

	s.False(result.IsError())
	s.Equal("55", result.Value.(*object.Number).Inspect())
}

func (s *PipelineSuite) TestCheck_ConvenienceFunction() {
	node := &ast.StringLiteral{Value: "test"}

	typ, err := Check(node)

	s.NoError(err)
	s.IsType(types.String{}, typ)
}

func (s *PipelineSuite) TestDeclareVariable() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()
	bindings.Set("x", object.NewNatural(42), evaluator.NamespaceGlobal)

	// Declare x as a Number type
	pipeline.DeclareVariable("x", types.Number{})

	node := &ast.Identifier{Value: "x"}
	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	s.Equal("42", result.Value.(*object.Number).Inspect())
}

func (s *PipelineSuite) TestEval_IfThenElse() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := &ast.ExpressionIfElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewIntLiteral(1),
		Else:      ast.NewIntLiteral(2),
	}

	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	s.Equal("1", result.Value.(*object.Number).Inspect())
}

func (s *PipelineSuite) TestEval_TupleLiteral() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		},
	}

	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	tuple := result.Value.(*object.Tuple)
	s.Equal(3, tuple.Len())
}

func (s *PipelineSuite) TestEval_RecordInstance() {
	pipeline := NewPipeline()
	bindings := evaluator.NewBindings()

	node := &ast.RecordInstance{
		Bindings: []*ast.FieldBinding{
			{
				Field:      &ast.Identifier{Value: "name"},
				Expression: &ast.StringLiteral{Value: "Alice"},
			},
			{
				Field:      &ast.Identifier{Value: "age"},
				Expression: ast.NewIntLiteral(30),
			},
		},
	}

	result := pipeline.Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal("Alice", record.Get("name").(*object.String).Value())
	s.Equal("30", record.Get("age").Inspect())
}
