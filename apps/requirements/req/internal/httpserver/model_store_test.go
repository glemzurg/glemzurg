package httpserver

import (
	"errors"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

func TestModelStoreErrorSlot(t *testing.T) {
	store := NewModelStore()

	// No error recorded initially.
	if _, ok := store.GetModelError("test_model"); ok {
		t.Fatal("expected no error initially")
	}

	// Recording an error makes it retrievable.
	store.SetModelError("test_model", errors.New("parse failed: line 8"))
	msg, ok := store.GetModelError("test_model")
	if !ok || msg != "parse failed: line 8" {
		t.Fatalf("expected recorded error, got ok=%v msg=%q", ok, msg)
	}

	// A nil error still records a non-empty message.
	store.SetModelError("other", nil)
	if msg, ok := store.GetModelError("other"); !ok || msg == "" {
		t.Fatalf("expected fallback message for nil error, got ok=%v msg=%q", ok, msg)
	}
}

func TestModelStoreSetModelClearsError(t *testing.T) {
	store := NewModelStore()

	store.SetModelError("test_model", errors.New("boom"))
	if _, ok := store.GetModelError("test_model"); !ok {
		t.Fatal("expected recorded error before recovery")
	}

	// A successful SetModel clears the recorded error.
	model := test_helper.GetTestModel()
	if err := store.SetModel("test_model", &model, nil); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}
	if _, ok := store.GetModelError("test_model"); ok {
		t.Fatal("expected error cleared after successful SetModel")
	}
}
