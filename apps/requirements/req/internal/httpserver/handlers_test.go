package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

// requestMD issues a GET for a model's .md page and returns status + body.
func requestMD(server *Server, path string) (int, string) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestRenderMDShowsErrorPage(t *testing.T) {
	store := NewModelStore()
	store.SetModelError("test_model", errors.New("failed to parse: line 8"))
	server := NewServer(store)

	code, body := requestMD(server, "/test_model/model.md")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if !strings.Contains(body, "ERROR: failed to parse: line 8") {
		t.Errorf("expected error message in body, got: %s", body)
	}
	if !strings.Contains(body, "color:#cc0000") || !strings.Contains(body, "font-weight:bold") {
		t.Errorf("expected red bold styling, got: %s", body)
	}
}

// A generation failure invalidates the whole model: every .md page shows the error.
func TestRenderMDErrorAppliesToEveryPage(t *testing.T) {
	store := NewModelStore()
	store.SetModelError("test_model", errors.New("boom"))
	server := NewServer(store)

	code, body := requestMD(server, "/test_model/class-some_class.md")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for non-root page, got %d", code)
	}
	if !strings.Contains(body, "ERROR: boom") {
		t.Errorf("expected error page for non-root .md page, got: %s", body)
	}
}

// Once the source is fixed, SetModel clears the error and real content renders.
func TestRenderMDRecoversAfterSetModel(t *testing.T) {
	store := NewModelStore()
	store.SetModelError("test_model", errors.New("boom"))
	server := NewServer(store)

	model := test_helper.GetTestModel()
	if err := store.SetModel("test_model", &model, nil); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	code, body := requestMD(server, "/test_model/model.md")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if strings.Contains(body, "ERROR: boom") {
		t.Errorf("expected real content after recovery, still saw error: %s", body)
	}
}

func TestRenderMDUsesModelWideEventSource(t *testing.T) {
	store := NewModelStore()
	model := test_helper.GetTestModel()
	if err := store.SetModel("test_model", &model, nil); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}
	server := NewServer(store)

	code, body := requestMD(server, "/test_model/model.md")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if !strings.Contains(body, `EventSource("/events/test_model")`) {
		t.Errorf("expected model-wide SSE endpoint, got: %s", body)
	}
	if strings.Contains(body, `EventSource("/events/test_model/model.md")`) {
		t.Errorf("expected per-model SSE, not per-file, got: %s", body)
	}
	if !strings.Contains(body, `pagehide`) {
		t.Errorf("expected pagehide listener to close EventSource, got: %s", body)
	}
}
