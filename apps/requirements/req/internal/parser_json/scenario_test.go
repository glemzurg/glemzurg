package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenarioInOutConversionRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, "usecase1")
	require.NoError(t, err)
	scenarioKey, err := identity.NewScenarioKey(useCaseKey, "scenario1")
	require.NoError(t, err)
	userObjKey, err := identity.NewScenarioObjectKey(scenarioKey, "user")
	require.NoError(t, err)
	systemObjKey, err := identity.NewScenarioObjectKey(scenarioKey, "system")
	require.NoError(t, err)
	userClassKey, err := identity.NewClassKey(subdomainKey, "user_class")
	require.NoError(t, err)
	systemClassKey, err := identity.NewClassKey(subdomainKey, "system_class")
	require.NoError(t, err)

	loginEventKey, err := identity.NewEventKey(userClassKey, "login")
	require.NoError(t, err)

	steps := model_scenario.Node{
		Description:   "User enters credentials",
		EventKey:      &loginEventKey,
		FromObjectKey: userObjKey,
		ToObjectKey:   systemObjKey,
	}
	original := model_scenario.Scenario{
		Key:     scenarioKey,
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps:   &steps,
		Objects: map[identity.Key]model_scenario.Object{
			userObjKey: {
				Key:          userObjKey,
				ObjectNumber: 1,
				Name:         "User",
				NameStyle:    "name",
				ClassKey:     userClassKey,
				Multi:        false,
			},
			systemObjKey: {
				Key:          systemObjKey,
				ObjectNumber: 2,
				Name:         "System",
				NameStyle:    "name",
				ClassKey:     systemClassKey,
				Multi:        false,
			},
		},
	}

	inOut := FromRequirementsScenario(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
