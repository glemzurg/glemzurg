package model_scenario

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// t_strPtr returns a pointer to the given string.
func t_strPtr(s string) *string { return &s }

type ScenarioStepsSuite struct {
	suite.Suite
	// Test keys created once for the suite
	domainKey    identity.Key
	subdomainKey identity.Key
	classKey     identity.Key
	useCaseKey   identity.Key
	scenarioKey  identity.Key
	fromObjKey   *identity.Key
	toObjKey     *identity.Key
	eventKey     *identity.Key
	scenarioRef  *identity.Key
	queryKey     *identity.Key
	// Step keys for building test steps.
	stepKeys []identity.Key
}

func TestScenarioStepsSuite(t *testing.T) {
	suite.Run(t, new(ScenarioStepsSuite))
}

func (suite *ScenarioStepsSuite) SetupSuite() {
	var err error
	suite.domainKey, err = identity.NewDomainKey("test_domain")
	require.NoError(suite.T(), err)
	suite.subdomainKey, err = identity.NewSubdomainKey(suite.domainKey, "default")
	require.NoError(suite.T(), err)
	suite.classKey, err = identity.NewClassKey(suite.subdomainKey, "test_class")
	require.NoError(suite.T(), err)
	suite.useCaseKey, err = identity.NewUseCaseKey(suite.subdomainKey, "test_use_case")
	require.NoError(suite.T(), err)
	suite.scenarioKey, err = identity.NewScenarioKey(suite.useCaseKey, "test_scenario")
	require.NoError(suite.T(), err)
	fromObjKey, err := identity.NewScenarioObjectKey(suite.scenarioKey, "from_obj")
	require.NoError(suite.T(), err)
	suite.fromObjKey = &fromObjKey
	toObjKey, err := identity.NewScenarioObjectKey(suite.scenarioKey, "to_obj")
	require.NoError(suite.T(), err)
	suite.toObjKey = &toObjKey
	eventKey, err := identity.NewEventKey(suite.classKey, "test_event")
	require.NoError(suite.T(), err)
	suite.eventKey = &eventKey
	scenarioRef, err := identity.NewScenarioKey(suite.useCaseKey, "ref_scenario")
	require.NoError(suite.T(), err)
	suite.scenarioRef = &scenarioRef
	queryKey, err := identity.NewQueryKey(suite.classKey, "test_query")
	require.NoError(suite.T(), err)
	suite.queryKey = &queryKey

	// Pre-create a pool of step keys.
	for i := 0; i < 20; i++ {
		k, err := identity.NewScenarioStepKey(suite.scenarioKey, fmt.Sprintf("%d", i))
		require.NoError(suite.T(), err)
		suite.stepKeys = append(suite.stepKeys, k)
	}
}

// stepKey returns the i-th pre-created step key.
func (suite *ScenarioStepsSuite) stepKey(i int) identity.Key {
	return suite.stepKeys[i]
}

func (suite *ScenarioStepsSuite) TestValidateSequence() {
	// Valid sequence
	step := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SEQUENCE,
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := step.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: empty statements
	step = Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_SEQUENCE,
		Statements: []Step{},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "sequence must have at least one statement")
}

func (suite *ScenarioStepsSuite) TestValidateSwitch() {
	// Valid switch with case children
	step := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SWITCH,
		Statements: []Step{
			{
				Key:       suite.stepKey(1),
				StepType:  STEP_TYPE_CASE,
				Condition: "cond1",
				Statements: []Step{
					{Key: suite.stepKey(2), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
				},
			},
		},
	}
	err := step.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no cases
	step = Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_SWITCH,
		Statements: []Step{},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "switch must have at least one case")

	// Invalid: non-case child
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SWITCH,
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "switch children must all be case steps")

	// Invalid: case without condition
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SWITCH,
		Statements: []Step{
			{
				Key:      suite.stepKey(1),
				StepType: STEP_TYPE_CASE,
				Statements: []Step{
					{Key: suite.stepKey(2), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), Description: "step", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
				},
			},
		},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "case must have a condition")
}

func (suite *ScenarioStepsSuite) TestValidateCase() {
	// Valid case
	step := Step{
		Key:       suite.stepKey(0),
		StepType:  STEP_TYPE_CASE,
		Condition: "some condition",
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := step.Validate()
	assert.Nil(suite.T(), err)

	// Valid case with no statements (empty case is allowed)
	step = Step{
		Key:       suite.stepKey(0),
		StepType:  STEP_TYPE_CASE,
		Condition: "some condition",
	}
	err = step.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no condition
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_CASE,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "case must have a condition")
}

func (suite *ScenarioStepsSuite) TestValidateLoop() {
	// Valid loop
	step := Step{
		Key:       suite.stepKey(0),
		StepType:  STEP_TYPE_LOOP,
		Condition: "while true",
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := step.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no condition
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_LOOP,
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "loop must have a condition")

	// Invalid: no statements
	step = Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_LOOP,
		Condition:  "while true",
		Statements: []Step{},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "loop must have at least one statement")
}

func (suite *ScenarioStepsSuite) TestValidateLeaf() {
	// Valid event leaf
	step := Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := step.Validate()
	assert.Nil(suite.T(), err)

	// Valid scenario leaf
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = step.Validate()
	assert.Nil(suite.T(), err)

	// Valid query leaf
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      suite.queryKey,
	}
	err = step.Validate()
	assert.NoError(suite.T(), err)

	// Valid delete leaf
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
	}
	err = step.Validate()
	assert.NoError(suite.T(), err)

	// Invalid: missing leaf_type
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a leaf_type")

	// Invalid: unknown leaf_type
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr("bogus"),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "unknown leaf type 'bogus'")

	// Invalid event: no from_object_key
	step = Step{
		Key:         suite.stepKey(0),
		StepType:    STEP_TYPE_LEAF,
		LeafType:    t_strPtr(LEAF_TYPE_EVENT),
		ToObjectKey: suite.toObjKey,
		EventKey:    suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")

	// Invalid event: no to_object_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a to_object_key")

	// Invalid event: no event_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have an event_key")

	// Invalid query: no query_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "query leaf must have a query_key")

	// Invalid scenario: no scenario_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "scenario leaf must have a scenario_key")

	// Invalid delete: has to_object_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have a to_object_key")

	// Invalid delete: has event_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have event_key, scenario_key, or query_key")

	// Invalid delete: no from_object_key
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_LEAF,
		LeafType: t_strPtr(LEAF_TYPE_DELETE),
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf must have a from_object_key")

	// Invalid query: no from_object_key
	step = Step{
		Key:         suite.stepKey(0),
		StepType:    STEP_TYPE_LEAF,
		LeafType:    t_strPtr(LEAF_TYPE_QUERY),
		ToObjectKey: suite.toObjKey,
		QueryKey:    suite.queryKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "query leaf must have a from_object_key")

	// Invalid query: no to_object_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		FromObjectKey: suite.fromObjKey,
		QueryKey:      suite.queryKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "query leaf must have a to_object_key")

	// Invalid scenario: no from_object_key
	step = Step{
		Key:         suite.stepKey(0),
		StepType:    STEP_TYPE_LEAF,
		LeafType:    t_strPtr(LEAF_TYPE_SCENARIO),
		ToObjectKey: suite.toObjKey,
		ScenarioKey: suite.scenarioRef,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "scenario leaf must have a from_object_key")

	// Invalid scenario: no to_object_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "scenario leaf must have a to_object_key")

	// Cross-validation: event leaf with scenario_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf cannot have scenario_key or query_key")

	// Cross-validation: event leaf with query_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		QueryKey:      suite.queryKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf cannot have scenario_key or query_key")

	// Cross-validation: query leaf with event_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      suite.queryKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "query leaf cannot have event_key or scenario_key")

	// Cross-validation: query leaf with scenario_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      suite.queryKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "query leaf cannot have event_key or scenario_key")

	// Cross-validation: scenario leaf with event_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "scenario leaf cannot have event_key or query_key")

	// Cross-validation: scenario leaf with query_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
		QueryKey:      suite.queryKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "scenario leaf cannot have event_key or query_key")

	// Cross-validation: delete leaf with scenario_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		FromObjectKey: suite.fromObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have event_key, scenario_key, or query_key")

	// Cross-validation: delete leaf with query_key
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		FromObjectKey: suite.fromObjKey,
		QueryKey:      suite.queryKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have event_key, scenario_key, or query_key")
}

func (suite *ScenarioStepsSuite) TestValidateLeafKeyTypes() {
	// Use domainKey as a wrong key type for all FK fields.
	wrongKey := &suite.domainKey

	// FromObjectKey with wrong key type.
	step := Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: wrongKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := step.Validate()
	assert.ErrorContains(suite.T(), err, "FromObjectKey: invalid key type 'domain' for scenario object")

	// ToObjectKey with wrong key type.
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   wrongKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "ToObjectKey: invalid key type 'domain' for scenario object")

	// EventKey with wrong key type.
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      wrongKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "EventKey: invalid key type 'domain' for event")

	// QueryKey with wrong key type.
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_QUERY),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      wrongKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "QueryKey: invalid key type 'domain' for query")

	// ScenarioKey with wrong key type.
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   wrongKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "ScenarioKey: invalid key type 'domain' for scenario")

	// FromObjectKey with wrong key type on delete leaf.
	step = Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_DELETE),
		FromObjectKey: wrongKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "FromObjectKey: invalid key type 'domain' for scenario object")
}

func (suite *ScenarioStepsSuite) TestValidateRecursiveChildFailure() {
	// A bad child inside a sequence propagates the error.
	badChild := Step{
		Key:      suite.stepKey(1),
		StepType: STEP_TYPE_LEAF,
		LeafType: t_strPtr(LEAF_TYPE_EVENT),
		// Missing from_object_key, to_object_key, event_key.
	}

	// Sequence with bad child.
	step := Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_SEQUENCE,
		Statements: []Step{badChild},
	}
	err := step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")

	// Loop with bad child.
	step = Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_LOOP,
		Condition:  "while true",
		Statements: []Step{badChild},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")

	// Case with bad child.
	step = Step{
		Key:        suite.stepKey(0),
		StepType:   STEP_TYPE_CASE,
		Condition:  "some condition",
		Statements: []Step{badChild},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")

	// Switch with bad case child (case itself is valid, but its child is bad).
	step = Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SWITCH,
		Statements: []Step{
			{
				Key:        suite.stepKey(2),
				StepType:   STEP_TYPE_CASE,
				Condition:  "cond",
				Statements: []Step{badChild},
			},
		},
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")
}

func (suite *ScenarioStepsSuite) TestValidateUnknownStepType() {
	step := Step{
		Key:      suite.stepKey(0),
		StepType: "bogus",
	}
	err := step.Validate()
	assert.ErrorContains(suite.T(), err, "unknown step type 'bogus'")
}

func (suite *ScenarioStepsSuite) TestValidateKeyErrors() {
	// Empty key
	step := Step{
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := step.Validate()
	assert.ErrorContains(suite.T(), err, "KeyType")

	// Wrong key type
	step = Step{
		Key:           suite.domainKey,
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err = step.Validate()
	assert.ErrorContains(suite.T(), err, "invalid key type 'domain' for scenario step")
}

func (suite *ScenarioStepsSuite) TestValidateWithParent() {
	// Valid: correct parent
	step := Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_EVENT),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := step.ValidateWithParent(&suite.scenarioKey)
	assert.NoError(suite.T(), err)

	// Invalid: wrong parent
	otherScenarioKey, err := identity.NewScenarioKey(suite.useCaseKey, "other_scenario")
	require.NoError(suite.T(), err)
	err = step.ValidateWithParent(&otherScenarioKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent")

	// Validate calls Validate (propagates validation error)
	badStep := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_LEAF,
		LeafType: t_strPtr(LEAF_TYPE_EVENT),
		// Missing from_object_key
		ToObjectKey: suite.toObjKey,
		EventKey:    suite.eventKey,
	}
	err = badStep.ValidateWithParent(&suite.scenarioKey)
	assert.ErrorContains(suite.T(), err, "event leaf must have a from_object_key")

	// ValidateWithParent recurses into children
	root := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SEQUENCE,
		Statements: []Step{
			{Key: suite.stepKey(1), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = root.ValidateWithParent(&suite.scenarioKey)
	assert.NoError(suite.T(), err)

	// Invalid: scenario leaf references its own scenario (self-referencing)
	selfRefStep := Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   &suite.scenarioKey,
	}
	err = selfRefStep.ValidateWithParent(&suite.scenarioKey)
	assert.ErrorContains(suite.T(), err, "scenario leaf cannot reference its own scenario")

	// Valid: scenario leaf references a different scenario (not self-referencing)
	validRefStep := Step{
		Key:           suite.stepKey(0),
		StepType:      STEP_TYPE_LEAF,
		LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = validRefStep.ValidateWithParent(&suite.scenarioKey)
	assert.NoError(suite.T(), err)

	// Invalid: self-referencing scenario leaf nested inside a sequence
	selfRefRoot := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SEQUENCE,
		Statements: []Step{
			{
				Key:           suite.stepKey(1),
				StepType:      STEP_TYPE_LEAF,
				LeafType:      t_strPtr(LEAF_TYPE_SCENARIO),
				FromObjectKey: suite.fromObjKey,
				ToObjectKey:   suite.toObjKey,
				ScenarioKey:   &suite.scenarioKey,
			},
		},
	}
	err = selfRefRoot.ValidateWithParent(&suite.scenarioKey)
	assert.ErrorContains(suite.T(), err, "scenario leaf cannot reference its own scenario")
}

func (suite *ScenarioStepsSuite) TestJSON() {
	// Complex structure: sequence > [switch > [case > [leaf]], loop > [leaf]]
	root := Step{
		Key:      suite.stepKey(0),
		StepType: STEP_TYPE_SEQUENCE,
		Statements: []Step{
			{
				Key:      suite.stepKey(1),
				StepType: STEP_TYPE_SWITCH,
				Statements: []Step{
					{
						Key:       suite.stepKey(2),
						StepType:  STEP_TYPE_CASE,
						Condition: "if x > 0",
						Statements: []Step{
							{Key: suite.stepKey(3), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_EVENT), Description: "positive", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
						},
					},
				},
			},
			{
				Key:       suite.stepKey(4),
				StepType:  STEP_TYPE_LOOP,
				Condition: "for i in range(10)",
				Statements: []Step{
					{Key: suite.stepKey(5), StepType: STEP_TYPE_LEAF, LeafType: t_strPtr(LEAF_TYPE_SCENARIO), Description: "loop body", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, ScenarioKey: suite.scenarioRef},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(root)
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), data)

	// Unmarshal back
	var unmarshaled Step
	err = json.Unmarshal(data, &unmarshaled)
	assert.Nil(suite.T(), err)

	// Validate
	err = unmarshaled.Validate()
	assert.Nil(suite.T(), err)

	// Check structure
	assert.Equal(suite.T(), STEP_TYPE_SEQUENCE, unmarshaled.StepType)
	assert.Len(suite.T(), unmarshaled.Statements, 2)
	assert.Equal(suite.T(), STEP_TYPE_SWITCH, unmarshaled.Statements[0].StepType)
	assert.Equal(suite.T(), STEP_TYPE_LOOP, unmarshaled.Statements[1].StepType)
}

func (suite *ScenarioStepsSuite) TestJSONRoundTrip() {
	scenarioKeyStr := suite.scenarioKey.String()
	// JSON literal with all structures - using valid identity key formats
	jsonLiteral := fmt.Sprintf(`{
		"key": "%[1]s/sstep/0",
		"step_type": "sequence",
		"statements": [
			{
				"key": "%[1]s/sstep/1",
				"step_type": "leaf",
				"leaf_type": "event",
				"description": "first step",
				"from_object_key": "%[1]s/sobject/from1",
				"to_object_key": "%[1]s/sobject/to1",
				"event_key": "domain/test_domain/subdomain/default/class/test_class/event/ev1"
			},
			{
				"key": "%[1]s/sstep/2",
				"step_type": "loop",
				"condition": "while condition",
				"statements": [
					{
						"key": "%[1]s/sstep/3",
						"step_type": "leaf",
						"leaf_type": "scenario",
						"description": "loop step",
						"from_object_key": "%[1]s/sobject/from2",
						"to_object_key": "%[1]s/sobject/to2",
						"scenario_key": "domain/test_domain/subdomain/default/usecase/test_use_case/scenario/sk2"
					}
				]
			},
			{
				"key": "%[1]s/sstep/4",
				"step_type": "switch",
				"statements": [
					{
						"key": "%[1]s/sstep/5",
						"step_type": "case",
						"condition": "case1",
						"statements": [
							{
								"key": "%[1]s/sstep/6",
								"step_type": "leaf",
								"leaf_type": "query",
								"description": "case1 step",
								"from_object_key": "%[1]s/sobject/from3",
								"to_object_key": "%[1]s/sobject/to3",
								"query_key": "domain/test_domain/subdomain/default/class/test_class/query/qk4"
							}
						]
					},
					{
						"key": "%[1]s/sstep/7",
						"step_type": "case",
						"condition": "case2",
						"statements": [
							{
								"key": "%[1]s/sstep/8",
								"step_type": "leaf",
								"leaf_type": "delete",
								"from_object_key": "%[1]s/sobject/from4"
							}
						]
					}
				]
			}
		]
	}`, scenarioKeyStr)

	// Parse into structure
	var step Step
	err := step.FromJSON(jsonLiteral)
	assert.Nil(suite.T(), err)

	// Export back to JSON
	jsonStr, err := step.ToJSON()
	assert.Nil(suite.T(), err)

	// Compare values are the same
	assert.Equal(suite.T(), t_OrderedJson(jsonLiteral), t_OrderedJson(jsonStr))
}

func (suite *ScenarioStepsSuite) TestYAMLRoundTrip() {
	scenarioKeyStr := suite.scenarioKey.String()
	// YAML literal with all structures - using valid identity key formats
	yamlLiteral := fmt.Sprintf(`key: %[1]s/sstep/0
step_type: sequence
statements:
    - key: %[1]s/sstep/1
      step_type: leaf
      leaf_type: event
      description: first step
      from_object_key: %[1]s/sobject/from1
      to_object_key: %[1]s/sobject/to1
      event_key: domain/test_domain/subdomain/default/class/test_class/event/ev1
    - key: %[1]s/sstep/2
      step_type: loop
      condition: while condition
      statements:
        - key: %[1]s/sstep/3
          step_type: leaf
          leaf_type: scenario
          description: loop step
          from_object_key: %[1]s/sobject/from2
          to_object_key: %[1]s/sobject/to2
          scenario_key: domain/test_domain/subdomain/default/usecase/test_use_case/scenario/sk2
    - key: %[1]s/sstep/4
      step_type: switch
      statements:
        - key: %[1]s/sstep/5
          step_type: case
          condition: case1
          statements:
            - key: %[1]s/sstep/6
              step_type: leaf
              leaf_type: query
              description: case1 step
              from_object_key: %[1]s/sobject/from3
              to_object_key: %[1]s/sobject/to3
              query_key: domain/test_domain/subdomain/default/class/test_class/query/qk4
        - key: %[1]s/sstep/7
          step_type: case
          condition: case2
          statements:
            - key: %[1]s/sstep/8
              step_type: leaf
              leaf_type: delete
              from_object_key: %[1]s/sobject/from4
`, scenarioKeyStr)

	// Parse into structure
	var step Step
	err := step.FromYAML(yamlLiteral)
	assert.Nil(suite.T(), err)

	// Export back to YAML
	yamlStr, err := step.ToYAML()
	assert.Nil(suite.T(), err)

	// Parse the exported YAML back into a second struct for comparison
	// (This avoids string comparison issues with YAML key ordering)
	var step2 Step
	err = step2.FromYAML(yamlStr)
	assert.Nil(suite.T(), err)

	// Compare the two structures
	assert.Equal(suite.T(), step, step2)
}

func t_OrderedJson(value string) (sorted string) {

	var data interface{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		panic(err)
	}

	sorted, err := t_ToSortedJSON(data)
	if err != nil {
		panic(err)
	}

	return sorted
}

func t_ToSortedJSON(v interface{}) (string, error) {
	switch vv := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(vv))
		for k := range vv {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			valStr, err := t_ToSortedJSON(vv[k])
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("%q:%s", k, valStr))
		}
		return "{" + strings.Join(parts, ",") + "}", nil
	case []interface{}:
		var parts []string
		for _, item := range vv {
			itemStr, err := t_ToSortedJSON(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, itemStr)
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case string:
		return strconv.Quote(vv), nil
	case float64:
		// Handles both integers and floats as parsed by json.Unmarshal.
		if vv == float64(int64(vv)) {
			return strconv.FormatInt(int64(vv), 10), nil
		}
		return strconv.FormatFloat(vv, 'g', -1, 64), nil
	case bool:
		if vv {
			return "true", nil
		}
		return "false", nil
	case nil:
		return "null", nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}
