package parser_ai

import (
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

// TestConvertFromModelMinimal tests converting a minimal valid core.Model to inputModel.
func (suite *ConvertSuite) TestConvertFromModelMinimal() {
	m := core.NewModel("testmodel", "Test Model", "Model details", nil, nil, nil)
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

	actor := model_actor.NewActor(actorKey, "Customer", "Customer details", "person", nil, nil, "")

	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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
	idAttr := helper.Must(model_class.NewAttribute(idAttrKey, "ID", "The order ID", "int", nil, false,
		model_class.AttributeAnnotations{IndexNums: []uint{0}}))
	statusAttr := helper.Must(model_class.NewAttribute(statusAttrKey, "Status", "", "string", nil, false,
		model_class.AttributeAnnotations{}))

	// Build class
	orderClass := model_class.NewClass(classKey, "Order", "Order details", &actorKey, nil, nil, "")
	orderClass.SetAttributes(map[identity.Key]model_class.Attribute{
		idAttrKey:     idAttr,
		statusAttrKey: statusAttr,
	})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build actor
	actor := model_actor.NewActor(actorKey, "Customer", "", "person", nil, nil, "")

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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
	suite.Require().Contains(class.Attributes, "id")
	suite.Equal("ID", class.Attributes["id"].Name)
	suite.Equal("int", class.Attributes["id"].DataTypeRules)
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
	transition := model_state.NewTransition(transitionKey, &stateKey1, eventKey, &guardKey, &actionKey, &stateKey2, "")

	// Build action
	action := model_state.NewAction(actionKey, "Process", "Process the order", nil, nil, nil, nil)

	// Build class
	orderClass := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
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
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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
	orderClass := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))
	orderClass.SetQueries(map[identity.Key]model_state.Query{
		queryKey: query,
	})

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: orderClass,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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
	productClass := model_class.NewClass(productKey, "Product", "", nil, &genKey, nil, "")
	productClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	bookClass := model_class.NewClass(bookKey, "Book", "", nil, nil, &genKey, "")
	bookClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build generalization
	gen := model_class.NewGeneralization(genKey, "Product Types", "Types of products", false, false, "")

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		productKey: productClass,
		bookKey:    bookClass,
	}
	subdomain.Generalizations = map[identity.Key]model_class.Generalization{
		genKey: gen,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Products", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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
	orderClass := model_class.NewClass(orderKey, "Order", "", nil, nil, nil, "")
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	lineItemClass := model_class.NewClass(lineItemKey, "Line Item", "", nil, nil, nil, "")
	lineItemClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build association
	assoc := model_class.NewAssociation(
		assocKey, "Order Lines", "",
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: lineItemKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1..many"))},
		nil, "",
	)

	// Build subdomain
	subdomain := model_domain.NewSubdomain(subdomainKey, "Default", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		orderKey:    orderClass,
		lineItemKey: lineItemClass,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	inputSubdomain := input.Domains["orders"].Subdomains["default"]
	suite.Require().Contains(inputSubdomain.ClassAssociations, "order_lines")

	inputAssoc := inputSubdomain.ClassAssociations["order_lines"]
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
	suite.Require().Contains(class.Attributes, "id")
	suite.Equal("ID", class.Attributes["id"].Name)
	suite.Equal("int", class.Attributes["id"].DataTypeRules)

	// Verify state machine
	suite.Require().NotNil(class.StateMachine)
	suite.Require().Contains(class.StateMachine.States, "pending")
	suite.Require().Contains(class.StateMachine.Events, "confirm")

	// Verify generalization
	suite.Require().Contains(subdomain.ClassGeneralizations, "product_types")
	gen := subdomain.ClassGeneralizations["product_types"]
	suite.Equal("product", gen.SuperclassKey)

	// Verify association
	suite.Require().Contains(subdomain.ClassAssociations, "order_lines")
	assoc := subdomain.ClassAssociations["order_lines"]
	suite.Equal("order", assoc.FromClassKey)
	suite.Equal("1..*", assoc.ToMultiplicity)
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
	suite.Require().Error(err)
	var pe *ParseError
	suite.Require().ErrorAs(err, &pe, "error should be a ParseError")
	suite.Equal(ErrConvModelValidation, pe.Code, "missing wrapping context should fall back to internal error")
	suite.Contains(pe.Message, "internal error")
	suite.Contains(pe.Hint, "CLASS_ACTOR_NOTFOUND")
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
	orderClass := model_class.NewClass(orderKey, "Order", "", nil, nil, nil, "")
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	shipmentClass := model_class.NewClass(shipmentKey, "Shipment", "", nil, nil, nil, "")
	shipmentClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build subdomains
	subdomain1 := model_domain.NewSubdomain(subdomain1Key, "Core", "", "")
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := model_domain.NewSubdomain(subdomain2Key, "Shipping", "", "")
	subdomain2.Classes = map[identity.Key]model_class.Class{shipmentKey: shipmentClass}

	// Build association
	assoc := model_class.NewAssociation(
		assocKey, "Order Shipments", "",
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: shipmentKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		nil, "",
	)

	// Build domain
	domain := model_domain.NewDomain(domainKey, "Orders", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
		subdomain2Key: subdomain2,
	}
	domain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
	m.Actors = make(map[identity.Key]model_actor.Actor)
	m.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	model := &m

	input, err := ConvertFromModel(model)
	suite.Require().NoError(err)

	inputDomain := input.Domains["orders"]
	suite.Require().Contains(inputDomain.ClassAssociations, "order_shipments")

	inputAssoc := inputDomain.ClassAssociations["order_shipments"]
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
	orderClass := model_class.NewClass(orderKey, "Order", "", nil, nil, nil, "")
	orderClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	productClass := model_class.NewClass(productKey, "Product", "", nil, nil, nil, "")
	productClass.SetAttributes(make(map[identity.Key]model_class.Attribute))

	// Build subdomains
	subdomain1 := model_domain.NewSubdomain(subdomain1Key, "Core", "", "")
	subdomain1.Classes = map[identity.Key]model_class.Class{orderKey: orderClass}

	subdomain2 := model_domain.NewSubdomain(subdomain2Key, "Products", "", "")
	subdomain2.Classes = map[identity.Key]model_class.Class{productKey: productClass}

	// Build domains
	domain1 := model_domain.NewDomain(domain1Key, "Orders", "", false, "")
	domain1.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain1Key: subdomain1,
	}

	domain2 := model_domain.NewDomain(domain2Key, "Inventory", "", false, "")
	domain2.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomain2Key: subdomain2,
	}

	// Build association
	assoc := model_class.NewAssociation(
		assocKey, "Order Products", "",
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: productKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		nil, "",
	)

	// Build model
	m := core.NewModel("testmodel", "Test Model", "", nil, nil, nil)
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

	suite.Require().Contains(input.ClassAssociations, "order_products")

	inputAssoc := input.ClassAssociations["order_products"]
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
	suite.Require().NoError(err)

	suite.Require().Len(model.ClassAssociations, 1)

	var assoc model_class.Association
	for _, a := range model.ClassAssociations {
		assoc = a
		break
	}
	suite.Equal("Order Products", assoc.Name)
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
