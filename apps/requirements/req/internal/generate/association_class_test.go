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

// nodeIDFor mirrors the template "nodeid" func: sanitize key string for mermaid.
func nodeIDFor(idtype string, key identity.Key) string {
	keyStr := key.String()
	keyStr = strings.ReplaceAll(keyStr, "/", "_")
	keyStr = strings.ReplaceAll(keyStr, "-", "_")
	keyStr = strings.ReplaceAll(keyStr, ".", "_")
	return idtype + "_" + keyStr
}

// buildAssocClassTestModel returns a minimal model with three classes (A, B, C)
// and an association A→B carrying AssociationClassKey = C.
func buildAssocClassTestModel(t *testing.T) (model core.Model, aKey, bKey, cKey identity.Key) {
	t.Helper()
	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	aKey = helper.Must(identity.NewClassKey(subdomainKey, "a_class"))
	bKey = helper.Must(identity.NewClassKey(subdomainKey, "b_class"))
	cKey = helper.Must(identity.NewClassKey(subdomainKey, "c_class"))

	one := helper.Must(model_class.NewMultiplicity("1"))
	assocName := "links"
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, aKey, bKey, assocName))
	assoc := model_class.NewAssociation(
		assocKey, assocName, "details",
		model_class.AssociationEnd{ClassKey: aKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: bKey, Multiplicity: one},
		&cKey,
		"",
	)

	subdomain := model_domain.Subdomain{
		Key:  subdomainKey,
		Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			aKey: model_class.NewClass(aKey, "A", "", nil, nil, nil, ""),
			bKey: model_class.NewClass(bKey, "B", "", nil, nil, nil, ""),
			cKey: model_class.NewClass(cKey, "C", "", nil, nil, nil, ""),
		},
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: assoc,
		},
	}
	domain := model_domain.Domain{
		Key:        domainKey,
		Name:       "D X",
		Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain},
	}
	model = core.Model{
		Key:     "test_assoc_class",
		Name:    "Test",
		Details: "",
		Domains: map[identity.Key]model_domain.Domain{domainKey: domain},
	}
	return model, aKey, bKey, cKey
}

// The class diagram on an endpoint's page renders the association class with
// dashed lines to BOTH endpoints, and the association class's label carries
// the «association class» annotation.
func TestGenerateAssociationClassMermaid(t *testing.T) {
	model, aKey, bKey, cKey := buildAssocClassTestModel(t)

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter: %v", err)
	}

	aFile := convertKeyToFilename("class", aKey.String(), "", ".md")
	body, ok := writer.md[aFile]
	if !ok {
		t.Fatalf("expected page for class A (%s)", aFile)
	}
	got := string(body)

	aNode := nodeIDFor("class", aKey)
	bNode := nodeIDFor("class", bKey)
	cNode := nodeIDFor("class", cKey)

	// Two dashed connections — one to each endpoint of the association.
	wantFrom := cNode + " ..> " + aNode
	wantTo := cNode + " ..> " + bNode
	if !strings.Contains(got, wantFrom) {
		t.Errorf("missing dashed connection to from-endpoint: want %q in:\n%s", wantFrom, got)
	}
	if !strings.Contains(got, wantTo) {
		t.Errorf("missing dashed connection to to-endpoint: want %q in:\n%s", wantTo, got)
	}

	// The association class's node label carries the annotation.
	wantLabel := `«association class» C`
	if !strings.Contains(got, wantLabel) {
		t.Errorf("expected %q on the association class node, got:\n%s", wantLabel, got)
	}

	// A and B (the endpoints) must NOT carry the annotation.
	for _, name := range []string{`"«association class» A"`, `"«association class» B"`} {
		if strings.Contains(got, name) {
			t.Errorf("endpoint class should not be tagged as association class: %s", name)
		}
	}
}

// Plain classes (no AssociationClassKey anywhere) get no annotation.
func TestGenerateNoAssociationClassAnnotation(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	xKey := helper.Must(identity.NewClassKey(subdomainKey, "x_class"))

	subdomain := model_domain.Subdomain{
		Key: subdomainKey, Name: "S X",
		Classes: map[identity.Key]model_class.Class{
			xKey: model_class.NewClass(xKey, "X", "", nil, nil, nil, ""),
		},
	}
	domain := model_domain.Domain{
		Key: domainKey, Name: "D X",
		Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain},
	}
	model := core.Model{
		Key:     "test_no_assoc_class",
		Name:    "Test",
		Domains: map[identity.Key]model_domain.Domain{domainKey: domain},
	}

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter: %v", err)
	}
	xFile := convertKeyToFilename("class", xKey.String(), "", ".md")
	if strings.Contains(string(writer.md[xFile]), "«association class»") {
		t.Errorf("class with no association-class role should not carry the annotation")
	}
}
