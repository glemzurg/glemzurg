package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderAssociationLinkNodeMermaid(t *testing.T) {
	t.Parallel()

	one := helper.Must(model_class.NewMultiplicity("1"))
	fromKey := helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "from"))
	toKey := helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to"))
	acKey := helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "ac"))
	assocKey := helper.Must(identity.NewClassAssociationKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), fromKey, toKey, "relates"))

	direct := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "relates", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		model_class.AssociationOptions{},
	)
	withComment := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "relates", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		model_class.AssociationOptions{UmlComment: "very import to users"},
	)
	withAC := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "relates", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		model_class.AssociationOptions{AssociationClassKey: &acKey},
	)

	assert.False(t, renderAssociationLinkNodeMermaid(direct))
	assert.True(t, renderAssociationLinkNodeMermaid(withComment))
	assert.True(t, renderAssociationLinkNodeMermaid(withAC))
}

func TestGenerateAssociationClassMermaidEndpointMultiplicities(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "from_class"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "to_class"))
	acKey := helper.Must(identity.NewClassKey(subdomainKey, "ac_class"))

	fromMult := helper.Must(model_class.NewMultiplicity("3"))
	toMult := helper.Must(model_class.NewMultiplicity("many..many"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "relates"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "relates", Details: ""}, model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: &acKey, UmlComment: ""})

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			fromKey: model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "From"}),
			toKey:   model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "To"}),
			acKey:   model_class.NewClass(acKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Assoc"}),
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}
	model := core.Model{
		Key:  "test_ac_mult",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:        domainKey,
				Name:       "D X",
				Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain},
			},
		},
	}

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter: %v", err)
	}

	fromFile := convertKeyToFilename("class", fromKey.String(), "", ".md")
	got := string(writer.md[fromFile])

	fromNode := nodeIDFor("class", fromKey)
	acNode := nodeIDFor("class", acKey)
	toNode := nodeIDFor("class", toKey)
	linkNode := nodeIDFor("assoc", assocKey)

	wantFrom := fromNode + ` "3" -- ` + linkNode
	wantTo := linkNode + ` --> "*" ` + toNode
	wantACLink := acNode + ` .. ` + linkNode
	if !strings.Contains(got, wantFrom) {
		t.Errorf("missing from leg with endpoint multiplicity: want %q in:\n%s", wantFrom, got)
	}
	if !strings.Contains(got, wantTo) {
		t.Errorf("missing directed to leg with endpoint multiplicity: want %q in:\n%s", wantTo, got)
	}
	if strings.Contains(got, wantTo+` : relates`) {
		t.Errorf("association name should not label endpoint legs:\n%s", got)
	}
	if !strings.Contains(got, wantACLink) {
		t.Errorf("missing dotted association-class link: want %q in:\n%s", wantACLink, got)
	}
}

func TestGenerateDirectAssociationUmlCommentUsesLinkNode(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "player"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	any := helper.Must(model_class.NewMultiplicity("any"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "has_customers"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Has Customers", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: any},
		model_class.AssociationOptions{UmlComment: "very import to users"},
	)

	subdomain := model_domain.Subdomain{
		Key: subdomainKey,
		Classes: map[identity.Key]model_class.Class{
			fromKey: model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"}),
			toKey:   model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Player"}),
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}
	model := core.Model{
		Key: "test_direct_assoc_comment",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {Key: domainKey, Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	fromFile := convertKeyToFilename("class", fromKey.String(), "", ".md")
	got := string(writer.md[fromFile])

	fromNode := nodeIDFor("class", fromKey)
	toNode := nodeIDFor("class", toKey)
	linkNode := nodeIDFor("assoc", assocKey)

	assert.Contains(t, got, `class `+linkNode+`["Has Customers"]`)
	assert.Contains(t, got, `<<association>> `+linkNode)
	assert.Contains(t, got, `note for `+linkNode+` "very import to users"`)
	assert.Contains(t, got, fromNode+` "1" -- `+linkNode)
	assert.Contains(t, got, linkNode+` --> "*" `+toNode)
	assert.NotContains(t, got, fromNode+` "1" --> "*" `+toNode+` : Has Customers`)
	assert.Contains(t, got, "style "+linkNode+" stroke:#333,stroke-dasharray:5 5")
}
