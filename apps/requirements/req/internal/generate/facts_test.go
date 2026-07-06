package generate

import (
	"strings"
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

	facts := modelfacts.FactsForSubdomain(model, subdomain)
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
	factsLink := "[Model facts](" + factsFile + ")"
	subdomainText := string(subdomainBody)
	assert.Contains(t, subdomainText, factsLink)
	classesIdx := strings.Index(subdomainText, "## Classes")
	factsIdx := strings.Index(subdomainText, factsLink)
	require.Positive(t, classesIdx)
	require.Positive(t, factsIdx)
	assert.Greater(t, factsIdx, classesIdx, "Model facts should follow the Classes section")
	generalizationsIdx := strings.Index(subdomainText, "### Generalizations")
	if generalizationsIdx >= 0 {
		assert.Less(t, factsIdx, generalizationsIdx, "Model facts should precede generalizations")
	}
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
	factsLink := "[Model facts](" + factsFile + ")"
	domainText := string(domainBody)
	assert.Contains(t, domainText, factsLink)
	classesIdx := strings.Index(domainText, "## Classes")
	factsIdx := strings.Index(domainText, factsLink)
	require.Positive(t, classesIdx)
	require.Positive(t, factsIdx)
	assert.Greater(t, factsIdx, classesIdx, "Model facts should follow the Classes section")

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
