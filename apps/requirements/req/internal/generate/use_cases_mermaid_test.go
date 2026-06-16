package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUseCasesMermaidContents_UsesGuillemetsForStereotypes(t *testing.T) {
	reqs := req_flat.NewRequirements(test_helper.GetTestModel())
	reqs.PrepLookups()

	domainLookup, _ := reqs.DomainLookup()
	require.NotEmpty(t, domainLookup)

	var domainKey string
	var useCases []model_use_case.UseCase
	for key, ucs := range reqs.DomainUseCasesLookup() {
		if len(ucs) > 0 {
			domainKey = key
			useCases = ucs
			break
		}
	}
	require.NotEmpty(t, domainKey)
	domain := domainLookup[domainKey]

	relevantUseCases, relevantActors, err := reqs.RegardingUseCases(useCases)
	require.NoError(t, err)
	require.NotEmpty(t, relevantActors)

	contents, err := generateUseCasesMermaidContents(reqs, domain, relevantUseCases, relevantActors)
	require.NoError(t, err)

	assert.Contains(t, contents, "«person»")
	assert.NotContains(t, contents, "<<")
	assert.NotContains(t, contents, ">>")
}
