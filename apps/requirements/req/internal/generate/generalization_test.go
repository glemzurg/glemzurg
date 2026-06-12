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

// Classes participating in a generalization must be rendered as their own
// class boxes (not just referenced by node id in the inheritance arrow).
// Regression for the bug where RegardingClasses filtered them out.
func TestGenerateClassBoxesForGeneralizationParticipants(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("dg"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sg"))

	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "shape_types"))
	parentKey := helper.Must(identity.NewClassKey(subdomainKey, "shape"))
	childKey := helper.Must(identity.NewClassKey(subdomainKey, "circle"))

	parent := model_class.NewClass(parentKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: &genKey, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Shape", Details: "", UnfinishedNotes: "", UmlComment: ""})
	child := model_class.NewClass(childKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: &genKey}, model_class.ClassDetails{Name: "Circle", Details: "", UnfinishedNotes: "", UmlComment: ""})
	gen := model_class.NewGeneralization(genKey, "Shape Types", "", "", true, true, "")

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S G",
		Classes: map[identity.Key]model_class.Class{
			parentKey: parent,
			childKey:  child,
		},
		Generalizations: map[identity.Key]model_class.Generalization{
			genKey: gen,
		},
	}
	domain := model_domain.Domain{
		Key:        domainKey,
		Name:       "D G",
		Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain},
	}
	model := core.Model{
		Key:     "test_generalization",
		Name:    "Test",
		Domains: map[identity.Key]model_domain.Domain{domainKey: domain},
	}

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter: %v", err)
	}

	parentFile := convertKeyToFilename("class", parentKey.String(), "", ".md")
	body, ok := writer.md[parentFile]
	if !ok {
		t.Fatalf("expected page for parent class (%s)", parentFile)
	}
	got := string(body)

	parentNode := nodeIDFor(parentKey)
	childNode := nodeIDFor(childKey)

	// Both classes get their own labelled boxes.
	if !strings.Contains(got, `class `+parentNode+`["Shape"]`) {
		t.Errorf("missing labelled class box for parent (Shape) in:\n%s", got)
	}
	if !strings.Contains(got, `class `+childNode+`["Circle"]`) {
		t.Errorf("missing labelled class box for child (Circle) in:\n%s", got)
	}
	// The inheritance arrow connects them.
	if !strings.Contains(got, parentNode+" <|-- "+childNode) {
		t.Errorf("missing inheritance arrow %s <|-- %s in:\n%s", parentNode, childNode, got)
	}
}
