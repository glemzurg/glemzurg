package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioSuite(t *testing.T) {
	suite.Run(t, new(ScenarioSuite))
}

type ScenarioSuite struct {
	suite.Suite
}

func (suite *ScenarioSuite) TestObjectMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputObject
	}{
		{
			name: "basic named object",
			input: inputObject{
				ObjectNumber: 1,
				Name:         "myOrder",
				NameStyle:    "name",
				ClassKey:     "order",
			},
		},
		{
			name: "unnamed object",
			input: inputObject{
				ObjectNumber: 2,
				NameStyle:    "unnamed",
				ClassKey:     "payment",
			},
		},
		{
			name: "id-style multi object with comment",
			input: inputObject{
				ObjectNumber: 3,
				Name:         "123",
				NameStyle:    "id",
				ClassKey:     "line_item",
				Multi:        true,
				UmlComment:   "multiple items",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputObject
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}

func (suite *ScenarioSuite) TestStepMarshalUnmarshal() {
	eventType := "event"
	queryType := "query"
	scenarioType := "scenario"
	deleteType := "delete"

	tests := []struct {
		name  string
		input inputStep
	}{
		{
			name: "leaf event step",
			input: inputStep{
				StepType:      "leaf",
				LeafType:      &eventType,
				Description:   "Customer submits order",
				FromObjectKey: strPtr("customer_obj"),
				ToObjectKey:   strPtr("order_obj"),
				EventKey:      strPtr("submit"),
			},
		},
		{
			name: "leaf query step",
			input: inputStep{
				StepType:      "leaf",
				LeafType:      &queryType,
				Description:   "Get order total",
				FromObjectKey: strPtr("customer_obj"),
				ToObjectKey:   strPtr("order_obj"),
				QueryKey:      strPtr("get_total"),
			},
		},
		{
			name: "leaf scenario step",
			input: inputStep{
				StepType:      "leaf",
				LeafType:      &scenarioType,
				Description:   "Login sub-flow",
				FromObjectKey: strPtr("customer_obj"),
				ToObjectKey:   strPtr("auth_obj"),
				ScenarioKey:   strPtr("login_scenario"),
			},
		},
		{
			name: "leaf delete step",
			input: inputStep{
				StepType:      "leaf",
				LeafType:      &deleteType,
				Description:   "Delete order",
				FromObjectKey: strPtr("order_obj"),
			},
		},
		{
			name: "sequence step",
			input: inputStep{
				StepType: "sequence",
				Statements: []inputStep{
					{
						StepType:      "leaf",
						LeafType:      &eventType,
						Description:   "Step 1",
						FromObjectKey: strPtr("a"),
						ToObjectKey:   strPtr("b"),
						EventKey:      strPtr("e1"),
					},
					{
						StepType:      "leaf",
						LeafType:      &eventType,
						Description:   "Step 2",
						FromObjectKey: strPtr("b"),
						ToObjectKey:   strPtr("c"),
						EventKey:      strPtr("e2"),
					},
				},
			},
		},
		{
			name: "loop step",
			input: inputStep{
				StepType:  "loop",
				Condition: "for each item",
				Statements: []inputStep{
					{
						StepType:      "leaf",
						LeafType:      &eventType,
						Description:   "Process item",
						FromObjectKey: strPtr("a"),
						ToObjectKey:   strPtr("b"),
						EventKey:      strPtr("process"),
					},
				},
			},
		},
		{
			name: "switch with cases",
			input: inputStep{
				StepType: "switch",
				Statements: []inputStep{
					{
						StepType:  "case",
						Condition: "if paid",
						Statements: []inputStep{
							{
								StepType:      "leaf",
								LeafType:      &eventType,
								Description:   "Ship order",
								FromObjectKey: strPtr("a"),
								ToObjectKey:   strPtr("b"),
								EventKey:      strPtr("ship"),
							},
						},
					},
					{
						StepType:  "case",
						Condition: "if not paid",
						Statements: []inputStep{
							{
								StepType:      "leaf",
								LeafType:      &eventType,
								Description:   "Cancel order",
								FromObjectKey: strPtr("a"),
								ToObjectKey:   strPtr("b"),
								EventKey:      strPtr("cancel"),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputStep
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}

func (suite *ScenarioSuite) TestScenarioMarshalUnmarshal() {
	eventType := "event"

	tests := []struct {
		name  string
		input inputScenario
	}{
		{
			name: "basic scenario",
			input: inputScenario{
				Name: "Happy Path",
			},
		},
		{
			name: "full scenario",
			input: inputScenario{
				Name:    "Order Placement",
				Details: "Customer places an order successfully",
				Objects: map[string]*inputObject{
					"customer_obj": {
						ObjectNumber: 1,
						Name:         "customer",
						NameStyle:    "name",
						ClassKey:     "customer",
					},
					"order_obj": {
						ObjectNumber: 2,
						Name:         "order",
						NameStyle:    "name",
						ClassKey:     "order",
					},
				},
				Steps: &inputStep{
					StepType: "sequence",
					Statements: []inputStep{
						{
							StepType:      "leaf",
							LeafType:      &eventType,
							Description:   "Customer creates order",
							FromObjectKey: strPtr("customer_obj"),
							ToObjectKey:   strPtr("order_obj"),
							EventKey:      strPtr("create"),
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputScenario
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
