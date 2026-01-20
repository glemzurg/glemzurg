package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, "usecase1")
	require.NoError(t, err)
	scenarioKey, err := identity.NewScenarioKey(useCaseKey, "scenario1")
	require.NoError(t, err)
	objKey, err := identity.NewScenarioObjectKey(scenarioKey, "obj1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)

	original := model_scenario.Object{
		Key:          objKey,
		ObjectNumber: 1,
		Name:         "Object1",
		NameStyle:    "name",
		ClassKey:     classKey,
		Multi:        true,
		UmlComment:   "comment",
	}

	inOut := FromRequirementsObject(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
