package associationfacts

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
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

			assoc := model_class.NewAssociation(
				helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, tc.assocName)),
				tc.assocName,
				tc.details,
				model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult},
				model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult},
				nil,
				"",
			)

			got := FormatAssociationFact(
				assoc,
				model_class.NewClass(fromKey, tc.fromName, "", "", nil, nil, nil, ""),
				model_class.NewClass(toKey, tc.toName, "", "", nil, nil, nil, ""),
				nil,
			)

			for _, substr := range tc.wantSubstr {
				assert.Contains(t, got, substr)
			}
		})
	}
}

func TestFactsForSubdomain_testModel(t *testing.T) {
	t.Parallel()

	model := test_helper.GetTestModel()
	subdomain, err := FindSubdomain(model, SubdomainPath{
		DomainSubKey:    "domain_a",
		SubdomainSubKey: "subdomain_a",
	})
	require.NoError(t, err)

	facts := FactsForSubdomain(subdomain)
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
