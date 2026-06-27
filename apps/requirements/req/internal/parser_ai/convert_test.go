package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

// ConvertSuite tests the conversion functions between req_model and inputModel.
type ConvertSuite struct {
	suite.Suite
}

func TestConvertSuite(t *testing.T) {
	suite.Run(t, new(ConvertSuite))
}

// inputAttributesFrom builds an attribute slice from a key-to-attribute map for tests.
func inputAttributesFrom(entries map[string]inputAttribute) []inputAttribute {
	attrs := make([]inputAttribute, 0, len(entries))
	for key, attr := range entries {
		attr.Key = key
		attrs = append(attrs, attr)
	}
	return attrs
}

// TestConvertFromModelMinimal tests converting a minimal valid core.Model to inputModel.
func (suite *ConvertSuite) TestConvertFromModelMinimal() {
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: "Model details"}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = make(map[identity.Key]model_domain.Domain)
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)
	suite.Equal("Test Model", input.Name)
	suite.Equal("Model details", input.Details)
	suite.Empty(input.Actors)
	suite.Empty(input.Domains)
	suite.Empty(input.ClassAssociations)
}

// TestConvertToModelMinimal tests converting a minimal inputModel to core.Model.
func (suite *ConvertSuite) TestConvertToModelMinimal() {
	input := &inputModel{
		Name:              "Test Model",
		Details:           "Model details",
		Actors:            make(map[string]*inputActor),
		Domains:           make(map[string]*inputDomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)
	suite.Equal("testmodel", model.Key)
	suite.Equal("Test Model", model.Name)
	suite.Equal("Model details", model.Details)
	suite.Empty(model.Actors)
	suite.Empty(model.Domains)
}

// TestConvertFromModelWithActor tests converting an actor.
func (suite *ConvertSuite) TestConvertFromModelWithActor() {
	actorKey := helper.Must(identity.NewActorKey("customer"))

	actor := model_actor.NewActor(actorKey, "person", model_actor.GeneralizationRefs{SuperclassOfKey: nil, SubclassOfKey: nil}, model_actor.ActorDetails{Name: "Customer", Details: "Customer details", UnfinishedNotes: "", UmlComment: ""})

	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = map[identity.Key]model_actor.Actor{
		actorKey: actor,
	}
	m.Domains = make(map[identity.Key]model_domain.Domain)
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)
	suite.Require().Contains(input.Actors, "customer")
	suite.Equal("Customer", input.Actors["customer"].Name)
	suite.Equal("person", input.Actors["customer"].Type)
	suite.Equal("Customer details", input.Actors["customer"].Details)
}

// TestConvertToModelWithActor tests converting an actor.
func (suite *ConvertSuite) TestConvertToModelWithActor() {
	input := &inputModel{
		Name: "Test Model",
		Actors: map[string]*inputActor{
			"customer": {
				Name:    "Customer",
				Type:    "person",
				Details: "Customer details",
			},
		},
		Domains:           make(map[string]*inputDomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)
	suite.Require().Len(model.Actors, 1)

	// Find the actor by checking the key's SubKey
	var foundActor model_actor.Actor
	for key, actor := range model.Actors {
		if key.SubKey == "customer" {
			foundActor = actor
			break
		}
	}
	suite.Equal("Customer", foundActor.Name)
	suite.Equal("person", foundActor.Type)
}

// TestConvertFromModelWithClass tests converting a class with attributes.
func (suite *ConvertSuite) TestConvertFromModelWithClass() {
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	actorKey := helper.Must(identity.NewActorKey("customer"))

	idAttrKey := helper.Must(identity.NewAttributeKey(classKey, "id"))
	statusAttrKey := helper.Must(identity.NewAttributeKey(classKey, "status"))

	// Build attributes
	idAttr := helper.Must(model_class.NewAttribute(idAttrKey, model_class.AttributeDetails{Name: "ID", Details: "The order ID"}, "int", nil, false, model_class.AttributeAnnotations{IndexNums: []uint{0}}))
	statusAttr := helper.Must(model_class.NewAttribute(statusAttrKey, model_class.AttributeDetails{Name: "Status", Details: ""}, "string", nil, false, model_class.AttributeAnnotations{}))

	// Build class
	orderClass := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: &actorKey, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "Order details", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes([]model_class.Attribute{idAttr, statusAttr})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build actor
	actor := model_actor.NewActor(actorKey, "person", model_actor.GeneralizationRefs{SuperclassOfKey: nil, SubclassOfKey: nil}, model_actor.ActorDetails{Name: "Customer", Details: "", UnfinishedNotes: "", UmlComment: ""})

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = map[identity.Key]model_actor.Actor{
		actorKey: actor,
	}
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	suite.Require().Contains(input.Domains, "orders")
	suite.Require().Contains(input.Domains["orders"].Subdomains, "default")
	suite.Require().Contains(input.Domains["orders"].Subdomains["default"].Classes, "order")

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	suite.Equal("Order", class.Name)
	suite.Equal("Order details", class.Details)
	suite.Equal("customer", class.ActorKey)
	suite.True(class.hasAttributeKey("id"))
	inputIDAttr, ok := class.attributeByKey("id")
	suite.Require().True(ok)
	suite.Equal("ID", inputIDAttr.Name)
	suite.Equal("int", inputIDAttr.DataTypeRules)
}

// TestConvertToModelWithClass tests converting a class with attributes.
func (suite *ConvertSuite) TestConvertToModelWithClass() {
	input := &inputModel{
		Name: "Test Model",
		Actors: map[string]*inputActor{
			"customer": {Name: "Customer", Type: "person"},
		},
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order": {
								Name:     "Order",
								Details:  "Order details",
								ActorKey: "customer",
								Attributes: inputAttributesFrom(map[string]inputAttribute{
									"id":     {Name: "ID", DataTypeRules: "int"},
									"status": {Name: "Status", DataTypeRules: "string"},
								}),
								Indexes: [][]string{{"id"}},
							},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Find domain
	var domain model_domain.Domain
	for key, d := range model.Domains {
		if key.SubKey == "orders" {
			domain = d
			break
		}
	}
	suite.Require().NotEmpty(domain.Name)

	// Find subdomain
	var subdomain model_domain.Subdomain
	for key, s := range domain.Subdomains {
		if key.SubKey == "default" {
			subdomain = s
			break
		}
	}
	suite.Require().NotEmpty(subdomain.Name)

	// Find class
	var class model_class.Class
	for key, c := range subdomain.Classes {
		if key.SubKey == "order" {
			class = c
			break
		}
	}
	suite.Equal("Order", class.Name)
	suite.Equal("Order details", class.Details)
}

// TestConvertFromModelWithStateMachine tests converting a state machine.
func (suite *ConvertSuite) TestConvertFromModelWithStateMachine() {
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	stateKey1 := helper.Must(identity.NewStateKey(classKey, "pending"))
	stateKey2 := helper.Must(identity.NewStateKey(classKey, "confirmed"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "confirm"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "has_items"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "process"))
	transitionKey := helper.Must(identity.NewTransitionKey(classKey, "pending", "confirm", "has_items", "process", "confirmed"))

	// Build states
	state1 := model_state.NewState(stateKey1, "Pending", "", "")
	state2 := model_state.NewState(stateKey2, "Confirmed", "", "")

	// Build event
	event := model_state.NewEvent(eventKey, "confirm", "", nil)

	// Build guard with logic
	guardLogic := model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Check if order has items", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	guard := model_state.NewGuard(guardKey, "has_items", guardLogic)

	// Build transition
	transition := model_state.NewTransition(transitionKey, eventKey, model_state.TransitionStateKeys{FromStateKey: &stateKey1, ToStateKey: &stateKey2}, model_state.TransitionLogicKeys{GuardKey: &guardKey, ActionKey: &actionKey}, "")

	// Build action
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Process", Details: "Process the order"}, nil, nil, nil, nil)

	// Build class
	orderClass := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes(nil)
	orderClass.SetStates(map[identity.Key]model_state.State{
		stateKey1: state1,
		stateKey2: state2,
	})
	orderClass.SetEvents(map[identity.Key]model_state.Event{
		eventKey: event,
	})
	orderClass.SetGuards(map[identity.Key]model_state.Guard{
		guardKey: guard,
	})
	orderClass.SetTransitions(map[identity.Key]model_state.Transition{
		transitionKey: transition,
	})
	orderClass.SetActions(map[identity.Key]model_state.Action{
		actionKey: action,
	})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	suite.Require().NotNil(class.StateMachine)

	sm := class.StateMachine
	suite.Require().Contains(sm.States, "pending")
	suite.Require().Contains(sm.States, "confirmed")
	suite.Require().Contains(sm.Events, "confirm")
	suite.Require().Contains(sm.Guards, "has_items")
	suite.Require().Len(sm.Transitions, 1)

	trans := sm.Transitions[0]
	suite.Equal("pending", *trans.FromStateKey)
	suite.Equal("confirmed", *trans.ToStateKey)
	suite.Equal("confirm", trans.EventKey)
	suite.Equal("has_items", *trans.GuardKey)
	suite.Equal("process", *trans.ActionKey)
}

// TestConvertFromModelDeterministicExport verifies repeated exports produce identical JSON.
func (suite *ConvertSuite) TestConvertFromModelDeterministicExport() {
	domainKey := helper.Must(identity.NewDomainKey("users"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "managed_user"))

	activeKey := helper.Must(identity.NewStateKey(classKey, "active"))
	unprovisionedKey := helper.Must(identity.NewStateKey(classKey, "unprovisioned"))
	deactivatedKey := helper.Must(identity.NewStateKey(classKey, "deactivated"))

	deactivateEventKey := helper.Must(identity.NewEventKey(classKey, "deactivate_user"))
	createEventKey := helper.Must(identity.NewEventKey(classKey, "create_managed_user"))
	selfElevationEventKey := helper.Must(identity.NewEventKey(classKey, "self_elevation_blocked"))

	deactivateActionKey := helper.Must(identity.NewActionKey(classKey, "deactivate_user"))
	createActionKey := helper.Must(identity.NewActionKey(classKey, "create_managed_user"))
	selfElevationActionKey := helper.Must(identity.NewActionKey(classKey, "self_elevation_blocked"))

	deactivateTransitionKey := helper.Must(identity.NewTransitionKey(classKey, "active", "deactivate_user", "", "deactivate_user", "deactivated"))
	createTransitionKey := helper.Must(identity.NewTransitionKey(classKey, "unprovisioned", "create_managed_user", "", "create_managed_user", "active"))
	selfElevationTransitionKey := helper.Must(identity.NewTransitionKey(classKey, "unprovisioned", "self_elevation_blocked", "", "self_elevation_blocked", "unprovisioned"))

	userClass := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Managed User"})
	userClass.SetAttributes(nil)
	userClass.SetStates(map[identity.Key]model_state.State{
		activeKey:        model_state.NewState(activeKey, "Active", "", ""),
		unprovisionedKey: model_state.NewState(unprovisionedKey, "Unprovisioned", "", ""),
		deactivatedKey:   model_state.NewState(deactivatedKey, "Deactivated", "", ""),
	})
	userClass.SetEvents(map[identity.Key]model_state.Event{
		deactivateEventKey:    model_state.NewEvent(deactivateEventKey, "deactivate_user", "", nil),
		createEventKey:        model_state.NewEvent(createEventKey, "create_managed_user", "", nil),
		selfElevationEventKey: model_state.NewEvent(selfElevationEventKey, "self_elevation_blocked", "", nil),
	})
	userClass.SetGuards(make(map[identity.Key]model_state.Guard))
	userClass.SetActions(map[identity.Key]model_state.Action{
		deactivateActionKey:    model_state.NewAction(deactivateActionKey, model_state.ActionDetails{Name: "deactivate_user", Details: ""}, nil, nil, nil, nil),
		createActionKey:        model_state.NewAction(createActionKey, model_state.ActionDetails{Name: "create_managed_user", Details: ""}, nil, nil, nil, nil),
		selfElevationActionKey: model_state.NewAction(selfElevationActionKey, model_state.ActionDetails{Name: "self_elevation_blocked", Details: ""}, nil, nil, nil, nil),
	})
	userClass.SetTransitions(map[identity.Key]model_state.Transition{
		deactivateTransitionKey:    model_state.NewTransition(deactivateTransitionKey, deactivateEventKey, model_state.TransitionStateKeys{FromStateKey: &activeKey, ToStateKey: &deactivatedKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &deactivateActionKey}, ""),
		createTransitionKey:        model_state.NewTransition(createTransitionKey, createEventKey, model_state.TransitionStateKeys{FromStateKey: &unprovisionedKey, ToStateKey: &activeKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &createActionKey}, ""),
		selfElevationTransitionKey: model_state.NewTransition(selfElevationTransitionKey, selfElevationEventKey, model_state.TransitionStateKeys{FromStateKey: &unprovisionedKey, ToStateKey: &unprovisionedKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &selfElevationActionKey}, ""),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: userClass}

	domain := model_domain.NewDomain(domainKey, "Users", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	model := &m

	marshalStateMachine := func() []byte {
		input, err := ConvertFromModel(model)
		suite.Require().NoError(err)
		sm := input.Domains["users"].Subdomains["default"].Classes["managed_user"].StateMachine
		suite.Require().NotNil(sm)
		data, err := json.Marshal(sm)
		suite.Require().NoError(err)
		return data
	}

	first := marshalStateMachine()
	second := marshalStateMachine()
	suite.Equal(first, second, "repeated ConvertFromModel must produce identical state machine JSON")

	sm := helper.Must(ConvertFromModel(model)).Domains["users"].Subdomains["default"].Classes["managed_user"].StateMachine
	suite.Require().Len(sm.Transitions, 3)
	suite.Equal("deactivate_user", sm.Transitions[0].EventKey)
	suite.Equal("create_managed_user", sm.Transitions[1].EventKey)
	suite.Equal("self_elevation_blocked", sm.Transitions[2].EventKey)
}

// TestConvertToModelWithStateMachine tests converting a state machine.
func (suite *ConvertSuite) TestConvertToModelWithStateMachine() {
	fromState := "pending"
	toState := "confirmed"
	guardKey := "has_items"
	actionKey := "process"

	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order": {
								Name:       "Order",
								Attributes: nil,
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"pending":   {Name: "Pending"},
										"confirmed": {Name: "Confirmed"},
									},
									Events: map[string]*inputEvent{
										"confirm": {Name: "confirm"},
									},
									Guards: map[string]*inputGuard{
										"has_items": {Name: "has_items", Logic: inputLogic{Description: "Check if order has items", Notation: model_logic.NotationTLAPlus}},
									},
									Transitions: []inputTransition{
										{
											FromStateKey: &fromState,
											ToStateKey:   &toState,
											EventKey:     "confirm",
											GuardKey:     &guardKey,
											ActionKey:    &actionKey,
										},
									},
								},
								Actions: map[string]*inputAction{
									"process": {Name: "Process"},
								},
							},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Navigate to the class
	var class model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for key, c := range subdomain.Classes {
				if key.SubKey == "order" {
					class = c
					break
				}
			}
		}
	}
	suite.Len(class.States, 2)
	suite.Len(class.Events, 1)
	suite.Len(class.Guards, 1)
	suite.Len(class.Transitions, 1)
}

// TestConvertFromModelWithQueries tests converting queries.
func (suite *ConvertSuite) TestConvertFromModelWithQueries() {
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "get_total"))

	// Build query logic
	requireKey := helper.Must(identity.NewQueryRequireKey(queryKey, "0"))
	guaranteeKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	requireLogic := model_logic.NewLogic(requireKey, model_logic.LogicTypeAssessment, "order must exist", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	guaranteeLogic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeQuery, "returns total amount", "total", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)

	// Build query
	query := model_state.NewQuery(queryKey, "Get Total", "Get order total",
		[]model_logic.Logic{requireLogic},
		[]model_logic.Logic{guaranteeLogic},
		nil,
	)

	// Build class
	orderClass := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes(nil)
	orderClass.SetQueries(map[identity.Key]model_state.Query{
		queryKey: query,
	})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	suite.Require().Contains(class.Queries, "get_total")

	inputQuery := class.Queries["get_total"]
	suite.Equal("Get Total", inputQuery.Name)
	suite.Equal("Get order total", inputQuery.Details)
	suite.Require().Len(inputQuery.Requires, 1)
	suite.Equal("order must exist", inputQuery.Requires[0].Description)
	suite.Require().Len(inputQuery.Guarantees, 1)
	suite.Equal("returns total amount", inputQuery.Guarantees[0].Description)
}

// TestConvertToModelWithQueries tests converting queries.
func (suite *ConvertSuite) TestConvertToModelWithQueries() {
	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order": {
								Name:       "Order",
								Attributes: nil,
								Queries: map[string]*inputQuery{
									"get_total": {
										Name:    "Get Total",
										Details: "Get order total",
										Requires: []inputLogic{
											{Description: "order must exist", Notation: model_logic.NotationTLAPlus},
										},
										Guarantees: []inputLogic{
											{Description: "returns total amount", Target: "total", Notation: model_logic.NotationTLAPlus},
										},
									},
								},
							},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Navigate to the class
	var class model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for key, c := range subdomain.Classes {
				if key.SubKey == "order" {
					class = c
					break
				}
			}
		}
	}
	suite.Require().Len(class.Queries, 1)

	var query model_state.Query
	for _, q := range class.Queries {
		query = q
		break
	}
	suite.Equal("Get Total", query.Name)
	suite.Equal("Get order total", query.Details)
	suite.Require().Len(query.Requires, 1)
	suite.Equal("order must exist", query.Requires[0].Description)
	suite.Require().Len(query.Guarantees, 1)
	suite.Equal("returns total amount", query.Guarantees[0].Description)
}

// TestConvertFromModelWithGeneralization tests converting a generalization.
func (suite *ConvertSuite) TestConvertFromModelWithGeneralization() {
	domainKey := helper.Must(identity.NewDomainKey("products"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	productKey := helper.Must(identity.NewClassKey(subdomainKey, "product"))
	bookKey := helper.Must(identity.NewClassKey(subdomainKey, "book"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "product_types"))

	// Build classes
	productClass := model_class.NewClass(productKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: &genKey, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	productClass.SetAttributes(nil)

	bookClass := model_class.NewClass(bookKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: &genKey}, model_class.ClassDetails{Name: "Book", Details: "", UnfinishedNotes: "", UmlComment: ""})
	bookClass.SetAttributes(nil)

	// Build generalization
	gen := model_class.NewGeneralization(genKey, model_class.GeneralizationDetails{Name: "Product Types", Details: "Types of products"}, "", model_class.GeneralizationTraits{IsComplete: false, IsStatic: false}, "")

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		productKey: productClass,
		bookKey:    bookClass,
	}
	subdomain.Generalizations = map[identity.Key]model_class.Generalization{
		genKey: gen,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Products", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	inputSubdomain := input.Domains["products"].Subdomains["default"]
	suite.Require().Contains(inputSubdomain.ClassGeneralizations, "product_types")

	inputGen := inputSubdomain.ClassGeneralizations["product_types"]
	suite.Equal("Product Types", inputGen.Name)
	suite.Equal("product", inputGen.SuperclassKey)
	suite.Equal([]string{"book"}, inputGen.SubclassKeys)
}

// TestConvertToModelWithGeneralization tests converting a generalization.
func (suite *ConvertSuite) TestConvertToModelWithGeneralization() {
	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"products": {
				Name: "Products",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"product": {Name: "Product", Attributes: nil},
							"book":    {Name: "Book", Attributes: nil},
						},
						ClassGeneralizations: map[string]*inputClassGeneralization{
							"product_types": {
								Name:          "Product Types",
								SuperclassKey: "product",
								SubclassKeys:  []string{"book"},
							},
						},
						ClassAssociations: make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Navigate to the subdomain
	var subdomain model_domain.Subdomain
	for _, domain := range model.Domains {
		for key, s := range domain.Subdomains {
			if key.SubKey == "default" {
				subdomain = s
				break
			}
		}
	}
	suite.Require().Len(subdomain.Generalizations, 1)

	var gen model_class.Generalization
	for _, g := range subdomain.Generalizations {
		gen = g
		break
	}
	suite.Equal("Product Types", gen.Name)

	// In req_model, classes have back-references to their generalization
	var productClass, bookClass model_class.Class
	for key, c := range subdomain.Classes {
		if key.SubKey == "product" {
			productClass = c
		}
		if key.SubKey == "book" {
			bookClass = c
		}
	}
	// Product class should be the superclass of the generalization
	suite.Require().NotNil(productClass.SuperclassOfKey)
	suite.Equal(gen.Key, *productClass.SuperclassOfKey)
	// Book class should be a subclass of the generalization
	suite.Require().NotNil(bookClass.SubclassOfKey)
	suite.Equal(gen.Key, *bookClass.SubclassOfKey)
}

// TestConvertFromModelWithSubdomainAssociation tests converting a subdomain-level association.
func (suite *ConvertSuite) TestConvertFromModelWithSubdomainAssociation() {
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	orderKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	lineItemKey := helper.Must(identity.NewClassKey(subdomainKey, "line_item"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, orderKey, lineItemKey, "order_lines"))

	// Build classes
	orderClass := model_class.NewClass(orderKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes(nil)

	lineItemClass := model_class.NewClass(lineItemKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Line Item", Details: "", UnfinishedNotes: "", UmlComment: ""})
	lineItemClass.SetAttributes(nil)

	// Build association
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "Order Lines", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))}, model_class.AssociationEnd{ClassKey: lineItemKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1..many"))}, model_class.Multiplicity{}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		orderKey:    orderClass,
		lineItemKey: lineItemClass,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	inputSubdomain := input.Domains["orders"].Subdomains["default"]
	suite.Require().Contains(inputSubdomain.ClassAssociations, "order--line_item--order_lines")

	inputAssoc := inputSubdomain.ClassAssociations["order--line_item--order_lines"]
	suite.Equal("Order Lines", inputAssoc.Name)
	suite.Equal("order", inputAssoc.FromClassKey)
	suite.Equal("1", inputAssoc.FromMultiplicity)
	suite.Equal("line_item", inputAssoc.ToClassKey)
	suite.Equal("1..*", inputAssoc.ToMultiplicity)
}

// TestConvertToModelWithSubdomainAssociation tests converting a subdomain-level association.
func (suite *ConvertSuite) TestConvertToModelWithSubdomainAssociation() {
	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order":     {Name: "Order", Attributes: nil},
							"line_item": {Name: "Line Item", Attributes: nil},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations: map[string]*inputClassAssociation{
							"order_lines": {
								Name:             "Order Lines",
								FromClassKey:     "order",
								FromMultiplicity: "1",
								ToClassKey:       "line_item",
								ToMultiplicity:   "1..*",

								Uniqueness: "any",
							},
						},
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Navigate to the subdomain
	var subdomain model_domain.Subdomain
	for _, domain := range model.Domains {
		for key, s := range domain.Subdomains {
			if key.SubKey == "default" {
				subdomain = s
				break
			}
		}
	}
	suite.Require().Len(subdomain.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range subdomain.ClassAssociations {
		assoc = a
		break
	}
	suite.Equal("Order Lines", assoc.Name)
	suite.Equal("order", assoc.FromClassKey.SubKey)
	suite.Equal("line_item", assoc.ToClassKey.SubKey)
}

// TestRoundTripMinimal tests that a minimal model survives roundtrip conversion.
func (suite *ConvertSuite) TestRoundTripMinimal() {
	original := &inputModel{
		Name:              "Test Model",
		Details:           "Model details",
		Actors:            make(map[string]*inputActor),
		Domains:           make(map[string]*inputDomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	// Convert to req_model
	model, err := ConvertToModel(original, "testmodel")
	suite.Require().NoError(err)

	// Convert back to inputModel
	result, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	suite.Equal(original.Name, result.Name)
	suite.Equal(original.Details, result.Details)
}

// TestRoundTripComplete tests that a complete model survives roundtrip conversion.
func (suite *ConvertSuite) TestRoundTripComplete() {
	fromState := "pending"
	toState := "confirmed"
	guardKey := "has_items"
	actionKey := "calculate_total"

	original := &inputModel{
		Name:    "Complete Model",
		Details: "A complete model for testing",
		Actors: map[string]*inputActor{
			"customer": {Name: "Customer", Type: "person", Details: "A customer actor"},
		},
		Domains: map[string]*inputDomain{
			"orders": {
				Name:    "Orders",
				Details: "Orders domain",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name:    "Default",
						Details: "Default subdomain",
						Classes: map[string]*inputClass{
							"order": {
								Name:     "Order",
								Details:  "An order class",
								ActorKey: "customer",
								Attributes: inputAttributesFrom(map[string]inputAttribute{
									"id":     {Name: "ID", Details: "Order ID", DataTypeRules: "int"},
									"status": {Name: "Status", DataTypeRules: "string"},
								}),
								Indexes: [][]string{{"id"}},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"pending":   {Name: "Pending", Details: "Order is pending"},
										"confirmed": {Name: "Confirmed", Details: "Order is confirmed"},
									},
									Events: map[string]*inputEvent{
										"confirm": {Name: "confirm"},
									},
									Guards: map[string]*inputGuard{
										"has_items": {Name: "has_items", Logic: inputLogic{Description: "Order has items", Notation: model_logic.NotationTLAPlus}},
									},
									Transitions: []inputTransition{
										{
											FromStateKey: &fromState,
											ToStateKey:   &toState,
											EventKey:     "confirm",
											GuardKey:     &guardKey,
											ActionKey:    &actionKey,
										},
									},
								},
								Actions: map[string]*inputAction{
									"calculate_total": {Name: "Calculate Total", Details: "Calculate order total"},
								},
								Queries: map[string]*inputQuery{},
							},
							"line_item": {
								Name:       "Line Item",
								Attributes: nil,
							},
							"product": {
								Name:       "Product",
								Attributes: nil,
							},
							"book": {
								Name:       "Book",
								Attributes: nil,
							},
						},
						ClassGeneralizations: map[string]*inputClassGeneralization{
							"product_types": {
								Name:          "Product Types",
								SuperclassKey: "product",
								SubclassKeys:  []string{"book"},
							},
						},
						ClassAssociations: map[string]*inputClassAssociation{
							"order_lines": {
								Name:             "Order Lines",
								FromClassKey:     "order",
								FromMultiplicity: "1",
								ToClassKey:       "line_item",
								ToMultiplicity:   "1..*",

								Uniqueness: "any",
							},
						},
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	// Convert to req_model
	model, err := ConvertToModel(original, "testmodel")
	suite.Require().NoError(err)

	// Convert back to inputModel
	result, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	// Verify top-level fields
	suite.Equal(original.Name, result.Name)
	suite.Equal(original.Details, result.Details)

	// Verify actor
	suite.Require().Contains(result.Actors, "customer")
	suite.Equal(original.Actors["customer"].Name, result.Actors["customer"].Name)
	suite.Equal(original.Actors["customer"].Type, result.Actors["customer"].Type)

	// Verify domain structure
	suite.Require().Contains(result.Domains, "orders")
	suite.Equal(original.Domains["orders"].Name, result.Domains["orders"].Name)

	// Verify subdomain
	suite.Require().Contains(result.Domains["orders"].Subdomains, "default")
	subdomain := result.Domains["orders"].Subdomains["default"]
	suite.Equal("Default", subdomain.Name)

	// Verify class
	suite.Require().Contains(subdomain.Classes, "order")
	class := subdomain.Classes["order"]
	suite.Equal("Order", class.Name)
	suite.Equal("customer", class.ActorKey)

	// Verify attributes
	suite.True(class.hasAttributeKey("id"))
	inputIDAttr, ok := class.attributeByKey("id")
	suite.Require().True(ok)
	suite.Equal("ID", inputIDAttr.Name)
	suite.Equal("int", inputIDAttr.DataTypeRules)

	// Verify state machine
	suite.Require().NotNil(class.StateMachine)
	suite.Require().Contains(class.StateMachine.States, "pending")
	suite.Require().Contains(class.StateMachine.Events, "confirm")

	// Verify generalization
	suite.Require().Contains(subdomain.ClassGeneralizations, "product_types")
	gen := subdomain.ClassGeneralizations["product_types"]
	suite.Equal("product", gen.SuperclassKey)

	// Verify association
	suite.Require().Contains(subdomain.ClassAssociations, "order--line_item--order_lines")
	assoc := subdomain.ClassAssociations["order--line_item--order_lines"]
	suite.Equal("order", assoc.FromClassKey)
	suite.Equal("1..*", assoc.ToMultiplicity)
}

// TestConvertParameterTypeSpecRoundTrip verifies action parameter type_spec survives model conversion.
func (suite *ConvertSuite) TestConvertParameterTypeSpecRoundTrip() {
	original := &inputModel{
		Name:              "Test Model",
		Actors:            make(map[string]*inputActor),
		ClassAssociations: make(map[string]*inputClassAssociation),
		Domains: map[string]*inputDomain{
			"test": {
				Name: "Test",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"widget": {
								Name:       "Widget",
								Attributes: nil,
								Actions: map[string]*inputAction{
									"adjust": {
										Name: "Adjust",
										Parameters: []inputParameter{
											{
												Name:          "amount",
												DataTypeRules: "unconstrained",
												TypeSpec:      "Nat",
											},
										},
									},
								},
								Queries: make(map[string]*inputQuery),
							},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
	}

	model, err := ConvertToModel(original, "testmodel")
	suite.Require().NoError(err)

	var action model_state.Action
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, a := range class.Actions {
					action = a
				}
			}
		}
	}
	suite.Require().Len(action.Parameters, 1)
	suite.Require().NotNil(action.Parameters[0].DataType)
	suite.Require().NotNil(action.Parameters[0].DataType.TypeSpec)
	suite.Equal("Nat", action.Parameters[0].DataType.TypeSpec.Specification)

	result, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	actionInput := result.Domains["test"].Subdomains["default"].Classes["widget"].Actions["adjust"]
	suite.Require().Len(actionInput.Parameters, 1)
	suite.Equal("Nat", actionInput.Parameters[0].TypeSpec)
}

// TestConvertFromModelValidationError tests that validation errors from source model are returned.
func (suite *ConvertSuite) TestConvertFromModelValidationError() {
	model := &core.Model{
		Key:  "", // Invalid - empty key
		Name: "Test Model",
	}

	_, err := ConvertFromModel(model)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "validation failed")
}

// TestConvertToModelValidationError tests that req_model validation catches errors
// when there are issues not caught by tree validation (safety net).
// Note: Since tree validation now runs in readModelTree before ConvertToModel is called,
// the error here comes from core.Validate() as a safety net.
func (suite *ConvertSuite) TestConvertToModelValidationError() {
	input := &inputModel{
		Name: "Test Model",
		Actors: map[string]*inputActor{
			"customer": {Name: "Customer", Type: "person"},
		},
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order": {
								Name:       "Order",
								ActorKey:   "nonexistent_actor", // Invalid - references missing actor
								Attributes: nil,
							},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	_, err := ConvertToModel(input, "testmodel")
	suite.Require().Error(err)
	var pe *ParseError
	suite.Require().ErrorAs(err, &pe, "error should be a ParseError")
	suite.Equal(ErrConvReferenceNotFound, pe.Code, "CLASS_ACTOR_NOTFOUND should map to ErrConvReferenceNotFound")
	suite.Contains(pe.Message, "non-existent actor")
}

// TestConvertFromModelWithDomainAssociation tests converting a domain-level association.
func (suite *ConvertSuite) TestConvertFromModelWithDomainAssociation() {
	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "core"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domainKey, "shipping"))
	orderKey := helper.Must(identity.NewClassKey(subdomain1Key, "order"))
	shipmentKey := helper.Must(identity.NewClassKey(subdomain2Key, "shipment"))
	assocKey := helper.Must(identity.NewClassAssociationKey(domainKey, orderKey, shipmentKey, "order_shipments"))

	// Build classes
	orderClass := model_class.NewClass(orderKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes(nil)

	shipmentClass := model_class.NewClass(shipmentKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Shipment", Details: "", UnfinishedNotes: "", UmlComment: ""})
	shipmentClass.SetAttributes(nil)

	// Build subdomains
	subdomain1 := model_domain.NewSubdomain(subdomain1Key, "Core", "", "", "")
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := model_domain.NewSubdomain(subdomain2Key, "Shipping", "", "", "")
	subdomain2.Classes = map[identity.Key]model_class.Class{shipmentKey: shipmentClass}

	// Build association
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "Order Shipments", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))}, model_class.AssociationEnd{ClassKey: shipmentKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))}, model_class.Multiplicity{}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
		subdomain2Key: subdomain2,
	}
	domain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	inputDomain := input.Domains["orders"]
	suite.Require().Contains(inputDomain.ClassAssociations, "core.order--shipping.shipment--order_shipments")

	inputAssoc := inputDomain.ClassAssociations["core.order--shipping.shipment--order_shipments"]
	suite.Equal("Order Shipments", inputAssoc.Name)
	suite.Equal("core/order", inputAssoc.FromClassKey)
	suite.Equal("shipping/shipment", inputAssoc.ToClassKey)
}

// TestConvertToModelWithDomainAssociation tests converting a domain-level association.
func (suite *ConvertSuite) TestConvertToModelWithDomainAssociation() {
	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"core": {
						Name: "Core",
						Classes: map[string]*inputClass{
							"order": {Name: "Order", Attributes: nil},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
					"shipping": {
						Name: "Shipping",
						Classes: map[string]*inputClass{
							"shipment": {Name: "Shipment", Attributes: nil},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: map[string]*inputClassAssociation{
					"order_shipments": {
						Name:             "Order Shipments",
						FromClassKey:     "core/order",
						FromMultiplicity: "1",
						ToClassKey:       "shipping/shipment",
						ToMultiplicity:   "*",

						Uniqueness: "any",
					},
				},
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	// Find the domain-level association
	var domain model_domain.Domain
	for key, d := range model.Domains {
		if key.SubKey == "orders" {
			domain = d
			break
		}
	}
	suite.Require().Len(domain.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range domain.ClassAssociations {
		assoc = a
		break
	}
	suite.Equal("Order Shipments", assoc.Name)
}

// TestConvertFromModelWithModelAssociation tests converting a model-level association.
func (suite *ConvertSuite) TestConvertFromModelWithModelAssociation() {
	domain1Key := helper.Must(identity.NewDomainKey("orders"))
	domain2Key := helper.Must(identity.NewDomainKey("inventory"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "default"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "default"))
	orderKey := helper.Must(identity.NewClassKey(subdomain1Key, "order"))
	productKey := helper.Must(identity.NewClassKey(subdomain2Key, "product"))
	assocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, orderKey, productKey, "order_products"))

	// Build classes
	orderClass := model_class.NewClass(orderKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	orderClass.SetAttributes(nil)

	productClass := model_class.NewClass(productKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	productClass.SetAttributes(nil)

	// Build subdomains
	subdomain1 := model_domain.NewSubdomain(subdomain1Key, "Core", "", "", "")
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := model_domain.NewSubdomain(subdomain2Key, "Products", "", "", "")
	subdomain2.Classes = map[identity.Key]model_class.Class{productKey: productClass}

	// Build domains
	domain1 := model_domain.NewDomain(domain1Key, "Orders", "", "", false, "")
	domain1.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
	}

	domain2 := model_domain.NewDomain(domain2Key, "Inventory", "", "", false, "")
	domain2.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain2Key: subdomain2,
	}

	// Build association
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "Order Products", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))}, model_class.AssociationEnd{ClassKey: productKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))}, model_class.Multiplicity{}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	// Build model
	m := core.NewModel("testmodel", core.ModelDetails{Name: "Test Model", Details: ""}, "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domain1Key: domain1,
		domain2Key: domain2,
	}
	m.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	suite.Require().Contains(input.ClassAssociations, "orders.default.order--inventory.default.product--order_products")

	inputAssoc := input.ClassAssociations["orders.default.order--inventory.default.product--order_products"]
	suite.Equal("Order Products", inputAssoc.Name)
	suite.Equal("orders/default/order", inputAssoc.FromClassKey)
	suite.Equal("inventory/default/product", inputAssoc.ToClassKey)
}

// TestConvertToModelWithModelAssociation tests converting a model-level association.
func (suite *ConvertSuite) TestConvertToModelWithModelAssociation() {
	input := &inputModel{
		Name:   "Test Model",
		Actors: make(map[string]*inputActor),
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Core",
						Classes: map[string]*inputClass{
							"order": {Name: "Order", Attributes: nil},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
			"inventory": {
				Name: "Inventory",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Products",
						Classes: map[string]*inputClass{
							"product": {Name: "Product", Attributes: nil},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
				},
				ClassAssociations: make(map[string]*inputClassAssociation),
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{
			"order_products": {
				Name:             "Order Products",
				FromClassKey:     "orders/default/order",
				FromMultiplicity: "1",
				ToClassKey:       "inventory/default/product",
				ToMultiplicity:   "*",

				Uniqueness: "any",
			},
		},
	}

	model, err := ConvertToModel(input, "testmodel")
	suite.Require().NoError(err)

	suite.Require().Len(model.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range model.ClassAssociations {
		assoc = a
		break
	}
	suite.Equal("Order Products", assoc.Name)
}

// TestConvertTransitionFromModelInitial tests that initial transitions (nil FromStateKey) produce no from_state_key.
func (suite *ConvertSuite) TestConvertTransitionFromModelInitial() {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(
			helper.Must(identity.NewDomainKey("d")), "s")), "c"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "start"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "active"))
	transitionKey := helper.Must(identity.NewTransitionKey(classKey, "", "start", "", "", "active"))

	transition := model_state.NewTransition(transitionKey, eventKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &toStateKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")
	result := convertTransitionFromModel(&transition)

	suite.Nil(result.FromStateKey, "initial transition should have nil FromStateKey")
	suite.Require().NotNil(result.ToStateKey)
	suite.Equal("active", *result.ToStateKey)
	suite.Equal("start", result.EventKey)
}

// TestConvertTransitionFromModelFinal tests that final transitions (nil ToStateKey) produce no to_state_key.
func (suite *ConvertSuite) TestConvertTransitionFromModelFinal() {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(
			helper.Must(identity.NewDomainKey("d")), "s")), "c"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "active"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "stop"))
	transitionKey := helper.Must(identity.NewTransitionKey(classKey, "active", "stop", "", "", ""))

	transition := model_state.NewTransition(transitionKey, eventKey, model_state.TransitionStateKeys{FromStateKey: &fromStateKey, ToStateKey: nil}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")
	result := convertTransitionFromModel(&transition)

	suite.Require().NotNil(result.FromStateKey)
	suite.Equal("active", *result.FromStateKey)
	suite.Nil(result.ToStateKey, "final transition should have nil ToStateKey")
	suite.Equal("stop", result.EventKey)
}

// TestConvertTransitionFromModelNamedInitialState tests that a real state named "initial" is preserved.
func (suite *ConvertSuite) TestConvertTransitionFromModelNamedInitialState() {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(
			helper.Must(identity.NewDomainKey("d")), "s")), "c"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "initial"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "go"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "running"))
	transitionKey := helper.Must(identity.NewTransitionKey(classKey, "initial", "go", "", "", "running"))

	transition := model_state.NewTransition(transitionKey, eventKey, model_state.TransitionStateKeys{FromStateKey: &fromStateKey, ToStateKey: &toStateKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")
	result := convertTransitionFromModel(&transition)

	suite.Require().NotNil(result.FromStateKey, "real state named 'initial' must not be stripped")
	suite.Equal("initial", *result.FromStateKey)
	suite.Require().NotNil(result.ToStateKey)
	suite.Equal("running", *result.ToStateKey)
}

// TestConvertMultiplicityFormats tests various multiplicity format conversions.
func (suite *ConvertSuite) TestConvertMultiplicityFormats() {
	tests := []struct {
		mult     model_class.Multiplicity
		expected string
	}{
		{helper.Must(model_class.NewMultiplicity("1")), "1"},
		{helper.Must(model_class.NewMultiplicity("0..1")), "0..1"},
		{helper.Must(model_class.NewMultiplicity("any")), "*"},
		{helper.Must(model_class.NewMultiplicity("1..many")), "1..*"},
		{helper.Must(model_class.NewMultiplicity("2..5")), "2..5"},
		{helper.Must(model_class.NewMultiplicity("3")), "3"},
	}

	for _, tt := range tests {
		suite.Run(tt.expected, func() {
			suite.Equal(tt.expected, tt.mult.String())
		})
	}
}
