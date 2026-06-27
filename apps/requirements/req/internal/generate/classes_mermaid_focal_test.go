package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestGenerateClassesMermaidHighlightsFocalClassOnClassPage(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "from_class"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "to_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "relates"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "relates", Details: ""}, model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: one}, model_class.AssociationEnd{ClassKey: toKey, Multiplicity: one}, model_class.Multiplicity{}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			fromKey: model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "From"}),
			toKey:   model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "To"}),
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}
	model := core.Model{
		Key:  "test_focal_class",
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
	toFile := convertKeyToFilename("class", toKey.String(), "", ".md")
	fromBody := string(writer.md[fromFile])
	toBody := string(writer.md[toFile])

	fromNode := nodeIDFor("class", fromKey)
	toNode := nodeIDFor("class", toKey)
	fromStyle := "style " + fromNode + " " + classesMermaidFocalClassStyle
	toStyle := "style " + toNode + " " + classesMermaidFocalClassStyle

	if !strings.Contains(fromBody, fromStyle) {
		t.Errorf("from class page should highlight focal class, want %q in:\n%s", fromStyle, fromBody)
	}
	if strings.Contains(fromBody, toStyle) {
		t.Errorf("from class page should not highlight related class %q in:\n%s", toStyle, fromBody)
	}
	if !strings.Contains(toBody, toStyle) {
		t.Errorf("to class page should highlight focal class, want %q in:\n%s", toStyle, toBody)
	}
	if strings.Contains(toBody, fromStyle) {
		t.Errorf("to class page should not highlight related class %q in:\n%s", fromStyle, toBody)
	}
}
