package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDomainAssociationSuite(t *testing.T) {
	suite.Run(t, new(DomainAssociationSuite))
}

type DomainAssociationSuite struct {
	suite.Suite
}

func (suite *DomainAssociationSuite) TestMarshalUnmarshal() {
	tests := []struct {
		name  string
		input inputDomainAssociation
	}{
		{
			name: "basic domain association",
			input: inputDomainAssociation{
				ProblemDomainKey:  "billing",
				SolutionDomainKey: "payment_processing",
			},
		},
		{
			name: "domain association with comment",
			input: inputDomainAssociation{
				ProblemDomainKey:  "billing",
				SolutionDomainKey: "payment_processing",
				UmlComment:        "billing enforces payment requirements",
			},
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.Nil(t, err)

			var result inputDomainAssociation
			err = json.Unmarshal(data, &result)
			assert.Nil(t, err)

			assert.Equal(t, tc.input, result)
		})
	}
}
