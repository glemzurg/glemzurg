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
	if strings.Contains(body, "This Model Has Parse Errors") {
		t.Errorf("class page should not show model.md hub banner, got: %s", body)
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
	if !strings.Contains(modelBody, "This Model Has Parse Errors") {
		t.Errorf("expected parse error hub on model.md, got: %s", modelBody)
	}
	if !strings.Contains(modelBody, "\n\n# ") {
		t.Errorf("expected blank line between parse banner and model title heading, got: %s", modelBody[:min(200, len(modelBody))])
	}
	classHref := convertKeyToFilename("class", target.Key.String(), "", ".md")
	if !strings.Contains(modelBody, `href="`+classHref+`"`) {
		t.Errorf("expected model.md to link to class %s, got: %s", classHref, modelBody)
	}
	if !strings.Contains(modelBody, target.Name) {
		t.Errorf("expected model.md to name class %q, got: %s", target.Name, modelBody)
	}
	if strings.Contains(modelBody, "class invariant 0:") {
		t.Errorf("model.md should link to classes, not duplicate class error details, got: %s", modelBody)
	}
}

func TestModelSummaryBannerHub(t *testing.T) {
	model := test_helper.GetTestModel()

	var classKey identity.Key
	var className string
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for ck, class := range subdomain.Classes {
				classKey = ck
				className = class.Name
				goto found
			}
		}
	}
found:
	if classKey.KeyType == "" {
		t.Skip("test model has no classes")
	}

	invKey, err := identity.NewClassInvariantKey(classKey, "99")
	if err != nil {
		t.Fatalf("NewClassInvariantKey: %v", err)
	}
	spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "(((", nil)
	if err != nil {
		t.Fatalf("NewExpressionSpec: %v", err)
	}
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[classKey]; ok {
				class.SetInvariants(append(class.Invariants, model_logic.NewLogic(
					invKey,
					model_logic.LogicTypeAssessment,
					"broken invariant",
					"",
					spec,
					nil,
				)))
				subdomain.Classes[classKey] = class
				domain.Subdomains[sKey] = subdomain
				model.Domains[dKey] = domain
			}
		}
	}

	classHref := convertKeyToFilename("class", classKey.String(), "", ".md")

	tests := []struct {
		name       string
		fileErrors map[string]string
	}{
		{name: "class expression error only", fileErrors: nil},
		{
			name: "class file error only",
			fileErrors: map[string]string{
				classKey.String(): "yaml: line 1: syntax error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			idx := BuildParseIssueIndex(&model, tc.fileErrors)
			banner := idx.ModelSummaryBanner(&model)
			if banner == "" {
				t.Fatal("expected model.md hub banner")
			}
			if !strings.Contains(banner, "This Model Has Parse Errors") {
				t.Errorf("expected hub heading, got: %s", banner)
			}
			if !strings.Contains(banner, `href="`+classHref+`"`) {
				t.Errorf("expected link to %s, got: %s", classHref, banner)
			}
			if !strings.Contains(banner, className) {
				t.Errorf("expected class name %q in hub, got: %s", className, banner)
			}
			if strings.Contains(banner, "yaml: line 1") || strings.Contains(banner, "class invariant 0:") {
				t.Errorf("hub should not duplicate error details, got: %s", banner)
			}
		})
	}
}
