package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSuite))
}

type UseCaseSuite struct {
	suite.Suite
}

func (suite *UseCaseSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputUseCase
	}{
		{
			name: "basic use case",
			input: inputUseCase{
				Name:  "Place Order",
				Level: "sea",
			},
		},
		{
			name: "full use case",
			input: inputUseCase{
				Name:     "Place Order",
				Details:  "Customer places an order for products",
				Level:    "sea",
				ReadOnly: false,
				Actors: map[string]*inputUseCaseActor{
					"customer_class": {UmlComment: "initiates"},
				},
			},
		},
		{
			name: "sky level use case",
			input: inputUseCase{
				Name:  "Manage Orders",
				Level: "sky",
			},
		},
		{
			name: "mud level use case",
			input: inputUseCase{
				Name:  "Login Flow",
				Level: "mud",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputUseCase
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			// Scenarios are json:"-", won't round-trip via JSON
			assert.Equal(t, tc.input.Name, result.Name)
			assert.Equal(t, tc.input.Details, result.Details)
			assert.Equal(t, tc.input.Level, result.Level)
			assert.Equal(t, tc.input.ReadOnly, result.ReadOnly)
			assert.Equal(t, tc.input.Actors, result.Actors)
		})
	}
}

func (suite *UseCaseSuite) TestUseCaseActorMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputUseCaseActor
	}{
		{
			name:  "empty actor",
			input: inputUseCaseActor{},
		},
		{
			name:  "actor with comment",
			input: inputUseCaseActor{UmlComment: "primary actor"},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputUseCaseActor
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
