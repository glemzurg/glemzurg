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

func TestClassesMermaidClassBoxStyle(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "order"))
	otherKey := helper.Must(identity.NewClassKey(subdomainKey, "other"))
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order"})

	tests := []struct {
		name          string
		marked        bool
		focalClassKey *identity.Key
		want          string
	}{
		{
			name:   "default unmarked non-focal",
			marked: false,
			want:   "",
		},
		{
			name:   "marked only",
			marked: true,
			want:   classesMermaidMarkedClassStyle,
		},
		{
			name:          "focal only",
			marked:        false,
			focalClassKey: &classKey,
			want:          classesMermaidFocalClassStyle,
		},
		{
			name:          "marked and focal combine",
			marked:        true,
			focalClassKey: &classKey,
			want:          classesMermaidMarkedClassStyle + "," + classesMermaidFocalClassStyle,
		},
		{
			name:          "marked with other class as focal",
			marked:        true,
			focalClassKey: &otherKey,
			want:          classesMermaidMarkedClassStyle,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := class
			c.SetMarked(tc.marked)
			assert.Equal(t, tc.want, classesMermaidClassBoxStyle(c, tc.focalClassKey))
		})
	}
}

func TestGenerateClassesMermaidYellowFillForMarkedClass(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	markedKey := helper.Must(identity.NewClassKey(subdomainKey, "marked_class"))
	plainKey := helper.Must(identity.NewClassKey(subdomainKey, "plain_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, markedKey, plainKey, "relates"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "relates", Details: ""}, model_class.AssociationEnd{ClassKey: markedKey, Multiplicity: one}, model_class.AssociationEnd{ClassKey: plainKey, Multiplicity: one}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	markedClass := model_class.NewClass(markedKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Marked"})
	markedClass.SetMarked(true)
	plainClass := model_class.NewClass(plainKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Plain"})

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			markedKey: markedClass,
			plainKey:  plainClass,
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}
	model := core.Model{
		Key:  "test_marked_class_color",
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
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	// Class pages embed the relation diagram; use the plain class page so the
	// marked peer is non-focal and only the yellow fill should apply.
	plainFile := convertKeyToFilename("class", plainKey.String(), "", ".md")
	body := string(writer.md[plainFile])

	markedNode := nodeIDFor("class", markedKey)
	plainNode := nodeIDFor("class", plainKey)
	markedStyle := "style " + markedNode + " " + classesMermaidMarkedClassStyle
	plainMarkedStyle := "style " + plainNode + " " + classesMermaidMarkedClassStyle

	assert.Contains(t, body, markedStyle, "marked class should get yellow fill style")
	assert.NotContains(t, body, plainMarkedStyle, "unmarked class should keep default (no yellow style line)")
}

func TestGenerateClassesMermaidMarkedAndFocalCombineOnClassPage(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	markedKey := helper.Must(identity.NewClassKey(subdomainKey, "marked_class"))
	plainKey := helper.Must(identity.NewClassKey(subdomainKey, "plain_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, markedKey, plainKey, "relates"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "relates", Details: ""}, model_class.AssociationEnd{ClassKey: markedKey, Multiplicity: one}, model_class.AssociationEnd{ClassKey: plainKey, Multiplicity: one}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	markedClass := model_class.NewClass(markedKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Marked"})
	markedClass.SetMarked(true)

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			markedKey: markedClass,
			plainKey:  model_class.NewClass(plainKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Plain"}),
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}
	model := core.Model{
		Key:  "test_marked_focal_combine",
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
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	classFile := convertKeyToFilename("class", markedKey.String(), "", ".md")
	body := string(writer.md[classFile])

	markedNode := nodeIDFor("class", markedKey)
	combined := "style " + markedNode + " " + classesMermaidMarkedClassStyle + "," + classesMermaidFocalClassStyle
	assert.Contains(t, body, combined, "class page for marked focal class should combine yellow fill and focal stroke")
}
