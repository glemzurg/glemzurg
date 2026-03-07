package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		suite.Run(tc.name, func() {
			t := suite.T()
			data, err := json.Marshal(tc.input)
			require.NoError(t, err)

			var result inputUseCaseShared
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
