package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGlobalFunctionSuite(t *testing.T) {
	suite.Run(t, new(GlobalFunctionSuite))
}

type GlobalFunctionSuite struct {
	suite.Suite
}

func (suite *GlobalFunctionSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name     string
		input    inputGlobalFunction
		expected string
	}{
		{
			name: "basic global function",
			input: inputGlobalFunction{
				Name: "_Max",
				Parameters: []string{"x", "y"},
				Logic: inputLogic{
					Description: "Returns the maximum of two values",
					Notation:    "tla_plus",
					Specification: "IF x > y THEN x ELSE y",
				},
			},
		},
		{
			name: "global function without parameters",
			input: inputGlobalFunction{
				Name: "_SetOfValues",
				Logic: inputLogic{
					Description: "A set of valid values",
					Notation:    "tla_plus",
				},
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			// Unmarshal back
			var result inputGlobalFunction
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
