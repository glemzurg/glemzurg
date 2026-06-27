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

func TestAssociationUniquenessMermaidTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantTag string
	}{
		{name: "any omitted", value: "any", wantTag: ""},
		{name: "exactly one", value: "1", wantTag: "{unique}"},
		{name: "zero or one", value: "0..1", wantTag: "{0..1}"},
		{name: "lower bound only", value: "3", wantTag: "{3}"},
		{name: "bounded range", value: "2..5", wantTag: "{2..5}"},
		{name: "lower with any upper", value: "3..many", wantTag: "{3..any}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := helper.Must(model_class.NewMultiplicity(tc.value))
			assert.Equal(t, tc.wantTag, associationUniquenessMermaidTag(m))
		})
	}
}

func TestGenerateClassesMermaidDirectAssociationUniqueness(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "from_class"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "to_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	uniqOne := helper.Must(model_class.NewMultiplicity("1"))
	uniqBounded := helper.Must(model_class.NewMultiplicity("0..1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "owns"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "owns", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		uniqBounded,
		model_class.AssociationOptions{},
	)

	model := core.Model{
		Key:  "test_direct_uniq",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D X",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S X",
						Classes: map[identity.Key]model_class.Class{
							fromKey: model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "From"}),
							toKey:   model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "To"}),
						},
						ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
					},
				},
			},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	fromFile := convertKeyToFilename("class", fromKey.String(), "", ".md")
	got := string(writer.md[fromFile])

	fromNode := nodeIDFor("class", fromKey)
	toNode := nodeIDFor("class", toKey)
	want := fromNode + ` "1" --> "1" ` + toNode + ` : owns<br/>{0..1}`
	assert.Contains(t, got, want)

	assocExact := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "owns", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		uniqOne,
		model_class.AssociationOptions{},
	)
	model.Domains[domainKey].Subdomains[subdomainKey].ClassAssociations[assocKey] = assocExact

	writer = newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))
	got = string(writer.md[fromFile])
	want = fromNode + ` "1" --> "1" ` + toNode + ` : owns<br/>{unique}`
	assert.Contains(t, got, want)
}

func TestGenerateClassesMermaidAssociationClassUniqueness(t *testing.T) {
	t.Parallel()

	model, aKey, _, _ := buildAssocClassTestModel(t)
	assocKey := helper.Must(identity.NewClassAssociationKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("dx")), "sx")),
		aKey,
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("dx")), "sx")), "b_class")),
		"links",
	))

	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("dx")), "sx"))
	bKey := helper.Must(identity.NewClassKey(subdomainKey, "b_class"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	uniq := helper.Must(model_class.NewMultiplicity("1"))
	cKey := helper.Must(identity.NewClassKey(subdomainKey, "c_class"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "links", Details: "details"},
		model_class.AssociationEnd{ClassKey: aKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: bKey, Multiplicity: one},
		uniq,
		model_class.AssociationOptions{AssociationClassKey: &cKey},
	)
	domainKey := helper.Must(identity.NewDomainKey("dx"))
	model.Domains[domainKey].Subdomains[subdomainKey].ClassAssociations[assocKey] = assoc

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	aFile := convertKeyToFilename("class", aKey.String(), "", ".md")
	got := string(writer.md[aFile])

	linkNode := nodeIDFor("assoc", assocKey)
	wantLinkNode := `class ` + linkNode + `["links<br/>{unique}"]`
	assert.Contains(t, got, wantLinkNode)
	assert.Contains(t, got, `<<association>>`)
	if idx := strings.Index(got, wantLinkNode); idx >= 0 {
		block := got[idx:]
		if end := strings.Index(block, "\n    }"); end >= 0 {
			block = block[:end]
		}
		assert.Contains(t, block, `<<association>>`, "stereotype should stay in class body after title")
	}
}

func TestGenerateClassesMermaidOmitsAnyUniqueness(t *testing.T) {
	t.Parallel()

	model, aKey, _, _ := buildAssocClassTestModel(t)
	aFile := convertKeyToFilename("class", aKey.String(), "", ".md")

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))
	got := string(writer.md[aFile])

	assert.NotContains(t, got, `{unique}`)
	assert.NotContains(t, got, `{any}`)
	assert.Contains(t, got, `["links"]`)
}
