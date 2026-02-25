package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTypesSuite(t *testing.T) {
	suite.Run(t, new(TypesSuite))
}

type TypesSuite struct {
	suite.Suite
}

func (s *TypesSuite) SetupTest() {
	ResetTypeVarCounter()
}

// === Basic Type String Representations ===

func (s *TypesSuite) TestBoolean_String() {
	assert.Equal(s.T(), "Boolean", Boolean{}.String())
}

func (s *TypesSuite) TestNumber_String() {
	assert.Equal(s.T(), "Number", Number{}.String())
}

func (s *TypesSuite) TestString_String() {
	assert.Equal(s.T(), "String", String{}.String())
}

func (s *TypesSuite) TestSet_String() {
	t := Set{Element: Number{}}
	assert.Equal(s.T(), "Set[Number]", t.String())
}

func (s *TypesSuite) TestSet_Nested_String() {
	t := Set{Element: Set{Element: Boolean{}}}
	assert.Equal(s.T(), "Set[Set[Boolean]]", t.String())
}

func (s *TypesSuite) TestTuple_String() {
	t := Tuple{Element: Number{}}
	assert.Equal(s.T(), "Tuple[Number]", t.String())
}

func (s *TypesSuite) TestRecord_String() {
	t := Record{Fields: map[string]Type{
		"x": Number{},
		"y": Boolean{},
	}}
	// Fields are sorted alphabetically
	assert.Equal(s.T(), "Record{x: Number, y: Boolean}", t.String())
}

func (s *TypesSuite) TestRecord_Empty_String() {
	t := Record{Fields: map[string]Type{}}
	assert.Equal(s.T(), "Record{}", t.String())
}

func (s *TypesSuite) TestBag_String() {
	t := Bag{Element: String{}}
	assert.Equal(s.T(), "Bag[String]", t.String())
}

func (s *TypesSuite) TestFunction_SingleParam_String() {
	t := Function{Params: []Type{Number{}}, Return: Boolean{}}
	assert.Equal(s.T(), "Number → Boolean", t.String())
}

func (s *TypesSuite) TestFunction_MultipleParams_String() {
	t := Function{Params: []Type{Number{}, String{}}, Return: Boolean{}}
	assert.Equal(s.T(), "(Number, String) → Boolean", t.String())
}

func (s *TypesSuite) TestFunction_NoParams_String() {
	t := Function{Params: []Type{}, Return: Number{}}
	assert.Equal(s.T(), "() → Number", t.String())
}

func (s *TypesSuite) TestTypeVar_String() {
	tv := TypeVar{ID: 0, Name: "a"}
	assert.Equal(s.T(), "a", tv.String())
}

func (s *TypesSuite) TestTypeVar_NoName_String() {
	tv := TypeVar{ID: 42, Name: ""}
	assert.Equal(s.T(), "t42", tv.String())
}

func (s *TypesSuite) TestAny_String() {
	assert.Equal(s.T(), "Any", Any{}.String())
}

// === Type Equality ===

func (s *TypesSuite) TestBoolean_Equals() {
	assert.True(s.T(), Boolean{}.Equals(Boolean{}))
	assert.False(s.T(), Boolean{}.Equals(Number{}))
}

func (s *TypesSuite) TestNumber_Equals() {
	assert.True(s.T(), Number{}.Equals(Number{}))
	assert.False(s.T(), Number{}.Equals(Boolean{}))
}

func (s *TypesSuite) TestString_Equals() {
	assert.True(s.T(), String{}.Equals(String{}))
	assert.False(s.T(), String{}.Equals(Number{}))
}

func (s *TypesSuite) TestSet_Equals() {
	s1 := Set{Element: Number{}}
	s2 := Set{Element: Number{}}
	s3 := Set{Element: Boolean{}}

	assert.True(s.T(), s1.Equals(s2))
	assert.False(s.T(), s1.Equals(s3))
	assert.False(s.T(), s1.Equals(Number{}))
}

func (s *TypesSuite) TestTuple_Equals() {
	t1 := Tuple{Element: Number{}}
	t2 := Tuple{Element: Number{}}
	t3 := Tuple{Element: Boolean{}}

	assert.True(s.T(), t1.Equals(t2))
	assert.False(s.T(), t1.Equals(t3))
}

func (s *TypesSuite) TestRecord_Equals() {
	r1 := Record{Fields: map[string]Type{"x": Number{}, "y": Boolean{}}}
	r2 := Record{Fields: map[string]Type{"x": Number{}, "y": Boolean{}}}
	r3 := Record{Fields: map[string]Type{"x": Number{}}}
	r4 := Record{Fields: map[string]Type{"x": Number{}, "z": Boolean{}}}

	assert.True(s.T(), r1.Equals(r2))
	assert.False(s.T(), r1.Equals(r3))  // Different field count
	assert.False(s.T(), r1.Equals(r4))  // Different field names
}

func (s *TypesSuite) TestBag_Equals() {
	b1 := Bag{Element: Number{}}
	b2 := Bag{Element: Number{}}
	b3 := Bag{Element: String{}}

	assert.True(s.T(), b1.Equals(b2))
	assert.False(s.T(), b1.Equals(b3))
}

func (s *TypesSuite) TestFunction_Equals() {
	f1 := Function{Params: []Type{Number{}}, Return: Boolean{}}
	f2 := Function{Params: []Type{Number{}}, Return: Boolean{}}
	f3 := Function{Params: []Type{String{}}, Return: Boolean{}}
	f4 := Function{Params: []Type{Number{}}, Return: String{}}
	f5 := Function{Params: []Type{Number{}, Number{}}, Return: Boolean{}}

	assert.True(s.T(), f1.Equals(f2))
	assert.False(s.T(), f1.Equals(f3))  // Different param type
	assert.False(s.T(), f1.Equals(f4))  // Different return type
	assert.False(s.T(), f1.Equals(f5))  // Different param count
}

func (s *TypesSuite) TestTypeVar_Equals() {
	tv1 := TypeVar{ID: 0, Name: "a"}
	tv2 := TypeVar{ID: 0, Name: "a"}
	tv3 := TypeVar{ID: 1, Name: "b"}

	assert.True(s.T(), tv1.Equals(tv2))
	assert.False(s.T(), tv1.Equals(tv3))
	assert.False(s.T(), tv1.Equals(Number{}))
}

func (s *TypesSuite) TestAny_Equals() {
	assert.True(s.T(), Any{}.Equals(Any{}))
	assert.False(s.T(), Any{}.Equals(Number{}))
}

// === Free Type Variables ===

func (s *TypesSuite) TestBoolean_FreeTypeVars() {
	assert.Nil(s.T(), Boolean{}.FreeTypeVars())
}

func (s *TypesSuite) TestNumber_FreeTypeVars() {
	assert.Nil(s.T(), Number{}.FreeTypeVars())
}

func (s *TypesSuite) TestTypeVar_FreeTypeVars() {
	tv := TypeVar{ID: 42, Name: "a"}
	free := tv.FreeTypeVars()
	assert.Len(s.T(), free, 1)
	_, exists := free[42]
	assert.True(s.T(), exists)
}

func (s *TypesSuite) TestSet_FreeTypeVars() {
	tv := TypeVar{ID: 1, Name: "a"}
	t := Set{Element: tv}
	free := t.FreeTypeVars()
	assert.Len(s.T(), free, 1)
	_, exists := free[1]
	assert.True(s.T(), exists)
}

func (s *TypesSuite) TestFunction_FreeTypeVars() {
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	f := Function{Params: []Type{tv1}, Return: tv2}

	free := f.FreeTypeVars()
	assert.Len(s.T(), free, 2)
	_, exists1 := free[1]
	_, exists2 := free[2]
	assert.True(s.T(), exists1)
	assert.True(s.T(), exists2)
}

func (s *TypesSuite) TestRecord_FreeTypeVars() {
	tv := TypeVar{ID: 5, Name: "a"}
	r := Record{Fields: map[string]Type{"x": tv, "y": Number{}}}

	free := r.FreeTypeVars()
	assert.Len(s.T(), free, 1)
	_, exists := free[5]
	assert.True(s.T(), exists)
}

func (s *TypesSuite) TestScheme_FreeTypeVars() {
	// ∀a. a → b  (b is free, a is bound)
	scheme := Scheme{
		TypeVars: []int{1},
		Type: Function{
			Params: []Type{TypeVar{ID: 1, Name: "a"}},
			Return: TypeVar{ID: 2, Name: "b"},
		},
	}

	free := scheme.FreeTypeVars()
	assert.Len(s.T(), free, 1)
	_, existsA := free[1]
	_, existsB := free[2]
	assert.False(s.T(), existsA) // a is bound
	assert.True(s.T(), existsB)  // b is free
}

// === Substitution ===

func (s *TypesSuite) TestSubstitution_Apply_TypeVar() {
	tv := TypeVar{ID: 1, Name: "a"}
	subst := Substitution{1: Number{}}

	result := subst.Apply(tv)
	assert.True(s.T(), result.Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Unbound() {
	tv := TypeVar{ID: 1, Name: "a"}
	subst := Substitution{2: Number{}} // Different ID

	result := subst.Apply(tv)
	assert.True(s.T(), result.Equals(tv)) // Unchanged
}

func (s *TypesSuite) TestSubstitution_Apply_Set() {
	tv := TypeVar{ID: 1, Name: "a"}
	t := Set{Element: tv}
	subst := Substitution{1: Number{}}

	result := subst.Apply(t)
	expected := Set{Element: Number{}}
	assert.True(s.T(), result.Equals(expected))
}

func (s *TypesSuite) TestSubstitution_Apply_Function() {
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	f := Function{Params: []Type{tv1}, Return: tv2}
	subst := Substitution{1: Number{}, 2: Boolean{}}

	result := subst.Apply(f)
	expected := Function{Params: []Type{Number{}}, Return: Boolean{}}
	assert.True(s.T(), result.Equals(expected))
}

func (s *TypesSuite) TestSubstitution_Apply_Record() {
	tv := TypeVar{ID: 1, Name: "a"}
	r := Record{Fields: map[string]Type{"x": tv, "y": Number{}}}
	subst := Substitution{1: String{}}

	result := subst.Apply(r).(Record)
	assert.True(s.T(), result.Fields["x"].Equals(String{}))
	assert.True(s.T(), result.Fields["y"].Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Chained() {
	// a → b, then b → Number
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	subst := Substitution{1: tv2, 2: Number{}}

	result := subst.Apply(tv1)
	assert.True(s.T(), result.Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Nil() {
	subst := Substitution{1: Number{}}
	result := subst.Apply(nil)
	assert.Nil(s.T(), result)
}

func (s *TypesSuite) TestSubstitution_Compose() {
	// s1: a → b
	// s2: b → Number
	// Composed: a → Number, b → Number
	s1 := Substitution{1: TypeVar{ID: 2, Name: "b"}}
	s2 := Substitution{2: Number{}}

	composed := s1.Compose(s2)

	tvA := TypeVar{ID: 1, Name: "a"}
	tvB := TypeVar{ID: 2, Name: "b"}

	assert.True(s.T(), composed.Apply(tvA).Equals(Number{}))
	assert.True(s.T(), composed.Apply(tvB).Equals(Number{}))
}

// === Scheme String ===

func (s *TypesSuite) TestScheme_String_Monomorphic() {
	scheme := Monotype(Number{})
	assert.Equal(s.T(), "Number", scheme.String())
}

func (s *TypesSuite) TestScheme_String_Polymorphic() {
	// ∀a. a → a
	tv := TypeVar{ID: 1, Name: "a"}
	scheme := Scheme{
		TypeVars: []int{1},
		Type:     Function{Params: []Type{tv}, Return: tv},
	}
	assert.Equal(s.T(), "∀a. a → a", scheme.String())
}

// === NewTypeVar ===

func (s *TypesSuite) TestNewTypeVar() {
	ResetTypeVarCounter()

	tv1 := NewTypeVar("a")
	tv2 := NewTypeVar("b")
	tv3 := NewTypeVar("")

	assert.Equal(s.T(), 0, tv1.ID)
	assert.Equal(s.T(), "a", tv1.Name)
	assert.Equal(s.T(), 1, tv2.ID)
	assert.Equal(s.T(), "b", tv2.Name)
	assert.Equal(s.T(), 2, tv3.ID)
	assert.Equal(s.T(), "", tv3.Name)
}

// === Helper Functions ===

func (s *TypesSuite) TestMonotype() {
	scheme := Monotype(Boolean{})
	assert.Empty(s.T(), scheme.TypeVars)
	assert.True(s.T(), scheme.Type.Equals(Boolean{}))
}

func (s *TypesSuite) TestForAll() {
	tv := TypeVar{ID: 1, Name: "a"}
	scheme := ForAll([]int{1}, Function{Params: []Type{tv}, Return: tv})

	assert.Equal(s.T(), []int{1}, scheme.TypeVars)
	assert.IsType(s.T(), Function{}, scheme.Type)
}
