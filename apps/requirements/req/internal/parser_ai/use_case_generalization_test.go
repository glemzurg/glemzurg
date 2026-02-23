package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(UseCaseGeneralizationSuite))
}

type UseCaseGeneralizationSuite struct {
	suite.Suite
}

func (suite *UseCaseGeneralizationSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputUseCaseGeneralization
	}{
		{
			name: "basic use case generalization",
			input: inputUseCaseGeneralization{
				Name:          "Order Operations",
				SuperclassKey: "manage_order",
				SubclassKeys:  []string{"create_order", "cancel_order"},
			},
		},
		{
			name: "full use case generalization",
			input: inputUseCaseGeneralization{
				Name:          "Order Operations",
				Details:       "All order-related operations",
				SuperclassKey: "manage_order",
				SubclassKeys:  []string{"create_order", "cancel_order", "update_order"},
				IsComplete:    true,
				IsStatic:      false,
				UMLComment:    "order hierarchy",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputUseCaseGeneralization
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
