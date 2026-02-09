package typechecker

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/assert"
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
	assert.NotNil(s.T(), env)
	assert.Empty(s.T(), env.Names())
}

func (s *EnvSuite) TestBind_And_Lookup() {
	env := NewEnv()
	env.Bind("x", types.Monotype(types.Number{}))

	scheme, ok := env.Lookup("x")
	assert.True(s.T(), ok)
	assert.True(s.T(), scheme.Type.Equals(types.Number{}))
}

func (s *EnvSuite) TestBind_NotFound() {
	env := NewEnv()
	_, ok := env.Lookup("x")
	assert.False(s.T(), ok)
}

func (s *EnvSuite) TestBindMono() {
	env := NewEnv()
	env.BindMono("x", types.Boolean{})

	scheme, ok := env.Lookup("x")
	assert.True(s.T(), ok)
	assert.Empty(s.T(), scheme.TypeVars) // Monomorphic
	assert.True(s.T(), scheme.Type.Equals(types.Boolean{}))
}

func (s *EnvSuite) TestExtend() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("y", types.Boolean{})

	// Child can see both
	_, okX := child.Lookup("x")
	_, okY := child.Lookup("y")
	assert.True(s.T(), okX)
	assert.True(s.T(), okY)

	// Parent can only see x
	_, okX = parent.Lookup("x")
	_, okY = parent.Lookup("y")
	assert.True(s.T(), okX)
	assert.False(s.T(), okY)
}

func (s *EnvSuite) TestExtend_Shadowing() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("x", types.Boolean{})

	// Child sees Boolean
	schemeChild, _ := child.Lookup("x")
	assert.True(s.T(), schemeChild.Type.Equals(types.Boolean{}))

	// Parent still sees Number
	schemeParent, _ := parent.Lookup("x")
	assert.True(s.T(), schemeParent.Type.Equals(types.Number{}))
}

func (s *EnvSuite) TestContains() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	assert.True(s.T(), env.Contains("x"))
	assert.False(s.T(), env.Contains("y"))
}

func (s *EnvSuite) TestFreeTypeVars_Empty() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	free := env.FreeTypeVars()
	assert.Empty(s.T(), free)
}

func (s *EnvSuite) TestFreeTypeVars_WithTypeVar() {
	env := NewEnv()
	tv := types.TypeVar{ID: 42, Name: "a"}
	env.BindMono("x", tv)

	free := env.FreeTypeVars()
	_, exists := free[42]
	assert.True(s.T(), exists)
}

func (s *EnvSuite) TestFreeTypeVars_InParent() {
	parent := NewEnv()
	tv := types.TypeVar{ID: 10, Name: "a"}
	parent.BindMono("x", tv)

	child := parent.Extend()
	child.BindMono("y", types.Number{})

	free := child.FreeTypeVars()
	_, exists := free[10]
	assert.True(s.T(), exists)
}

func (s *EnvSuite) TestApply() {
	env := NewEnv()
	tv := types.TypeVar{ID: 1, Name: "a"}
	env.BindMono("x", tv)

	subst := types.Substitution{1: types.Number{}}
	newEnv := env.Apply(subst)

	scheme, _ := newEnv.Lookup("x")
	assert.True(s.T(), scheme.Type.Equals(types.Number{}))

	// Original unchanged
	origScheme, _ := env.Lookup("x")
	assert.True(s.T(), origScheme.Type.Equals(tv))
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
	assert.True(s.T(), schemeX.Type.Equals(types.Number{}))
	assert.True(s.T(), schemeY.Type.Equals(types.Boolean{}))
}

func (s *EnvSuite) TestClone() {
	env := NewEnv()
	env.BindMono("x", types.Number{})

	clone := env.Clone()
	clone.BindMono("y", types.Boolean{})

	// Clone has both
	assert.True(s.T(), clone.Contains("x"))
	assert.True(s.T(), clone.Contains("y"))

	// Original only has x
	assert.True(s.T(), env.Contains("x"))
	assert.False(s.T(), env.Contains("y"))
}

func (s *EnvSuite) TestNames() {
	env := NewEnv()
	env.BindMono("x", types.Number{})
	env.BindMono("y", types.Boolean{})

	names := env.Names()
	assert.Len(s.T(), names, 2)
	assert.Contains(s.T(), names, "x")
	assert.Contains(s.T(), names, "y")
}

func (s *EnvSuite) TestAllNames() {
	parent := NewEnv()
	parent.BindMono("x", types.Number{})

	child := parent.Extend()
	child.BindMono("y", types.Boolean{})

	// Names() only returns local bindings
	assert.Len(s.T(), child.Names(), 1)

	// AllNames() includes parent
	allNames := child.AllNames()
	assert.Len(s.T(), allNames, 2)
	assert.Contains(s.T(), allNames, "x")
	assert.Contains(s.T(), allNames, "y")
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
	assert.Equal(s.T(), 1, count)
}
