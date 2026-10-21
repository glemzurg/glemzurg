package parser_json

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/require"
)

func TestReadModelIgnoresAnnotationFiles(t *testing.T) {
	model := test_helper.GetTestModel()
	modelDir := filepath.Join(t.TempDir(), "sample")
	require.NoError(t, WriteModel(model, modelDir))

	before, err := ReadModel(modelDir)
	require.NoError(t, err)

	classFile := findTreePath(t, modelDir, func(path string, isDir bool) bool {
		return !isDir && strings.HasSuffix(path, string(filepath.Separator)+"class.json")
	})
	classAnnot := strings.TrimSuffix(classFile, ".json") + ".annot.json"
	require.NoError(t, os.WriteFile(classAnnot, []byte(`{"note":"later"}`), 0o600))

	actionsDir := findTreePath(t, modelDir, func(path string, isDir bool) bool {
		return isDir && strings.HasSuffix(path, string(filepath.Separator)+"actions")
	})
	actionAnnot := filepath.Join(actionsDir, "do.annot.json")
	require.NoError(t, os.WriteFile(actionAnnot, []byte(`{"note":"later"}`), 0o600))

	after, err := ReadModel(modelDir)
	require.NoError(t, err)
	require.Equal(t, before, after)
	for _, key := range actionKeys(after) {
		require.NotContains(t, key, "annot")
	}

	require.NoError(t, WriteModel(before, modelDir))
	_, err = os.Stat(classAnnot)
	require.NoError(t, err)
	_, err = os.Stat(actionAnnot)
	require.NoError(t, err)
}

func findTreePath(t *testing.T, root string, match func(path string, isDir bool) bool) string {
	t.Helper()
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found == "" && match(path, d.IsDir()) {
			found = path
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, found)
	return found
}

func actionKeys(model core.Model) []string {
	var keys []string
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for key := range class.Actions {
					keys = append(keys, key.String())
				}
			}
		}
	}
	return keys
}
