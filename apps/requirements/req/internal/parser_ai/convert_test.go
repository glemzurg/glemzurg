package parser_ai

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConvertSuite tests the conversion functions between req_model and inputModel.
type ConvertSuite struct {
	suite.Suite
}

func TestConvertSuite(t *testing.T) {
	suite.Run(t, new(ConvertSuite))
}

// TestConvertFromModelMinimal tests converting a minimal valid req_model.Model to inputModel.
func (suite *ConvertSuite) TestConvertFromModelMinimal() {
	t := suite.T()

	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "Model details", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = make(map[identity.Key]model_domain.Domain)
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)
	assert.Equal(t, "Test Model", input.Name)
	assert.Equal(t, "Model details", input.Details)
	assert.Empty(t, input.Actors)
	assert.Empty(t, input.Domains)
	assert.Empty(t, input.ClassAssociations)
}

// TestConvertToModelMinimal tests converting a minimal inputModel to req_model.Model.
func (suite *ConvertSuite) TestConvertToModelMinimal() {
	t := suite.T()

	input := &inputModel{
		Name:              "Test Model",
		Details:           "Model details",
		Actors:            make(map[string]*inputActor),
		Domains:           make(map[string]*inputDomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	require.NoError(t, err)
	assert.Equal(t, "testmodel", model.Key)
	assert.Equal(t, "Test Model", model.Name)
	assert.Equal(t, "Model details", model.Details)
	assert.Empty(t, model.Actors)
	assert.Empty(t, model.Domains)
}

// TestConvertFromModelWithActor tests converting an actor.
func (suite *ConvertSuite) TestConvertFromModelWithActor() {
	t := suite.T()

	actorKey := helper.Must(identity.NewActorKey("customer"))

	actor := helper.Must(model_actor.NewActor(actorKey, "Customer", "Customer details", "person", nil, nil, ""))

	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = map[identity.Key]model_actor.Actor{
		actorKey: actor,
	}
	m.Domains = make(map[identity.Key]model_domain.Domain)
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)
	require.Contains(t, input.Actors, "customer")
	assert.Equal(t, "Customer", input.Actors["customer"].Name)
	assert.Equal(t, "person", input.Actors["customer"].Type)
	assert.Equal(t, "Customer details", input.Actors["customer"].Details)
}

// TestConvertToModelWithActor tests converting an actor.
func (suite *ConvertSuite) TestConvertToModelWithActor() {
	t := suite.T()

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
	require.NoError(t, err)
	require.Len(t, model.Actors, 1)

	// Find the actor by checking the key's SubKey
	var foundActor model_actor.Actor
	for key, actor := range model.Actors {
		if key.SubKey == "customer" {
			foundActor = actor
			break
		}
	}
	assert.Equal(t, "Customer", foundActor.Name)
	assert.Equal(t, "person", foundActor.Type)
}

// TestConvertFromModelWithClass tests converting a class with attributes.
func (suite *ConvertSuite) TestConvertFromModelWithClass() {
	t := suite.T()

	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	actorKey := helper.Must(identity.NewActorKey("customer"))

	idAttrKey := helper.Must(identity.NewAttributeKey(classKey, "id"))
	statusAttrKey := helper.Must(identity.NewAttributeKey(classKey, "status"))

	// Build attributes
	idAttr := helper.Must(model_class.NewAttribute(idAttrKey, "ID", "The order ID", "int", nil, false, "", []uint{0}))
	statusAttr := helper.Must(model_class.NewAttribute(statusAttrKey, "Status", "", "string", nil, false, "", nil))

	// Build class
	orderClass := helper.Must(model_class.NewClass(classKey, "Order", "Order details", &actorKey, nil, nil, ""))
	orderClass.SetAttributes(map[identity.Key]model_class.Attribute{
		idAttrKey:     idAttr,
		statusAttrKey: statusAttr,
	})

	// Build subdomain
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Default", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build actor
	actor := helper.Must(model_actor.NewActor(actorKey, "Customer", "", "person", nil, nil, ""))

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = map[identity.Key]model_actor.Actor{
		actorKey: actor,
	}
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	require.Contains(t, input.Domains, "orders")
	require.Contains(t, input.Domains["orders"].Subdomains, "default")
	require.Contains(t, input.Domains["orders"].Subdomains["default"].Classes, "order")

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	assert.Equal(t, "Order", class.Name)
	assert.Equal(t, "Order details", class.Details)
	assert.Equal(t, "customer", class.ActorKey)
	require.Contains(t, class.Attributes, "id")
	assert.Equal(t, "ID", class.Attributes["id"].Name)
	assert.Equal(t, "int", class.Attributes["id"].DataTypeRules)
}

// TestConvertToModelWithClass tests converting a class with attributes.
func (suite *ConvertSuite) TestConvertToModelWithClass() {
	t := suite.T()

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
								Attributes: map[string]*inputAttribute{
									"id":     {Name: "ID", DataTypeRules: "int"},
									"status": {Name: "Status", DataTypeRules: "string"},
								},
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
	require.NoError(t, err)

	// Find domain
	var domain model_domain.Domain
	for key, d := range model.Domains {
		if key.SubKey == "orders" {
			domain = d
			break
		}
	}
	require.NotEmpty(t, domain.Name)

	// Find subdomain
	var subdomain model_domain.Subdomain
	for key, s := range domain.Subdomains {
		if key.SubKey == "default" {
			subdomain = s
			break
		}
	}
	require.NotEmpty(t, subdomain.Name)

	// Find class
	var class model_class.Class
	for key, c := range subdomain.Classes {
		if key.SubKey == "order" {
			class = c
			break
		}
	}
	assert.Equal(t, "Order", class.Name)
	assert.Equal(t, "Order details", class.Details)
}

// TestConvertFromModelWithStateMachine tests converting a state machine.
func (suite *ConvertSuite) TestConvertFromModelWithStateMachine() {
	t := suite.T()

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
	state1 := helper.Must(model_state.NewState(stateKey1, "Pending", "", ""))
	state2 := helper.Must(model_state.NewState(stateKey2, "Confirmed", "", ""))

	// Build event
	event := helper.Must(model_state.NewEvent(eventKey, "confirm", "", nil))

	// Build guard with logic
	guardLogic := helper.Must(model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Check if order has items", "", model_logic.NotationTLAPlus, "", nil))
	guard := helper.Must(model_state.NewGuard(guardKey, "has_items", guardLogic))

	// Build transition
	transition := helper.Must(model_state.NewTransition(transitionKey, &stateKey1, eventKey, &guardKey, &actionKey, &stateKey2, ""))

	// Build action
	action := helper.Must(model_state.NewAction(actionKey, "Process", "Process the order", nil, nil, nil, nil))

	// Build class
	orderClass := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))
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
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Default", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	require.NotNil(t, class.StateMachine)

	sm := class.StateMachine
	require.Contains(t, sm.States, "pending")
	require.Contains(t, sm.States, "confirmed")
	require.Contains(t, sm.Events, "confirm")
	require.Contains(t, sm.Guards, "has_items")
	require.Len(t, sm.Transitions, 1)

	trans := sm.Transitions[0]
	assert.Equal(t, "pending", *trans.FromStateKey)
	assert.Equal(t, "confirmed", *trans.ToStateKey)
	assert.Equal(t, "confirm", trans.EventKey)
	assert.Equal(t, "has_items", *trans.GuardKey)
	assert.Equal(t, "process", *trans.ActionKey)
}

// TestConvertToModelWithStateMachine tests converting a state machine.
func (suite *ConvertSuite) TestConvertToModelWithStateMachine() {
	t := suite.T()

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
								Attributes: make(map[string]*inputAttribute),
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
	require.NoError(t, err)

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
	assert.Len(t, class.States, 2)
	assert.Len(t, class.Events, 1)
	assert.Len(t, class.Guards, 1)
	assert.Len(t, class.Transitions, 1)
}

// TestConvertFromModelWithQueries tests converting queries.
func (suite *ConvertSuite) TestConvertFromModelWithQueries() {
	t := suite.T()

	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "get_total"))

	// Build query logic
	requireKey := helper.Must(identity.NewQueryRequireKey(queryKey, "0"))
	guaranteeKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	requireLogic := helper.Must(model_logic.NewLogic(requireKey, model_logic.LogicTypeAssessment, "order must exist", "", model_logic.NotationTLAPlus, "", nil))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeQuery, "returns total amount", "total", model_logic.NotationTLAPlus, "", nil))

	// Build query
	query := helper.Must(model_state.NewQuery(queryKey, "Get Total", "Get order total",
		[]model_logic.Logic{requireLogic},
		[]model_logic.Logic{guaranteeLogic},
		nil,
	))

	// Build class
	orderClass := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))
	orderClass.SetQueries(map[identity.Key]model_state.Query{
		queryKey: query,
	})

	// Build subdomain
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Default", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	class := input.Domains["orders"].Subdomains["default"].Classes["order"]
	require.Contains(t, class.Queries, "get_total")

	inputQuery := class.Queries["get_total"]
	assert.Equal(t, "Get Total", inputQuery.Name)
	assert.Equal(t, "Get order total", inputQuery.Details)
	require.Len(t, inputQuery.Requires, 1)
	assert.Equal(t, "order must exist", inputQuery.Requires[0].Description)
	require.Len(t, inputQuery.Guarantees, 1)
	assert.Equal(t, "returns total amount", inputQuery.Guarantees[0].Description)
}

// TestConvertToModelWithQueries tests converting queries.
func (suite *ConvertSuite) TestConvertToModelWithQueries() {
	t := suite.T()

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
								Attributes: make(map[string]*inputAttribute),
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
	require.NoError(t, err)

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
	require.Len(t, class.Queries, 1)

	var query model_state.Query
	for _, q := range class.Queries {
		query = q
		break
	}
	assert.Equal(t, "Get Total", query.Name)
	assert.Equal(t, "Get order total", query.Details)
	require.Len(t, query.Requires, 1)
	assert.Equal(t, "order must exist", query.Requires[0].Description)
	require.Len(t, query.Guarantees, 1)
	assert.Equal(t, "returns total amount", query.Guarantees[0].Description)
}

// TestConvertFromModelWithGeneralization tests converting a generalization.
func (suite *ConvertSuite) TestConvertFromModelWithGeneralization() {
	t := suite.T()

	domainKey := helper.Must(identity.NewDomainKey("products"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	productKey := helper.Must(identity.NewClassKey(subdomainKey, "product"))
	bookKey := helper.Must(identity.NewClassKey(subdomainKey, "book"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "product_types"))

	// Build classes
	productClass := helper.Must(model_class.NewClass(productKey, "Product", "", nil, &genKey, nil, ""))
	productClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	bookClass := helper.Must(model_class.NewClass(bookKey, "Book", "", nil, nil, &genKey, ""))
	bookClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build generalization
	gen := helper.Must(model_class.NewGeneralization(genKey, "Product Types", "Types of products", false, false, ""))

	// Build subdomain
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Default", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		productKey: productClass,
		bookKey:    bookClass,
	}
	subdomain.Generalizations = map[identity.Key]model_class.Generalization{
		genKey: gen,
	}

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Products", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	inputSubdomain := input.Domains["products"].Subdomains["default"]
	require.Contains(t, inputSubdomain.ClassGeneralizations, "product_types")

	inputGen := inputSubdomain.ClassGeneralizations["product_types"]
	assert.Equal(t, "Product Types", inputGen.Name)
	assert.Equal(t, "product", inputGen.SuperclassKey)
	assert.Equal(t, []string{"book"}, inputGen.SubclassKeys)
}

// TestConvertToModelWithGeneralization tests converting a generalization.
func (suite *ConvertSuite) TestConvertToModelWithGeneralization() {
	t := suite.T()

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
							"product": {Name: "Product", Attributes: make(map[string]*inputAttribute)},
							"book":    {Name: "Book", Attributes: make(map[string]*inputAttribute)},
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
	require.NoError(t, err)

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
	require.Len(t, subdomain.Generalizations, 1)

	var gen model_class.Generalization
	for _, g := range subdomain.Generalizations {
		gen = g
		break
	}
	assert.Equal(t, "Product Types", gen.Name)

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
	require.NotNil(t, productClass.SuperclassOfKey)
	assert.Equal(t, gen.Key, *productClass.SuperclassOfKey)
	// Book class should be a subclass of the generalization
	require.NotNil(t, bookClass.SubclassOfKey)
	assert.Equal(t, gen.Key, *bookClass.SubclassOfKey)
}

// TestConvertFromModelWithSubdomainAssociation tests converting a subdomain-level association.
func (suite *ConvertSuite) TestConvertFromModelWithSubdomainAssociation() {
	t := suite.T()

	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	orderKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	lineItemKey := helper.Must(identity.NewClassKey(subdomainKey, "line_item"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, orderKey, lineItemKey, "order_lines"))

	// Build classes
	orderClass := helper.Must(model_class.NewClass(orderKey, "Order", "", nil, nil, nil, ""))
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	lineItemClass := helper.Must(model_class.NewClass(lineItemKey, "Line Item", "", nil, nil, nil, ""))
	lineItemClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build association
	assoc := helper.Must(model_class.NewAssociation(
		assocKey, "Order Lines", "",
		orderKey, helper.Must(model_class.NewMultiplicity("1")),
		lineItemKey, helper.Must(model_class.NewMultiplicity("1..many")),
		nil, "",
	))

	// Build subdomain
	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "Default", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		orderKey:    orderClass,
		lineItemKey: lineItemClass,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	inputSubdomain := input.Domains["orders"].Subdomains["default"]
	require.Contains(t, inputSubdomain.ClassAssociations, "order_lines")

	inputAssoc := inputSubdomain.ClassAssociations["order_lines"]
	assert.Equal(t, "Order Lines", inputAssoc.Name)
	assert.Equal(t, "order", inputAssoc.FromClassKey)
	assert.Equal(t, "1", inputAssoc.FromMultiplicity)
	assert.Equal(t, "line_item", inputAssoc.ToClassKey)
	assert.Equal(t, "1..*", inputAssoc.ToMultiplicity)
}

// TestConvertToModelWithSubdomainAssociation tests converting a subdomain-level association.
func (suite *ConvertSuite) TestConvertToModelWithSubdomainAssociation() {
	t := suite.T()

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
							"order":     {Name: "Order", Attributes: make(map[string]*inputAttribute)},
							"line_item": {Name: "Line Item", Attributes: make(map[string]*inputAttribute)},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations: map[string]*inputClassAssociation{
							"order_lines": {
								Name:             "Order Lines",
								FromClassKey:     "order",
								FromMultiplicity: "1",
								ToClassKey:       "line_item",
								ToMultiplicity:   "1..*",
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
	require.NoError(t, err)

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
	require.Len(t, subdomain.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range subdomain.ClassAssociations {
		assoc = a
		break
	}
	assert.Equal(t, "Order Lines", assoc.Name)
	assert.Equal(t, "order", assoc.FromClassKey.SubKey)
	assert.Equal(t, "line_item", assoc.ToClassKey.SubKey)
}

// TestRoundTripMinimal tests that a minimal model survives roundtrip conversion.
func (suite *ConvertSuite) TestRoundTripMinimal() {
	t := suite.T()

	original := &inputModel{
		Name:              "Test Model",
		Details:           "Model details",
		Actors:            make(map[string]*inputActor),
		Domains:           make(map[string]*inputDomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	// Convert to req_model
	model, err := ConvertToModel(original, "testmodel")
	require.NoError(t, err)

	// Convert back to inputModel
	result, err := ConvertFromModel(model)
	require.NoError(t, err)

	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Details, result.Details)
}

// TestRoundTripComplete tests that a complete model survives roundtrip conversion.
func (suite *ConvertSuite) TestRoundTripComplete() {
	t := suite.T()

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
								Attributes: map[string]*inputAttribute{
									"id":     {Name: "ID", Details: "Order ID", DataTypeRules: "int"},
									"status": {Name: "Status", DataTypeRules: "string"},
								},
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
								Attributes: make(map[string]*inputAttribute),
							},
							"product": {
								Name:       "Product",
								Attributes: make(map[string]*inputAttribute),
							},
							"book": {
								Name:       "Book",
								Attributes: make(map[string]*inputAttribute),
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
	require.NoError(t, err)

	// Convert back to inputModel
	result, err := ConvertFromModel(model)
	require.NoError(t, err)

	// Verify top-level fields
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Details, result.Details)

	// Verify actor
	require.Contains(t, result.Actors, "customer")
	assert.Equal(t, original.Actors["customer"].Name, result.Actors["customer"].Name)
	assert.Equal(t, original.Actors["customer"].Type, result.Actors["customer"].Type)

	// Verify domain structure
	require.Contains(t, result.Domains, "orders")
	assert.Equal(t, original.Domains["orders"].Name, result.Domains["orders"].Name)

	// Verify subdomain
	require.Contains(t, result.Domains["orders"].Subdomains, "default")
	subdomain := result.Domains["orders"].Subdomains["default"]
	assert.Equal(t, "Default", subdomain.Name)

	// Verify class
	require.Contains(t, subdomain.Classes, "order")
	class := subdomain.Classes["order"]
	assert.Equal(t, "Order", class.Name)
	assert.Equal(t, "customer", class.ActorKey)

	// Verify attributes
	require.Contains(t, class.Attributes, "id")
	assert.Equal(t, "ID", class.Attributes["id"].Name)
	assert.Equal(t, "int", class.Attributes["id"].DataTypeRules)

	// Verify state machine
	require.NotNil(t, class.StateMachine)
	require.Contains(t, class.StateMachine.States, "pending")
	require.Contains(t, class.StateMachine.Events, "confirm")

	// Verify generalization
	require.Contains(t, subdomain.ClassGeneralizations, "product_types")
	gen := subdomain.ClassGeneralizations["product_types"]
	assert.Equal(t, "product", gen.SuperclassKey)

	// Verify association
	require.Contains(t, subdomain.ClassAssociations, "order_lines")
	assoc := subdomain.ClassAssociations["order_lines"]
	assert.Equal(t, "order", assoc.FromClassKey)
	assert.Equal(t, "1..*", assoc.ToMultiplicity)
}

// TestConvertFromModelValidationError tests that validation errors from source model are returned.
func (suite *ConvertSuite) TestConvertFromModelValidationError() {
	t := suite.T()

	model := &req_model.Model{
		Key:  "", // Invalid - empty key
		Name: "Test Model",
	}

	_, err := ConvertFromModel(model)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// TestConvertToModelValidationError tests that req_model validation catches errors
// when there are issues not caught by tree validation (safety net).
// Note: Since tree validation now runs in readModelTree before ConvertToModel is called,
// the error here comes from req_model.Validate() as a safety net.
func (suite *ConvertSuite) TestConvertToModelValidationError() {
	t := suite.T()

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
								Attributes: make(map[string]*inputAttribute),
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// TestConvertFromModelWithDomainAssociation tests converting a domain-level association.
func (suite *ConvertSuite) TestConvertFromModelWithDomainAssociation() {
	t := suite.T()

	domainKey := helper.Must(identity.NewDomainKey("orders"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "core"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domainKey, "shipping"))
	orderKey := helper.Must(identity.NewClassKey(subdomain1Key, "order"))
	shipmentKey := helper.Must(identity.NewClassKey(subdomain2Key, "shipment"))
	assocKey := helper.Must(identity.NewClassAssociationKey(domainKey, orderKey, shipmentKey, "order_shipments"))

	// Build classes
	orderClass := helper.Must(model_class.NewClass(orderKey, "Order", "", nil, nil, nil, ""))
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	shipmentClass := helper.Must(model_class.NewClass(shipmentKey, "Shipment", "", nil, nil, nil, ""))
	shipmentClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build subdomains
	subdomain1 := helper.Must(model_domain.NewSubdomain(subdomain1Key, "Core", "", ""))
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := helper.Must(model_domain.NewSubdomain(subdomain2Key, "Shipping", "", ""))
	subdomain2.Classes = map[identity.Key]model_class.Class{shipmentKey: shipmentClass}

	// Build association
	assoc := helper.Must(model_class.NewAssociation(
		assocKey, "Order Shipments", "",
		orderKey, helper.Must(model_class.NewMultiplicity("1")),
		shipmentKey, helper.Must(model_class.NewMultiplicity("any")),
		nil, "",
	))

	// Build domain
	domain := helper.Must(model_domain.NewDomain(domainKey, "Orders", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
		subdomain2Key: subdomain2,
	}
	domain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	require.NoError(t, err)

	inputDomain := input.Domains["orders"]
	require.Contains(t, inputDomain.ClassAssociations, "order_shipments")

	inputAssoc := inputDomain.ClassAssociations["order_shipments"]
	assert.Equal(t, "Order Shipments", inputAssoc.Name)
	assert.Equal(t, "core/order", inputAssoc.FromClassKey)
	assert.Equal(t, "shipping/shipment", inputAssoc.ToClassKey)
}

// TestConvertToModelWithDomainAssociation tests converting a domain-level association.
func (suite *ConvertSuite) TestConvertToModelWithDomainAssociation() {
	t := suite.T()

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
							"order": {Name: "Order", Attributes: make(map[string]*inputAttribute)},
						},
						ClassGeneralizations: make(map[string]*inputClassGeneralization),
						ClassAssociations:    make(map[string]*inputClassAssociation),
					},
					"shipping": {
						Name: "Shipping",
						Classes: map[string]*inputClass{
							"shipment": {Name: "Shipment", Attributes: make(map[string]*inputAttribute)},
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
					},
				},
			},
		},
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	model, err := ConvertToModel(input, "testmodel")
	require.NoError(t, err)

	// Find the domain-level association
	var domain model_domain.Domain
	for key, d := range model.Domains {
		if key.SubKey == "orders" {
			domain = d
			break
		}
	}
	require.Len(t, domain.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range domain.ClassAssociations {
		assoc = a
		break
	}
	assert.Equal(t, "Order Shipments", assoc.Name)
}

// TestConvertFromModelWithModelAssociation tests converting a model-level association.
func (suite *ConvertSuite) TestConvertFromModelWithModelAssociation() {
	t := suite.T()

	domain1Key := helper.Must(identity.NewDomainKey("orders"))
	domain2Key := helper.Must(identity.NewDomainKey("inventory"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "default"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "default"))
	orderKey := helper.Must(identity.NewClassKey(subdomain1Key, "order"))
	productKey := helper.Must(identity.NewClassKey(subdomain2Key, "product"))
	assocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, orderKey, productKey, "order_products"))

	// Build classes
	orderClass := helper.Must(model_class.NewClass(orderKey, "Order", "", nil, nil, nil, ""))
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	productClass := helper.Must(model_class.NewClass(productKey, "Product", "", nil, nil, nil, ""))
	productClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build subdomains
	subdomain1 := helper.Must(model_domain.NewSubdomain(subdomain1Key, "Core", "", ""))
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := helper.Must(model_domain.NewSubdomain(subdomain2Key, "Products", "", ""))
	subdomain2.Classes = map[identity.Key]model_class.Class{productKey: productClass}

	// Build domains
	domain1 := helper.Must(model_domain.NewDomain(domain1Key, "Orders", "", false, ""))
	domain1.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
	}

	domain2 := helper.Must(model_domain.NewDomain(domain2Key, "Inventory", "", false, ""))
	domain2.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain2Key: subdomain2,
	}

	// Build association
	assoc := helper.Must(model_class.NewAssociation(
		assocKey, "Order Products", "",
		orderKey, helper.Must(model_class.NewMultiplicity("1")),
		productKey, helper.Must(model_class.NewMultiplicity("any")),
		nil, "",
	))

	// Build model
	m := helper.Must(req_model.NewModel("testmodel", "Test Model", "", nil, nil))
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
	require.NoError(t, err)

	require.Contains(t, input.ClassAssociations, "order_products")

	inputAssoc := input.ClassAssociations["order_products"]
	assert.Equal(t, "Order Products", inputAssoc.Name)
	assert.Equal(t, "orders/default/order", inputAssoc.FromClassKey)
	assert.Equal(t, "inventory/default/product", inputAssoc.ToClassKey)
}

// TestConvertToModelWithModelAssociation tests converting a model-level association.
func (suite *ConvertSuite) TestConvertToModelWithModelAssociation() {
	t := suite.T()

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
							"order": {Name: "Order", Attributes: make(map[string]*inputAttribute)},
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
							"product": {Name: "Product", Attributes: make(map[string]*inputAttribute)},
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
			},
		},
	}

	model, err := ConvertToModel(input, "testmodel")
	require.NoError(t, err)

	require.Len(t, model.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range model.ClassAssociations {
		assoc = a
		break
	}
	assert.Equal(t, "Order Products", assoc.Name)
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
		suite.T().Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mult.String())
		})
	}
}
