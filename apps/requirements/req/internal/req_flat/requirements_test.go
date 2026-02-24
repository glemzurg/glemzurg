package req_flat

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
	"github.com/stretchr/testify/suite"
)

// ============================================================
// Test Keys
// ============================================================

var (
	tDomainKey    = helper.Must(identity.NewDomainKey("d"))
	tSubdomainKey = helper.Must(identity.NewSubdomainKey(tDomainKey, "s"))
	tClassKey     = helper.Must(identity.NewClassKey(tSubdomainKey, "order"))
	tClass2Key    = helper.Must(identity.NewClassKey(tSubdomainKey, "item"))

	tStateOpenKey   = helper.Must(identity.NewStateKey(tClassKey, "open"))
	tStateClosedKey = helper.Must(identity.NewStateKey(tClassKey, "closed"))
	tEventCreateKey = helper.Must(identity.NewEventKey(tClassKey, "create"))
	tEventCloseKey  = helper.Must(identity.NewEventKey(tClassKey, "close"))
	tGuardKey       = helper.Must(identity.NewGuardKey(tClassKey, "is_valid"))
	tActionKey      = helper.Must(identity.NewActionKey(tClassKey, "do_close"))
	tQueryKey       = helper.Must(identity.NewQueryKey(tClassKey, "get_total"))
	tTransCreateKey = helper.Must(identity.NewTransitionKey(tClassKey, "", "create", "", "", "open"))
	tTransCloseKey  = helper.Must(identity.NewTransitionKey(tClassKey, "open", "close", "", "do_close", "closed"))
	tAttributeKey   = helper.Must(identity.NewAttributeKey(tClassKey, "amount"))
	tSActionKey     = helper.Must(identity.NewStateActionKey(tStateOpenKey, "do", "do_close"))

	tActorKey  = helper.Must(identity.NewActorKey("user"))
	tActor2Key = helper.Must(identity.NewActorKey("admin"))
	tActor3Key = helper.Must(identity.NewActorKey("guest"))
	tActorGenKey = helper.Must(identity.NewActorGeneralizationKey("user_type"))

	tGenKey   = helper.Must(identity.NewGeneralizationKey(tSubdomainKey, "vehicle"))
	tAssocKey = helper.Must(identity.NewClassAssociationKey(tSubdomainKey, tClassKey, tClass2Key, "order_items"))

	tDomainAssocKey = helper.Must(identity.NewDomainAssociationKey(tDomainKey, tDomainKey))

	tUseCaseKey     = helper.Must(identity.NewUseCaseKey(tSubdomainKey, "place_order"))
	tUseCase2Key    = helper.Must(identity.NewUseCaseKey(tSubdomainKey, "login"))
	tUseCase3Key    = helper.Must(identity.NewUseCaseKey(tSubdomainKey, "checkout"))
	tUseCaseGenKey  = helper.Must(identity.NewUseCaseGeneralizationKey(tSubdomainKey, "order_flow"))
	tScenarioKey    = helper.Must(identity.NewScenarioKey(tUseCaseKey, "happy_path"))
	tObjectKey      = helper.Must(identity.NewScenarioObjectKey(tScenarioKey, "order1"))

	tGlobalFuncKey = helper.Must(identity.NewGlobalFunctionKey("_Max"))

	tActionGuaranteeKey = helper.Must(identity.NewActionGuaranteeKey(tActionKey, "0"))
	tQueryGuaranteeKey  = helper.Must(identity.NewQueryGuaranteeKey(tQueryKey, "0"))
	tInvariantKey       = helper.Must(identity.NewInvariantKey("0"))
)

// ============================================================
// Test Helpers
// ============================================================

// buildTestModel builds a comprehensive model with all entity types populated.
func buildTestModel() req_model.Model {
	// Logics.
	guardLogic := helper.Must(model_logic.NewLogic(tGuardKey, "Guard logic.", model_logic.NotationTLAPlus, "self.amount > 0"))
	actionGuarantee := helper.Must(model_logic.NewLogic(tActionGuaranteeKey, "Postcondition.", model_logic.NotationTLAPlus, "self.amount' = 0"))
	queryGuarantee := helper.Must(model_logic.NewLogic(tQueryGuaranteeKey, "Query result.", model_logic.NotationTLAPlus, "result = self.amount"))
	invariant := helper.Must(model_logic.NewLogic(tInvariantKey, "Always true.", model_logic.NotationTLAPlus, "TRUE"))
	gfLogic := helper.Must(model_logic.NewLogic(tGlobalFuncKey, "Max function.", model_logic.NotationTLAPlus, "IF a > b THEN a ELSE b"))

	// Global function.
	globalFunc := helper.Must(model_logic.NewGlobalFunction(tGlobalFuncKey, "_Max", []string{"a", "b"}, gfLogic))

	// State machine elements.
	eventCreate := helper.Must(model_state.NewEvent(tEventCreateKey, "create", "", nil))
	eventClose := helper.Must(model_state.NewEvent(tEventCloseKey, "close", "", nil))
	guard := helper.Must(model_state.NewGuard(tGuardKey, "is_valid", guardLogic))
	action := helper.Must(model_state.NewAction(tActionKey, "DoClose", "", nil, []model_logic.Logic{actionGuarantee}, nil, nil))
	query := helper.Must(model_state.NewQuery(tQueryKey, "GetTotal", "", nil, []model_logic.Logic{queryGuarantee}, nil))

	// Class.
	class := helper.Must(model_class.NewClass(tClassKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{
		tAttributeKey: {Key: tAttributeKey, Name: "amount"},
	}
	class.States = map[identity.Key]model_state.State{
		tStateOpenKey: {
			Key: tStateOpenKey, Name: "Open",
			Actions: []model_state.StateAction{
				{Key: tSActionKey, ActionKey: tActionKey, When: "do"},
			},
		},
		tStateClosedKey: {Key: tStateClosedKey, Name: "Closed"},
	}
	class.Events = map[identity.Key]model_state.Event{
		tEventCreateKey: eventCreate,
		tEventCloseKey:  eventClose,
	}
	class.Guards = map[identity.Key]model_state.Guard{
		tGuardKey: guard,
	}
	class.Actions = map[identity.Key]model_state.Action{
		tActionKey: action,
	}
	class.Queries = map[identity.Key]model_state.Query{
		tQueryKey: query,
	}
	class.Transitions = map[identity.Key]model_state.Transition{
		tTransCreateKey: {
			Key:          tTransCreateKey,
			FromStateKey: nil,
			EventKey:     tEventCreateKey,
			ToStateKey:   &tStateOpenKey,
		},
		tTransCloseKey: {
			Key:          tTransCloseKey,
			FromStateKey: &tStateOpenKey,
			EventKey:     tEventCloseKey,
			GuardKey:     &tGuardKey,
			ActionKey:    &tActionKey,
			ToStateKey:   &tStateClosedKey,
		},
	}

	// Second class (minimal).
	class2 := helper.Must(model_class.NewClass(tClass2Key, "Item", "", nil, nil, nil, ""))
	class2.Attributes = map[identity.Key]model_class.Attribute{}
	class2.States = map[identity.Key]model_state.State{}
	class2.Events = map[identity.Key]model_state.Event{}
	class2.Guards = map[identity.Key]model_state.Guard{}
	class2.Actions = map[identity.Key]model_state.Action{}
	class2.Queries = map[identity.Key]model_state.Query{}
	class2.Transitions = map[identity.Key]model_state.Transition{}

	// Generalization.
	gen := helper.Must(model_class.NewGeneralization(tGenKey, "Vehicle", "", false, false, ""))

	// Use case.
	useCase := helper.Must(model_use_case.NewUseCase(tUseCaseKey, "PlaceOrder", "", "sea", false, nil, nil, ""))
	scenario := helper.Must(model_scenario.NewScenario(tScenarioKey, "HappyPath", ""))
	obj := helper.Must(model_scenario.NewObject(tObjectKey, 1, "order1", "name", tClassKey, false, ""))
	scenario.Objects = map[identity.Key]model_scenario.Object{
		tObjectKey: obj,
	}
	useCase.Scenarios = map[identity.Key]model_scenario.Scenario{
		tScenarioKey: scenario,
	}
	useCase.Actors = map[identity.Key]model_use_case.Actor{}

	// Use case generalization.
	ucGen := helper.Must(model_use_case.NewGeneralization(tUseCaseGenKey, "OrderFlow", "", false, false, ""))

	// Subdomain.
	subdomain := helper.Must(model_domain.NewSubdomain(tSubdomainKey, "S", "", ""))
	subdomain.Generalizations = map[identity.Key]model_class.Generalization{
		tGenKey: gen,
	}
	subdomain.UseCaseGeneralizations = map[identity.Key]model_use_case.Generalization{
		tUseCaseGenKey: ucGen,
	}
	subdomain.Classes = map[identity.Key]model_class.Class{
		tClassKey:  class,
		tClass2Key: class2,
	}
	subdomain.UseCases = map[identity.Key]model_use_case.UseCase{
		tUseCaseKey: useCase,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		tAssocKey: {
			Key:              tAssocKey,
			Name:             "order_items",
			FromClassKey:     tClassKey,
			ToClassKey:       tClass2Key,
			FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 0},
		},
	}

	// Domain.
	domain := helper.Must(model_domain.NewDomain(tDomainKey, "D", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		tSubdomainKey: subdomain,
	}

	// Actors.
	actor := helper.Must(model_actor.NewActor(tActorKey, "User", "", "person", nil, nil, ""))

	// Actor generalization.
	actorGen := helper.Must(model_actor.NewGeneralization(tActorGenKey, "UserType", "", false, false, ""))

	// Model.
	model := helper.Must(req_model.NewModel("test", "Test", "", []model_logic.Logic{invariant}, map[identity.Key]model_logic.GlobalFunction{
		tGlobalFuncKey: globalFunc,
	}))
	model.Actors = map[identity.Key]model_actor.Actor{
		tActorKey: actor,
	}
	model.ActorGeneralizations = map[identity.Key]model_actor.Generalization{
		tActorGenKey: actorGen,
	}
	model.Domains = map[identity.Key]model_domain.Domain{
		tDomainKey: domain,
	}
	model.DomainAssociations = map[identity.Key]model_domain.Association{
		tDomainAssocKey: {
			Key:               tDomainAssocKey,
			ProblemDomainKey:  tDomainKey,
			SolutionDomainKey: tDomainKey,
		},
	}
	model.ClassAssociations = map[identity.Key]model_class.Association{}

	return model
}

// ============================================================
// Suite
// ============================================================

type RequirementsSuite struct {
	suite.Suite
}

func TestRequirementsSuite(t *testing.T) {
	suite.Run(t, new(RequirementsSuite))
}

// ============================================================
// Flattening Tests
// ============================================================

func (s *RequirementsSuite) TestFlattenModel_Actors() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Actors, 1)
	s.Contains(reqs.Actors, tActorKey)
	s.Equal("User", reqs.Actors[tActorKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_ActorGeneralizations() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.ActorGeneralizations, 1)
	s.Contains(reqs.ActorGeneralizations, tActorGenKey)
	s.Equal("UserType", reqs.ActorGeneralizations[tActorGenKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Domains() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Domains, 1)
	s.Contains(reqs.Domains, tDomainKey)
	s.Equal("D", reqs.Domains[tDomainKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Subdomains() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Subdomains, 1)
	s.Contains(reqs.Subdomains, tSubdomainKey)
	s.Equal("S", reqs.Subdomains[tSubdomainKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_DomainAssociations() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.DomainAssociations, 1)
	s.Contains(reqs.DomainAssociations, tDomainAssocKey)
}

func (s *RequirementsSuite) TestFlattenModel_Generalizations() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Generalizations, 1)
	s.Contains(reqs.Generalizations, tGenKey)
	s.Equal("Vehicle", reqs.Generalizations[tGenKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Classes() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Classes, 2)
	s.Contains(reqs.Classes, tClassKey)
	s.Contains(reqs.Classes, tClass2Key)
	s.Equal("Order", reqs.Classes[tClassKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Attributes() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Attributes, 1)
	s.Contains(reqs.Attributes, tAttributeKey)
	s.Equal("amount", reqs.Attributes[tAttributeKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_ClassAssociations() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.ClassAssociations, 1)
	s.Contains(reqs.ClassAssociations, tAssocKey)
	s.Equal("order_items", reqs.ClassAssociations[tAssocKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_States() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.States, 2)
	s.Contains(reqs.States, tStateOpenKey)
	s.Contains(reqs.States, tStateClosedKey)
}

func (s *RequirementsSuite) TestFlattenModel_Events() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Events, 2)
	s.Contains(reqs.Events, tEventCreateKey)
	s.Contains(reqs.Events, tEventCloseKey)
}

func (s *RequirementsSuite) TestFlattenModel_Guards() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Guards, 1)
	s.Contains(reqs.Guards, tGuardKey)
	s.Equal("is_valid", reqs.Guards[tGuardKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Actions() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Actions, 1)
	s.Contains(reqs.Actions, tActionKey)
	s.Equal("DoClose", reqs.Actions[tActionKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Queries() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Queries, 1)
	s.Contains(reqs.Queries, tQueryKey)
	s.Equal("GetTotal", reqs.Queries[tQueryKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Transitions() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Transitions, 2)
	s.Contains(reqs.Transitions, tTransCreateKey)
	s.Contains(reqs.Transitions, tTransCloseKey)
}

func (s *RequirementsSuite) TestFlattenModel_StateActions() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.StateActions, 1)
	s.Contains(reqs.StateActions, tSActionKey)
	s.Equal("do", reqs.StateActions[tSActionKey].When)
}

func (s *RequirementsSuite) TestFlattenModel_GlobalFunctions() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.GlobalFunctions, 1)
	s.Contains(reqs.GlobalFunctions, tGlobalFuncKey)
	s.Equal("_Max", reqs.GlobalFunctions[tGlobalFuncKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_UseCases() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.UseCases, 1)
	s.Contains(reqs.UseCases, tUseCaseKey)
	s.Equal("PlaceOrder", reqs.UseCases[tUseCaseKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_UseCaseGeneralizations() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.UseCaseGeneralizations, 1)
	s.Contains(reqs.UseCaseGeneralizations, tUseCaseGenKey)
	s.Equal("OrderFlow", reqs.UseCaseGeneralizations[tUseCaseGenKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Scenarios() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Scenarios, 1)
	s.Contains(reqs.Scenarios, tScenarioKey)
	s.Equal("HappyPath", reqs.Scenarios[tScenarioKey].Name)
}

func (s *RequirementsSuite) TestFlattenModel_Objects() {
	reqs := NewRequirements(buildTestModel())
	s.Len(reqs.Objects, 1)
	s.Contains(reqs.Objects, tObjectKey)
	s.Equal("order1", reqs.Objects[tObjectKey].Name)
}

// ============================================================
// Simple Lookup Tests
// ============================================================

func (s *RequirementsSuite) TestActorLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ActorLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tActorKey.String())
	s.Equal("User", lookup[tActorKey.String()].Name)
}

func (s *RequirementsSuite) TestActorGeneralizationLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ActorGeneralizationLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tActorGenKey.String())
	s.Equal("UserType", lookup[tActorGenKey.String()].Name)
}

func (s *RequirementsSuite) TestDomainLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup, assocs := reqs.DomainLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tDomainKey.String())
	s.Len(assocs, 1)
}

func (s *RequirementsSuite) TestGeneralizationLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.GeneralizationLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tGenKey.String())
}

func (s *RequirementsSuite) TestClassLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup, assocs := reqs.ClassLookup()
	s.Len(lookup, 2)
	s.Contains(lookup, tClassKey.String())
	s.Contains(lookup, tClass2Key.String())
	s.Len(assocs, 1)
}

func (s *RequirementsSuite) TestStateLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.StateLookup()
	s.Len(lookup, 2)
	s.Contains(lookup, tStateOpenKey.String())
	s.Contains(lookup, tStateClosedKey.String())
}

func (s *RequirementsSuite) TestEventLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.EventLookup()
	s.Len(lookup, 2)
	s.Contains(lookup, tEventCreateKey.String())
	s.Contains(lookup, tEventCloseKey.String())
}

func (s *RequirementsSuite) TestGuardLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.GuardLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tGuardKey.String())
}

func (s *RequirementsSuite) TestActionLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ActionLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tActionKey.String())
}

func (s *RequirementsSuite) TestQueryLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.QueryLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tQueryKey.String())
	s.Equal("GetTotal", lookup[tQueryKey.String()].Name)
}

func (s *RequirementsSuite) TestGlobalFunctionLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.GlobalFunctionLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tGlobalFuncKey.String())
	s.Equal("_Max", lookup[tGlobalFuncKey.String()].Name)
}

func (s *RequirementsSuite) TestUseCaseLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.UseCaseLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tUseCaseKey.String())
}

func (s *RequirementsSuite) TestUseCaseGeneralizationLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.UseCaseGeneralizationLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tUseCaseGenKey.String())
	s.Equal("OrderFlow", lookup[tUseCaseGenKey.String()].Name)
}

func (s *RequirementsSuite) TestScenarioLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ScenarioLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tScenarioKey.String())
}

func (s *RequirementsSuite) TestObjectLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ObjectLookup()
	s.Len(lookup, 1)
	s.Contains(lookup, tObjectKey.String())
}

// ============================================================
// Cross-Reference Lookup Tests
// ============================================================

func (s *RequirementsSuite) TestClassDomainLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ClassDomainLookup()
	s.Contains(lookup, tClassKey.String())
	s.Equal("D", lookup[tClassKey.String()].Name)
}

func (s *RequirementsSuite) TestClassSubdomainLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ClassSubdomainLookup()
	s.Contains(lookup, tClassKey.String())
	s.Equal("S", lookup[tClassKey.String()].Name)
}

func (s *RequirementsSuite) TestDomainClassesLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.DomainClassesLookup()
	s.Contains(lookup, tDomainKey.String())
	s.Len(lookup[tDomainKey.String()], 2)
}

func (s *RequirementsSuite) TestDomainUseCasesLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.DomainUseCasesLookup()
	s.Contains(lookup, tDomainKey.String())
	s.Len(lookup[tDomainKey.String()], 1)
}

func (s *RequirementsSuite) TestUseCaseDomainLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.UseCaseDomainLookup()
	s.Contains(lookup, tUseCaseKey.String())
	s.Equal("D", lookup[tUseCaseKey.String()].Name)
}

func (s *RequirementsSuite) TestUseCaseSubdomainLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.UseCaseSubdomainLookup()
	s.Contains(lookup, tUseCaseKey.String())
	s.Equal("S", lookup[tUseCaseKey.String()].Name)
}

func (s *RequirementsSuite) TestGeneralizationSuperclassLookup() {
	// Build a model where a class is a superclass of a generalization.
	model := buildTestModel()

	// Modify Order to be superclass of the Vehicle generalization.
	class := model.Domains[tDomainKey].Subdomains[tSubdomainKey].Classes[tClassKey]
	genKey := tGenKey
	class.SuperclassOfKey = &genKey

	subdomain := model.Domains[tDomainKey].Subdomains[tSubdomainKey]
	subdomain.Classes[tClassKey] = class
	domain := model.Domains[tDomainKey]
	domain.Subdomains[tSubdomainKey] = subdomain
	model.Domains[tDomainKey] = domain

	reqs := NewRequirements(model)
	lookup := reqs.GeneralizationSuperclassLookup()
	s.Contains(lookup, tGenKey.String())
	s.Equal("Order", lookup[tGenKey.String()].Name)
}

func (s *RequirementsSuite) TestGeneralizationSubclassesLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.GeneralizationSubclassesLookup()
	// No classes reference the generalization, so should have empty slice.
	s.Contains(lookup, tGenKey.String())
	s.Empty(lookup[tGenKey.String()])
}

func (s *RequirementsSuite) TestActionTransitionsLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ActionTransitionsLookup()
	s.Contains(lookup, tActionKey.String())
	s.Len(lookup[tActionKey.String()], 1)
	s.Equal(tTransCloseKey, lookup[tActionKey.String()][0].Key)
}

func (s *RequirementsSuite) TestActionStateActionsLookup() {
	reqs := NewRequirements(buildTestModel())
	lookup := reqs.ActionStateActionsLookup()
	s.Contains(lookup, tActionKey.String())
	s.Len(lookup[tActionKey.String()], 1)
	s.Equal(tSActionKey, lookup[tActionKey.String()][0].Key)
}

func (s *RequirementsSuite) TestActorGeneralizationSuperclassLookup() {
	// Build model with actor superclass.
	model := buildTestModel()

	genKey := tActorGenKey
	superActor := helper.Must(model_actor.NewActor(tActorKey, "User", "", "person", &genKey, nil, ""))
	model.Actors[tActorKey] = superActor

	reqs := NewRequirements(model)
	lookup := reqs.ActorGeneralizationSuperclassLookup()
	s.Contains(lookup, tActorGenKey.String())
	s.Equal("User", lookup[tActorGenKey.String()].Name)
}

func (s *RequirementsSuite) TestActorGeneralizationSubclassesLookup() {
	// Build model with actor subclasses.
	model := buildTestModel()

	genKey := tActorGenKey
	subActor1 := helper.Must(model_actor.NewActor(tActor2Key, "Admin", "", "person", nil, &genKey, ""))
	subActor2 := helper.Must(model_actor.NewActor(tActor3Key, "Guest", "", "person", nil, &genKey, ""))
	model.Actors[tActor2Key] = subActor1
	model.Actors[tActor3Key] = subActor2

	reqs := NewRequirements(model)
	lookup := reqs.ActorGeneralizationSubclassesLookup()
	s.Contains(lookup, tActorGenKey.String())
	s.Len(lookup[tActorGenKey.String()], 2)
	// Sorted by key.
	s.Equal("Admin", lookup[tActorGenKey.String()][0].Name)
	s.Equal("Guest", lookup[tActorGenKey.String()][1].Name)
}

func (s *RequirementsSuite) TestUseCaseGeneralizationSuperclassLookup() {
	model := buildTestModel()

	genKey := tUseCaseGenKey
	superUC := helper.Must(model_use_case.NewUseCase(tUseCaseKey, "PlaceOrder", "", "sea", false, &genKey, nil, ""))
	superUC.Actors = map[identity.Key]model_use_case.Actor{}
	superUC.Scenarios = map[identity.Key]model_scenario.Scenario{
		tScenarioKey: helper.Must(model_scenario.NewScenario(tScenarioKey, "HappyPath", "")),
	}

	subdomain := model.Domains[tDomainKey].Subdomains[tSubdomainKey]
	subdomain.UseCases[tUseCaseKey] = superUC
	domain := model.Domains[tDomainKey]
	domain.Subdomains[tSubdomainKey] = subdomain
	model.Domains[tDomainKey] = domain

	reqs := NewRequirements(model)
	lookup := reqs.UseCaseGeneralizationSuperclassLookup()
	s.Contains(lookup, tUseCaseGenKey.String())
	s.Equal("PlaceOrder", lookup[tUseCaseGenKey.String()].Name)
}

func (s *RequirementsSuite) TestUseCaseGeneralizationSubclassesLookup() {
	model := buildTestModel()

	genKey := tUseCaseGenKey
	subUC1 := helper.Must(model_use_case.NewUseCase(tUseCase2Key, "Login", "", "mud", false, nil, &genKey, ""))
	subUC1.Actors = map[identity.Key]model_use_case.Actor{}
	subUC1.Scenarios = map[identity.Key]model_scenario.Scenario{}
	subUC2 := helper.Must(model_use_case.NewUseCase(tUseCase3Key, "Checkout", "", "mud", false, nil, &genKey, ""))
	subUC2.Actors = map[identity.Key]model_use_case.Actor{}
	subUC2.Scenarios = map[identity.Key]model_scenario.Scenario{}

	subdomain := model.Domains[tDomainKey].Subdomains[tSubdomainKey]
	subdomain.UseCases[tUseCase2Key] = subUC1
	subdomain.UseCases[tUseCase3Key] = subUC2
	domain := model.Domains[tDomainKey]
	domain.Subdomains[tSubdomainKey] = subdomain
	model.Domains[tDomainKey] = domain

	reqs := NewRequirements(model)
	lookup := reqs.UseCaseGeneralizationSubclassesLookup()
	s.Contains(lookup, tUseCaseGenKey.String())
	s.Len(lookup[tUseCaseGenKey.String()], 2)
	// Sorted by key.
	s.Equal("Checkout", lookup[tUseCaseGenKey.String()][0].Name)
	s.Equal("Login", lookup[tUseCaseGenKey.String()][1].Name)
}

// ============================================================
// DomainHasMultipleSubdomains Tests
// ============================================================

func (s *RequirementsSuite) TestDomainHasMultipleSubdomains_SingleNonDefault() {
	reqs := NewRequirements(buildTestModel())
	// Subdomain "S" is not "default".
	s.True(reqs.DomainHasMultipleSubdomains(tDomainKey))
}

func (s *RequirementsSuite) TestDomainHasMultipleSubdomains_SingleDefault() {
	model := buildTestModel()

	// Replace subdomain with one named "default".
	defaultSubKey := helper.Must(identity.NewSubdomainKey(tDomainKey, "default"))
	defaultSub := helper.Must(model_domain.NewSubdomain(defaultSubKey, "Default", "", ""))
	defaultSub.Generalizations = map[identity.Key]model_class.Generalization{}
	defaultSub.UseCaseGeneralizations = map[identity.Key]model_use_case.Generalization{}
	defaultSub.Classes = map[identity.Key]model_class.Class{}
	defaultSub.UseCases = map[identity.Key]model_use_case.UseCase{}
	defaultSub.ClassAssociations = map[identity.Key]model_class.Association{}

	domain := model.Domains[tDomainKey]
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		defaultSubKey: defaultSub,
	}
	model.Domains[tDomainKey] = domain

	reqs := NewRequirements(model)
	s.False(reqs.DomainHasMultipleSubdomains(tDomainKey))
}

func (s *RequirementsSuite) TestDomainHasMultipleSubdomains_UnknownDomain() {
	reqs := NewRequirements(buildTestModel())
	unknownKey := helper.Must(identity.NewDomainKey("unknown"))
	s.False(reqs.DomainHasMultipleSubdomains(unknownKey))
}

// ============================================================
// Empty Model Test
// ============================================================

func (s *RequirementsSuite) TestFlattenModel_EmptyModel() {
	model := helper.Must(req_model.NewModel("empty", "Empty", "", nil, nil))
	reqs := NewRequirements(model)

	s.Empty(reqs.Actors)
	s.Empty(reqs.ActorGeneralizations)
	s.Empty(reqs.Domains)
	s.Empty(reqs.Subdomains)
	s.Empty(reqs.DomainAssociations)
	s.Empty(reqs.Generalizations)
	s.Empty(reqs.Classes)
	s.Empty(reqs.Attributes)
	s.Empty(reqs.ClassAssociations)
	s.Empty(reqs.States)
	s.Empty(reqs.Events)
	s.Empty(reqs.Guards)
	s.Empty(reqs.Actions)
	s.Empty(reqs.Queries)
	s.Empty(reqs.Transitions)
	s.Empty(reqs.StateActions)
	s.Empty(reqs.GlobalFunctions)
	s.Empty(reqs.UseCases)
	s.Empty(reqs.UseCaseGeneralizations)
	s.Empty(reqs.Scenarios)
	s.Empty(reqs.Objects)
}
