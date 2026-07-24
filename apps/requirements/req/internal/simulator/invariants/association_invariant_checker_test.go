package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type AssociationInvariantCheckerSuite struct {
	suite.Suite
}

func TestAssociationInvariantCheckerSuite(t *testing.T) {
	suite.Run(t, new(AssociationInvariantCheckerSuite))
}

func (s *AssociationInvariantCheckerSuite) buildChecker() (*AssociationInvariantChecker, identity.Key, identity.Key, identity.Key) {
	partnerClass, partnerKey := associationInvPartnerClass()
	jurisdictionClass, jurisdictionKey := associationInvJurisdictionClass()
	assocKey := associationInvAssocKey(partnerKey, jurisdictionKey)

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	invKey := helper.Must(identity.NewClassAssociationInvariantKey(assocKey, "0"))
	assoc.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Unique jurisdiction codes.", "", parsedSpec(
			`∀ j1 ∈ self.Configures : ∀ j2 ∈ self.Configures : ((j1 ≠ j2) ⇒ (j1.Code ≠ j2.Code))`,
		), nil),
	})

	model := multiplicityTestModel(classEntry(partnerClass, partnerKey), classEntry(jurisdictionClass, jurisdictionKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	checker, err := NewAssociationInvariantChecker(schema.New(model))
	s.Require().NoError(err)
	return checker, partnerKey, jurisdictionKey, assocKey
}

func (s *AssociationInvariantCheckerSuite) TestPassesWhenInvariantHolds() {
	checker, partnerKey, jurisdictionKey, assocKey := s.buildChecker()

	simState := instance.NewState(emptySchema())
	partner := simState.CreateInstance(partnerKey, object.NewRecord())
	j1 := simState.CreateInstance(jurisdictionKey, object.NewRecord())
	j1.Attributes.Set("Code", object.NewString("US"))
	j2 := simState.CreateInstance(jurisdictionKey, object.NewRecord())
	j2.Attributes.Set("Code", object.NewString("UK"))

	bb := state.NewBindingsBuilder(simState)
	bb.AddAssociation(assocKey, "Configures", partnerKey, jurisdictionKey,
		evaluator.Multiplicity{}, evaluator.Multiplicity{HigherBound: 0})
	bb.RelationContext().CreateLink(evaluator.AssociationKey(assocKey.String()), partner.Attributes, j1.Attributes)
	bb.RelationContext().CreateLink(evaluator.AssociationKey(assocKey.String()), partner.Attributes, j2.Attributes)

	s.Empty(checker.CheckState(simState, bb))
}

func (s *AssociationInvariantCheckerSuite) TestFailsWhenAssessmentIsFalse() {
	partnerClass, partnerKey := associationInvPartnerClass()
	jurisdictionClass, jurisdictionKey := associationInvJurisdictionClass()
	assocKey := associationInvAssocKey(partnerKey, jurisdictionKey)

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	invKey := helper.Must(identity.NewClassAssociationInvariantKey(assocKey, "0"))
	assoc.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Always false.", "", parsedSpec("FALSE"), nil),
	})

	model := multiplicityTestModel(classEntry(partnerClass, partnerKey), classEntry(jurisdictionClass, jurisdictionKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	checker, err := NewAssociationInvariantChecker(schema.New(model))
	s.Require().NoError(err)

	simState := instance.NewState(emptySchema())
	simState.CreateInstance(partnerKey, object.NewRecord())

	bb := state.NewBindingsBuilder(simState)
	bb.AddAssociation(assocKey, "Configures", partnerKey, jurisdictionKey,
		evaluator.Multiplicity{}, evaluator.Multiplicity{HigherBound: 0})

	violations := checker.CheckState(simState, bb)
	s.Len(violations, 1)
	s.Equal(ViolationTypeAssociationInvariant, violations[0].Type)
}

func associationInvPartnerClass() (model_class.Class, identity.Key) {
	key := multiplicityMustKey("domain/d/subdomain/s/class/partner")
	return model_class.NewClass(key, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}), key
}

func associationInvJurisdictionClass() (model_class.Class, identity.Key) {
	key := multiplicityMustKey("domain/d/subdomain/s/class/jurisdiction")
	return model_class.NewClass(key, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"}), key
}

func associationInvAssocKey(fromKey, toKey identity.Key) identity.Key {
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	return helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "configures"))
}
