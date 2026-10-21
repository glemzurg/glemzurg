package load

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRejectsMissingModelFile(t *testing.T) {
	modelDir := filepath.Join(t.TempDir(), "sample")
	require.NoError(t, os.MkdirAll(modelDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(modelDir, "model.json"), []byte(`{"name":"Shop Books"}`), 0o600))

	annotDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(annotDir, "missing.annot.json"), []byte(`{}`), 0o600))

	_, err := Load(modelDir, annotDir)
	require.ErrorContains(t, err, "annotation model file missing: missing.json")
}
