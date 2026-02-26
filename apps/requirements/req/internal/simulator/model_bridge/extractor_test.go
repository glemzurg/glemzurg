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
	"github.com/stretchr/testify/suite"
)

type ExtractorTestSuite struct {
	suite.Suite
}

func TestExtractorSuite(t *testing.T) {
	suite.Run(t, new(ExtractorTestSuite))
}

func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

// =============================================================================
// Model Invariants
// =============================================================================

func (s *ExtractorTestSuite) TestExtractModelInvariants() {
	invKey0 := helper.Must(identity.NewInvariantKey("0"))
	inv0 := helper.Must(model_logic.NewLogic(invKey0, model_logic.LogicTypeAssessment, "Item quantities positive.", "", model_logic.NotationTLAPlus, "∀ x ∈ Items : x.quantity > 0", nil))

	invKey1 := helper.Must(identity.NewInvariantKey("1"))
	inv1 := helper.Must(model_logic.NewLogic(invKey1, model_logic.LogicTypeAssessment, "Order count limit.", "", model_logic.NotationTLAPlus, "Cardinality(Orders) < 1000", nil))

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{inv0, inv1}, nil))

	expressions := ExtractFromModel(&model)

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
	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", []model_logic.Logic{}, nil))

	expressions := ExtractFromModel(&model)

	s.Len(expressions, 0)
}

// =============================================================================
// Global Functions
// =============================================================================

func (s *ExtractorTestSuite) TestExtractGlobalFunctions() {
	gfuncMaxKey := helper.Must(identity.NewGlobalFunctionKey("_Max"))
	gfuncMaxLogic := helper.Must(model_logic.NewLogic(gfuncMaxKey, model_logic.LogicTypeValue, "Max of two values.", "", model_logic.NotationTLAPlus, "IF x > y THEN x ELSE y", nil))
	gfuncMax := helper.Must(model_logic.NewGlobalFunction(gfuncMaxKey, "_Max", []string{"x", "y"}, gfuncMaxLogic))

	gfuncStatusKey := helper.Must(identity.NewGlobalFunctionKey("_ValidStatuses"))
	gfuncStatusLogic := helper.Must(model_logic.NewLogic(gfuncStatusKey, model_logic.LogicTypeValue, "Valid status set.", "", model_logic.NotationTLAPlus, `{"pending", "active", "complete"}`, nil))
	gfuncStatus := helper.Must(model_logic.NewGlobalFunction(gfuncStatusKey, "_ValidStatuses", []string{}, gfuncStatusLogic))

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		gfuncMaxKey:    gfuncMax,
		gfuncStatusKey: gfuncStatus,
	}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, globalFunctions))

	expressions := ExtractFromModel(&model)

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
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "management"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "place_order"))

	actionReqKey0 := helper.Must(identity.NewActionRequireKey(actionKey, "0"))
	actionReq0 := helper.Must(model_logic.NewLogic(actionReqKey0, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, `self.status = "pending"`, nil))

	actionReqKey1 := helper.Must(identity.NewActionRequireKey(actionKey, "1"))
	actionReq1 := helper.Must(model_logic.NewLogic(actionReqKey1, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.items # {}", nil))

	actionGuarKey0 := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	actionGuar0 := helper.Must(model_logic.NewLogic(actionGuarKey0, model_logic.LogicTypeStateChange, "Postcondition.", "status", model_logic.NotationTLAPlus, `self'.status = "placed"`, nil))

	action := helper.Must(model_state.NewAction(actionKey, "PlaceOrder", "", []model_logic.Logic{actionReq0, actionReq1}, []model_logic.Logic{actionGuar0}, nil, nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Actions = map[identity.Key]model_state.Action{actionKey: action}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	expressions := ExtractFromModel(&model)

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
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "management"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "find_pending"))

	queryReqKey0 := helper.Must(identity.NewQueryRequireKey(queryKey, "0"))
	queryReq0 := helper.Must(model_logic.NewLogic(queryReqKey0, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, `user.role = "admin"`, nil))

	queryGuarKey0 := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	queryGuar0 := helper.Must(model_logic.NewLogic(queryGuarKey0, model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, `∀ o ∈ result : o.status = "pending"`, nil))

	queryGuarKey1 := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "1"))
	queryGuar1 := helper.Must(model_logic.NewLogic(queryGuarKey1, model_logic.LogicTypeQuery, "Postcondition.", "subset", model_logic.NotationTLAPlus, "result ⊆ Orders", nil))

	query := helper.Must(model_state.NewQuery(queryKey, "FindPending", "", []model_logic.Logic{queryReq0}, []model_logic.Logic{queryGuar0, queryGuar1}, nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Queries = map[identity.Key]model_state.Query{queryKey: query}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	expressions := ExtractFromModel(&model)

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
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "management"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "can_ship"))

	guardLogic := helper.Must(model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Order can be shipped", "", model_logic.NotationTLAPlus, `self.status = "paid" /\ self.items # {}`, nil))

	guard := helper.Must(model_state.NewGuard(guardKey, "CanShip", guardLogic))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Guards = map[identity.Key]model_state.Guard{guardKey: guard}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Management", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := helper.Must(req_model.NewModel("test_model", "Test Model", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	expressions := ExtractFromModel(&model)

	s.Len(expressions, 1)

	s.Equal(SourceGuardCondition, expressions[0].Source)
	s.Equal("self.status = \"paid\" /\\ self.items # {}", expressions[0].Expression)
	s.NotNil(expressions[0].ScopeKey)
	s.Equal(guardKey, *expressions[0].ScopeKey)
	s.Equal("CanShip", expressions[0].Name)
	s.Equal(0, expressions[0].Index)
}

// =============================================================================
// Combined Extraction
// =============================================================================

func (s *ExtractorTestSuite) TestExtractFromModel_Combined() {
	domainKey := helper.Must(identity.NewDomainKey("shop"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "inventory"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "product"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "restock"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "find_low_stock"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "low_stock"))

	// Model invariant
	invKey0 := helper.Must(identity.NewInvariantKey("0"))
	inv0 := helper.Must(model_logic.NewLogic(invKey0, model_logic.LogicTypeAssessment, "Stock non-negative.", "", model_logic.NotationTLAPlus, "∀ p ∈ Products : p.stock >= 0", nil))

	// Global function
	gfuncKey := helper.Must(identity.NewGlobalFunctionKey("_LowStockThreshold"))
	gfuncLogic := helper.Must(model_logic.NewLogic(gfuncKey, model_logic.LogicTypeValue, "Low stock threshold.", "", model_logic.NotationTLAPlus, "10", nil))
	gfunc := helper.Must(model_logic.NewGlobalFunction(gfuncKey, "_LowStockThreshold", nil, gfuncLogic))

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		gfuncKey: gfunc,
	}

	// Action
	actionReqKey0 := helper.Must(identity.NewActionRequireKey(actionKey, "0"))
	actionReq0 := helper.Must(model_logic.NewLogic(actionReqKey0, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "quantity > 0", nil))

	actionGuarKey0 := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	actionGuar0 := helper.Must(model_logic.NewLogic(actionGuarKey0, model_logic.LogicTypeStateChange, "Postcondition.", "stock", model_logic.NotationTLAPlus, "self'.stock = self.stock + quantity", nil))

	action := helper.Must(model_state.NewAction(actionKey, "Restock", "", []model_logic.Logic{actionReq0}, []model_logic.Logic{actionGuar0}, nil, nil))

	// Query
	queryReqKey0 := helper.Must(identity.NewQueryRequireKey(queryKey, "0"))
	queryReq0 := helper.Must(model_logic.NewLogic(queryReqKey0, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "TRUE", nil))

	queryGuarKey0 := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	queryGuar0 := helper.Must(model_logic.NewLogic(queryGuarKey0, model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, "∀ p ∈ result : p.stock < _LowStockThreshold", nil))

	query := helper.Must(model_state.NewQuery(queryKey, "FindLowStock", "", []model_logic.Logic{queryReq0}, []model_logic.Logic{queryGuar0}, nil))

	// Guard
	guardLogic := helper.Must(model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Stock is low", "", model_logic.NotationTLAPlus, "self.stock < _LowStockThreshold", nil))
	guard := helper.Must(model_state.NewGuard(guardKey, "LowStock", guardLogic))

	// Assemble class
	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Actions = map[identity.Key]model_state.Action{actionKey: action}
	class.Queries = map[identity.Key]model_state.Query{queryKey: query}
	class.Guards = map[identity.Key]model_state.Guard{guardKey: guard}

	// Assemble subdomain
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Inventory", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	// Assemble domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Shop", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	// Assemble model
	model := helper.Must(req_model.NewModel("shop_model", "Shop Model", "", []model_logic.Logic{inv0}, globalFunctions))
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	expressions := ExtractFromModel(&model)

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
