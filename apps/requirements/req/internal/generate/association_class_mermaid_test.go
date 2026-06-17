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
	assoc := model_class.NewAssociation(
		assocKey, "relates", "",
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult},
		&acKey,
		"",
	)

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

	fromNode := nodeIDFor(fromKey)
	acNode := nodeIDFor(acKey)
	toNode := nodeIDFor(toKey)

	wantFrom := fromNode + ` "3" -- ` + acNode
	wantTo := acNode + ` --> "*" ` + toNode + ` : relates`
	if !strings.Contains(got, wantFrom) {
		t.Errorf("missing from leg with endpoint multiplicity: want %q in:\n%s", wantFrom, got)
	}
	if !strings.Contains(got, wantTo) {
		t.Errorf("missing directed to leg with endpoint multiplicity: want %q in:\n%s", wantTo, got)
	}
}
