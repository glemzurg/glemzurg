package parser_ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/require"
)

func TestRejectDomainLevelModelContent(t *testing.T) {
	t.Parallel()
	domainDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(domainDir, "classes"), 0755))

	err := rejectDomainLevelModelContent(domainDir)
	require.Error(t, err)
	var pe *ParseError
	require.ErrorAs(t, err, &pe)
	require.Equal(t, ErrDomainDirInvalid, pe.Code)
}

func TestWriteModelUsesDefaultSubdomainDirectory(t *testing.T) {
	model := test_helper.GetTestModel()
	modelDir := filepath.Join(t.TempDir(), model.Key)
	require.NoError(t, WriteModel(model, modelDir))

	domainBDir := filepath.Join(modelDir, "domains", "domain_b")
	require.DirExists(t, filepath.Join(domainBDir, "subdomains", "default"))
	require.FileExists(t, filepath.Join(domainBDir, "subdomains", "default", "subdomain.json"))
	require.DirExists(t, filepath.Join(domainBDir, "subdomains", "default", "classes"))
	require.NoDirExists(t, filepath.Join(domainBDir, "classes"))
}

func TestReadModelRejectsDomainLevelClassesDirectory(t *testing.T) {
	modelDir := t.TempDir()
	domainDir := filepath.Join(modelDir, "domains", "sales")
	require.NoError(t, os.MkdirAll(filepath.Join(domainDir, "classes", "order"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(domainDir, "domain.json"),
		[]byte(`{"name":"Sales"}`),
		0600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(modelDir, "model.json"),
		[]byte(`{"name":"Test"}`),
		0600,
	))

	_, err := ReadModel(modelDir)
	require.Error(t, err)
	var pe *ParseError
	require.ErrorAs(t, err, &pe)
	require.Equal(t, ErrDomainDirInvalid, pe.Code)
}
