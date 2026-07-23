package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type AssociationUniquenessCheckerSuite struct {
	suite.Suite
}

func TestAssociationUniquenessCheckerSuite(t *testing.T) {
	suite.Run(t, new(AssociationUniquenessCheckerSuite))
}

func (s *AssociationUniquenessCheckerSuite) buildModel() (*core.Model, identity.Key, identity.Key, identity.Key, identity.Key) {
	fromClass, fromKey := associationUniquenessPartnerClass()
	toClass, toKey := associationUniquenessJurisdictionClass()
	acClass, acKey := associationUniquenessTestLinkClass()

	assocKey := multiplicityTestAssocKey(fromKey, toKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	jurisdictionAttrKey := multiplicityMustKey("domain/d/subdomain/s/class/jurisdiction/attribute/jurisdiction_code")
	uniqueness := model_class.NewAssociationUniqueness(nil, []identity.Key{jurisdictionAttrKey})

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult},
		model_class.AssociationOptions{
			AssociationClassKey: &acKey,
			Uniqueness:          &uniqueness,
		},
	)

	model := multiplicityTestModel(
		classEntry(fromClass, fromKey),
		classEntry(toClass, toKey),
		classEntry(acClass, acKey),
	)
	domainKey := multiplicityMustKey("domain/d")
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	domain := model.Domains[domainKey]
	subdomain := domain.Subdomains[subdomainKey]
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
	return model, assocKey, fromKey, toKey, acKey
}

func (s *AssociationUniquenessCheckerSuite) TestDistinctCodesNoViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	fromInst := simState.CreateInstance(fromKey, object.NewRecord())
	toInst1 := simState.CreateInstance(toKey, object.NewRecordFromFields(map[string]object.Object{
		"jurisdiction_code": object.NewString("US-NJ"),
	}))
	toInst2 := simState.CreateInstance(toKey, object.NewRecordFromFields(map[string]object.Object{
		"jurisdiction_code": object.NewString("US-PA"),
	}))
	link1 := simState.CreateInstance(acKey, object.NewRecord())
	link2 := simState.CreateInstance(acKey, object.NewRecord())
	s.Require().NoError(simState.AddAssociationLink(assocKey, fromInst.ID, toInst1.ID, link1.ID))
	s.Require().NoError(simState.AddAssociationLink(assocKey, fromInst.ID, toInst2.ID, link2.ID))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationUniquenessCheckerSuite) TestDuplicateCodeReportsViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	fromInst := simState.CreateInstance(fromKey, object.NewRecord())
	toInst1 := simState.CreateInstance(toKey, object.NewRecordFromFields(map[string]object.Object{
		"jurisdiction_code": object.NewString("US-NJ"),
	}))
	toInst2 := simState.CreateInstance(toKey, object.NewRecordFromFields(map[string]object.Object{
		"jurisdiction_code": object.NewString("US-NJ"),
	}))
	link1 := simState.CreateInstance(acKey, object.NewRecord())
	link2 := simState.CreateInstance(acKey, object.NewRecord())
	s.Require().NoError(simState.AddAssociationLink(assocKey, fromInst.ID, toInst1.ID, link1.ID))
	s.Require().NoError(simState.AddAssociationLink(assocKey, fromInst.ID, toInst2.ID, link2.ID))

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationUniqueness, violations[0].Type)
}

func (s *AssociationUniquenessCheckerSuite) TestPlainToOnlyDistinctNoViolation() {
	model, assocKey, orderKey, customerKey := associationUniquenessPlainToOnlyModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	order := simState.CreateInstance(orderKey, object.NewRecord())
	customer1 := simState.CreateInstance(customerKey, object.NewRecordFromFields(map[string]object.Object{
		"customer_code": object.NewString("C-100"),
	}))
	customer2 := simState.CreateInstance(customerKey, object.NewRecordFromFields(map[string]object.Object{
		"customer_code": object.NewString("C-200"),
	}))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, customer1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, customer2.ID))

	s.Empty(checker.CheckState(simState))
}

func (s *AssociationUniquenessCheckerSuite) TestPlainToOnlyDuplicateReportsViolation() {
	model, assocKey, orderKey, customerKey := associationUniquenessPlainToOnlyModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	order := simState.CreateInstance(orderKey, object.NewRecord())
	customer1 := simState.CreateInstance(customerKey, object.NewRecordFromFields(map[string]object.Object{
		"customer_code": object.NewString("C-100"),
	}))
	customer2 := simState.CreateInstance(customerKey, object.NewRecordFromFields(map[string]object.Object{
		"customer_code": object.NewString("C-100"),
	}))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, customer1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, customer2.ID))

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationUniqueness, violations[0].Type)
}

func (s *AssociationUniquenessCheckerSuite) TestPlainFromOnlyDistinctNoViolation() {
	model, assocKey, productKey, shelfKey := associationUniquenessPlainFromOnlyModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	product1 := simState.CreateInstance(productKey, object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Widget"),
	}))
	product2 := simState.CreateInstance(productKey, object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Gadget"),
	}))
	shelf1 := simState.CreateInstance(shelfKey, object.NewRecord())
	shelf2 := simState.CreateInstance(shelfKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, product1.ID, shelf1.ID))
	s.Require().NoError(simState.AddLink(assocKey, product2.ID, shelf2.ID))

	s.Empty(checker.CheckState(simState))
}

func (s *AssociationUniquenessCheckerSuite) TestPlainFromOnlyDuplicateReportsViolation() {
	model, assocKey, productKey, shelfKey := associationUniquenessPlainFromOnlyModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	product1 := simState.CreateInstance(productKey, object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Widget"),
	}))
	product2 := simState.CreateInstance(productKey, object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Widget"),
	}))
	shelf := simState.CreateInstance(shelfKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, product1.ID, shelf.ID))
	s.Require().NoError(simState.AddLink(assocKey, product2.ID, shelf.ID))

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationUniqueness, violations[0].Type)
}

func (s *AssociationUniquenessCheckerSuite) TestPlainBothSidesDistinctNoViolation() {
	model, assocKey, orderKey, shipmentKey := associationUniquenessPlainBothSidesModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	order1 := simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"order_date": object.NewString("2026-01-01"),
	}))
	order2 := simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"order_date": object.NewString("2026-01-02"),
	}))
	shipment1 := simState.CreateInstance(shipmentKey, object.NewRecordFromFields(map[string]object.Object{
		"tracking_id": object.NewString("TRK-1"),
	}))
	shipment2 := simState.CreateInstance(shipmentKey, object.NewRecordFromFields(map[string]object.Object{
		"tracking_id": object.NewString("TRK-2"),
	}))
	s.Require().NoError(simState.AddLink(assocKey, order1.ID, shipment1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order2.ID, shipment2.ID))

	s.Empty(checker.CheckState(simState))
}

func (s *AssociationUniquenessCheckerSuite) TestPlainBothSidesDuplicateReportsViolation() {
	model, assocKey, orderKey, shipmentKey := associationUniquenessPlainBothSidesModel()
	checker := NewAssociationUniquenessChecker(model)

	simState := instance.NewState(nil)
	order1 := simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"order_date": object.NewString("2026-01-01"),
	}))
	order2 := simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"order_date": object.NewString("2026-01-01"),
	}))
	shipment1 := simState.CreateInstance(shipmentKey, object.NewRecordFromFields(map[string]object.Object{
		"tracking_id": object.NewString("TRK-1"),
	}))
	shipment2 := simState.CreateInstance(shipmentKey, object.NewRecordFromFields(map[string]object.Object{
		"tracking_id": object.NewString("TRK-1"),
	}))
	s.Require().NoError(simState.AddLink(assocKey, order1.ID, shipment1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order2.ID, shipment2.ID))

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationUniqueness, violations[0].Type)
}

func associationUniquenessPartnerClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/partner")
	return model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}), classKey
}

func associationUniquenessJurisdictionClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/jurisdiction")
	attrKey := multiplicityMustKey("domain/d/subdomain/s/class/jurisdiction/attribute/jurisdiction_code")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Jurisdiction Code"}, "unconstrained", nil, true, model_class.AttributeAnnotations{})),
	})
	return class, classKey
}

func associationUniquenessTestLinkClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/link")
	stateActiveKey := multiplicityMustKey("domain/d/subdomain/s/class/link/state/active")
	eventCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/link/event/create")
	transCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/link/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Link"})
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})
	return class, classKey
}

func associationUniquenessPlainToOnlyModel() (*core.Model, identity.Key, identity.Key, identity.Key) {
	orderClass, orderKey := associationUniquenessOrderClass()
	customerClass, customerKey := associationUniquenessCustomerClass()
	assocKey := multiplicityTestAssocKey(orderKey, customerKey)
	customerCodeKey := multiplicityMustKey("domain/d/subdomain/s/class/customer/attribute/customer_code")
	uniqueness := model_class.NewAssociationUniqueness(nil, []identity.Key{customerCodeKey})
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "order belongs to customer", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: customerKey, Multiplicity: toMult},
		model_class.AssociationOptions{Uniqueness: &uniqueness},
	)
	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(customerClass, customerKey))
	associationUniquenessAttachAssoc(model, assocKey, assoc)
	return model, assocKey, orderKey, customerKey
}

func associationUniquenessPlainFromOnlyModel() (*core.Model, identity.Key, identity.Key, identity.Key) {
	productClass, productKey := associationUniquenessProductClass()
	shelfClass, shelfKey := associationUniquenessShelfClass()
	assocKey := multiplicityTestAssocKey(productKey, shelfKey)
	productNameKey := multiplicityMustKey("domain/d/subdomain/s/class/product/attribute/name")
	uniqueness := model_class.NewAssociationUniqueness([]identity.Key{productNameKey}, nil)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "product stored on shelf", Details: ""},
		model_class.AssociationEnd{ClassKey: productKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: shelfKey, Multiplicity: toMult},
		model_class.AssociationOptions{Uniqueness: &uniqueness},
	)
	model := multiplicityTestModel(classEntry(productClass, productKey), classEntry(shelfClass, shelfKey))
	associationUniquenessAttachAssoc(model, assocKey, assoc)
	return model, assocKey, productKey, shelfKey
}

func associationUniquenessPlainBothSidesModel() (*core.Model, identity.Key, identity.Key, identity.Key) {
	orderClass, orderKey := associationUniquenessOrderClass()
	shipmentClass, shipmentKey := associationUniquenessShipmentClass()
	assocKey := multiplicityTestAssocKey(orderKey, shipmentKey)
	orderDateKey := multiplicityMustKey("domain/d/subdomain/s/class/order/attribute/order_date")
	trackingKey := multiplicityMustKey("domain/d/subdomain/s/class/shipment/attribute/tracking_id")
	uniqueness := model_class.NewAssociationUniqueness(
		[]identity.Key{orderDateKey},
		[]identity.Key{trackingKey},
	)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "order has shipment", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: shipmentKey, Multiplicity: toMult},
		model_class.AssociationOptions{Uniqueness: &uniqueness},
	)
	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(shipmentClass, shipmentKey))
	associationUniquenessAttachAssoc(model, assocKey, assoc)
	return model, assocKey, orderKey, shipmentKey
}

func associationUniquenessAttachAssoc(model *core.Model, assocKey identity.Key, assoc model_class.Association) {
	domainKey := multiplicityMustKey("domain/d")
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	domain := model.Domains[domainKey]
	subdomain := domain.Subdomains[subdomainKey]
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
}

func associationUniquenessOrderClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/order")
	attrKey := multiplicityMustKey("domain/d/subdomain/s/class/order/attribute/order_date")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Order Date"}, "unconstrained", nil, true, model_class.AttributeAnnotations{})),
	})
	return class, classKey
}

func associationUniquenessCustomerClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/customer")
	attrKey := multiplicityMustKey("domain/d/subdomain/s/class/customer/attribute/customer_code")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Customer"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Customer Code"}, "unconstrained", nil, true, model_class.AttributeAnnotations{})),
	})
	return class, classKey
}

func associationUniquenessProductClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/product")
	attrKey := multiplicityMustKey("domain/d/subdomain/s/class/product/attribute/name")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Product"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Name"}, "unconstrained", nil, true, model_class.AttributeAnnotations{})),
	})
	return class, classKey
}

func associationUniquenessShelfClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/shelf")
	return model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Shelf"}), classKey
}

func associationUniquenessShipmentClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/shipment")
	attrKey := multiplicityMustKey("domain/d/subdomain/s/class/shipment/attribute/tracking_id")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Shipment"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Tracking ID"}, "unconstrained", nil, true, model_class.AttributeAnnotations{})),
	})
	return class, classKey
}
