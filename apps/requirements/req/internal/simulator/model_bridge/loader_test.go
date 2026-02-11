package model_bridge

import (
	"testing"

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
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Always true.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
			{Key: "inv_1", Description: "Basic arithmetic.", Notation: model_logic.NotationTLAPlus, Specification: "1 + 1 = 2"},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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
	model := &req_model.Model{
		Key:        "test_model",
		Name:       "Test Model",
		Invariants: []model_logic.Logic{},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

	s.False(result.HasErrors())
	s.Equal(0, result.SuccessCount())
	s.Equal(0, result.Registry.Count())
}

func (s *LoaderTestSuite) TestLoadModelInvariants_ParseError() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Always true.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
			{Key: "inv_1", Description: "Invalid expression.", Notation: model_logic.NotationTLAPlus, Specification: "THIS IS NOT VALID TLA+"},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

	s.True(result.HasErrors())
	s.Equal(1, result.SuccessCount())
	s.Equal(1, result.ErrorCount())
}

// =============================================================================
// Global Functions Loading
// =============================================================================

func (s *LoaderTestSuite) TestLoadGlobalFunctions() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		GlobalFunctions: map[string]model_logic.GlobalFunction{
			"_Max": {
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: model_logic.Logic{
					Key:           "spec_max",
					Description:   "Max of two values.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: "IF x > y THEN x ELSE y",
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		GlobalFunctions: map[string]model_logic.GlobalFunction{
			"_StatusSet": {
				Name:       "_StatusSet",
				Parameters: []string{},
				Specification: model_logic.Logic{
					Key:           "spec_statuses",
					Description:   "Status set.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: `{"pending", "active"}`,
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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

	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "Orders",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "Management",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:  classKey,
								Name: "Order",
								Actions: map[identity.Key]model_state.Action{
									actionKey: {
										Key:  actionKey,
										Name: "PlaceOrder",
										Requires: []model_logic.Logic{
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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

	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "Orders",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "Management",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:  classKey,
								Name: "Order",
								Queries: map[identity.Key]model_state.Query{
									queryKey: {
										Key:  queryKey,
										Name: "FindPending",
										Requires: []model_logic.Logic{
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
											{Key: "guar_2", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "FALSE"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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

	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "Orders",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "Management",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:  classKey,
								Name: "Order",
								Guards: map[identity.Key]model_state.Guard{
									guardKey: {
										Key:     guardKey,
										Name:    "CanShip",
										Details: "Order can be shipped",
										TlaGuard: []string{
											"TRUE",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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

	model := &req_model.Model{
		Key:  "shop_model",
		Name: "Shop Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Always true.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
		},
		GlobalFunctions: map[string]model_logic.GlobalFunction{
			"_Threshold": {
				Name:       "_Threshold",
				Parameters: nil,
				Specification: model_logic.Logic{
					Key:           "spec_threshold",
					Description:   "Threshold value.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: "10",
				},
			},
		},
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "Shop",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "Inventory",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:  classKey,
								Name: "Product",
								Actions: map[identity.Key]model_state.Action{
									actionKey: {
										Key:  actionKey,
										Name: "Restock",
										Requires: []model_logic.Logic{
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	loader := NewLoader()
	result := loader.LoadFromModel(model)

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
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Always true.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
		},
	}

	loader := NewLoader()
	result, err := loader.LoadFromModelStrict(model)

	s.NoError(err)
	s.False(result.HasErrors())
	s.Equal(1, result.SuccessCount())
}

func (s *LoaderTestSuite) TestLoadFromModelStrict_Error() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Invalid syntax.", Notation: model_logic.NotationTLAPlus, Specification: "INVALID SYNTAX HERE"},
		},
	}

	loader := NewLoader()
	result, err := loader.LoadFromModelStrict(model)

	s.Error(err)
	s.True(result.HasErrors())
}

// =============================================================================
// MustLoadFromModel
// =============================================================================

func (s *LoaderTestSuite) TestMustLoadFromModel_Success() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Always true.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
		},
	}

	loader := NewLoader()

	s.NotPanics(func() {
		result := loader.MustLoadFromModel(model)
		s.Equal(1, result.SuccessCount())
	})
}

func (s *LoaderTestSuite) TestMustLoadFromModel_Panics() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Invalid syntax.", Notation: model_logic.NotationTLAPlus, Specification: "INVALID SYNTAX HERE"},
		},
	}

	loader := NewLoader()

	s.Panics(func() {
		loader.MustLoadFromModel(model)
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
