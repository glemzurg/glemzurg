package instance

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	checker := NewAssociationInstancePairChecker(schema.New(model, schema.RunScopeAll()))

	simState := NewState(emptySchema())
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item1 := simState.CreateInstance(itemKey, object.NewRecord())
	item2 := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.GetID(), item1.GetID()))
	s.Require().NoError(simState.AddLink(assocKey, order.GetID(), item2.GetID()))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationInstancePairCheckerSuite) TestDuplicatePairReportsViolation() {
	// Write path rejects duplicates; unit-test the pair-count logic with synthetic rows.
	assocKey := multiplicityTestAssocKey(
		multiplicityMustKey("domain/d/subdomain/s/class/order"),
		multiplicityMustKey("domain/d/subdomain/s/class/item"),
	)
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "OrderItem", Details: ""},
		model_class.AssociationEnd{
			ClassKey:     multiplicityMustKey("domain/d/subdomain/s/class/order"),
			Multiplicity: helper.Must(model_class.NewMultiplicity("any")),
		},
		model_class.AssociationEnd{
			ClassKey:     multiplicityMustKey("domain/d/subdomain/s/class/item"),
			Multiplicity: helper.Must(model_class.NewMultiplicity("any")),
		},
		model_class.AssociationOptions{},
	)
	dup := associationLinkEndpoints{fromID: 1, toID: 2}
	violations := checkAssociationInstancePairs(assoc, []associationLinkEndpoints{dup, dup})
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationDuplicateLink, violations[0].Type)
}

func (s *AssociationInstancePairCheckerSuite) TestAssociationClassDuplicatePairReportsViolation() {
	fromKey := multiplicityMustKey("domain/d/subdomain/s/class/partner")
	toKey := multiplicityMustKey("domain/d/subdomain/s/class/jurisdiction")
	acKey := multiplicityMustKey("domain/d/subdomain/s/class/test_link")
	assocKey := multiplicityTestAssocKey(fromKey, toKey)
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{AssociationClassKey: &acKey},
	)
	dup := associationLinkEndpoints{fromID: 10, toID: 20}
	violations := checkAssociationInstancePairs(assoc, []associationLinkEndpoints{dup, dup})
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationDuplicateLink, violations[0].Type)
}
