package generate

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
)

func TestDumpTestModel(t *testing.T) {

	t.Skip("DumpTestModel is a utility for dumping the test model to a directory for manual inspection; not a real test")

	model := test_helper.GetTestModel()

	// Write to the dump folder within this package for manual inspection.
	outputDir := "/workspaces/glemzurg/test_model_dump"
	err := GenerateMdFromModel(true, outputDir, model)
	assert.NoError(t, err, "GenerateMdFromModel should succeed")

	fmt.Printf("Model written to: %s\n", outputDir)
}
