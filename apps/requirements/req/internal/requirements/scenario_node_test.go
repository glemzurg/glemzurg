package requirements

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ScenarioStepsSuite struct {
	suite.Suite
}

func TestScenarioStepsSuite(t *testing.T) {
	suite.Run(t, new(ScenarioStepsSuite))
}

func (suite *ScenarioStepsSuite) TestInferredLeafType() {
	// Event leaf
	node := Node{
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		EventKey:      "ek",
	}
	assert.Equal(suite.T(), LEAF_TYPE_EVENT, node.InferredLeafType())

	// Scenario leaf
	node = Node{
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		ScenarioKey:   "sk",
	}
	assert.Equal(suite.T(), LEAF_TYPE_SCENARIO, node.InferredLeafType())

	// Attribute leaf
	node = Node{
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		AttributeKey:  "ak",
	}
	assert.Equal(suite.T(), LEAF_TYPE_ATTRIBUTE, node.InferredLeafType())

	// Delete leaf
	node = Node{
		FromObjectKey: "fk",
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
			{Description: "step1", FromObjectKey: "fk1", ToObjectKey: "tk1", EventKey: "ek1"},
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
					{Description: "step1", FromObjectKey: "fk1", ToObjectKey: "tk1", EventKey: "ek1"},
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
				Statements: []Node{{Description: "step", FromObjectKey: "fk", ToObjectKey: "tk", EventKey: "ek"}},
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
			{Description: "step1", FromObjectKey: "fk1", ToObjectKey: "tk1", EventKey: "ek1"},
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
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		EventKey:      "ek",
	}
	err := node.Validate()
	assert.Nil(suite.T(), err)

	// Valid scenario leaf
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		ScenarioKey:   "sk",
	}
	err = node.Validate()
	assert.Nil(suite.T(), err)

	// Invalid: no from object key
	node = Node{
		Description: "desc",
		ToObjectKey: "tk",
		EventKey:    "ek",
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a from_object_key")

	// Invalid: no to object key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		EventKey:      "ek",
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have a to_object_key")

	// Invalid: both event_key and scenario_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		EventKey:      "ek",
		ScenarioKey:   "sk",
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or attribute_key")

	// Invalid: neither event_key nor scenario_key nor attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf must have one of event_key, scenario_key, or attribute_key")

	// Valid: attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		AttributeKey:  "attrk",
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid: event_key and attribute_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		EventKey:      "ek",
		AttributeKey:  "attrk",
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "leaf cannot have more than one of event_key, scenario_key, or attribute_key")

	// Valid delete leaf
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		IsDelete:      true,
	}
	err = node.Validate()
	assert.NoError(suite.T(), err)

	// Invalid delete: has to_object_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		ToObjectKey:   "tk",
		IsDelete:      true,
	}
	err = node.Validate()
	assert.ErrorContains(suite.T(), err, "delete leaf cannot have a to_object_key")

	// Invalid delete: has event_key
	node = Node{
		Description:   "desc",
		FromObjectKey: "fk",
		EventKey:      "ek",
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
							{Description: "positive", FromObjectKey: "fk1", ToObjectKey: "tk1", EventKey: "ek1"},
						},
					},
				},
			},
			{
				Loop: "for i in range(10)",
				Statements: []Node{
					{Description: "loop body", FromObjectKey: "fk2", ToObjectKey: "tk2", ScenarioKey: "scenario1"},
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
	// JSON literal with all structures
	jsonLiteral := `{
		"statements": [
			{
				"description": "first step",
				"from_object_key": "fk1",
				"to_object_key": "tk1",
				"event_key": "ek1"
			},
			{
				"loop": "while condition",
				"statements": [
					{
						"description": "loop step",
						"from_object_key": "fk2",
						"to_object_key": "tk2",
						"scenario_key": "sk2"
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
								"from_object_key": "fk3",
								"to_object_key": "tk3",
								"attribute_key": "ak4"
							}
						]
					},
					{
						"condition": "case2",
						"statements": [ 
							{
								"from_object_key": "fk4",
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
	// YAML literal with all structures
	yamlLiteral := `statements:
    - description: first step
      from_object_key: fk1
      to_object_key: tk1
      event_key: ek1
    - statements:
        - description: loop step
          from_object_key: fk2
          to_object_key: tk2
          scenario_key: sk2
      loop: while condition
    - cases:
        - condition: case1
          statements:
            - description: case1 step
              from_object_key: fk3
              to_object_key: tk3
              attribute_key: ak4
        - condition: case2
          statements:
            - from_object_key: fk4
              is_delete: true
`

	// Parse into structure
	var node Node
	err := node.FromYAML(yamlLiteral)
	assert.Nil(suite.T(), err)

	// Export back to YAML
	yamlStr, err := node.ToYAML()
	assert.Nil(suite.T(), err)

	// Compare values are the same
	assert.Equal(suite.T(), yamlLiteral, yamlStr)
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
