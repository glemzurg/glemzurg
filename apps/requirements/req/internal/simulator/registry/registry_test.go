package registry

import (
	"math/big"
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/suite"
)

func big0(v int64) *big.Int {
	return big.NewInt(v)
}

// RegistryTestSuite tests the registry package.
type RegistryTestSuite struct {
	suite.Suite
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

// irInt creates a simple IR integer literal for testing.
func irInt(v int64) me.Expression {
	return &me.IntLiteral{Value: big0(v)}
}

// irBool creates a simple IR boolean literal for testing.
func irBool(v bool) me.Expression {
	return &me.BoolLiteral{Value: v}
}

// =============================================================================
// Registry Tests
// =============================================================================

func (s *RegistryTestSuite) TestNewRegistry() {
	r := NewRegistry()
	s.NotNil(r)
	s.Equal(0, r.Count())
	s.Equal(uint64(0), r.Version())
}

func (s *RegistryTestSuite) TestRegisterClassFunction() {
	r := NewRegistry()

	body := irInt(42)

	def, err := r.RegisterClassFunction(
		"DomainA", "SubdomainB", "ClassC", "Func",
		body,
		[]Parameter{
			{Name: "x", Type: types.Number{}},
		},
	)

	s.Require().NoError(err)
	s.NotNil(def)
	s.Equal(DefinitionKey("DomainA!SubdomainB!ClassC!Func"), def.Key)
	s.Equal(KindClassFunction, def.Kind)
	s.Equal(ScopePath("DomainA!SubdomainB!ClassC"), def.Scope)
	s.Equal("Func", def.LocalName)
	s.Len(def.Parameters, 1)
	s.Equal("x", def.Parameters[0].Name)
	s.Equal(uint64(1), def.Version)
	s.False(def.Validated)

	s.Equal(1, r.Count())
	s.Equal(uint64(1), r.Version())
}

func (s *RegistryTestSuite) TestRegisterGlobalFunction() {
	r := NewRegistry()

	body := irInt(100)

	def, err := r.RegisterGlobalFunction(
		"IsoCurrency",
		body,
		nil, // No parameters
	)

	s.Require().NoError(err)
	s.NotNil(def)
	s.Equal(DefinitionKey("_IsoCurrency"), def.Key)
	s.Equal(KindGlobalFunction, def.Kind)
	s.Equal(ScopePath(""), def.Scope)
	s.Equal("IsoCurrency", def.LocalName)
	s.Empty(def.Parameters)

	// Test global lookup
	retrieved, ok := r.GetGlobal("IsoCurrency")
	s.True(ok)
	s.Equal(def, retrieved)
}

func (s *RegistryTestSuite) TestRegisterDuplicateFails() {
	r := NewRegistry()
	body := irInt(1)

	_, err := r.RegisterGlobalFunction("Test", body, nil)
	s.Require().NoError(err)

	_, err = r.RegisterGlobalFunction("Test", body, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "already exists")
}

func (s *RegistryTestSuite) TestGet() {
	r := NewRegistry()
	body := irInt(1)

	_, err := r.RegisterClassFunction("A", "B", "C", "F", body, nil)
	s.Require().NoError(err)

	def, ok := r.Get("A!B!C!F")
	s.True(ok)
	s.NotNil(def)

	_, ok = r.Get("NonExistent")
	s.False(ok)
}

func (s *RegistryTestSuite) TestUpdate() {
	r := NewRegistry()
	body1 := irInt(1)
	body2 := irInt(2)

	def, _ := r.RegisterGlobalFunction("Test", body1, nil)
	originalVersion := def.Version

	err := r.Update("_Test", body2, []Parameter{{Name: "y", Type: types.String{}}})
	s.Require().NoError(err)

	def, _ = r.Get("_Test")
	s.Equal(body2, def.Body)
	s.Len(def.Parameters, 1)
	s.Equal("y", def.Parameters[0].Name)
	s.Greater(def.Version, originalVersion)
	s.False(def.Validated)
}

func (s *RegistryTestSuite) TestDelete() {
	r := NewRegistry()
	body := irInt(1)

	_, err := r.RegisterGlobalFunction("Test", body, nil)
	s.Require().NoError(err)
	s.Equal(1, r.Count())

	err = r.Delete("_Test")
	s.Require().NoError(err)
	s.Equal(0, r.Count())

	_, ok := r.GetGlobal("Test")
	s.False(ok)
}

// =============================================================================
// ScopePath Tests
// =============================================================================

func (s *RegistryTestSuite) TestParseScopePath() {
	// Valid class scope
	path, err := ParseScopePath("A", "B", "C")
	s.Require().NoError(err)
	s.Equal(ScopePath("A!B!C"), path)

	// Valid global scope
	path, err = ParseScopePath("", "", "")
	s.Require().NoError(err)
	s.Equal(ScopePath(""), path)

	// Invalid partial scope
	_, err = ParseScopePath("A", "", "")
	s.Require().Error(err)

	_, err = ParseScopePath("A", "B", "")
	s.Require().Error(err)
}

func (s *RegistryTestSuite) TestScopePathParts() {
	path := ScopePath("A!B!C")
	parts := path.Parts()
	s.Equal([]string{"A", "B", "C"}, parts)

	s.Equal("A", path.Domain())
	s.Equal("B", path.Subdomain())
	s.Equal("C", path.Class())

	// Empty path
	emptyPath := ScopePath("")
	s.Nil(emptyPath.Parts())
	s.Empty(emptyPath.Domain())
}

// =============================================================================
// DefinitionKey Tests
// =============================================================================

func (s *RegistryTestSuite) TestDefinitionKeyIsGlobal() {
	s.True(DefinitionKey("_Global").IsGlobal())
	s.False(DefinitionKey("A!B!C!F").IsGlobal())
}

func (s *RegistryTestSuite) TestDefinitionKeyLocalName() {
	s.Equal("Global", DefinitionKey("_Global").LocalName())
	s.Equal("Func", DefinitionKey("A!B!C!Func").LocalName())
}

// =============================================================================
// ScopeContext Tests
// =============================================================================

func (s *RegistryTestSuite) TestNewScopeContext() {
	r := NewRegistry()

	// Global scope
	ctx := NewGlobalScopeContext(r)
	s.Equal(ScopeLevelGlobal, ctx.Level)
	s.Empty(ctx.Domain)

	// Domain scope
	ctx = NewDomainScopeContext(r, "DomainA")
	s.Equal(ScopeLevelDomain, ctx.Level)
	s.Equal("DomainA", ctx.Domain)

	// Subdomain scope
	ctx = NewSubdomainScopeContext(r, "DomainA", "SubB")
	s.Equal(ScopeLevelSubdomain, ctx.Level)
	s.Equal("DomainA", ctx.Domain)
	s.Equal("SubB", ctx.Subdomain)

	// Class scope
	ctx = NewClassScopeContext(r, "DomainA", "SubB", "ClassC")
	s.Equal(ScopeLevelClass, ctx.Level)
	s.Equal("DomainA", ctx.Domain)
	s.Equal("SubB", ctx.Subdomain)
	s.Equal("ClassC", ctx.Class)
}

// =============================================================================
// Dependency Tests
// =============================================================================

func (s *RegistryTestSuite) TestAddDependency() {
	r := NewRegistry()
	body := irInt(1)

	_, err := r.RegisterGlobalFunction("A", body, nil)
	s.Require().NoError(err)
	_, err = r.RegisterGlobalFunction("B", body, nil)
	s.Require().NoError(err)

	r.AddDependency("_A", "_B") // A depends on B

	deps := r.GetDependencies("_A")
	s.Contains(deps, DefinitionKey("_B"))

	dependents := r.GetDependents("_B")
	s.Contains(dependents, DefinitionKey("_A"))
}

func (s *RegistryTestSuite) TestFindTransitiveDependents() {
	r := NewRegistry()
	body := irInt(1)

	// A -> B -> C (A depends on B, B depends on C)
	_, err := r.RegisterGlobalFunction("A", body, nil)
	s.Require().NoError(err)
	_, err = r.RegisterGlobalFunction("B", body, nil)
	s.Require().NoError(err)
	_, err = r.RegisterGlobalFunction("C", body, nil)
	s.Require().NoError(err)

	r.AddDependency("_A", "_B")
	r.AddDependency("_B", "_C")

	// Dependents of C should include both A and B
	dependents := r.FindTransitiveDependents("_C")
	s.Len(dependents, 2)
	s.Contains(dependents, DefinitionKey("_A"))
	s.Contains(dependents, DefinitionKey("_B"))
}

func (s *RegistryTestSuite) TestInvalidateDefinition() {
	r := NewRegistry()
	body := irInt(1)

	_, err := r.RegisterGlobalFunction("A", body, nil)
	s.Require().NoError(err)
	_, err = r.RegisterGlobalFunction("B", body, nil)
	s.Require().NoError(err)

	r.AddDependency("_A", "_B")

	// Invalidate B - should also invalidate A
	invalidated := r.InvalidateDefinition("_B")

	s.Contains(invalidated.Keys, DefinitionKey("_B"))
	s.Contains(invalidated.Keys, DefinitionKey("_A"))

	// Both should need validation
	defA, _ := r.Get("_A")
	defB, _ := r.Get("_B")
	s.False(defA.Validated)
	s.False(defB.Validated)
}

// =============================================================================
// InvalidationSet Tests
// =============================================================================

func (s *RegistryTestSuite) TestInvalidationSet() {
	set := NewInvalidationSet()
	s.NotNil(set)
	s.Empty(set.Keys)

	set.Add("_A", 1)
	s.True(set.Contains("_A"))
	s.False(set.Contains("_B"))

	set.Add("_A", 2) // Duplicate - should not add
	s.Len(set.Keys, 1)
}

func (s *RegistryTestSuite) TestInvalidationSetMerge() {
	set1 := NewInvalidationSet()
	set1.Add("_A", 1)

	set2 := NewInvalidationSet()
	set2.Add("_B", 2)

	set1.Merge(set2)

	s.True(set1.Contains("_A"))
	s.True(set1.Contains("_B"))
}
