package model_bridge

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/suite"
)

type ExtractorTestSuite struct {
	suite.Suite
}

func TestExtractorSuite(t *testing.T) {
	suite.Run(t, new(ExtractorTestSuite))
}

// =============================================================================
// Model Invariants
// =============================================================================

func (s *ExtractorTestSuite) TestExtractModelInvariants() {
	model := &req_model.Model{
		Key:  "test_model",
		Name: "Test Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Item quantities positive.", Notation: model_logic.NotationTLAPlus, Specification: "∀ x ∈ Items : x.quantity > 0"},
			{Key: "inv_1", Description: "Order count limit.", Notation: model_logic.NotationTLAPlus, Specification: "Cardinality(Orders) < 1000"},
		},
	}

	expressions := ExtractFromModel(model)

	s.Len(expressions, 2)

	s.Equal(SourceModelInvariant, expressions[0].Source)
	s.Equal("∀ x ∈ Items : x.quantity > 0", expressions[0].Expression)
	s.Nil(expressions[0].ScopeKey)
	s.Equal("Item quantities positive.", expressions[0].Name)
	s.Equal(0, expressions[0].Index)

	s.Equal(SourceModelInvariant, expressions[1].Source)
	s.Equal("Cardinality(Orders) < 1000", expressions[1].Expression)
	s.Nil(expressions[1].ScopeKey)
	s.Equal("Order count limit.", expressions[1].Name)
	s.Equal(1, expressions[1].Index)
}

func (s *ExtractorTestSuite) TestExtractModelInvariants_Empty() {
	model := &req_model.Model{
		Key:           "test_model",
		Name:          "Test Model",
		Invariants: []model_logic.Logic{},
	}

	expressions := ExtractFromModel(model)

	s.Len(expressions, 0)
}

// =============================================================================
// Global Functions
// =============================================================================

func (s *ExtractorTestSuite) TestExtractGlobalFunctions() {
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
			"_ValidStatuses": {
				Name:       "_ValidStatuses",
				Parameters: []string{},
				Specification: model_logic.Logic{
					Key:           "spec_statuses",
					Description:   "Valid status set.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: `{"pending", "active", "complete"}`,
				},
			},
		},
	}

	expressions := ExtractFromModel(model)

	s.Len(expressions, 2)

	// Find the _Max definition (map iteration order is not guaranteed)
	var maxExpr, statusExpr *ExtractedExpression
	for i := range expressions {
		if expressions[i].Name == "_Max" {
			maxExpr = &expressions[i]
		} else if expressions[i].Name == "_ValidStatuses" {
			statusExpr = &expressions[i]
		}
	}

	s.Require().NotNil(maxExpr)
	s.Equal(SourceTlaDefinition, maxExpr.Source)
	s.Equal("IF x > y THEN x ELSE y", maxExpr.Expression)
	s.Nil(maxExpr.ScopeKey)
	s.Equal("_Max", maxExpr.Name)
	s.Equal([]string{"x", "y"}, maxExpr.Parameters)
	s.Equal(0, maxExpr.Index)

	s.Require().NotNil(statusExpr)
	s.Equal(SourceTlaDefinition, statusExpr.Source)
	s.Equal(`{"pending", "active", "complete"}`, statusExpr.Expression)
	s.Nil(statusExpr.ScopeKey)
	s.Equal("_ValidStatuses", statusExpr.Name)
	s.Equal([]string{}, statusExpr.Parameters)
}

// =============================================================================
// Action Expressions
// =============================================================================

func (s *ExtractorTestSuite) TestExtractActionExpressions() {
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
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.status = \"pending\""},
											{Key: "req_2", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.items # {}"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self'.status = \"placed\""},
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

	expressions := ExtractFromModel(model)

	s.Len(expressions, 3)

	// Check requires expressions
	var requires1, requires2, guarantees *ExtractedExpression
	for i := range expressions {
		switch {
		case expressions[i].Source == SourceActionRequires && expressions[i].Index == 0:
			requires1 = &expressions[i]
		case expressions[i].Source == SourceActionRequires && expressions[i].Index == 1:
			requires2 = &expressions[i]
		case expressions[i].Source == SourceActionGuarantees:
			guarantees = &expressions[i]
		}
	}

	s.Require().NotNil(requires1)
	s.Equal(SourceActionRequires, requires1.Source)
	s.Equal(`self.status = "pending"`, requires1.Expression)
	s.NotNil(requires1.ScopeKey)
	s.Equal(actionKey, *requires1.ScopeKey)
	s.Equal("PlaceOrder", requires1.Name)
	s.Equal(0, requires1.Index)

	s.Require().NotNil(requires2)
	s.Equal(SourceActionRequires, requires2.Source)
	s.Equal("self.items # {}", requires2.Expression)
	s.Equal(1, requires2.Index)

	s.Require().NotNil(guarantees)
	s.Equal(SourceActionGuarantees, guarantees.Source)
	s.Equal(`self'.status = "placed"`, guarantees.Expression)
	s.NotNil(guarantees.ScopeKey)
	s.Equal(actionKey, *guarantees.ScopeKey)
	s.Equal("PlaceOrder", guarantees.Name)
	s.Equal(0, guarantees.Index)
}

// =============================================================================
// Query Expressions
// =============================================================================

func (s *ExtractorTestSuite) TestExtractQueryExpressions() {
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
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "user.role = \"admin\""},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "∀ o ∈ result : o.status = \"pending\""},
											{Key: "guar_2", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "result ⊆ Orders"},
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

	expressions := ExtractFromModel(model)

	s.Len(expressions, 3)

	// Check requires and guarantees expressions
	var requires, guarantees1, guarantees2 *ExtractedExpression
	for i := range expressions {
		switch {
		case expressions[i].Source == SourceQueryRequires:
			requires = &expressions[i]
		case expressions[i].Source == SourceQueryGuarantees && expressions[i].Index == 0:
			guarantees1 = &expressions[i]
		case expressions[i].Source == SourceQueryGuarantees && expressions[i].Index == 1:
			guarantees2 = &expressions[i]
		}
	}

	s.Require().NotNil(requires)
	s.Equal(SourceQueryRequires, requires.Source)
	s.Equal(`user.role = "admin"`, requires.Expression)
	s.NotNil(requires.ScopeKey)
	s.Equal(queryKey, *requires.ScopeKey)
	s.Equal("FindPending", requires.Name)
	s.Equal(0, requires.Index)

	s.Require().NotNil(guarantees1)
	s.Equal(SourceQueryGuarantees, guarantees1.Source)
	s.Equal(`∀ o ∈ result : o.status = "pending"`, guarantees1.Expression)
	s.Equal(0, guarantees1.Index)

	s.Require().NotNil(guarantees2)
	s.Equal(SourceQueryGuarantees, guarantees2.Source)
	s.Equal("result ⊆ Orders", guarantees2.Expression)
	s.Equal(1, guarantees2.Index)
}

// =============================================================================
// Guard Expressions
// =============================================================================

func (s *ExtractorTestSuite) TestExtractGuardExpressions() {
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
											"self.status = \"paid\"",
											"self.items # {}",
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

	expressions := ExtractFromModel(model)

	s.Len(expressions, 2)

	s.Equal(SourceGuardCondition, expressions[0].Source)
	s.Equal(`self.status = "paid"`, expressions[0].Expression)
	s.NotNil(expressions[0].ScopeKey)
	s.Equal(guardKey, *expressions[0].ScopeKey)
	s.Equal("CanShip", expressions[0].Name)
	s.Equal(0, expressions[0].Index)

	s.Equal(SourceGuardCondition, expressions[1].Source)
	s.Equal("self.items # {}", expressions[1].Expression)
	s.Equal(1, expressions[1].Index)
}

// =============================================================================
// Combined Extraction
// =============================================================================

func (s *ExtractorTestSuite) TestExtractFromModel_Combined() {
	domainKey, err := identity.NewDomainKey("shop")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "inventory")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, "product")
	s.Require().NoError(err)
	actionKey, err := identity.NewActionKey(classKey, "restock")
	s.Require().NoError(err)
	queryKey, err := identity.NewQueryKey(classKey, "find_low_stock")
	s.Require().NoError(err)
	guardKey, err := identity.NewGuardKey(classKey, "low_stock")
	s.Require().NoError(err)

	model := &req_model.Model{
		Key:  "shop_model",
		Name: "Shop Model",
		Invariants: []model_logic.Logic{
			{Key: "inv_0", Description: "Stock non-negative.", Notation: model_logic.NotationTLAPlus, Specification: "∀ p ∈ Products : p.stock >= 0"},
		},
		GlobalFunctions: map[string]model_logic.GlobalFunction{
			"_LowStockThreshold": {
				Name:       "_LowStockThreshold",
				Parameters: nil,
				Specification: model_logic.Logic{
					Key:           "spec_threshold",
					Description:   "Low stock threshold.",
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
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "quantity > 0"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self'.stock = self.stock + quantity"},
										},
									},
								},
								Queries: map[identity.Key]model_state.Query{
									queryKey: {
										Key:  queryKey,
										Name: "FindLowStock",
										Requires: []model_logic.Logic{
											{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
										},
										Guarantees: []model_logic.Logic{
											{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "∀ p ∈ result : p.stock < _LowStockThreshold"},
										},
									},
								},
								Guards: map[identity.Key]model_state.Guard{
									guardKey: {
										Key:      guardKey,
										Name:     "LowStock",
										Details:  "Stock is low",
										TlaGuard: []string{"self.stock < _LowStockThreshold"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expressions := ExtractFromModel(model)

	// Count by source type
	counts := make(map[ExpressionSource]int)
	for _, expr := range expressions {
		counts[expr.Source]++
	}

	s.Equal(1, counts[SourceModelInvariant])
	s.Equal(1, counts[SourceTlaDefinition])
	s.Equal(1, counts[SourceActionRequires])
	s.Equal(1, counts[SourceActionGuarantees])
	s.Equal(1, counts[SourceQueryRequires])
	s.Equal(1, counts[SourceQueryGuarantees])
	s.Equal(1, counts[SourceGuardCondition])
	s.Len(expressions, 7)
}

// =============================================================================
// ExpressionSource String
// =============================================================================

func (s *ExtractorTestSuite) TestExpressionSource_String() {
	s.Equal("model_invariant", SourceModelInvariant.String())
	s.Equal("tla_definition", SourceTlaDefinition.String())
	s.Equal("action_requires", SourceActionRequires.String())
	s.Equal("action_guarantees", SourceActionGuarantees.String())
	s.Equal("query_requires", SourceQueryRequires.String())
	s.Equal("query_guarantees", SourceQueryGuarantees.String())
	s.Equal("guard_condition", SourceGuardCondition.String())
	s.Equal("unknown", ExpressionSource(99).String())
}
