package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestClassMarkdownRendersAssociationInvariants(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))
	invKey := helper.Must(identity.NewClassAssociationInvariantKey(assocKey, "0"))
	anyMult := helper.Must(model_class.NewMultiplicity("any"))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{
			Name:    "Configures Customers For",
			Details: "The partner configures jurisdictional wallet behavior for its customers.",
		},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: anyMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: anyMult},
		nil,
		"",
	)
	assoc.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"A partner cannot configure two jurisdictions with the same jurisdiction code.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: `∀ j1 ∈ self.ConfiguresCustomersFor : ∀ j2 ∈ self.ConfiguresCustomersFor : ((j1 ≠ j2) ⇒ (j1.jurisdiction_code ≠ j2.jurisdiction_code))`,
			},
			nil,
		),
	})

	partner := model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"})
	jurisdiction := model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerKey:      partner,
		jurisdictionKey: jurisdiction,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, partner, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "## Association Invariants")
	require.Contains(t, contents, "### Configures Customers For")
	require.Contains(t, contents, "A partner cannot configure two jurisdictions with the same jurisdiction code.")
	require.Contains(t, contents, "self.ConfiguresCustomersFor")
	require.NotContains(t, contents, "## Association Invariants\n\n## Association Invariants")
}

func TestClassMarkdownOmitsAssociationInvariantsWhenNone(t *testing.T) {
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
		nil,
		"",
	)

	partner := model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"})
	jurisdiction := model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerKey:      partner,
		jurisdictionKey: jurisdiction,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, partner, "", "")
	require.NoError(t, err)
	require.NotContains(t, contents, "## Association Invariants")
}

func TestClassOutgoingAssociationsWithInvariantsOnlyFromClass(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	outgoingKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "outgoing"))
	incomingKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, toKey, fromKey, "incoming"))
	invKey := helper.Must(identity.NewClassAssociationInvariantKey(incomingKey, "0"))
	anyMult := helper.Must(model_class.NewMultiplicity("any"))

	outgoing := model_class.NewAssociation(
		outgoingKey,
		model_class.AssociationDetails{Name: "Outgoing", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: anyMult},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: anyMult},
		nil,
		"",
	)
	incoming := model_class.NewAssociation(
		incomingKey,
		model_class.AssociationDetails{Name: "Incoming", Details: ""},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: anyMult},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: anyMult},
		nil,
		"",
	)
	incoming.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"Only on to-class anchor.",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
			nil,
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		outgoingKey: outgoing,
		incomingKey: incoming,
	}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	got := reqs.ClassOutgoingAssociationsWithInvariants(fromKey)
	require.Empty(t, got)

	got = reqs.ClassOutgoingAssociationsWithInvariants(toKey)
	require.Len(t, got, 1)
	require.Equal(t, "Incoming", got[0].Name)
}
