package model_bridge

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
	"github.com/stretchr/testify/suite"
)

type LoaderTestSuite struct {
	suite.Suite
}

func TestLoaderSuite(t *testing.T) {
	suite.Run(t, new(LoaderTestSuite))
}

// =============================================================================
// Model Invariants Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadModelInvariants() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	invKey1, err := identity.NewInvariantKey("1")
	s.Require().NoError(err)

	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Always true.", model_logic.NotationTLAPlus, "TRUE"))
	inv1 := helper.Must(model_logic.NewLogic(invKey1, "Basic arithmetic.", model_logic.NotationTLAPlus, "1 + 1 = 2"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0, inv1}, nil))

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(2, result.SuccessCount())
	s.Equal(0, result.ErrorCount())

	// Check that definitions are registered at global scope
	def0, ok := result.Registry.GetGlobal("Invariant0")
	s.True(ok)
	s.NotNil(def0)
	s.Equal(registry.KindGlobalFunction, def0.Kind)

	def1, ok := result.Registry.GetGlobal("Invariant1")
	s.True(ok)
	s.NotNil(def1)
}

func (s *LoaderTestSuite) TestLoadModelInvariants_Empty() {
	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{}, nil))

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(0, result.SuccessCount())
	s.Equal(0, result.Registry.Count())
}

func (s *LoaderTestSuite) TestLoadModelInvariants_ParseError() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	invKey1, err := identity.NewInvariantKey("1")
	s.Require().NoError(err)

	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Always true.", model_logic.NotationTLAPlus, "TRUE"))
	inv1 := helper.Must(model_logic.NewLogic(invKey1, "Invalid expression.", model_logic.NotationTLAPlus, "THIS IS NOT VALID TLA+"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0, inv1}, nil))

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.True(result.HasErrors())
	s.Equal(1, result.SuccessCount())
	s.Equal(1, result.ErrorCount())
}

// =============================================================================
// Global Functions Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadGlobalFunctions() {
	gfuncKey, err := identity.NewGlobalFunctionKey("_Max")
	s.Require().NoError(err)

	gfuncLogic := helper.Must(model_logic.NewLogic(gfuncKey, "Max of two values.", model_logic.NotationTLAPlus, "IF x > y THEN x ELSE y"))
	gfunc := helper.Must(model_logic.NewGlobalFunction(gfuncKey, "_Max", []string{"x", "y"}, gfuncLogic))

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		gfuncKey: gfunc,
	}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, globalFunctions))

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(1, result.SuccessCount())

	// Check that definition is registered at global scope with correct name
	def, ok := result.Registry.GetGlobal("_Max")
	s.True(ok)
	s.NotNil(def)
	s.Equal(registry.KindGlobalFunction, def.Kind)
	s.Len(def.Parameters, 2)
	s.Equal("x", def.Parameters[0].Name)
	s.Equal("y", def.Parameters[1].Name)
}

func (s *LoaderTestSuite) TestLoadGlobalFunctions_NoParams() {
	gfuncKey, err := identity.NewGlobalFunctionKey("_StatusSet")
	s.Require().NoError(err)

	gfuncLogic := helper.Must(model_logic.NewLogic(gfuncKey, "Status set.", model_logic.NotationTLAPlus, `{"pending", "active"}`))
	gfunc := helper.Must(model_logic.NewGlobalFunction(gfuncKey, "_StatusSet", []string{}, gfuncLogic))

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		gfuncKey: gfunc,
	}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, globalFunctions))

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())

	def, ok := result.Registry.GetGlobal("_StatusSet")
	s.True(ok)
	s.NotNil(def)
	s.Len(def.Parameters, 0)
}

// =============================================================================
// Action Expressions Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadActionExpressions() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "place_order")
	s.Require().NoError(err)

	// Build action requires logic
	actionReqKey, err := identity.NewActionRequireKey(actionKey, "0")
	s.Require().NoError(err)
	actionReq := helper.Must(model_logic.NewLogic(actionReqKey, "Precondition.", model_logic.NotationTLAPlus, "TRUE"))

	// Build action guarantees logic
	actionGuarKey, err := identity.NewActionGuaranteeKey(actionKey, "0")
	s.Require().NoError(err)
	actionGuar := helper.Must(model_logic.NewLogic(actionGuarKey, "Postcondition.", model_logic.NotationTLAPlus, "TRUE"))

	// Build the action
	action := helper.Must(model_state.NewAction(actionKey, "PlaceOrder", "", []model_logic.Logic{actionReq}, []model_logic.Logic{actionGuar}, nil, nil))

	// Build class, subdomain, domain using constructors then set children
	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Actions = map[identity.Key]model_state.Action{actionKey: action}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(2, result.SuccessCount())

	// Check that definitions are registered at class scope
	requiresKey := registry.DefinitionKey("orders!management!order!PlaceOrder_Requires0")
	def, ok := result.Registry.Get(requiresKey)
	s.True(ok)
	s.NotNil(def)
	s.Equal(registry.KindClassFunction, def.Kind)
	s.Equal("PlaceOrder_Requires0", def.LocalName)

	guaranteesKey := registry.DefinitionKey("orders!management!order!PlaceOrder_Guarantees0")
	def, ok = result.Registry.Get(guaranteesKey)
	s.True(ok)
	s.NotNil(def)
	s.Equal("PlaceOrder_Guarantees0", def.LocalName)
}

// =============================================================================
// Query Expressions Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadQueryExpressions() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	queryKey, err := identity.NewQueryKey(classKey, "find_pending")
	s.Require().NoError(err)

	// Build query requires logic
	queryReqKey, err := identity.NewQueryRequireKey(queryKey, "0")
	s.Require().NoError(err)
	queryReq := helper.Must(model_logic.NewLogic(queryReqKey, "Precondition.", model_logic.NotationTLAPlus, "TRUE"))

	// Build query guarantees logic
	queryGuarKey0, err := identity.NewQueryGuaranteeKey(queryKey, "0")
	s.Require().NoError(err)
	queryGuar0 := helper.Must(model_logic.NewLogic(queryGuarKey0, "Postcondition.", model_logic.NotationTLAPlus, "TRUE"))

	queryGuarKey1, err := identity.NewQueryGuaranteeKey(queryKey, "1")
	s.Require().NoError(err)
	queryGuar1 := helper.Must(model_logic.NewLogic(queryGuarKey1, "Postcondition.", model_logic.NotationTLAPlus, "FALSE"))

	// Build the query
	query := helper.Must(model_state.NewQuery(queryKey, "FindPending", "", []model_logic.Logic{queryReq}, []model_logic.Logic{queryGuar0, queryGuar1}, nil))

	// Build class, subdomain, domain using constructors then set children
	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Queries = map[identity.Key]model_state.Query{queryKey: query}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(3, result.SuccessCount())

	// Check query definitions
	requiresKey := registry.DefinitionKey("orders!management!order!FindPending_Requires0")
	def, ok := result.Registry.Get(requiresKey)
	s.True(ok)
	s.NotNil(def)

	guarantees0Key := registry.DefinitionKey("orders!management!order!FindPending_Guarantees0")
	def, ok = result.Registry.Get(guarantees0Key)
	s.True(ok)
	s.NotNil(def)

	guarantees1Key := registry.DefinitionKey("orders!management!order!FindPending_Guarantees1")
	def, ok = result.Registry.Get(guarantees1Key)
	s.True(ok)
	s.NotNil(def)
}

// =============================================================================
// Guard Expressions Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadGuardExpressions() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	guardKey, err := identity.NewGuardKey(classKey, "can_ship")
	s.Require().NoError(err)

	// Build guard logic using the guard key itself
	guardLogic := helper.Must(model_logic.NewLogic(guardKey, "Order can be shipped", model_logic.NotationTLAPlus, "TRUE"))

	// Build the guard
	guard := helper.Must(model_state.NewGuard(guardKey, "CanShip", guardLogic))

	// Build class, subdomain, domain using constructors then set children
	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Guards = map[identity.Key]model_state.Guard{guardKey: guard}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(1, result.SuccessCount())

	// Check guard definition
	guardDefKey := registry.DefinitionKey("orders!management!order!CanShip_Guard0")
	def, ok := result.Registry.Get(guardDefKey)
	s.True(ok)
	s.NotNil(def)
	s.Equal("CanShip_Guard0", def.LocalName)
}

// =============================================================================
// Combined Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadCombined() {
	domainKey, err := identity.NewDomainKey("shop")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "inventory")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "product")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "restock")
	s.Require().NoError(err)

	// Build invariant
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Always true.", model_logic.NotationTLAPlus, "TRUE"))

	// Build global function
	gfuncKey, err := identity.NewGlobalFunctionKey("_Threshold")
	s.Require().NoError(err)
	gfuncLogic := helper.Must(model_logic.NewLogic(gfuncKey, "Threshold value.", model_logic.NotationTLAPlus, "10"))
	gfunc := helper.Must(model_logic.NewGlobalFunction(gfuncKey, "_Threshold", nil, gfuncLogic))

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		gfuncKey: gfunc,
	}

	// Build action requires logic
	actionReqKey, err := identity.NewActionRequireKey(actionKey, "0")
	s.Require().NoError(err)
	actionReq := helper.Must(model_logic.NewLogic(actionReqKey, "Precondition.", model_logic.NotationTLAPlus, "TRUE"))

	// Build action guarantees logic
	actionGuarKey, err := identity.NewActionGuaranteeKey(actionKey, "0")
	s.Require().NoError(err)
	actionGuar := helper.Must(model_logic.NewLogic(actionGuarKey, "Postcondition.", model_logic.NotationTLAPlus, "TRUE"))

	// Build the action
	action := helper.Must(model_state.NewAction(actionKey, "Restock", "", []model_logic.Logic{actionReq}, []model_logic.Logic{actionGuar}, nil, nil))

	// Build class, subdomain, domain using constructors then set children
	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Actions = map[identity.Key]model_state.Action{actionKey: action}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Inventory", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Shop", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("shop_model", "Shop Model", "", []model_logic.Logic{inv0}, globalFunctions))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	loader := NewLoader()
	result := loader.LoadFromModel(&model)

	s.False(result.HasErrors())
	s.Equal(4, result.SuccessCount())

	// Check definitions by source
	bySource := result.DefinitionsBySource()
	s.Len(bySource[SourceModelInvariant], 1)
	s.Len(bySource[SourceTlaDefinition], 1)
	s.Len(bySource[SourceActionRequires], 1)
	s.Len(bySource[SourceActionGuarantees], 1)
}

// =============================================================================
// LoadFromExpressions
// =============================================================================

func (s *LoaderTestSuite) TestLoadFromExpressions() {
	expressions := []ExtractedExpression{
		{
			Source:     SourceModelInvariant,
			Expression: "TRUE",
			ScopeKey:   nil,
			Name:       "",
			Index:      0,
		},
		{
			Source:     SourceModelInvariant,
			Expression: "FALSE",
			ScopeKey:   nil,
			Name:       "",
			Index:      1,
		},
	}

	loader := NewLoader()
	result := loader.LoadFromExpressions(expressions)

	s.False(result.HasErrors())
	s.Equal(2, result.SuccessCount())
	s.Equal(2, result.Registry.Count())
}

// =============================================================================
// LoadIntoRegistry
// =============================================================================

func (s *LoaderTestSuite) TestLoadIntoRegistry() {
	// Create registry with existing definition
	reg := registry.NewRegistry()
	_, err := reg.RegisterGlobalFunction("Existing", nil, nil)
	s.Require().NoError(err)
	s.Equal(1, reg.Count())

	expressions := []ExtractedExpression{
		{
			Source:     SourceModelInvariant,
			Expression: "TRUE",
			ScopeKey:   nil,
			Name:       "",
			Index:      0,
		},
	}

	loader := NewLoader()
	result := loader.LoadIntoRegistry(expressions, reg)

	s.False(result.HasErrors())
	s.Equal(1, result.SuccessCount())
	s.Equal(2, reg.Count()) // Existing + new
	s.Same(reg, result.Registry)
}

// =============================================================================
// LoadFromModelStrict
// =============================================================================

func (s *LoaderTestSuite) TestLoadFromModelStrict_Success() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Always true.", model_logic.NotationTLAPlus, "TRUE"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0}, nil))

	loader := NewLoader()
	result, err := loader.LoadFromModelStrict(&model)

	s.NoError(err)
	s.False(result.HasErrors())
	s.Equal(1, result.SuccessCount())
}

func (s *LoaderTestSuite) TestLoadFromModelStrict_Error() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Invalid syntax.", model_logic.NotationTLAPlus, "INVALID SYNTAX HERE"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0}, nil))

	loader := NewLoader()
	result, err := loader.LoadFromModelStrict(&model)

	s.Error(err)
	s.True(result.HasErrors())
}

// =============================================================================
// MustLoadFromModel
// =============================================================================

func (s *LoaderTestSuite) TestMustLoadFromModel_Success() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Always true.", model_logic.NotationTLAPlus, "TRUE"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0}, nil))

	loader := NewLoader()

	s.NotPanics(func() {
		result := loader.MustLoadFromModel(&model)
		s.Equal(1, result.SuccessCount())
	})
}

func (s *LoaderTestSuite) TestMustLoadFromModel_Panics() {
	invKey0, err := identity.NewInvariantKey("0")
	s.Require().NoError(err)
	inv0 := helper.Must(model_logic.NewLogic(invKey0, "Invalid syntax.", model_logic.NotationTLAPlus, "INVALID SYNTAX HERE"))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0}, nil))

	loader := NewLoader()

	s.Panics(func() {
		loader.MustLoadFromModel(&model)
	})
}

// =============================================================================
// Definition Builder Tests
// =============================================================================

func (s *LoaderTestSuite) TestDefinitionBuilder_BuildResult() {
	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	expr := ExtractedExpression{
		Source:     SourceModelInvariant,
		Expression: "TRUE",
		ScopeKey:   nil,
		Name:       "",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.True(result.IsSuccess())
	s.Nil(result.Error)
	s.NotNil(result.Definition)
	s.Equal(expr, result.Source)
}

func (s *LoaderTestSuite) TestDefinitionBuilder_BuildResult_Error() {
	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	expr := ExtractedExpression{
		Source:     SourceModelInvariant,
		Expression: "INVALID SYNTAX",
		ScopeKey:   nil,
		Name:       "",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.False(result.IsSuccess())
	s.NotNil(result.Error)
	s.Nil(result.Definition)
}

func (s *LoaderTestSuite) TestDefinitionBuilder_UnsupportedSource() {
	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	expr := ExtractedExpression{
		Source:     ExpressionSource(99), // Invalid source
		Expression: "TRUE",
		ScopeKey:   nil,
		Name:       "",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.False(result.IsSuccess())
	s.NotNil(result.Error)
	s.Contains(result.Error.Error(), "unsupported expression source")
}

// =============================================================================
// Guarantee Classification Tests
// =============================================================================

func (s *LoaderTestSuite) TestGuaranteeClassification_PrimedAssignment() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "update_status")
	s.Require().NoError(err)

	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	// Test primed assignment: self.status' = "shipped"
	expr := ExtractedExpression{
		Source:     SourceActionGuarantees,
		Expression: `self.status' = "shipped"`,
		ScopeKey:   &actionKey,
		Name:       "UpdateStatus",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.True(result.IsSuccess())
	s.Equal(GuaranteePrimedAssignment, result.GuaranteeKind)
}

func (s *LoaderTestSuite) TestGuaranteeClassification_PostCondition() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "update_count")
	s.Require().NoError(err)

	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	// Test post-condition: count' > count (a check that must be TRUE)
	expr := ExtractedExpression{
		Source:     SourceActionGuarantees,
		Expression: `count' > count`,
		ScopeKey:   &actionKey,
		Name:       "UpdateCount",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.True(result.IsSuccess())
	s.Equal(GuaranteePostCondition, result.GuaranteeKind)
}

func (s *LoaderTestSuite) TestGuaranteeClassification_SimpleTRUE() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "noop")
	s.Require().NoError(err)

	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	// Test simple TRUE (a trivial post-condition)
	expr := ExtractedExpression{
		Source:     SourceActionGuarantees,
		Expression: `TRUE`,
		ScopeKey:   &actionKey,
		Name:       "Noop",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.True(result.IsSuccess())
	s.Equal(GuaranteePostCondition, result.GuaranteeKind)
}

func (s *LoaderTestSuite) TestGuaranteeClassification_QueryResultPrimed() {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	queryKey, err := identity.NewQueryKey(classKey, "find_all")
	s.Require().NoError(err)

	builder := NewDefinitionBuilder()
	reg := registry.NewRegistry()

	// Test query result assignment: result' = Orders
	expr := ExtractedExpression{
		Source:     SourceQueryGuarantees,
		Expression: `result' = Orders`,
		ScopeKey:   &queryKey,
		Name:       "FindAll",
		Index:      0,
	}

	result := builder.Build(expr, reg)

	s.True(result.IsSuccess())
	s.Equal(GuaranteePrimedAssignment, result.GuaranteeKind)
}

func (s *LoaderTestSuite) TestGuaranteeKind_String() {
	s.Equal("primed_assignment", GuaranteePrimedAssignment.String())
	s.Equal("post_condition", GuaranteePostCondition.String())
	s.Equal("unknown", GuaranteeUnknown.String())
}
