package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRoundTripSuite(t *testing.T) {
	suite.Run(t, new(RoundTripSuite))
}

type RoundTripSuite struct {
	suite.Suite
}

func (suite *RoundTripSuite) TestRoundTrip() {

	// -- Actor generalizations --
	genKeyA, err := identity.NewActorGeneralizationKey("gen_a")
	assert.Nil(suite.T(), err)
	genKeyB, err := identity.NewActorGeneralizationKey("gen_b")
	assert.Nil(suite.T(), err)

	genA, err := model_actor.NewGeneralization(genKeyA, "Generalization A", "## Generalization A\n\nDetails for gen A.", false, true, "")
	assert.Nil(suite.T(), err)
	genB, err := model_actor.NewGeneralization(genKeyB, "Generalization B", "## Generalization B\n\nDetails for gen B.", true, false, "uml comment for gen B")
	assert.Nil(suite.T(), err)

	// -- Actors --
	actorKeyA, err := identity.NewActorKey("alice")
	assert.Nil(suite.T(), err)
	actorKeyB, err := identity.NewActorKey("bob")
	assert.Nil(suite.T(), err)

	// Alice is the superclass of gen_a.
	actorA, err := model_actor.NewActor(actorKeyA, "Alice", "# Alice\n\nA person actor.", "person", &genKeyA, nil, "")
	assert.Nil(suite.T(), err)
	// Bob is the subclass of gen_a, and superclass of gen_b.
	actorB, err := model_actor.NewActor(actorKeyB, "Bob", "# Bob\n\nA system actor.", "system", &genKeyB, &genKeyA, "uml comment for bob")
	assert.Nil(suite.T(), err)

	// -- Domains --
	domainKeyA, err := identity.NewDomainKey("ordering")
	assert.Nil(suite.T(), err)
	domainKeyB, err := identity.NewDomainKey("shipping")
	assert.Nil(suite.T(), err)

	domainA, err := model_domain.NewDomain(domainKeyA, "Ordering", "# Ordering\n\nThe ordering domain.", true, "")
	assert.Nil(suite.T(), err)
	domainB, err := model_domain.NewDomain(domainKeyB, "Shipping", "# Shipping\n\nThe shipping domain.", false, "uml comment for shipping")
	assert.Nil(suite.T(), err)

	// Each domain gets a default subdomain (the parser creates this automatically).
	defaultSubKeyA, err := identity.NewSubdomainKey(domainKeyA, "default")
	assert.Nil(suite.T(), err)
	defaultSubA, err := model_domain.NewSubdomain(defaultSubKeyA, "Default", "", "")
	assert.Nil(suite.T(), err)

	// Add an explicit subdomain to the ordering domain.
	explicitSubKey, err := identity.NewSubdomainKey(domainKeyA, "fulfillment")
	assert.Nil(suite.T(), err)
	explicitSub, err := model_domain.NewSubdomain(explicitSubKey, "Fulfillment", "# Fulfillment\n\nOrder fulfillment subdomain.", "uml comment for fulfillment")
	assert.Nil(suite.T(), err)

	// -- Class generalizations --
	// Add a class generalization to the fulfillment subdomain.
	classGenKey, err := identity.NewGeneralizationKey(explicitSubKey, "order_type")
	assert.Nil(suite.T(), err)
	classGen, err := model_class.NewGeneralization(classGenKey, "Order Type", "## Order Type\n\nGeneralization of order types.", false, true, "")
	assert.Nil(suite.T(), err)
	explicitSub.Generalizations = map[identity.Key]model_class.Generalization{
		classGenKey: classGen,
	}

	// -- Classes --
	// Add a class to the fulfillment subdomain, referencing an actor and generalization.
	classKeyOrder, err := identity.NewClassKey(explicitSubKey, "order")
	assert.Nil(suite.T(), err)
	classOrder, err := model_class.NewClass(classKeyOrder, "Order", "## Order\n\nAn order placed by a customer.", &actorKeyA, nil, &classGenKey, "uml comment for order")
	assert.Nil(suite.T(), err)

	// Add an attribute to the order class.
	attrKeyStatus, err := identity.NewAttributeKey(classKeyOrder, "status")
	assert.Nil(suite.T(), err)
	attrStatus, err := model_class.NewAttribute(attrKeyStatus, "Status", "Current order status.", "string that is 3-20 chars long", nil, false, "", nil)
	assert.Nil(suite.T(), err)
	classOrder.SetAttributes(map[identity.Key]model_class.Attribute{attrKeyStatus: attrStatus})
	classOrder.SetStates(map[identity.Key]model_state.State{})
	classOrder.SetEvents(map[identity.Key]model_state.Event{})
	classOrder.SetGuards(map[identity.Key]model_state.Guard{})
	classOrder.SetActions(map[identity.Key]model_state.Action{})
	classOrder.SetTransitions(map[identity.Key]model_state.Transition{})

	// Add a second class "line_item" in the fulfillment subdomain (for subdomain-level association).
	classKeyLineItem, err := identity.NewClassKey(explicitSubKey, "line_item")
	assert.Nil(suite.T(), err)
	classLineItem, err := model_class.NewClass(classKeyLineItem, "Line Item", "## Line Item\n\nA line item in an order.", nil, nil, nil, "")
	assert.Nil(suite.T(), err)
	classLineItem.SetAttributes(map[identity.Key]model_class.Attribute{})
	classLineItem.SetStates(map[identity.Key]model_state.State{})
	classLineItem.SetEvents(map[identity.Key]model_state.Event{})
	classLineItem.SetGuards(map[identity.Key]model_state.Guard{})
	classLineItem.SetActions(map[identity.Key]model_state.Action{})
	classLineItem.SetTransitions(map[identity.Key]model_state.Transition{})

	explicitSub.Classes = map[identity.Key]model_class.Class{
		classKeyOrder:    classOrder,
		classKeyLineItem: classLineItem,
	}

	// Add a class "customer" in the default subdomain of ordering (for domain-level association).
	classKeyCustomer, err := identity.NewClassKey(defaultSubKeyA, "customer")
	assert.Nil(suite.T(), err)
	classCustomer, err := model_class.NewClass(classKeyCustomer, "Customer", "## Customer\n\nA customer who places orders.", nil, nil, nil, "")
	assert.Nil(suite.T(), err)
	classCustomer.SetAttributes(map[identity.Key]model_class.Attribute{})
	classCustomer.SetStates(map[identity.Key]model_state.State{})
	classCustomer.SetEvents(map[identity.Key]model_state.Event{})
	classCustomer.SetGuards(map[identity.Key]model_state.Guard{})
	classCustomer.SetActions(map[identity.Key]model_state.Action{})
	classCustomer.SetTransitions(map[identity.Key]model_state.Transition{})

	defaultSubA.Classes = map[identity.Key]model_class.Class{
		classKeyCustomer: classCustomer,
	}

	domainA.Subdomains = map[identity.Key]model_domain.Subdomain{
		defaultSubKeyA: defaultSubA,
		explicitSubKey: explicitSub,
	}

	// Add a class "shipment" in the default subdomain of shipping (for model-level association).
	defaultSubKeyB, err := identity.NewSubdomainKey(domainKeyB, "default")
	assert.Nil(suite.T(), err)
	defaultSubB, err := model_domain.NewSubdomain(defaultSubKeyB, "Default", "", "")
	assert.Nil(suite.T(), err)

	classKeyShipment, err := identity.NewClassKey(defaultSubKeyB, "shipment")
	assert.Nil(suite.T(), err)
	classShipment, err := model_class.NewClass(classKeyShipment, "Shipment", "## Shipment\n\nA shipment for an order.", nil, nil, nil, "")
	assert.Nil(suite.T(), err)
	classShipment.SetAttributes(map[identity.Key]model_class.Attribute{})
	classShipment.SetStates(map[identity.Key]model_state.State{})
	classShipment.SetEvents(map[identity.Key]model_state.Event{})
	classShipment.SetGuards(map[identity.Key]model_state.Guard{})
	classShipment.SetActions(map[identity.Key]model_state.Action{})
	classShipment.SetTransitions(map[identity.Key]model_state.Transition{})

	defaultSubB.Classes = map[identity.Key]model_class.Class{
		classKeyShipment: classShipment,
	}

	domainB.Subdomains = map[identity.Key]model_domain.Subdomain{defaultSubKeyB: defaultSubB}

	// -- Class associations at all three levels --
	// Subdomain-level: order -> line_item (both in fulfillment subdomain).
	multOne, err := model_class.NewMultiplicity("1")
	assert.Nil(suite.T(), err)
	multMany, err := model_class.NewMultiplicity("1..many")
	assert.Nil(suite.T(), err)
	multOptional, err := model_class.NewMultiplicity("0..1")
	assert.Nil(suite.T(), err)

	subdomainAssocKey, err := identity.NewClassAssociationKey(explicitSubKey, classKeyOrder, classKeyLineItem, "contains")
	assert.Nil(suite.T(), err)
	subdomainAssoc, err := model_class.NewAssociation(subdomainAssocKey, "Contains", "Order contains line items.", classKeyOrder, multOne, classKeyLineItem, multMany, nil, "")
	assert.Nil(suite.T(), err)

	// Domain-level: order -> customer (different subdomains in ordering domain).
	domainClassAssocKey, err := identity.NewClassAssociationKey(domainKeyA, classKeyOrder, classKeyCustomer, "placed by")
	assert.Nil(suite.T(), err)
	domainClassAssoc, err := model_class.NewAssociation(domainClassAssocKey, "Placed By", "Order placed by a customer.", classKeyOrder, multMany, classKeyCustomer, multOne, nil, "uml comment for placed by")
	assert.Nil(suite.T(), err)

	// Model-level: order -> shipment (different domains).
	modelClassAssocKey, err := identity.NewClassAssociationKey(identity.Key{}, classKeyOrder, classKeyShipment, "ships via")
	assert.Nil(suite.T(), err)
	modelClassAssoc, err := model_class.NewAssociation(modelClassAssocKey, "Ships Via", "", classKeyOrder, multOptional, classKeyShipment, multOne, nil, "")
	assert.Nil(suite.T(), err)

	// -- Domain associations --
	// Ordering is the problem domain, shipping is the solution domain.
	domainAssocKey, err := identity.NewDomainAssociationKey(domainKeyA, domainKeyB)
	assert.Nil(suite.T(), err)
	domainAssoc, err := model_domain.NewAssociation(domainAssocKey, domainKeyA, domainKeyB, "shipping solves ordering")
	assert.Nil(suite.T(), err)

	// -- Model --
	input := req_model.Model{
		Key:     "test_model",
		Name:    "Test Model",
		Details: "# Test Model\n\nTest model details in markdown.",
		Actors: map[identity.Key]model_actor.Actor{
			actorKeyA: actorA,
			actorKeyB: actorB,
		},
		ActorGeneralizations: map[identity.Key]model_actor.Generalization{
			genKeyA: genA,
			genKeyB: genB,
		},
		Domains: map[identity.Key]model_domain.Domain{
			domainKeyA: domainA,
			domainKeyB: domainB,
		},
		DomainAssociations: map[identity.Key]model_domain.Association{
			domainAssocKey: domainAssoc,
		},
	}

	// Use SetClassAssociations to automatically distribute associations to the correct level.
	allClassAssocs := map[identity.Key]model_class.Association{
		subdomainAssocKey:   subdomainAssoc,
		domainClassAssocKey: domainClassAssoc,
		modelClassAssocKey:  modelClassAssoc,
	}
	err = input.SetClassAssociations(allClassAssocs)
	assert.Nil(suite.T(), err, "setting class associations should succeed")

	// Validate the model before writing.
	err = input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Write to a temporary folder.
	tempDir := suite.T().TempDir()
	err = Write(input, tempDir)
	assert.Nil(suite.T(), err, "writing model should succeed")

	// Read from the temporary folder.
	output, err := Parse(tempDir)
	assert.Nil(suite.T(), err, "parsing model should succeed")

	// The parsed model's Key will be the tempDir path, not our original key.
	// Overwrite it for comparison since the parser uses the modelPath as the key.
	output.Key = input.Key

	// Compare the model values.
	assert.Equal(suite.T(), input, output)
}
