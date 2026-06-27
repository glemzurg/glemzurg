package modelfacts

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestAssociationInvariantFactsIncludeClassInvariantTaggedOverAssociation(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))
	anyMult := helper.Must(model_class.NewMultiplicity("any"))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: anyMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: anyMult},
		model_class.Multiplicity{},
		model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)

	invKey := helper.Must(identity.NewClassInvariantKey(jurisdictionKey, "0"))
	spec := `∀ p1 ∈ self._ConfiguresCustomersFor : ∀ p2 ∈ self._ConfiguresCustomersFor : ((p1 ≠ p2) ⇒ (p1.JurisdictionalWalletDefinition ≠ p2.JurisdictionalWalletDefinition))`
	jurisdiction := model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	jurisdiction.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"A jurisdiction cannot be linked to a given partner more than once.",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: spec},
			nil,
		),
	})
	jurisdiction.Invariants[0].SetOverAssociationKey(&assocKey)

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerKey:      model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}),
		jurisdictionKey: jurisdiction,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	facts := AssociationInvariantFactsForSubdomain(subdomain)
	require.Len(t, facts, 1)
	require.Equal(t, "Jurisdiction (configures customers for)", facts[0].Label)
	require.Equal(t, "A jurisdiction cannot be linked to a given partner more than once.", facts[0].Description)
	require.Equal(t, spec, facts[0].Spec)
}
