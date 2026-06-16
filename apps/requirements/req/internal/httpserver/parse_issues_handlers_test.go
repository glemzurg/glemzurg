package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

func addBrokenInvariantToFirstClass(model *core.Model) model_class.Class {
	var target model_class.Class
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				target = class
				invKey, err := identity.NewClassInvariantKey(classKey, "99")
				if err != nil {
					panic(err)
				}
				spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "(((", nil)
				if err != nil {
					panic(err)
				}
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
				return target
			}
		}
	}
	return target
}

func TestRenderMDShowsParseIssueBanner(t *testing.T) {
	store := NewModelStore()
	model := test_helper.GetTestModel()
	if addBrokenInvariantToFirstClass(&model).Key.KeyType == "" {
		t.Skip("test model has no classes")
	}

	if err := store.SetModel("test_model", &model, nil); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	server := NewServer(store)
	code, body := requestMD(server, "/test_model/model.md")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if !strings.Contains(body, "This model has parse errors") {
		t.Errorf("expected global parse banner, got: %s", body)
	}
}

func TestHomeHandlerShowsParseIssueMarker(t *testing.T) {
	store := NewModelStore()
	model := test_helper.GetTestModel()
	if addBrokenInvariantToFirstClass(&model).Key.KeyType == "" {
		t.Skip("test model has no classes")
	}
	if err := store.SetModel("test_model", &model, nil); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	server := NewServer(store)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "parse-error-marker") {
		t.Errorf("expected warning marker on home page, got: %s", body)
	}
}
