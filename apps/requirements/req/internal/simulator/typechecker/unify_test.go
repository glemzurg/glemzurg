package typechecker

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/suite"
)

func TestUnifySuite(t *testing.T) {
	suite.Run(t, new(UnifySuite))
}

type UnifySuite struct {
	suite.Suite
}

func (s *UnifySuite) SetupTest() {
	types.ResetTypeVarCounter()
}

// === Basic Unification ===

func (s *UnifySuite) TestUnify_SameType() {
	subst, err := Unify(types.Number{}, types.Number{})
	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_DifferentPrimitives() {
	_, err := Unify(types.Number{}, types.Boolean{})
	s.Error(err)
	var target *UnificationError
	s.ErrorAs(err, &target)
}

func (s *UnifySuite) TestUnify_TypeVarWithPrimitive() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	subst, err := Unify(tv, types.Number{})

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

func (s *UnifySuite) TestUnify_PrimitiveWithTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	subst, err := Unify(types.Number{}, tv)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

func (s *UnifySuite) TestUnify_TwoTypeVars() {
	tv1 := types.TypeVar{ID: 1, Name: "a"}
	tv2 := types.TypeVar{ID: 2, Name: "b"}
	subst, err := Unify(tv1, tv2)

	s.NoError(err)
	// One should be mapped to the other
	s.Len(subst, 1)
}

func (s *UnifySuite) TestUnify_SameTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	subst, err := Unify(tv, tv)

	s.NoError(err)
	s.Empty(subst)
}

// === Set Unification ===

func (s *UnifySuite) TestUnify_Set_SameElement() {
	s1 := types.Set{Element: types.Number{}}
	s2 := types.Set{Element: types.Number{}}
	subst, err := Unify(s1, s2)

	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Set_DifferentElements() {
	s1 := types.Set{Element: types.Number{}}
	s2 := types.Set{Element: types.Boolean{}}
	_, err := Unify(s1, s2)

	s.Error(err)
}

func (s *UnifySuite) TestUnify_Set_WithTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	s1 := types.Set{Element: tv}
	s2 := types.Set{Element: types.Number{}}
	subst, err := Unify(s1, s2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

func (s *UnifySuite) TestUnify_Set_TypeVarWithSet() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	set := types.Set{Element: types.Number{}}
	subst, err := Unify(tv, set)

	s.NoError(err)
	s.True(subst[1].Equals(set))
}

// === Tuple Unification ===

func (s *UnifySuite) TestUnify_Tuple_SameElement() {
	t1 := types.Tuple{Element: types.String{}}
	t2 := types.Tuple{Element: types.String{}}
	subst, err := Unify(t1, t2)

	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Tuple_WithTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	t1 := types.Tuple{Element: tv}
	t2 := types.Tuple{Element: types.Number{}}
	subst, err := Unify(t1, t2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

// === Function Unification ===

func (s *UnifySuite) TestUnify_Function_Same() {
	f1 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	f2 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	subst, err := Unify(f1, f2)

	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Function_ArityMismatch() {
	f1 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	f2 := types.Function{Params: []types.Type{types.Number{}, types.Number{}}, Return: types.Boolean{}}
	_, err := Unify(f1, f2)

	s.Error(err)
	s.Contains(err.Error(), "arity mismatch")
}

func (s *UnifySuite) TestUnify_Function_DifferentParam() {
	f1 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	f2 := types.Function{Params: []types.Type{types.String{}}, Return: types.Boolean{}}
	_, err := Unify(f1, f2)

	s.Error(err)
}

func (s *UnifySuite) TestUnify_Function_DifferentReturn() {
	f1 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	f2 := types.Function{Params: []types.Type{types.Number{}}, Return: types.String{}}
	_, err := Unify(f1, f2)

	s.Error(err)
}

func (s *UnifySuite) TestUnify_Function_WithTypeVars() {
	// a → b unifies with Number → Boolean
	tvA := types.TypeVar{ID: 1, Name: "a"}
	tvB := types.TypeVar{ID: 2, Name: "b"}
	f1 := types.Function{Params: []types.Type{tvA}, Return: tvB}
	f2 := types.Function{Params: []types.Type{types.Number{}}, Return: types.Boolean{}}
	subst, err := Unify(f1, f2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
	s.True(subst[2].Equals(types.Boolean{}))
}

// === Record Unification ===

func (s *UnifySuite) TestUnify_Record_Same() {
	r1 := types.Record{Fields: map[string]types.Type{"x": types.Number{}, "y": types.Boolean{}}}
	r2 := types.Record{Fields: map[string]types.Type{"x": types.Number{}, "y": types.Boolean{}}}
	subst, err := Unify(r1, r2)

	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Record_FieldCountMismatch() {
	r1 := types.Record{Fields: map[string]types.Type{"x": types.Number{}}}
	r2 := types.Record{Fields: map[string]types.Type{"x": types.Number{}, "y": types.Boolean{}}}
	_, err := Unify(r1, r2)

	s.Error(err)
	s.Contains(err.Error(), "field count")
}

func (s *UnifySuite) TestUnify_Record_MissingField() {
	r1 := types.Record{Fields: map[string]types.Type{"x": types.Number{}}}
	r2 := types.Record{Fields: map[string]types.Type{"y": types.Number{}}}
	_, err := Unify(r1, r2)

	s.Error(err)
	s.Contains(err.Error(), "missing field")
}

func (s *UnifySuite) TestUnify_Record_WithTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	r1 := types.Record{Fields: map[string]types.Type{"x": tv}}
	r2 := types.Record{Fields: map[string]types.Type{"x": types.Number{}}}
	subst, err := Unify(r1, r2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

// === Bag Unification ===

func (s *UnifySuite) TestUnify_Bag_Same() {
	b1 := types.Bag{Element: types.Number{}}
	b2 := types.Bag{Element: types.Number{}}
	subst, err := Unify(b1, b2)

	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Bag_WithTypeVar() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	b1 := types.Bag{Element: tv}
	b2 := types.Bag{Element: types.String{}}
	subst, err := Unify(b1, b2)

	s.NoError(err)
	s.True(subst[1].Equals(types.String{}))
}

// === Any Type ===

func (s *UnifySuite) TestUnify_Any_WithPrimitive() {
	subst, err := Unify(types.Any{}, types.Number{})
	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Primitive_WithAny() {
	subst, err := Unify(types.Number{}, types.Any{})
	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnify_Any_WithComplex() {
	subst, err := Unify(types.Any{}, types.Set{Element: types.Number{}})
	s.NoError(err)
	s.Empty(subst)
}

// === Occurs Check ===

func (s *UnifySuite) TestUnify_OccursCheck_Simple() {
	// a = Set[a] is an infinite type
	tv := types.TypeVar{ID: 1, Name: "a"}
	set := types.Set{Element: tv}
	_, err := Unify(tv, set)

	s.Error(err)
	s.Contains(err.Error(), "infinite type")
}

func (s *UnifySuite) TestUnify_OccursCheck_Nested() {
	// a = Tuple[Set[a]] is infinite
	tv := types.TypeVar{ID: 1, Name: "a"}
	nested := types.Tuple{Element: types.Set{Element: tv}}
	_, err := Unify(tv, nested)

	s.Error(err)
	s.Contains(err.Error(), "infinite type")
}

// === UnifyAll ===

func (s *UnifySuite) TestUnifyAll_Empty() {
	subst, err := UnifyAll([][2]types.Type{})
	s.NoError(err)
	s.Empty(subst)
}

func (s *UnifySuite) TestUnifyAll_MultiplePairs() {
	tv1 := types.TypeVar{ID: 1, Name: "a"}
	tv2 := types.TypeVar{ID: 2, Name: "b"}
	pairs := [][2]types.Type{
		{tv1, types.Number{}},
		{tv2, types.Boolean{}},
	}
	subst, err := UnifyAll(pairs)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
	s.True(subst[2].Equals(types.Boolean{}))
}

func (s *UnifySuite) TestUnifyAll_Conflict() {
	tv := types.TypeVar{ID: 1, Name: "a"}
	pairs := [][2]types.Type{
		{tv, types.Number{}},
		{tv, types.Boolean{}}, // Conflict: a can't be both
	}
	_, err := UnifyAll(pairs)

	s.Error(err)
}

// === UnifyWithSubst ===

func (s *UnifySuite) TestUnifyWithSubst_ExtendsExisting() {
	tv2 := types.TypeVar{ID: 2, Name: "b"}

	existingSubst := types.Substitution{1: types.Number{}}
	newSubst, err := UnifyWithSubst(tv2, types.Boolean{}, existingSubst)

	s.NoError(err)
	s.True(newSubst[1].Equals(types.Number{}))
	s.True(newSubst[2].Equals(types.Boolean{}))

	// Original unchanged
	_, hasTwo := existingSubst[2]
	s.False(hasTwo)
}

func (s *UnifySuite) TestUnifyWithSubst_UsesExisting() {
	tv1 := types.TypeVar{ID: 1, Name: "a"}

	existingSubst := types.Substitution{1: types.Number{}}
	// tv1 is already Number, so this should succeed
	newSubst, err := UnifyWithSubst(tv1, types.Number{}, existingSubst)

	s.NoError(err)
	s.True(newSubst[1].Equals(types.Number{}))
}

func (s *UnifySuite) TestUnifyWithSubst_Conflict() {
	tv := types.TypeVar{ID: 1, Name: "a"}

	existingSubst := types.Substitution{1: types.Number{}}
	// tv1 is already Number, but we try to unify with Boolean
	_, err := UnifyWithSubst(tv, types.Boolean{}, existingSubst)

	s.Error(err)
}

// === Complex Scenarios ===

func (s *UnifySuite) TestUnify_NestedSets() {
	// Set[Set[a]] with Set[Set[Number]]
	tv := types.TypeVar{ID: 1, Name: "a"}
	s1 := types.Set{Element: types.Set{Element: tv}}
	s2 := types.Set{Element: types.Set{Element: types.Number{}}}
	subst, err := Unify(s1, s2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
}

func (s *UnifySuite) TestUnify_FunctionReturnsSet() {
	// (a → Set[b]) with (Number → Set[Boolean])
	tvA := types.TypeVar{ID: 1, Name: "a"}
	tvB := types.TypeVar{ID: 2, Name: "b"}
	f1 := types.Function{
		Params: []types.Type{tvA},
		Return: types.Set{Element: tvB},
	}
	f2 := types.Function{
		Params: []types.Type{types.Number{}},
		Return: types.Set{Element: types.Boolean{}},
	}
	subst, err := Unify(f1, f2)

	s.NoError(err)
	s.True(subst[1].Equals(types.Number{}))
	s.True(subst[2].Equals(types.Boolean{}))
}

func (s *UnifySuite) TestUnify_TransitiveTypeVars() {
	// a = b, b = Number => a = Number
	tvA := types.TypeVar{ID: 1, Name: "a"}
	tvB := types.TypeVar{ID: 2, Name: "b"}

	pairs := [][2]types.Type{
		{tvA, tvB},
		{tvB, types.Number{}},
	}
	subst, err := UnifyAll(pairs)

	s.NoError(err)
	// After applying substitution, both should resolve to Number
	resolvedA := subst.Apply(tvA)
	resolvedB := subst.Apply(tvB)
	s.True(resolvedA.Equals(types.Number{}))
	s.True(resolvedB.Equals(types.Number{}))
}
