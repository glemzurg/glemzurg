package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeInOutConversionRoundTrip(t *testing.T) {
	// Create proper identity keys for testing.
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "subdomain1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, "usecase1")
	require.NoError(t, err)
	scenarioKey, err := identity.NewScenarioKey(useCaseKey, "scenario1")
	require.NoError(t, err)
	clientObjKey, err := identity.NewScenarioObjectKey(scenarioKey, "client")
	require.NoError(t, err)
	serverObjKey, err := identity.NewScenarioObjectKey(scenarioKey, "server")
	require.NoError(t, err)
	eventKey, err := identity.NewEventKey(classKey, "request")
	require.NoError(t, err)
	nestedScenarioKey, err := identity.NewScenarioKey(useCaseKey, "nested_scenario")
	require.NoError(t, err)
	attributeKey, err := identity.NewAttributeKey(classKey, "status")
	require.NoError(t, err)

	original := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description: "First step",
			},
			{
				Description: "Second step",
			},
		},
		Cases: []model_scenario.Case{
			{
				Condition: "success",
			},
			{
				Condition: "alternative",
			},
		},
		Loop:          "while condition",
		Description:   "Main scenario",
		FromObjectKey: &clientObjKey,
		ToObjectKey:   &serverObjKey,
		EventKey:      &eventKey,
		ScenarioKey:   &nestedScenarioKey,
		AttributeKey:  &attributeKey,
		IsDelete:      true,
	}

	inOut := FromRequirementsNode(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
