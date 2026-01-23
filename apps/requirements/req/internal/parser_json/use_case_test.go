package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCaseInOutConversionRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, "usecase1")
	require.NoError(t, err)
	actorKey, err := identity.NewActorKey("user")
	require.NoError(t, err)
	scenario1Key, err := identity.NewScenarioKey(useCaseKey, "scenario1")
	require.NoError(t, err)
	scenario2Key, err := identity.NewScenarioKey(useCaseKey, "scenario2")
	require.NoError(t, err)

	original := model_use_case.UseCase{
		Key:        useCaseKey,
		Name:       "Login Use Case",
		Details:    "User logs into the system",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "Login flow",
		Actors: map[identity.Key]model_use_case.Actor{
			actorKey: {
				UmlComment: "The user",
			},
		},
		Scenarios: map[identity.Key]model_scenario.Scenario{
			scenario1Key: {
				Key:  scenario1Key,
				Name: "Scenario1",
			},
			scenario2Key: {
				Key:  scenario2Key,
				Name: "Scenario2",
			},
		},
	}

	inOut := FromRequirementsUseCase(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
