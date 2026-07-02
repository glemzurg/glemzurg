package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/modelfacts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSubdomainFactsRendersAssociationInvariants(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))
	invKey := helper.Must(identity.NewClassAssociationInvariantKey(assocKey, "0"))
	anyMult := helper.Must(model_class.NewMultiplicity("any"))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: anyMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: anyMult},
		model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	spec := `∀ j1 ∈ self.ConfiguresCustomersFor : ∀ j2 ∈ self.ConfiguresCustomersFor : ((j1 ≠ j2) ⇒ (j1.jurisdiction_code ≠ j2.jurisdiction_code))`
	assoc.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"A partner cannot configure two jurisdictions with the same jurisdiction code.",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: spec},
			nil,
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerKey:      model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}),
		jurisdictionKey: model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"}),
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	facts := modelfacts.FactsForSubdomain(subdomain)
	require.Len(t, facts.AssociationInvariants, 1)
	require.Equal(t, spec, facts.AssociationInvariants[0].Spec)

	factsFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "facts", ".md")
	factsBody := string(writer.md[factsFile])
	assert.Contains(t, factsBody, "## Association Invariants")
	assert.Contains(t, factsBody, "- Partner (configures customers for): A partner cannot configure two jurisdictions with the same jurisdiction code.")
	assert.Contains(t, factsBody, "    - **"+spec+"**")
}
