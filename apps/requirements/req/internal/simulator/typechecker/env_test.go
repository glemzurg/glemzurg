package typechecker

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/suite"
)

func TestEnvSuite(t *testing.T) {
	suite.Run(t, new(EnvSuite))
}

type EnvSuite struct {
	suite.Suite
}

func (s *EnvSuite) SetupTest() {
	types.ResetTypeVarCounter()
}

func (s *EnvSuite) TestNewEnv() {
	env := NewEnv()
	s.NotNil(env)
	s.Empty(env.Names())
}

func (s *EnvSuite) TestBind_And_Lookup() {
	env := NewEnv()
	env.Bind("x", types.Monotype(types.Number{}))

	scheme, ok := env.Lookup("x")
	s.True(ok)
	s.True(scheme.Type.Equals(types.Number{}))
}

func (s *EnvSuite) TestBind_NotFound() {
	env := NewEnv()
	_, ok := env.Lookup("x")
	s.False(ok)
}

func (s *EnvSuite) TestBindMono() {
	env := NewEnv()
	env.BindMono("x", types.Boolean{})

	scheme, ok := env.Lookup("x")
	s.True(ok)
	s.Empty(scheme.TypeVars) // Monomorphic
	s.True(scheme.Type.Equals(types.Boolean{}))
}

func (s *EnvSuite) TestExtend() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("y", types.Boolean{})

	// Child can see both
	_, okX := child.Lookup("x")
	_, okY := child.Lookup("y")
	s.True(okX)
	s.True(okY)

	// Parent can only see x
	_, okX = parent.Lookup("x")
	_, okY = parent.Lookup("y")
	s.True(okX)
	s.False(okY)
}

func (s *EnvSuite) TestExtend_Shadowing() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("x", types.Boolean{})

	// Child sees Boolean
	schemeChild, _ := child.Lookup("x")
	s.True(schemeChild.Type.Equals(types.Boolean{}))

	// Parent still sees Number
	schemeParent, _ := parent.Lookup("x")
	s.True(schemeParent.Type.Equals(types.Number{}))
}

func (s *EnvSuite) TestContains() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	s.True(env.Contains("x"))
	s.False(env.Contains("y"))
}

func (s *EnvSuite) TestFreeTypeVars_Empty() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	free := env.FreeTypeVars()
	s.Empty(free)
}

func (s *EnvSuite) TestFreeTypeVars_WithTypeVar() {
	env := NewEnv()
	tv := types.TypeVar{ID: 42, Name: "a"}
	env.BindMono("x", tv)

	free := env.FreeTypeVars()
	_, exists := free[42]
	s.True(exists)
}

func (s *EnvSuite) TestFreeTypeVars_InParent() {
	parent := NewEnv()
	tv := types.TypeVar{ID: 10, Name: "a"}
	parent.BindMono("x", tv)

	child := parent.Extend()
	child.BindMono("y", types.Number{})

	free := child.FreeTypeVars()
	_, exists := free[10]
	s.True(exists)
}

func (s *EnvSuite) TestApply() {
	env := NewEnv()
	tv := types.TypeVar{ID: 1, Name: "a"}
	env.BindMono("x", tv)

	subst := types.Substitution{1: types.Number{}}
	newEnv := env.Apply(subst)

	scheme, _ := newEnv.Lookup("x")
	s.True(scheme.Type.Equals(types.Number{}))

	// Original unchanged
	origScheme, _ := env.Lookup("x")
	s.True(origScheme.Type.Equals(tv))
}

func (s *EnvSuite) TestApply_WithParent() {
	parent := NewEnv()
	tv1 := types.TypeVar{ID: 1, Name: "a"}
	parent.BindMono("x", tv1)

	child := parent.Extend()
	tv2 := types.TypeVar{ID: 2, Name: "b"}
	child.BindMono("y", tv2)

	subst := types.Substitution{1: types.Number{}, 2: types.Boolean{}}
	newEnv := child.Apply(subst)

	schemeX, _ := newEnv.Lookup("x")
	schemeY, _ := newEnv.Lookup("y")
	s.True(schemeX.Type.Equals(types.Number{}))
	s.True(schemeY.Type.Equals(types.Boolean{}))
}

func (s *EnvSuite) TestClone() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	clone := env.Clone()
	clone.BindMono("y", types.Boolean{})

	// Clone has both
	s.True(clone.Contains("x"))
	s.True(clone.Contains("y"))

	// Original only has x
	s.True(env.Contains("x"))
	s.False(env.Contains("y"))
}

func (s *EnvSuite) TestNames() {
	env := NewEnv()
	env.BindMono("x", types.Number{})
	env.BindMono("y", types.Boolean{})

	names := env.Names()
	s.Len(names, 2)
	s.Contains(names, "x")
	s.Contains(names, "y")
}

func (s *EnvSuite) TestAllNames() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("y", types.Boolean{})

	// Names() only returns local bindings
	s.Len(child.Names(), 1)

	// AllNames() includes parent
	allNames := child.AllNames()
	s.Len(allNames, 2)
	s.Contains(allNames, "x")
	s.Contains(allNames, "y")
}

func (s *EnvSuite) TestAllNames_Shadowing() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("x", types.Boolean{}) // Shadow

	// Only one "x" in allNames
	allNames := child.AllNames()
	count := 0
	for _, name := range allNames {
		if name == "x" {
			count++
		}
	}
	s.Equal(1, count)
}
