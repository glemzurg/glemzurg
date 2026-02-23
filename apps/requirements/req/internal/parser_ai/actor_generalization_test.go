package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActorGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(ActorGeneralizationSuite))
}

type ActorGeneralizationSuite struct {
	suite.Suite
}

func (suite *ActorGeneralizationSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputActorGeneralization
	}{
		{
			name: "basic actor generalization",
			input: inputActorGeneralization{
				Name:          "User Types",
				SuperclassKey: "user",
				SubclassKeys:  []string{"admin", "customer"},
			},
		},
		{
			name: "full actor generalization",
			input: inputActorGeneralization{
				Name:          "User Types",
				Details:       "Categorizes system users",
				SuperclassKey: "user",
				SubclassKeys:  []string{"admin", "customer", "guest"},
				IsComplete:    true,
				IsStatic:      true,
				UMLComment:    "note on generalization",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputActorGeneralization
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
