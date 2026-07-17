package evaluator

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type GZBuiltinSuite struct {
	suite.Suite
}

func TestGZBuiltinSuite(t *testing.T) {
	suite.Run(t, new(GZBuiltinSuite))
}

func (s *GZBuiltinSuite) TestWhenNotNullTrueWhenDriverNull() {
	bindings := NewBindings()
	bindings.Set("ISO", EMPTY_SET, NamespaceLocal)
	call := &me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNotNull,
		Args: []me.Expression{
			&me.LocalVar{Name: "ISO"},
			&me.BoolLiteral{Value: false}, // would fail if evaluated
		},
	}
	result := Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())
}

func (s *GZBuiltinSuite) TestWhenNotNullEvaluatesEquationWhenSet() {
	bindings := NewBindings()
	bindings.Set("ISO", object.NewString("USD"), NamespaceLocal)
	call := &me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNotNull,
		Args: []me.Expression{
			&me.LocalVar{Name: "ISO"},
			&me.Compare{
				Op:    me.CompareEq,
				Left:  &me.LocalVar{Name: "ISO"},
				Right: &me.StringLiteral{Value: "USD"},
			},
		},
	}
	result := Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())
}

func (s *GZBuiltinSuite) TestWhenNullEvaluatesEquationWhenNull() {
	bindings := NewBindings()
	bindings.Set("ISO", EMPTY_SET, NamespaceLocal)
	bindings.Set("SocialOnly", object.NewBoolean(true), NamespaceLocal)
	call := &me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNull,
		Args: []me.Expression{
			&me.LocalVar{Name: "ISO"},
			&me.Compare{
				Op:    me.CompareEq,
				Left:  &me.LocalVar{Name: "SocialOnly"},
				Right: &me.BoolLiteral{Value: true},
			},
		},
	}
	result := Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())
}

func (s *GZBuiltinSuite) TestWhenNullTrueWhenDriverSet() {
	bindings := NewBindings()
	bindings.Set("ISO", object.NewString("USD"), NamespaceLocal)
	call := &me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNull,
		Args: []me.Expression{
			&me.LocalVar{Name: "ISO"},
			&me.BoolLiteral{Value: false},
		},
	}
	result := Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())
}

func (s *GZBuiltinSuite) TestWhenNullElsePicksBranch() {
	bindings := NewBindings()
	bindings.Set("ISO", EMPTY_SET, NamespaceLocal)
	bindings.Set("Abbr", object.NewString("SOC"), NamespaceLocal)
	call := &me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNullElse,
		Args: []me.Expression{
			&me.LocalVar{Name: "ISO"},
			&me.Compare{
				Op:    me.CompareEq,
				Left:  &me.LocalVar{Name: "Abbr"},
				Right: &me.StringLiteral{Value: "SOC"},
			},
			&me.BoolLiteral{Value: false},
		},
	}
	result := Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())

	bindings.Set("ISO", object.NewString("USD"), NamespaceLocal)
	bindings.Set("Abbr", object.NewString("USD"), NamespaceLocal)
	call.Args[2] = &me.Compare{
		Op:    me.CompareEq,
		Left:  &me.LocalVar{Name: "ISO"},
		Right: &me.LocalVar{Name: "Abbr"},
	}
	result = Eval(call, bindings)
	s.Require().False(result.IsError(), result.Error)
	boolVal, ok = result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())
}

func (s *GZBuiltinSuite) TestWhenNullElseArity() {
	result := Eval(&me.BuiltinCall{
		Module:   ModuleGZ,
		Function: GZWhenNullElse,
		Args:     []me.Expression{&me.LocalVar{Name: "ISO"}},
	}, NewBindings())
	s.True(result.IsError())
	s.Require().Contains(result.Error.Inspect(), "3 arguments")
}
