package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSharedSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSharedSuite))
}

type UseCaseSharedSuite struct {
	suite.Suite
}

func (suite *UseCaseSharedSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputUseCaseShared
	}{
		{
			name: "include share",
			input: inputUseCaseShared{
				ShareType: "include",
			},
		},
		{
			name: "extend share with comment",
			input: inputUseCaseShared{
				ShareType:  "extend",
				UmlComment: "optional login flow",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.NoError(t, err)

			var result inputUseCaseShared
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
