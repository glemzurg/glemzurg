package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

func TestGenerateShowsExpressionParseErrors(t *testing.T) {
	model := test_helper.GetTestModel()

	var target model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				target = class
				goto found
			}
		}
	}
found:
	if target.Key.KeyType == "" {
		t.Skip("test model has no classes")
	}

	invKey, err := identity.NewClassInvariantKey(target.Key, "99")
	if err != nil {
		t.Fatalf("NewClassInvariantKey: %v", err)
	}
	spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "(((", nil)
	if err != nil {
		t.Fatalf("NewExpressionSpec: %v", err)
	}
	target.SetInvariants(append(target.Invariants, model_logic.NewLogic(
		invKey,
		model_logic.LogicTypeAssessment,
		"broken invariant",
		"",
		spec,
		nil,
	)))

	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			if _, ok := subdomain.Classes[target.Key]; ok {
				subdomain.Classes[target.Key] = target
				domain.Subdomains[sKey] = subdomain
				model.Domains[dKey] = domain
			}
		}
	}

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter failed: %v", err)
	}

	classFile := convertKeyToFilename("class", target.Key.String(), "", ".md")
	body := string(writer.md[classFile])
	if !strings.Contains(body, "Expression Parse Errors") {
		t.Errorf("expected expression error banner on class page, got: %s", body)
	}
	if !strings.Contains(body, "parse-error-spec") {
		t.Errorf("expected red styling on broken expression, got: %s", body)
	}

	var subdomainFile string
	for _, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			if _, ok := subdomain.Classes[target.Key]; ok {
				subdomainFile = convertKeyToFilename("subdomain", subdomainKey.String(), "", ".md")
				break
			}
		}
		if subdomainFile != "" {
			break
		}
	}
	subdomainBody := string(writer.md[subdomainFile])
	if subdomainBody == "" {
		t.Fatalf("expected subdomain page %s listing class %s", subdomainFile, target.Name)
	}
	if !strings.Contains(subdomainBody, "parse-error-marker") {
		t.Errorf("expected parse warning marker in subdomain class list, got: %s", subdomainBody)
	}

	modelBody := string(writer.md["model.md"])
	if !strings.Contains(modelBody, "Model Parse Errors") {
		t.Errorf("expected model summary banner, got: %s", modelBody)
	}
	if !strings.Contains(modelBody, target.Name) {
		t.Errorf("expected broken class named in model summary, got: %s", modelBody)
	}
}
