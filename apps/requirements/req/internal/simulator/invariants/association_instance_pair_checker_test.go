package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type AssociationInstancePairCheckerSuite struct {
	suite.Suite
}

func TestAssociationInstancePairCheckerSuite(t *testing.T) {
	suite.Run(t, new(AssociationInstancePairCheckerSuite))
}

func (s *AssociationInstancePairCheckerSuite) buildPlainAssociationModel() (*core.Model, identity.Key, identity.Key, identity.Key) {
	orderClass, orderKey := multiplicityTestOrderClass()
	itemClass, itemKey := multiplicityTestItemClass()
	assocKey := multiplicityTestAssocKey(orderKey, itemKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "OrderItem", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult},
		model_class.AssociationOptions{},
	)

	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	domainKey := multiplicityMustKey("domain/d")
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	domain := model.Domains[domainKey]
	subdomain := domain.Subdomains[subdomainKey]
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
	return model, assocKey, orderKey, itemKey
}

func (s *AssociationInstancePairCheckerSuite) TestDistinctPairsNoViolation() {
	model, assocKey, orderKey, itemKey := s.buildPlainAssociationModel()
	checker := NewAssociationInstancePairChecker(schema.New(model))

	simState := instance.NewState(emptySchema())
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item1 := simState.CreateInstance(itemKey, object.NewRecord())
	item2 := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item2.ID))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationInstancePairCheckerSuite) TestDuplicatePairReportsViolation() {
	model, assocKey, orderKey, itemKey := s.buildPlainAssociationModel()
	checker := NewAssociationInstancePairChecker(schema.New(model))

	simState := instance.NewState(emptySchema())
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item.ID))

	// Bypass write-time rejection to exercise the checker directly.
	simState.Links().AppendLinkWithoutValidation(
		evaluator.AssociationKey(assocKey.String()),
		evaluator.ObjectID(order.ID),
		evaluator.ObjectID(item.ID),
	)

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationDuplicateLink, violations[0].Type)
}

func (s *AssociationInstancePairCheckerSuite) TestAssociationClassDuplicatePairReportsViolation() {
	model, assocKey, fromKey, toKey, acKey := associationUniquenessSuiteModelWithoutUniqueness()
	checker := NewAssociationInstancePairChecker(schema.New(model))

	simState := instance.NewState(emptySchema())
	fromInst := simState.CreateInstance(fromKey, object.NewRecord())
	toInst := simState.CreateInstance(toKey, object.NewRecord())
	link1 := simState.CreateInstance(acKey, object.NewRecord())
	link2 := simState.CreateInstance(acKey, object.NewRecord())
	s.Require().NoError(simState.AddAssociationLink(assocKey, fromInst.ID, toInst.ID, link1.ID))

	simState.AssociationLinks().AppendLinkWithoutValidation(instance.AssociationLink{
		HostAssocKey:   assocKey,
		FromEndpointID: fromInst.ID,
		ToEndpointID:   toInst.ID,
		LinkInstanceID: link2.ID,
	})

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationDuplicateLink, violations[0].Type)
}

func associationUniquenessSuiteModelWithoutUniqueness() (*core.Model, identity.Key, identity.Key, identity.Key, identity.Key) {
	fromClass, fromKey := associationUniquenessPartnerClass()
	toClass, toKey := associationUniquenessJurisdictionClass()
	acClass, acKey := associationUniquenessTestLinkClass()

	assocKey := multiplicityTestAssocKey(fromKey, toKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult},
		model_class.AssociationOptions{AssociationClassKey: &acKey},
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
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
	return model, assocKey, fromKey, toKey, acKey
}
