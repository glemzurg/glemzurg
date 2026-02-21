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
	// Step keys for building test nodes.
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

func (suite *ScenarioStepsSuite) TestInferredLeafType() {
	// Event leaf
	node := Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	assert.Equal(suite.T(), LEAF_TYPE_EVENT, node.InferredLeafType())

	// Scenario leaf
	node = Node{
		Key:           suite.stepKey(1),
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	assert.Equal(suite.T(), LEAF_TYPE_SCENARIO, node.InferredLeafType())

	// Query leaf
	node = Node{
		Key:           suite.stepKey(2),
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      suite.queryKey,
	}
	assert.Equal(suite.T(), LEAF_TYPE_QUERY, node.InferredLeafType())

	// Delete leaf
	node = Node{
		Key:           suite.stepKey(3),
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		IsDelete:      true,
	}
	assert.Equal(suite.T(), LEAF_TYPE_DELETE, node.InferredLeafType())

	// Not a leaf (panic)
	node = Node{
		Key:      suite.stepKey(4),
		NodeType: NODE_TYPE_SEQUENCE,
		Statements: []Node{
			{Key: suite.stepKey(5), NodeType: NODE_TYPE_LEAF, FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	assert.Panics(suite.T(), func() { node.InferredLeafType() })
}

func (suite *ScenarioStepsSuite) TestValidateSequence() {
	// Valid sequence
	node := Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SEQUENCE,
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: empty statements
	node = Node{
		Key:        suite.stepKey(0),
		NodeType:   NODE_TYPE_SEQUENCE,
		Statements: []Node{},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "sequence must have at least one statement")
}

func (suite *ScenarioStepsSuite) TestValidateSwitch() {
	// Valid switch with case children
	node := Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SWITCH,
		Statements: []Node{
			{
				Key:       suite.stepKey(1),
				NodeType:  NODE_TYPE_CASE,
				Condition: "cond1",
				Statements: []Node{
					{Key: suite.stepKey(2), NodeType: NODE_TYPE_LEAF, Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
				},
			},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no cases
	node = Node{
		Key:        suite.stepKey(0),
		NodeType:   NODE_TYPE_SWITCH,
		Statements: []Node{},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "switch must have at least one case")

	// Invalid: non-case child
	node = Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SWITCH,
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "switch children must all be case nodes")

	// Invalid: case without condition
	node = Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SWITCH,
		Statements: []Node{
			{
				Key:      suite.stepKey(1),
				NodeType: NODE_TYPE_CASE,
				Statements: []Node{
					{Key: suite.stepKey(2), NodeType: NODE_TYPE_LEAF, Description: "step", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
				},
			},
		},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "case must have a condition")
}

func (suite *ScenarioStepsSuite) TestValidateCase() {
	// Valid case
	node := Node{
		Key:       suite.stepKey(0),
		NodeType:  NODE_TYPE_CASE,
		Condition: "some condition",
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Valid case with no statements (empty case is allowed)
	node = Node{
		Key:       suite.stepKey(0),
		NodeType:  NODE_TYPE_CASE,
		Condition: "some condition",
	}
	err = node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no condition
	node = Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_CASE,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "case must have a condition")
}

func (suite *ScenarioStepsSuite) TestValidateLoop() {
	// Valid loop
	node := Node{
		Key:       suite.stepKey(0),
		NodeType:  NODE_TYPE_LOOP,
		Condition: "while true",
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no condition
	node = Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_LOOP,
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "loop must have a condition")

	// Invalid: no statements
	node = Node{
		Key:        suite.stepKey(0),
		NodeType:   NODE_TYPE_LOOP,
		Condition:  "while true",
		Statements: []Node{},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "loop must have at least one statement")
}

func (suite *ScenarioStepsSuite) TestValidateLeaf() {
	// Valid event leaf
	node := Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Valid scenario leaf
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = node.Validate()
	assert.Nil(suite.T(), err)

	// Valid query leaf
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		QueryKey:      suite.queryKey,
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid: no from object key
	node = Node{
		Key:         suite.stepKey(0),
		NodeType:    NODE_TYPE_LEAF,
		Description: "desc",
		ToObjectKey: suite.toObjKey,
		EventKey:    suite.eventKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a from_object_key")

	// Invalid: no to object key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a to_object_key")

	// Invalid: both event_key and scenario_key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or query_key")

	// Invalid: neither event_key nor scenario_key nor query_key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have one of event_key, scenario_key, or query_key")

	// Invalid: event_key and query_key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		QueryKey:      suite.queryKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or query_key")

	// Valid delete leaf
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid delete: has to_object_key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have a to_object_key")

	// Invalid delete: has event_key
	node = Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have event_key, scenario_key, or query_key")

	// Invalid delete: no from_object_key
	node = Node{
		Key:         suite.stepKey(0),
		NodeType:    NODE_TYPE_LEAF,
		Description: "desc",
		IsDelete:    true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf must have a from_object_key")
}

func (suite *ScenarioStepsSuite) TestValidateUnknownNodeType() {
	node := Node{
		Key:      suite.stepKey(0),
		NodeType: "bogus",
	}
	err := node.Validate()
	assert.ErrorContains(suite.T(), err, "unknown node type 'bogus'")
}

func (suite *ScenarioStepsSuite) TestValidateKeyErrors() {
	// Empty key
	node := Node{
		NodeType: NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := node.Validate()
	assert.ErrorContains(suite.T(), err, "KeyType")

	// Wrong key type
	node = Node{
		Key:           suite.domainKey,
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "invalid key type 'domain' for scenario step")
}

func (suite *ScenarioStepsSuite) TestValidateWithParent() {
	// Valid: correct parent
	node := Node{
		Key:           suite.stepKey(0),
		NodeType:      NODE_TYPE_LEAF,
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := node.ValidateWithParent(&suite.scenarioKey)
	assert.NoError(suite.T(), err)

	// Invalid: wrong parent
	otherScenarioKey, err := identity.NewScenarioKey(suite.useCaseKey, "other_scenario")
	require.NoError(suite.T(), err)
	err = node.ValidateWithParent(&otherScenarioKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent")

	// Validate calls Validate (propagates validation error)
	badNode := Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_LEAF,
		// Missing from_object_key
		ToObjectKey: suite.toObjKey,
		EventKey:    suite.eventKey,
	}
	err = badNode.ValidateWithParent(&suite.scenarioKey)
	assert.ErrorContains(suite.T(), err, "leaf must have a from_object_key")

	// ValidateWithParent recurses into children
	root := Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SEQUENCE,
		Statements: []Node{
			{Key: suite.stepKey(1), NodeType: NODE_TYPE_LEAF, FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err = root.ValidateWithParent(&suite.scenarioKey)
	assert.NoError(suite.T(), err)
}

func (suite *ScenarioStepsSuite) TestJSON() {
	// Complex structure: sequence > [switch > [case > [leaf], case > [leaf]], loop > [leaf]]
	root := Node{
		Key:      suite.stepKey(0),
		NodeType: NODE_TYPE_SEQUENCE,
		Statements: []Node{
			{
				Key:      suite.stepKey(1),
				NodeType: NODE_TYPE_SWITCH,
				Statements: []Node{
					{
						Key:       suite.stepKey(2),
						NodeType:  NODE_TYPE_CASE,
						Condition: "if x > 0",
						Statements: []Node{
							{Key: suite.stepKey(3), NodeType: NODE_TYPE_LEAF, Description: "positive", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
						},
					},
				},
			},
			{
				Key:       suite.stepKey(4),
				NodeType:  NODE_TYPE_LOOP,
				Condition: "for i in range(10)",
				Statements: []Node{
					{Key: suite.stepKey(5), NodeType: NODE_TYPE_LEAF, Description: "loop body", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, ScenarioKey: suite.scenarioRef},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(root)
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), data)

	// Unmarshal back
	var unmarshaled Node
	err = json.Unmarshal(data, &unmarshaled)
	assert.Nil(suite.T(), err)

	// Validate
	err = unmarshaled.Validate()
	assert.Nil(suite.T(), err)

	// Check structure
	assert.Equal(suite.T(), NODE_TYPE_SEQUENCE, unmarshaled.NodeType)
	assert.Len(suite.T(), unmarshaled.Statements, 2)
	assert.Equal(suite.T(), NODE_TYPE_SWITCH, unmarshaled.Statements[0].NodeType)
	assert.Equal(suite.T(), NODE_TYPE_LOOP, unmarshaled.Statements[1].NodeType)
}

func (suite *ScenarioStepsSuite) TestJSONRoundTrip() {
	scenarioKeyStr := suite.scenarioKey.String()
	// JSON literal with all structures - using valid identity key formats
	jsonLiteral := fmt.Sprintf(`{
		"key": "%[1]s/sstep/0",
		"node_type": "sequence",
		"statements": [
			{
				"key": "%[1]s/sstep/1",
				"node_type": "leaf",
				"description": "first step",
				"from_object_key": "%[1]s/sobject/from1",
				"to_object_key": "%[1]s/sobject/to1",
				"event_key": "domain/test_domain/subdomain/default/class/test_class/event/ev1"
			},
			{
				"key": "%[1]s/sstep/2",
				"node_type": "loop",
				"condition": "while condition",
				"statements": [
					{
						"key": "%[1]s/sstep/3",
						"node_type": "leaf",
						"description": "loop step",
						"from_object_key": "%[1]s/sobject/from2",
						"to_object_key": "%[1]s/sobject/to2",
						"scenario_key": "domain/test_domain/subdomain/default/usecase/test_use_case/scenario/sk2"
					}
				]
			},
			{
				"key": "%[1]s/sstep/4",
				"node_type": "switch",
				"statements": [
					{
						"key": "%[1]s/sstep/5",
						"node_type": "case",
						"condition": "case1",
						"statements": [
							{
								"key": "%[1]s/sstep/6",
								"node_type": "leaf",
								"description": "case1 step",
								"from_object_key": "%[1]s/sobject/from3",
								"to_object_key": "%[1]s/sobject/to3",
								"query_key": "domain/test_domain/subdomain/default/class/test_class/query/qk4"
							}
						]
					},
					{
						"key": "%[1]s/sstep/7",
						"node_type": "case",
						"condition": "case2",
						"statements": [
							{
								"key": "%[1]s/sstep/8",
								"node_type": "leaf",
								"from_object_key": "%[1]s/sobject/from4",
								"is_delete": true
							}
						]
					}
				]
			}
		]
	}`, scenarioKeyStr)

	// Parse into structure
	var node Node
	err := node.FromJSON(jsonLiteral)
	assert.Nil(suite.T(), err)

	// Export back to JSON
	jsonStr, err := node.ToJSON()
	assert.Nil(suite.T(), err)

	// Compare values are the same
	assert.Equal(suite.T(), t_OrderedJson(jsonLiteral), t_OrderedJson(jsonStr))
}

func (suite *ScenarioStepsSuite) TestYAMLRoundTrip() {
	scenarioKeyStr := suite.scenarioKey.String()
	// YAML literal with all structures - using valid identity key formats
	yamlLiteral := fmt.Sprintf(`key: %[1]s/sstep/0
node_type: sequence
statements:
    - key: %[1]s/sstep/1
      node_type: leaf
      description: first step
      from_object_key: %[1]s/sobject/from1
      to_object_key: %[1]s/sobject/to1
      event_key: domain/test_domain/subdomain/default/class/test_class/event/ev1
    - key: %[1]s/sstep/2
      node_type: loop
      condition: while condition
      statements:
        - key: %[1]s/sstep/3
          node_type: leaf
          description: loop step
          from_object_key: %[1]s/sobject/from2
          to_object_key: %[1]s/sobject/to2
          scenario_key: domain/test_domain/subdomain/default/usecase/test_use_case/scenario/sk2
    - key: %[1]s/sstep/4
      node_type: switch
      statements:
        - key: %[1]s/sstep/5
          node_type: case
          condition: case1
          statements:
            - key: %[1]s/sstep/6
              node_type: leaf
              description: case1 step
              from_object_key: %[1]s/sobject/from3
              to_object_key: %[1]s/sobject/to3
              query_key: domain/test_domain/subdomain/default/class/test_class/query/qk4
        - key: %[1]s/sstep/7
          node_type: case
          condition: case2
          statements:
            - key: %[1]s/sstep/8
              node_type: leaf
              from_object_key: %[1]s/sobject/from4
              is_delete: true
`, scenarioKeyStr)

	// Parse into structure
	var node Node
	err := node.FromYAML(yamlLiteral)
	assert.Nil(suite.T(), err)

	// Export back to YAML
	yamlStr, err := node.ToYAML()
	assert.Nil(suite.T(), err)

	// Parse the exported YAML back into a second struct for comparison
	// (This avoids string comparison issues with YAML key ordering)
	var node2 Node
	err = node2.FromYAML(yamlStr)
	assert.Nil(suite.T(), err)

	// Compare the two structures
	assert.Equal(suite.T(), node, node2)
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
