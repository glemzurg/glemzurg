package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
}

func (suite *LogicSuite) TestLogicMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputLogic
	}{
		{
			name: "basic logic",
			input: inputLogic{
				Description: "Order total must be positive",
			},
		},
		{
			name: "full logic",
			input: inputLogic{
				Description:   "Order total must be positive",
				Notation:      "tla_plus",
				Specification: "self.total > 0",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputLogic
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}

func (suite *LogicSuite) TestParameterMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputParameter
	}{
		{
			name: "parameter with name only",
			input: inputParameter{
				Name: "amount",
			},
		},
		{
			name: "parameter with data type rules",
			input: inputParameter{
				Name:          "amount",
				DataTypeRules: "positive integer",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputParameter
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
