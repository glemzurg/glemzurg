package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

// collectWriter is a ContentWriter that keeps generated files in memory.
type collectWriter struct {
	md map[string][]byte
}

func newCollectWriter() *collectWriter { return &collectWriter{md: map[string][]byte{}} }
func (c *collectWriter) WriteMarkdown(f string, b []byte) error {
	c.md[f] = b
	return nil
}
func (c *collectWriter) WriteSVG(string, []byte) error { return nil }
func (c *collectWriter) WriteCSS([]byte) error         { return nil }

func TestGenerateClassErrorBlock(t *testing.T) {
	model := test_helper.GetTestModel()

	// Pick any class from the model to mark as failed.
	var failed, healthy model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if failed.Key.KeyType == "" {
					failed = class
				} else if healthy.Key.KeyType == "" {
					healthy = class
				}
			}
		}
	}
	if failed.Key.KeyType == "" {
		t.Skip("test model has no classes")
	}

	classErrors := map[string]string{
		failed.Key.String(): "yaml: line 8: found unexpected end of stream",
	}

	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, classErrors); err != nil {
		t.Fatalf("GenerateMdToWriter failed: %v", err)
	}

	// The failed class's page is the red-bold error block.
	failedFile := convertKeyToFilename("class", failed.Key.String(), "", ".md")
	failedContent, ok := writer.md[failedFile]
	if !ok {
		t.Fatalf("expected a page for the failed class %q", failedFile)
	}
	body := string(failedContent)
	if !strings.Contains(body, "Failed to Parse") {
		t.Errorf("expected error heading on failed class page, got: %s", body)
	}
	if !strings.Contains(body, "found unexpected end of stream") {
		t.Errorf("expected the parse error on failed class page, got: %s", body)
	}
	if !strings.Contains(body, "color:#cc0000") {
		t.Errorf("expected red styling on failed class page, got: %s", body)
	}

	// A healthy class's page is unaffected — no error block.
	if healthy.Key.KeyType != "" {
		healthyFile := convertKeyToFilename("class", healthy.Key.String(), "", ".md")
		if hc, ok := writer.md[healthyFile]; ok {
			if strings.Contains(string(hc), "Failed to Parse") {
				t.Errorf("healthy class page should not contain an error block")
			}
		}
	}
}

// With no classErrors, every class page is generated normally.
func TestGenerateNoClassErrors(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	if err := GenerateMdToWriter(model, writer, nil); err != nil {
		t.Fatalf("GenerateMdToWriter failed: %v", err)
	}
	for name, content := range writer.md {
		if strings.HasPrefix(name, "class-") && strings.Contains(string(content), "Failed to Parse") {
			t.Errorf("class page %s unexpectedly contains an error block", name)
		}
	}
}
