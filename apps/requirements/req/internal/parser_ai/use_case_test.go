package parser_ai

import (
	"encoding/json"
	"testing"

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
		suite.Run(tc.name, func() {
			data, err := json.Marshal(tc.input)
			suite.Require().NoError(err)

			var result inputUseCase
			err = json.Unmarshal(data, &result)
			suite.Require().NoError(err)

			// Scenarios are json:"-", won't round-trip via JSON
			suite.Equal(tc.input.Name, result.Name)
			suite.Equal(tc.input.Details, result.Details)
			suite.Equal(tc.input.Level, result.Level)
			suite.Equal(tc.input.ReadOnly, result.ReadOnly)
			suite.Equal(tc.input.Actors, result.Actors)
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
		suite.Run(tc.name, func() {
			data, err := json.Marshal(tc.input)
			suite.Require().NoError(err)

			var result inputUseCaseActor
			err = json.Unmarshal(data, &result)
			suite.Require().NoError(err)

			suite.Equal(tc.input, result)
		})
	}
}
