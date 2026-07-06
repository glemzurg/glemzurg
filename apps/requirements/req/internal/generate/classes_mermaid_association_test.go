package generate

import (
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

	toKey := helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "jurisdiction"))
	jurisdictionAttrKey := helper.Must(identity.NewAttributeKey(toKey, "jurisdiction_code"))
	toClass := model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	toClass.SetAttributes([]model_class.Attribute{
		mermaidTestAttribute(jurisdictionAttrKey, "Jurisdiction Code"),
	})
	uniqueness := model_class.NewAssociationUniqueness(nil, []identity.Key{jurisdictionAttrKey})
	fromClass := model_class.NewClass(
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "partner")),
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Partner"},
	)
	assert.Equal(t, "{unique → Jurisdiction Code}", associationUniquenessMermaidTag(&uniqueness, fromClass, toClass))
}

func TestGenerateClassesMermaidDirectAssociationUniqueness(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "from_class"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "to_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	fromAttrKey := helper.Must(identity.NewAttributeKey(fromKey, "abbr"))
	toAttrKey := helper.Must(identity.NewAttributeKey(toKey, "jurisdiction_code"))
	uniqueness := model_class.NewAssociationUniqueness(
		[]identity.Key{fromAttrKey},
		[]identity.Key{toAttrKey},
	)
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "owns"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "owns", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one},
		model_class.AssociationOptions{Uniqueness: &uniqueness},
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
							fromKey: mermaidTestClassWithAttrs(fromKey, "From", []model_class.Attribute{
								mermaidTestAttribute(fromAttrKey, "Abbr"),
							}),
							toKey: mermaidTestClassWithAttrs(toKey, "To", []model_class.Attribute{
								mermaidTestAttribute(toAttrKey, "Jurisdiction Code"),
							}),
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
	want := fromNode + ` "1" --> "1" ` + toNode + ` : owns<br/>{unique Abbr → Jurisdiction Code}`
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
	toAttrKey := helper.Must(identity.NewAttributeKey(bKey, "code"))
	uniqueness := model_class.NewAssociationUniqueness(nil, []identity.Key{toAttrKey})
	bClass := mermaidTestClassWithAttrs(bKey, "B", []model_class.Attribute{
		mermaidTestAttribute(toAttrKey, "Code"),
	})
	cKey := helper.Must(identity.NewClassKey(subdomainKey, "c_class"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "links", Details: "details"},
		model_class.AssociationEnd{ClassKey: aKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: bKey, Multiplicity: one},
		model_class.AssociationOptions{
			AssociationClassKey: &cKey,
			Uniqueness:          &uniqueness,
		},
	)
	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomain := model.Domains[domainKey].Subdomains[subdomainKey]
	subdomain.Classes[bKey] = bClass
	subdomain.ClassAssociations[assocKey] = assoc
	model.Domains[domainKey].Subdomains[subdomainKey] = subdomain

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	aFile := convertKeyToFilename("class", aKey.String(), "", ".md")
	got := string(writer.md[aFile])

	linkNode := nodeIDFor("assoc", assocKey)
	wantLinkNode := `class ` + linkNode + `["links<br/>{unique → Code}"]`
	wantLinkStereotype := `<<association>> ` + linkNode
	assert.Contains(t, got, wantLinkNode)
	assert.Contains(t, got, wantLinkStereotype)
}

func TestGenerateClassesMermaidLeaderboardUniquenessLabel(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("platform"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "leaderboards"))
	leaderboardKey := helper.Must(identity.NewClassKey(subdomainKey, "leaderboard"))
	rowKey := helper.Must(identity.NewClassKey(subdomainKey, "leaderboard_row"))
	rankKey := helper.Must(identity.NewAttributeKey(rowKey, "rank"))
	tieBreakKey := helper.Must(identity.NewAttributeKey(rowKey, "tie_break_rank"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	any := helper.Must(model_class.NewMultiplicity("any"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, leaderboardKey, rowKey, "is_composed_of"))
	uniqueness := model_class.NewAssociationUniqueness(nil, []identity.Key{rankKey, tieBreakKey})
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Is Composed Of", Details: ""},
		model_class.AssociationEnd{ClassKey: leaderboardKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: rowKey, Multiplicity: any},
		model_class.AssociationOptions{Uniqueness: &uniqueness},
	)

	rowClass := mermaidTestClassWithAttrs(rowKey, "Leaderboard Row", []model_class.Attribute{
		mermaidTestAttribute(rankKey, "Rank"),
		mermaidTestAttribute(tieBreakKey, "Tie Break Rank"),
	})
	model := core.Model{
		Key: "leaderboards",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key: domainKey,
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key: subdomainKey,
						Classes: map[identity.Key]model_class.Class{
							leaderboardKey: mermaidTestClassWithAttrs(leaderboardKey, "Leaderboard", nil),
							rowKey:         rowClass,
						},
						ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
					},
				},
			},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	leaderboardFile := convertKeyToFilename("class", leaderboardKey.String(), "", ".md")
	got := string(writer.md[leaderboardFile])

	fromNode := nodeIDFor("class", leaderboardKey)
	toNode := nodeIDFor("class", rowKey)
	want := fromNode + ` "1" --> "*" ` + toNode + ` : Is Composed Of<br/>{unique → Rank, Tie Break Rank}`
	assert.Contains(t, got, want)
	assert.NotContains(t, got, `{unique:`)
	assert.NotContains(t, got, `unique:`)
}

func TestGenerateClassesMermaidOmitsAnyUniqueness(t *testing.T) {
	t.Parallel()

	model, aKey, _, _ := buildAssocClassTestModel(t)
	aFile := convertKeyToFilename("class", aKey.String(), "", ".md")

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))
	got := string(writer.md[aFile])

	assert.NotContains(t, got, `{unique:`)
	assert.Contains(t, got, `["links"]`)
}

func mermaidTestClassWithAttrs(classKey identity.Key, name string, attrs []model_class.Attribute) model_class.Class {
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: name})
	class.SetAttributes(attrs)
	return class
}

func mermaidTestAttribute(key identity.Key, name string) model_class.Attribute {
	attr, err := model_class.NewAttribute(key, model_class.AttributeDetails{Name: name, Details: ""}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	if err != nil {
		panic(err)
	}
	return attr
}
