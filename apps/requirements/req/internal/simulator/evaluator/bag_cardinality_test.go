package evaluator_test

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BagAndFiniteSetsSuite struct {
	suite.Suite
}

func TestBagAndFiniteSetsSuite(t *testing.T) {
	suite.Run(t, new(BagAndFiniteSetsSuite))
}

func (s *BagAndFiniteSetsSuite) TestBagCardinalityAcceptsOnlyBag() {
	bag := object.NewBag()
	bag.Add(object.NewInteger(5), 2)
	bag.Add(object.NewInteger(7), 1)

	set := object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)})
	ar := object.NewAssociationRelation(set, "LinkDef", nil)

	tests := []struct {
		name    string
		value   object.Object
		want    int64
		wantErr string
	}{
		{name: "bag", value: bag, want: 3},
		{name: "set rejected", value: set, wantErr: "requires Bag"},
		{name: "AssociationRelation rejected", value: ar, wantErr: "requires Bag"},
		{name: "number rejected", value: object.NewInteger(1), wantErr: "requires Bag"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			call := &me.BuiltinCall{
				Module:   "_Bags",
				Function: "BagCardinality",
				Args:     []me.Expression{&me.LocalVar{Name: "s"}},
			}
			bindings := evaluator.NewBindings()
			bindings.Set("s", tc.value, evaluator.NamespaceLocal)
			result := evaluator.Eval(call, bindings)
			if tc.wantErr != "" {
				s.Require().True(result.IsError())
				s.Contains(result.Error.Inspect(), tc.wantErr)
				return
			}
			s.Require().False(result.IsError(), result.Error)
			n, ok := result.Value.(*object.Number)
			s.Require().True(ok)
			s.Equal(0, n.Cmp(object.NewInteger(tc.want)))
		})
	}
}

func (s *BagAndFiniteSetsSuite) TestFiniteSetsCardinality() {
	set := object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)})
	ep := object.NewRecord()
	ep.Set("Code", object.NewString("US"))
	ar := object.NewAssociationRelation(
		object.NewSetFromElements([]object.Object{ep}),
		"LinkDef",
		map[*object.Record]*object.Record{ep: object.NewRecord()},
	)
	bag := object.NewBag()
	bag.Add(object.NewInteger(1), 1)

	tests := []struct {
		name    string
		value   object.Object
		want    int64
		wantErr string
	}{
		{name: "set", value: set, want: 2},
		{name: "empty set", value: object.NewSet(), want: 0},
		{name: "AssociationRelation endpoints", value: ar, want: 1},
		{name: "bag rejected", value: bag, wantErr: "requires Set, got Bag"},
		{name: "number rejected", value: object.NewInteger(3), wantErr: "requires Set"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			call := &me.BuiltinCall{
				Module:   "_FiniteSets",
				Function: "Cardinality",
				Args:     []me.Expression{&me.LocalVar{Name: "s"}},
			}
			bindings := evaluator.NewBindings()
			bindings.Set("s", tc.value, evaluator.NamespaceLocal)
			result := evaluator.Eval(call, bindings)
			if tc.wantErr != "" {
				s.Require().True(result.IsError())
				s.Contains(result.Error.Inspect(), tc.wantErr)
				return
			}
			s.Require().False(result.IsError(), result.Error)
			n, ok := result.Value.(*object.Number)
			s.Require().True(ok)
			s.Equal(0, n.Cmp(object.NewInteger(tc.want)))
		})
	}
}

func (s *BagAndFiniteSetsSuite) TestSetToBagRejectsBag() {
	bag := object.NewBag()
	bag.Add(object.NewInteger(1), 1)
	call := &me.BuiltinCall{
		Module:   "_Bags",
		Function: "SetToBag",
		Args:     []me.Expression{&me.LocalVar{Name: "b"}},
	}
	bindings := evaluator.NewBindings()
	bindings.Set("b", bag, evaluator.NamespaceLocal)
	result := evaluator.Eval(call, bindings)
	s.Require().True(result.IsError())
	s.Contains(result.Error.Inspect(), "got Bag")
}

func (s *BagAndFiniteSetsSuite) TestBagCardinalityAfterSetToBag() {
	set := object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)})
	composed := &me.BuiltinCall{
		Module:   "_Bags",
		Function: "BagCardinality",
		Args: []me.Expression{
			&me.BuiltinCall{
				Module:   "_Bags",
				Function: "SetToBag",
				Args:     []me.Expression{&me.LocalVar{Name: "s"}},
			},
		},
	}
	bindings := evaluator.NewBindings()
	bindings.Set("s", set, evaluator.NamespaceLocal)
	result := evaluator.Eval(composed, bindings)
	s.Require().False(result.IsError(), result.Error)
	n, ok := result.Value.(*object.Number)
	s.Require().True(ok)
	s.Equal(0, n.Cmp(object.NewInteger(2)))
}

// Keep legacy package-level name for -run filters.
func TestEvalBagCardinalitySetAndBag(t *testing.T) {
	// BagCardinality no longer accepts sets; composition still works.
	set := object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)})
	composed := &me.BuiltinCall{
		Module:   "_Bags",
		Function: "BagCardinality",
		Args: []me.Expression{
			&me.BuiltinCall{
				Module:   "_Bags",
				Function: "SetToBag",
				Args:     []me.Expression{&me.LocalVar{Name: "s"}},
			},
		},
	}
	bindings := evaluator.NewBindings()
	bindings.Set("s", set, evaluator.NamespaceLocal)
	result := evaluator.Eval(composed, bindings)
	require.False(t, result.IsError(), result.Error)
	count, ok := result.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, 0, count.Cmp(object.NewInteger(2)))

	bag := object.NewBag()
	bag.Add(object.NewInteger(5), 2)
	bag.Add(object.NewInteger(7), 1)
	bagCall := &me.BuiltinCall{Module: "_Bags", Function: "BagCardinality", Args: []me.Expression{
		&me.LocalVar{Name: "s"},
	}}
	bindings.Set("s", bag, evaluator.NamespaceLocal)
	bagResult := evaluator.Eval(bagCall, bindings)
	require.False(t, bagResult.IsError(), bagResult.Error)
	total, ok := bagResult.Value.(*object.Number)
	require.True(t, ok)
	require.Equal(t, 0, total.Cmp(object.NewInteger(3)))
}
