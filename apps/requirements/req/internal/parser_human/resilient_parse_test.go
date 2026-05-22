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

// findFirstClassFile returns the path of any one .class file under root.
func findFirstClassFile(t *testing.T, root string) string {
	t.Helper()
	var found string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && found == "" && strings.HasSuffix(path, ".class") {
			found = path
		}
		return nil
	})
	if found == "" {
		t.Fatal("expected at least one .class file in the written model")
	}
	return found
}

// A single unparseable .class file must not abort the model: it becomes a
// placeholder class and the failure is reported, while every other class parses.
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

	// Corrupt one class file: append an unterminated YAML quote.
	classPath := findFirstClassFile(t, tempDir)
	orig, err := os.ReadFile(classPath)
	if err != nil {
		t.Fatalf("read class file: %v", err)
	}
	if err := os.WriteFile(classPath, append(orig, []byte("\nbroken_key: \"unterminated\n")...), 0644); err != nil {
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
