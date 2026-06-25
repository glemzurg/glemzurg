package modelfacts

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSubdomainPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		path    string
		want    SubdomainPath
		wantErr bool
	}{
		{
			name: "domain and subdomain",
			path: "billing/ledger",
			want: SubdomainPath{DomainSubKey: "billing", SubdomainSubKey: "ledger"},
		},
		{
			name:    "missing subdomain",
			path:    "billing",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSubdomainPath(tc.path)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatAssociationFact(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("domain_a"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sub_a"))

	cases := []struct {
		name       string
		assocName  string
		details    string
		fromKey    string
		fromName   string
		toKey      string
		toName     string
		fromMult   string
		toMult     string
		wantSubstr []string
	}{
		{
			name:      "one to many named association",
			assocName: "Can Game With",
			details:   "Fixture detail text.",
			fromKey:   "actor_a",
			fromName:  "Actor A",
			toKey:     "resource_b",
			toName:    "Resource B",
			fromMult:  "1",
			toMult:    "1..many",
			wantSubstr: []string{
				"each actor a (can game with) links to one or more resource bs",
				"each resource b links to exactly one actor a",
				"(Fixture detail text.)",
			},
		},
		{
			name:      "one to any named association",
			assocName: "Is Subdivided Into",
			details:   "Fixture split detail.",
			fromKey:   "container",
			fromName:  "Container",
			toKey:     "part",
			toName:    "Part",
			fromMult:  "1",
			toMult:    "any",
			wantSubstr: []string{
				"each container (is subdivided into) links to any number of parts",
				"each part links to exactly one container",
				"(Fixture split detail.)",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fromKey := helper.Must(identity.NewClassKey(subdomainKey, tc.fromKey))
			toKey := helper.Must(identity.NewClassKey(subdomainKey, tc.toKey))
			fromMult := helper.Must(model_class.NewMultiplicity(tc.fromMult))
			toMult := helper.Must(model_class.NewMultiplicity(tc.toMult))

			assoc := model_class.NewAssociation(helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, tc.assocName)), model_class.AssociationDetails{Name: tc.assocName, Details: tc.details}, model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult}, nil, "")

			got := FormatAssociationFact(
				assoc,
				model_class.NewClass(fromKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: tc.fromName, Details: "", UnfinishedNotes: "", UmlComment: ""}),
				model_class.NewClass(toKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: tc.toName, Details: "", UnfinishedNotes: "", UmlComment: ""}),
				nil,
			)

			for _, substr := range tc.wantSubstr {
				assert.Contains(t, got, substr)
			}
		})
	}
}

func TestFormatAssociationInvariantFact(t *testing.T) {
	t.Parallel()

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
		nil,
		"",
	)
	partner := model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"})

	spec := `∀ j1 ∈ self.ConfiguresCustomersFor : ∀ j2 ∈ self.ConfiguresCustomersFor : ((j1 ≠ j2) ⇒ (j1.jurisdiction_code ≠ j2.jurisdiction_code))`
	got := FormatAssociationInvariantFact(
		assoc,
		partner,
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"A partner cannot configure two jurisdictions with the same jurisdiction code",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: spec},
			nil,
		),
	)
	assert.Equal(t, AssociationInvariantFact{
		Label:       "Partner (configures customers for)",
		Description: "A partner cannot configure two jurisdictions with the same jurisdiction code.",
		Spec:        spec,
	}, got)
}

func TestFormatAssociationInvariantFactUsesSpecWhenDescriptionMissing(t *testing.T) {
	t.Parallel()

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
		nil,
		"",
	)
	partner := model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"})

	got := FormatAssociationInvariantFact(
		assoc,
		partner,
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
			nil,
		),
	)
	assert.Equal(t, AssociationInvariantFact{
		Label:       "Partner (configures customers for)",
		Description: "TRUE.",
		Spec:        "",
	}, got)
}

func TestAssociationInvariantFactsForSubdomain(t *testing.T) {
	t.Parallel()

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
		nil,
		"",
	)
	assoc.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"Jurisdiction codes must be unique per partner.",
			"",
			logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"},
			nil,
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerKey:      model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}),
		jurisdictionKey: model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"}),
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	facts := AssociationInvariantFactsForSubdomain(subdomain)
	require.Len(t, facts, 1)
	assert.Equal(t, "Partner (configures customers for)", facts[0].Label)
	assert.Equal(t, "Jurisdiction codes must be unique per partner.", facts[0].Description)
	assert.Equal(t, "TRUE", facts[0].Spec)
}

func TestFormatIndexFact(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		className string
		attrNames []string
		want      string
	}{
		{
			name:      "key index single attribute",
			className: "Currency",
			attrNames: []string{"Abbr"},
			want:      "No currencies can share the same Abbr.",
		},
		{
			name:      "secondary index composite",
			className: "Widget",
			attrNames: []string{"Email", "Tenant"},
			want:      "No widgets can share the same Email and Tenant combination.",
		},
		{
			name:      "secondary index three attributes",
			className: "Order",
			attrNames: []string{"Alpha", "Beta", "Gamma"},
			want:      "No orders can share the same Alpha, Beta, and Gamma combination.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, FormatIndexFact(tc.className, tc.attrNames))
		})
	}
}

func TestAssociationFactsForSubdomain_testModel(t *testing.T) {
	t.Parallel()

	model := test_helper.GetTestModel()
	subdomain, err := FindSubdomain(model, SubdomainPath{
		DomainSubKey:    "domain_a",
		SubdomainSubKey: "subdomain_a",
	})
	require.NoError(t, err)

	facts := AssociationFactsForSubdomain(subdomain)
	require.Len(t, facts, 3)

	joined := strings.Join(facts, "\n")
	assert.Contains(t, joined, "each order (order contains products) links to one or more products")
	assert.Contains(t, joined, "each product links to exactly one order")
	assert.Contains(t, joined, "each order–product pairing is a line item")
	assert.Contains(t, joined, "each order (order belongs to customer) links to exactly one customer")
	assert.Contains(t, joined, "each customer links to one or more orders")
	assert.Contains(t, joined, "each product (product has line items) links to one or more line items")
	assert.Contains(t, joined, "each line item links to exactly one product")
}

func TestIndexFactsForSubdomain_testModel(t *testing.T) {
	t.Parallel()

	model := test_helper.GetTestModel()
	subdomain, err := FindSubdomain(model, SubdomainPath{
		DomainSubKey:    "domain_a",
		SubdomainSubKey: "subdomain_a",
	})
	require.NoError(t, err)

	facts := IndexFactsForSubdomain(subdomain)
	require.Len(t, facts, 2)

	joined := strings.Join(facts, "\n")
	assert.Contains(t, joined, "No orders can share the same Total.")
}
