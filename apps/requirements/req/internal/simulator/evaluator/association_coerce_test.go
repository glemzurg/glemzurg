package evaluator_test

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AssociationCoerceSuite struct {
	suite.Suite
}

func TestAssociationCoerceSuite(t *testing.T) {
	suite.Run(t, new(AssociationCoerceSuite))
}

func (s *AssociationCoerceSuite) newAR(endpoints ...*object.Record) *object.AssociationRelation {
	set := object.NewSet()
	links := make(map[*object.Record]*object.Record)
	for i, ep := range endpoints {
		set.Add(ep)
		link := object.NewRecord()
		link.Set("Fee", object.NewInteger(int64(i+1)))
		links[ep] = link
	}
	return object.NewAssociationRelation(set, "LinkDef", links)
}

func (s *AssociationCoerceSuite) TestSetToBagAcceptsAssociationRelation() {
	tests := []struct {
		name     string
		value    object.Object
		wantSize int
		wantErr  string
	}{
		{
			name:     "empty AR",
			value:    s.newAR(),
			wantSize: 0,
		},
		{
			name:     "one endpoint",
			value:    s.newAR(object.NewRecord()),
			wantSize: 1,
		},
		{
			name: "two endpoints",
			value: s.newAR(
				func() *object.Record { r := object.NewRecord(); r.Set("Code", object.NewString("US")); return r }(),
				func() *object.Record { r := object.NewRecord(); r.Set("Code", object.NewString("UK")); return r }(),
			),
			wantSize: 2,
		},
		{
			name:     "plain Set unchanged",
			value:    object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)}),
			wantSize: 2,
		},
		{
			name:    "rejects Number",
			value:   object.NewInteger(3),
			wantErr: "AssociationRelation",
		},
		{
			name:    "rejects Bag",
			value:   object.NewBag(),
			wantErr: "AssociationRelation",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			call := &me.BuiltinCall{
				Module:   "_Bags",
				Function: "SetToBag",
				Args:     []me.Expression{&me.LocalVar{Name: "v"}},
			}
			bindings := evaluator.NewBindings()
			bindings.Set("v", tc.value, evaluator.NamespaceLocal)
			result := evaluator.Eval(call, bindings)
			if tc.wantErr != "" {
				s.Require().True(result.IsError())
				s.Contains(result.Error.Inspect(), tc.wantErr)
				return
			}
			s.Require().False(result.IsError(), result.Error)
			bag, ok := result.Value.(*object.Bag)
			s.Require().True(ok)
			s.Equal(tc.wantSize, bag.Size())
		})
	}
}

func (s *AssociationCoerceSuite) TestBagCardinalityUsesCoerceToSet() {
	ep1 := object.NewRecord()
	ep1.Set("Code", object.NewString("US"))
	ep2 := object.NewRecord()
	ep2.Set("Code", object.NewString("UK"))
	ar := s.newAR(ep1, ep2)

	tests := []struct {
		name    string
		value   object.Object
		want    int64
		wantErr string
	}{
		{
			name:  "Bag",
			value: func() *object.Bag { b := object.NewBag(); b.Add(object.NewInteger(1), 2); return b }(),
			want:  2,
		},
		{
			name:  "Set",
			value: object.NewSetFromElements([]object.Object{object.NewInteger(1), object.NewInteger(2)}),
			want:  2,
		},
		{
			name:  "AssociationRelation",
			value: ar,
			want:  2,
		},
		{
			name:    "rejects Number",
			value:   object.NewInteger(9),
			wantErr: "AssociationRelation",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			call := &me.BuiltinCall{
				Module:   "_Bags",
				Function: "BagCardinality",
				Args:     []me.Expression{&me.LocalVar{Name: "v"}},
			}
			bindings := evaluator.NewBindings()
			bindings.Set("v", tc.value, evaluator.NamespaceLocal)
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

	// Composition: BagCardinality(SetToBag(AR)) equals endpoint size.
	composed := &me.BuiltinCall{
		Module:   "_Bags",
		Function: "BagCardinality",
		Args: []me.Expression{
			&me.BuiltinCall{
				Module:   "_Bags",
				Function: "SetToBag",
				Args:     []me.Expression{&me.LocalVar{Name: "v"}},
			},
		},
	}
	bindings := evaluator.NewBindings()
	bindings.Set("v", ar, evaluator.NamespaceLocal)
	result := evaluator.Eval(composed, bindings)
	s.Require().False(result.IsError(), result.Error)
	n, ok := result.Value.(*object.Number)
	s.Require().True(ok)
	s.Equal(0, n.Cmp(object.NewInteger(2)))
}

func (s *AssociationCoerceSuite) TestAssociationEqualsTruthTable() {
	ep1 := object.NewRecord()
	ep1.Set("Code", object.NewString("US"))
	ep2 := object.NewRecord()
	ep2.Set("Code", object.NewString("UK"))

	emptyAR := object.NewAssociationRelation(object.NewSet(), "LinkDef", nil)
	arTwo := s.newAR(ep1, ep2)
	endpointSet := object.NewSetFromElements([]object.Object{ep1, ep2})

	// Same endpoints, different link rows → AR↔AR full equality is false if links differ.
	linkA := object.NewRecord()
	linkA.Set("Fee", object.NewInteger(1))
	linkB := object.NewRecord()
	linkB.Set("Fee", object.NewInteger(99))
	arSameEndpointsDiffLinks := object.NewAssociationRelation(
		object.NewSetFromElements([]object.Object{ep1}),
		"LinkDef",
		map[*object.Record]*object.Record{ep1: linkA},
	)
	arSameEndpointsOtherLinks := object.NewAssociationRelation(
		object.NewSetFromElements([]object.Object{ep1}),
		"LinkDef",
		map[*object.Record]*object.Record{ep1: linkB},
	)

	tests := []struct {
		name  string
		left  object.Object
		right object.Object
		want  bool
	}{
		{name: "empty AR equals empty Set", left: emptyAR, right: object.NewSet(), want: true},
		{name: "AR equals endpoint Set", left: arTwo, right: endpointSet, want: true},
		{name: "endpoint Set equals AR", left: endpointSet, right: arTwo, want: true},
		{name: "AR equals self", left: arTwo, right: arTwo, want: true},
		{name: "AR vs AR different links same endpoints", left: arSameEndpointsDiffLinks, right: arSameEndpointsOtherLinks, want: false},
		{name: "AR vs Number", left: arTwo, right: object.NewInteger(0), want: false},
		{name: "AR vs empty Bag", left: emptyAR, right: object.NewBag(), want: false},
		{name: "AR inequality different endpoints", left: arTwo, right: s.newAR(ep1), want: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, evaluator.ObjectsEqual(tc.left, tc.right))
		})
	}
}

func (s *AssociationCoerceSuite) TestProjectSetOfRecordsField() {
	r1 := object.NewRecord()
	r1.Set("amount", object.NewInteger(10))
	r2 := object.NewRecord()
	r2.Set("amount", object.NewInteger(20))
	set := object.NewSetFromElements([]object.Object{r1, r2})

	access := &me.FieldAccess{
		Base:  &me.LocalVar{Name: "rows"},
		Field: "amount",
	}
	bindings := evaluator.NewBindings()
	bindings.Set("rows", set, evaluator.NamespaceLocal)
	result := evaluator.Eval(access, bindings)
	s.Require().False(result.IsError(), result.Error)
	out, ok := result.Value.(*object.Set)
	s.Require().True(ok)
	s.Equal(2, out.Size())
	s.True(out.Contains(object.NewInteger(10)))
	s.True(out.Contains(object.NewInteger(20)))

	// Empty set → empty set.
	bindings.Set("rows", object.NewSet(), evaluator.NamespaceLocal)
	emptyResult := evaluator.Eval(access, bindings)
	s.Require().False(emptyResult.IsError(), emptyResult.Error)
	emptyOut, ok := emptyResult.Value.(*object.Set)
	s.Require().True(ok)
	s.Equal(0, emptyOut.Size())

	// Missing field → error.
	bad := object.NewRecord()
	bad.Set("other", object.NewInteger(1))
	bindings.Set("rows", object.NewSetFromElements([]object.Object{bad}), evaluator.NamespaceLocal)
	missing := evaluator.Eval(access, bindings)
	s.Require().True(missing.IsError())
	s.Contains(missing.Error.Inspect(), "amount")
}

func TestSetToBagAssociationRelationSmoke(t *testing.T) {
	// Keep a package-level smoke for go test -run filters.
	require.NotNil(t, object.NewAssociationRelation(object.NewSet(), "LinkDef", nil))
}
