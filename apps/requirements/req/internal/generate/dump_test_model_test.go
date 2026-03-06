package generate

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
)

// discardWriter is a ContentWriter that discards all output.
type discardWriter struct{}

func (d discardWriter) WriteMarkdown(_ string, _ []byte) error { return nil }
func (d discardWriter) WriteSVG(_ string, _ []byte) error      { return nil }
func (d discardWriter) WriteCSS(_ []byte) error                { return nil }

// TestGenerateTemplates exercises all templates with the test model to catch
// runtime errors (missing fields, type mismatches) without writing files.
func TestGenerateTemplates(t *testing.T) {
	model := test_helper.GetTestModel()
	err := GenerateMdToWriter(false, model, discardWriter{})
	assert.NoError(t, err, "GenerateMdToWriter should succeed with test model")
}

func TestDumpTestModel(t *testing.T) {

	t.Skip("DumpTestModel is a utility for dumping the test model to a directory for manual inspection; not a real test")

	model := test_helper.GetTestModel()

	// Write to the dump folder within this package for manual inspection.
	outputDir := "/workspaces/glemzurg/test_model_dump"
	err := GenerateMdFromModel(true, outputDir, model)
	assert.NoError(t, err, "GenerateMdFromModel should succeed")

	fmt.Printf("Model written to: %s\n", outputDir)
}
