package model_scenario

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
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
	attrKey      *identity.Key
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
	attrKey, err := identity.NewAttributeKey(suite.classKey, "test_attr")
	require.NoError(suite.T(), err)
	suite.attrKey = &attrKey
}

func (suite *ScenarioStepsSuite) TestInferredLeafType() {
	// Event leaf
	node := Node{
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	assert.Equal(suite.T(), LEAF_TYPE_EVENT, node.InferredLeafType())

	// Scenario leaf
	node = Node{
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	assert.Equal(suite.T(), LEAF_TYPE_SCENARIO, node.InferredLeafType())

	// Attribute leaf
	node = Node{
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		AttributeKey:  suite.attrKey,
	}
	assert.Equal(suite.T(), LEAF_TYPE_ATTRIBUTE, node.InferredLeafType())

	// Delete leaf
	node = Node{
		FromObjectKey: suite.fromObjKey,
		IsDelete:      true,
	}
	assert.Equal(suite.T(), LEAF_TYPE_DELETE, node.InferredLeafType())

	// Not a leaf (panic)
	node = Node{
		Statements: []Node{{}},
	}
	assert.Panics(suite.T(), func() { node.InferredLeafType() })
}

func (suite *ScenarioStepsSuite) TestValidateSequence() {
	// Valid sequence
	node := Node{
		Statements: []Node{
			{Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)
}

func (suite *ScenarioStepsSuite) TestValidateSwitch() {
	// Valid switch
	node := Node{
		Cases: []Case{
			{
				Condition: "cond1",
				Statements: []Node{
					{Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
				},
			},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: case without condition
	node = Node{
		Cases: []Case{
			{
				Statements: []Node{{Description: "step", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey}},
			},
		},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "switch case must have a conditional description")
}

func (suite *ScenarioStepsSuite) TestValidateLoop() {
	// Valid loop
	node := Node{
		Loop: "while true",
		Statements: []Node{
			{Description: "step1", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
		},
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no statements
	node = Node{
		Loop:       "while true",
		Statements: []Node{},
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "loop must have at least one statement")
}

func (suite *ScenarioStepsSuite) TestValidateLeaf() {
	// Valid event leaf
	node := Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Valid scenario leaf
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no from object key
	node = Node{
		Description: "desc",
		ToObjectKey: suite.toObjKey,
		EventKey:    suite.eventKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a from_object_key")

	// Invalid: no to object key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a to_object_key")

	// Invalid: both event_key and scenario_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		ScenarioKey:   suite.scenarioRef,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or attribute_key")

	// Invalid: neither event_key nor scenario_key nor attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have one of event_key, scenario_key, or attribute_key")

	// Valid: attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		AttributeKey:  suite.attrKey,
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid: event_key and attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		EventKey:      suite.eventKey,
		AttributeKey:  suite.attrKey,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or attribute_key")

	// Valid delete leaf
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid delete: has to_object_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		ToObjectKey:   suite.toObjKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have a to_object_key")

	// Invalid delete: has event_key
	node = Node{
		Description:   "desc",
		FromObjectKey: suite.fromObjKey,
		EventKey:      suite.eventKey,
		IsDelete:      true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have event_key, scenario_key, or attribute_key")

	// Invalid delete: no from_object_key
	node = Node{
		Description: "desc",
		IsDelete:    true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf must have a from_object_key")
}

func (suite *ScenarioStepsSuite) TestJSON() {
	// Complex structure
	root := Node{
		Statements: []Node{
			{
				Cases: []Case{
					{
						Condition: "if x > 0",
						Statements: []Node{
							{Description: "positive", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, EventKey: suite.eventKey},
						},
					},
				},
			},
			{
				Loop: "for i in range(10)",
				Statements: []Node{
					{Description: "loop body", FromObjectKey: suite.fromObjKey, ToObjectKey: suite.toObjKey, ScenarioKey: suite.scenarioRef},
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
	assert.Equal(suite.T(), "sequence", unmarshaled.Inferredtype())
	assert.Len(suite.T(), unmarshaled.Statements, 2)
	assert.Equal(suite.T(), "switch", unmarshaled.Statements[0].Inferredtype())
	assert.Equal(suite.T(), "loop", unmarshaled.Statements[1].Inferredtype())
}

func (suite *ScenarioStepsSuite) TestJSONRoundTrip() {
	// JSON literal with all structures - using valid identity key formats
	jsonLiteral := `{
		"statements": [
			{
				"description": "first step",
				"from_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from1",
				"to_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to1",
				"event_key": "domain/test_domain/subdomain/default/class/test_class/event/ev1"
			},
			{
				"loop": "while condition",
				"statements": [
					{
						"description": "loop step",
						"from_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from2",
						"to_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to2",
						"scenario_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/sk2"
					}
				]
			},
			{
				"cases": [
					{
						"condition": "case1",
						"statements": [
							{
								"description": "case1 step",
								"from_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from3",
								"to_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to3",
								"attribute_key": "domain/test_domain/subdomain/default/class/test_class/attribute/ak4"
							}
						]
					},
					{
						"condition": "case2",
						"statements": [
							{
								"from_object_key": "domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from4",
								"is_delete": true
							}
						]
					}
				]
			}
		]
	}`

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
	// YAML literal with all structures - using valid identity key formats
	yamlLiteral := `statements:
    - description: first step
      from_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from1
      to_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to1
      event_key: domain/test_domain/subdomain/default/class/test_class/event/ev1
    - statements:
        - description: loop step
          from_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from2
          to_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to2
          scenario_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/sk2
      loop: while condition
    - cases:
        - condition: case1
          statements:
            - description: case1 step
              from_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from3
              to_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/to3
              attribute_key: domain/test_domain/subdomain/default/class/test_class/attribute/ak4
        - condition: case2
          statements:
            - from_object_key: domain/test_domain/subdomain/default/usecase/test_uc/scenario/test_sc/sobject/from4
              is_delete: true
`

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
