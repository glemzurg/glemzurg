package parser_human

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

// countClasses returns the total number of classes across the whole model.
func countClasses(t *testing.T, modelPath string) int {
	t.Helper()
	model, _, err := Parse(modelPath)
	if err != nil {
		t.Fatalf("baseline Parse failed: %v", err)
	}
	n := 0
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			n += len(subdomain.Classes)
		}
	}
	return n
}

// findIsolatedClassFile returns a .class file whose placeholder does not break
// model-wide validation (no generalization superclass/subclass links).
func findIsolatedClassFile(t *testing.T, root string) string {
	t.Helper()
	const isolatedClass = "warehouse.class"
	var found string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && found == "" && strings.HasSuffix(path, isolatedClass) {
			found = path
		}
		return nil
	})
	if found == "" {
		t.Fatalf("expected %s in the written model", isolatedClass)
	}
	return found
}

// A single unparseable .class file is isolated as a ParseFailure and placeholder
// when the rest of the model still passes model.Validate().
func TestParseIsolatesBrokenClass(t *testing.T) {
	tempDir := t.TempDir()
	model := test_helper.GetTestModel()
	if err := Write(model, tempDir); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	totalClasses := countClasses(t, tempDir)
	if totalClasses == 0 {
		t.Skip("test model has no classes; nothing to isolate")
	}

	// Corrupt an isolated class file: append an unterminated YAML quote.
	classPath := findIsolatedClassFile(t, tempDir)
	orig, err := os.ReadFile(classPath)
	if err != nil {
		t.Fatalf("read class file: %v", err)
	}
	if err := os.WriteFile(classPath, append(orig, []byte("\nbroken_key: \"unterminated\n")...), 0o644); err != nil { //nolint:gosec // test fixture corruption
		t.Fatalf("corrupt class file: %v", err)
	}

	parsed, failures, err := Parse(tempDir)
	if err != nil {
		t.Fatalf("one broken class should not abort the model, got error: %v", err)
	}

	// Exactly one failure, pointing at the corrupted file.
	if len(failures) != 1 {
		t.Fatalf("expected 1 parse failure, got %d", len(failures))
	}
	if failures[0].Err == "" || failures[0].Name == "" {
		t.Errorf("failure record incomplete: %+v", failures[0])
	}

	// The model still has the same class count — the broken class is a placeholder.
	got := 0
	var placeholderFound bool
	for _, domain := range parsed.Domains {
		for _, subdomain := range domain.Subdomains {
			got += len(subdomain.Classes)
			if _, ok := subdomain.Classes[failures[0].ClassKey]; ok {
				placeholderFound = true
			}
		}
	}
	if got != totalClasses {
		t.Errorf("expected %d classes (incl. placeholder), got %d", totalClasses, got)
	}
	if !placeholderFound {
		t.Errorf("expected placeholder class %s in the model", failures[0].ClassKey.String())
	}
}

// A multiplicity written as a bare number (from_multiplicity: 1 instead of "1")
// must not panic — it is isolated as a per-class parse failure with a clear
// message.
func TestParseIsolatesMalformedMultiplicity(t *testing.T) {
	tempDir := t.TempDir()
	model := test_helper.GetTestModel()
	if err := Write(model, tempDir); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Find a class file with a quoted multiplicity and unquote it.
	// Skip classes whose placeholders break model validation (scenario events or generalizations).
	unsafeCorruptTargets := []string{
		"order.class",
		"product.class",
		"customer_class.class",
		"line_item.class",
		"vehicle.class",
		"car.class",
	}
	var corrupted string
	_ = filepath.WalkDir(tempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || corrupted != "" || !strings.HasSuffix(path, ".class") {
			return nil
		}
		for _, target := range unsafeCorruptTargets {
			if strings.HasSuffix(path, target) {
				return nil
			}
		}
		body, readErr := os.ReadFile(path) //nolint:gosec // test fixture walk over temp dir
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(body), `from_multiplicity: "1"`) {
			fixed := strings.Replace(string(body), `from_multiplicity: "1"`, `from_multiplicity: 1`, 1)
			if os.WriteFile(path, []byte(fixed), 0o644) == nil { //nolint:gosec // test fixture corruption
				corrupted = path
			}
		}
		return nil
	})
	if corrupted == "" {
		t.Skip("test model has no corruptible class with from_multiplicity: \"1\" outside the unsafe set")
	}

	parsed, failures, err := Parse(tempDir)
	if err != nil {
		t.Fatalf("a malformed multiplicity should not abort the model, got: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 parse failure, got %d", len(failures))
	}
	if !strings.Contains(failures[0].Err, "from_multiplicity") {
		t.Errorf("expected a clear multiplicity error, got: %s", failures[0].Err)
	}
	// The class is still present as a placeholder.
	found := false
	for _, domain := range parsed.Domains {
		for _, subdomain := range domain.Subdomains {
			if _, ok := subdomain.Classes[failures[0].ClassKey]; ok {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected placeholder class %s in the model", failures[0].ClassKey.String())
	}
}

// A model with no broken files parses with zero failures.
func TestParseNoFailuresOnCleanModel(t *testing.T) {
	tempDir := t.TempDir()
	model := test_helper.GetTestModel()
	if err := Write(model, tempDir); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	_, failures, err := Parse(tempDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(failures) != 0 {
		t.Errorf("expected no failures on a clean model, got %d", len(failures))
	}
}
