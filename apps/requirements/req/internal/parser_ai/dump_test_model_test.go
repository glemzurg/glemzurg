package parser_ai

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
)

func TestDumpTestModel(t *testing.T) {

	t.Skip("DumpTestModel is a utility for dumping the test model to a directory for manual inspection; not a real test")

	model := test_helper.GetTestModel()

	// Convert to input model.
	input, err := ConvertFromModel(&model)
	assert.NoError(t, err, "ConvertFromModel should succeed")

	// Write to a fixed directory that won't be cleaned up.
	outputDir := "/workspaces/glemzurg/test_model_dump"
	err = writeModelTree(input, outputDir)
	assert.NoError(t, err, "writeModelTree should succeed")

	fmt.Printf("Model written to: %s\n", outputDir)
}
