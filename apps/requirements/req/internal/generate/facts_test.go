package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/modelfacts"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSubdomainFactsPages(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	subdomain, err := modelfacts.FindSubdomain(model, modelfacts.SubdomainPath{
		DomainSubKey:    "domain_a",
		SubdomainSubKey: "subdomain_a",
	})
	require.NoError(t, err)

	facts := modelfacts.FactsForSubdomain(subdomain)
	require.NotEmpty(t, facts.Associations)
	require.NotEmpty(t, facts.Indexes)

	factsFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "facts", ".md")
	factsBody, ok := writer.md[factsFile]
	require.True(t, ok, "expected facts page %s", factsFile)

	factsText := string(factsBody)
	assert.Contains(t, factsText, "# Model Facts — "+subdomain.Name)
	assert.Contains(t, factsText, "## Associations")
	assert.Contains(t, factsText, "## Indexes")
	for _, fact := range facts.Associations {
		assert.Contains(t, factsText, "- "+fact)
	}
	for _, fact := range facts.Indexes {
		assert.Contains(t, factsText, "- "+fact)
	}

	subdomainFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")
	subdomainBody, ok := writer.md[subdomainFile]
	require.True(t, ok, "expected subdomain page %s", subdomainFile)
	assert.Contains(t, string(subdomainBody), "[Model facts]("+factsFile+")")
}

func TestGenerateSingleSubdomainFactsOnDomainPage(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	domain, subdomain, ok := findSingleSubdomainDomain(model)
	require.True(t, ok, "test model should include a single-subdomain domain")

	factsFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "facts", ".md")
	_, ok = writer.md[factsFile]
	require.True(t, ok, "expected facts page %s for single-subdomain domain", factsFile)

	domainFile := convertKeyToFilename("domain", domain.Key.String(), "", ".md")
	domainBody, ok := writer.md[domainFile]
	require.True(t, ok, "expected domain page %s", domainFile)
	assert.Contains(t, string(domainBody), "[Model facts]("+factsFile+")")

	subdomainFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")
	_, hasSubdomainPage := writer.md[subdomainFile]
	assert.False(t, hasSubdomainPage, "single-subdomain domain should not emit %s", subdomainFile)
}

func TestGenerateSubdomainFactsBreadcrumb(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	subdomain, err := modelfacts.FindSubdomain(model, modelfacts.SubdomainPath{
		DomainSubKey:    "domain_a",
		SubdomainSubKey: "subdomain_a",
	})
	require.NoError(t, err)

	factsFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "facts", ".md")
	factsBody := string(writer.md[factsFile])
	subdomainFile := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")

	assert.Contains(t, factsBody, "["+subdomain.Name+"]("+subdomainFile+")")
	assert.NotContains(t, factsBody, subdomain.Name+"]("+subdomainFile+") /", "multi-subdomain facts breadcrumb should end at subdomain page")
}
