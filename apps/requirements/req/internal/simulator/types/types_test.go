package types

import (
	"testing"

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
	s.Equal("Boolean", Boolean{}.String())
}

func (s *TypesSuite) TestNumber_String() {
	s.Equal("Number", Number{}.String())
}

func (s *TypesSuite) TestString_String() {
	s.Equal("String", String{}.String())
}

func (s *TypesSuite) TestSet_String() {
	t := Set{Element: Number{}}
	s.Equal("Set[Number]", t.String())
}

func (s *TypesSuite) TestSet_Nested_String() {
	t := Set{Element: Set{Element: Boolean{}}}
	s.Equal("Set[Set[Boolean]]", t.String())
}

func (s *TypesSuite) TestTuple_String() {
	t := Tuple{Element: Number{}}
	s.Equal("Tuple[Number]", t.String())
}

func (s *TypesSuite) TestRecord_String() {
	t := Record{Fields: map[string]Type{
		"x": Number{},
		"y": Boolean{},
	}}
	// Fields are sorted alphabetically
	s.Equal("Record{x: Number, y: Boolean}", t.String())
}

func (s *TypesSuite) TestRecord_Empty_String() {
	t := Record{Fields: map[string]Type{}}
	s.Equal("Record{}", t.String())
}

func (s *TypesSuite) TestBag_String() {
	t := Bag{Element: String{}}
	s.Equal("Bag[String]", t.String())
}

func (s *TypesSuite) TestFunction_SingleParam_String() {
	t := Function{Params: []Type{Number{}}, Return: Boolean{}}
	s.Equal("Number → Boolean", t.String())
}

func (s *TypesSuite) TestFunction_MultipleParams_String() {
	t := Function{Params: []Type{Number{}, String{}}, Return: Boolean{}}
	s.Equal("(Number, String) → Boolean", t.String())
}

func (s *TypesSuite) TestFunction_NoParams_String() {
	t := Function{Params: []Type{}, Return: Number{}}
	s.Equal("() → Number", t.String())
}

func (s *TypesSuite) TestTypeVar_String() {
	tv := TypeVar{ID: 0, Name: "a"}
	s.Equal("a", tv.String())
}

func (s *TypesSuite) TestTypeVar_NoName_String() {
	tv := TypeVar{ID: 42, Name: ""}
	s.Equal("t42", tv.String())
}

func (s *TypesSuite) TestAny_String() {
	s.Equal("Any", Any{}.String())
}

// === Type Equality ===

func (s *TypesSuite) TestBoolean_Equals() {
	s.True(Boolean{}.Equals(Boolean{}))
	s.False(Boolean{}.Equals(Number{}))
}

func (s *TypesSuite) TestNumber_Equals() {
	s.True(Number{}.Equals(Number{}))
	s.False(Number{}.Equals(Boolean{}))
}

func (s *TypesSuite) TestString_Equals() {
	s.True(String{}.Equals(String{}))
	s.False(String{}.Equals(Number{}))
}

func (s *TypesSuite) TestSet_Equals() {
	s1 := Set{Element: Number{}}
	s2 := Set{Element: Number{}}
	s3 := Set{Element: Boolean{}}

	s.True(s1.Equals(s2))
	s.False(s1.Equals(s3))
	s.False(s1.Equals(Number{}))
}

func (s *TypesSuite) TestTuple_Equals() {
	t1 := Tuple{Element: Number{}}
	t2 := Tuple{Element: Number{}}
	t3 := Tuple{Element: Boolean{}}

	s.True(t1.Equals(t2))
	s.False(t1.Equals(t3))
}

func (s *TypesSuite) TestRecord_Equals() {
	r1 := Record{Fields: map[string]Type{"x": Number{}, "y": Boolean{}}}
	r2 := Record{Fields: map[string]Type{"x": Number{}, "y": Boolean{}}}
	r3 := Record{Fields: map[string]Type{"x": Number{}}}
	r4 := Record{Fields: map[string]Type{"x": Number{}, "z": Boolean{}}}

	s.True(r1.Equals(r2))
	s.False(r1.Equals(r3)) // Different field count
	s.False(r1.Equals(r4)) // Different field names
}

func (s *TypesSuite) TestBag_Equals() {
	b1 := Bag{Element: Number{}}
	b2 := Bag{Element: Number{}}
	b3 := Bag{Element: String{}}

	s.True(b1.Equals(b2))
	s.False(b1.Equals(b3))
}

func (s *TypesSuite) TestFunction_Equals() {
	f1 := Function{Params: []Type{Number{}}, Return: Boolean{}}
	f2 := Function{Params: []Type{Number{}}, Return: Boolean{}}
	f3 := Function{Params: []Type{String{}}, Return: Boolean{}}
	f4 := Function{Params: []Type{Number{}}, Return: String{}}
	f5 := Function{Params: []Type{Number{}, Number{}}, Return: Boolean{}}

	s.True(f1.Equals(f2))
	s.False(f1.Equals(f3)) // Different param type
	s.False(f1.Equals(f4)) // Different return type
	s.False(f1.Equals(f5)) // Different param count
}

func (s *TypesSuite) TestTypeVar_Equals() {
	tv1 := TypeVar{ID: 0, Name: "a"}
	tv2 := TypeVar{ID: 0, Name: "a"}
	tv3 := TypeVar{ID: 1, Name: "b"}

	s.True(tv1.Equals(tv2))
	s.False(tv1.Equals(tv3))
	s.False(tv1.Equals(Number{}))
}

func (s *TypesSuite) TestAny_Equals() {
	s.True(Any{}.Equals(Any{}))
	s.False(Any{}.Equals(Number{}))
}

// === Free Type Variables ===

func (s *TypesSuite) TestBoolean_FreeTypeVars() {
	s.Nil(Boolean{}.FreeTypeVars())
}

func (s *TypesSuite) TestNumber_FreeTypeVars() {
	s.Nil(Number{}.FreeTypeVars())
}

func (s *TypesSuite) TestTypeVar_FreeTypeVars() {
	tv := TypeVar{ID: 42, Name: "a"}
	free := tv.FreeTypeVars()
	s.Len(free, 1)
	_, exists := free[42]
	s.True(exists)
}

func (s *TypesSuite) TestSet_FreeTypeVars() {
	tv := TypeVar{ID: 1, Name: "a"}
	t := Set{Element: tv}
	free := t.FreeTypeVars()
	s.Len(free, 1)
	_, exists := free[1]
	s.True(exists)
}

func (s *TypesSuite) TestFunction_FreeTypeVars() {
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	f := Function{Params: []Type{tv1}, Return: tv2}

	free := f.FreeTypeVars()
	s.Len(free, 2)
	_, exists1 := free[1]
	_, exists2 := free[2]
	s.True(exists1)
	s.True(exists2)
}

func (s *TypesSuite) TestRecord_FreeTypeVars() {
	tv := TypeVar{ID: 5, Name: "a"}
	r := Record{Fields: map[string]Type{"x": tv, "y": Number{}}}

	free := r.FreeTypeVars()
	s.Len(free, 1)
	_, exists := free[5]
	s.True(exists)
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
	s.Len(free, 1)
	_, existsA := free[1]
	_, existsB := free[2]
	s.False(existsA) // a is bound
	s.True(existsB)  // b is free
}

// === Substitution ===

func (s *TypesSuite) TestSubstitution_Apply_TypeVar() {
	tv := TypeVar{ID: 1, Name: "a"}
	subst := Substitution{1: Number{}}

	result := subst.Apply(tv)
	s.True(result.Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Unbound() {
	tv := TypeVar{ID: 1, Name: "a"}
	subst := Substitution{2: Number{}} // Different ID

	result := subst.Apply(tv)
	s.True(result.Equals(tv)) // Unchanged
}

func (s *TypesSuite) TestSubstitution_Apply_Set() {
	tv := TypeVar{ID: 1, Name: "a"}
	t := Set{Element: tv}
	subst := Substitution{1: Number{}}

	result := subst.Apply(t)
	expected := Set{Element: Number{}}
	s.True(result.Equals(expected))
}

func (s *TypesSuite) TestSubstitution_Apply_Function() {
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	f := Function{Params: []Type{tv1}, Return: tv2}
	subst := Substitution{1: Number{}, 2: Boolean{}}

	result := subst.Apply(f)
	expected := Function{Params: []Type{Number{}}, Return: Boolean{}}
	s.True(result.Equals(expected))
}

func (s *TypesSuite) TestSubstitution_Apply_Record() {
	tv := TypeVar{ID: 1, Name: "a"}
	r := Record{Fields: map[string]Type{"x": tv, "y": Number{}}}
	subst := Substitution{1: String{}}

	result := subst.Apply(r).(Record)
	s.True(result.Fields["x"].Equals(String{}))
	s.True(result.Fields["y"].Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Chained() {
	// a → b, then b → Number
	tv1 := TypeVar{ID: 1, Name: "a"}
	tv2 := TypeVar{ID: 2, Name: "b"}
	subst := Substitution{1: tv2, 2: Number{}}

	result := subst.Apply(tv1)
	s.True(result.Equals(Number{}))
}

func (s *TypesSuite) TestSubstitution_Apply_Nil() {
	subst := Substitution{1: Number{}}
	result := subst.Apply(nil)
	s.Nil(result)
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

	s.True(composed.Apply(tvA).Equals(Number{}))
	s.True(composed.Apply(tvB).Equals(Number{}))
}

// === Scheme String ===

func (s *TypesSuite) TestScheme_String_Monomorphic() {
	scheme := Monotype(Number{})
	s.Equal("Number", scheme.String())
}

func (s *TypesSuite) TestScheme_String_Polymorphic() {
	// ∀a. a → a
	tv := TypeVar{ID: 1, Name: "a"}
	scheme := Scheme{
		TypeVars: []int{1},
		Type:     Function{Params: []Type{tv}, Return: tv},
	}
	s.Equal("∀a. a → a", scheme.String())
}

// === NewTypeVar ===

func (s *TypesSuite) TestNewTypeVar() {
	ResetTypeVarCounter()

	tv1 := NewTypeVar("a")
	tv2 := NewTypeVar("b")
	tv3 := NewTypeVar("")

	s.Equal(0, tv1.ID)
	s.Equal("a", tv1.Name)
	s.Equal(1, tv2.ID)
	s.Equal("b", tv2.Name)
	s.Equal(2, tv3.ID)
	s.Empty(tv3.Name)
}

// === Helper Functions ===

func (s *TypesSuite) TestMonotype() {
	scheme := Monotype(Boolean{})
	s.Empty(scheme.TypeVars)
	s.True(scheme.Type.Equals(Boolean{}))
}

func (s *TypesSuite) TestForAll() {
	tv := TypeVar{ID: 1, Name: "a"}
	scheme := ForAll([]int{1}, Function{Params: []Type{tv}, Return: tv})

	s.Equal([]int{1}, scheme.TypeVars)
	s.IsType(Function{}, scheme.Type)
}
