package req_model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) TestNew() {
	tests := []struct {
		key     string
		name    string
		details string
		obj     Model
		errstr  string
	}{
		// OK.
		{
			key:     "model1",
			name:    "Name",
			details: "Details",
			obj: Model{
				Key:     "model1",
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			key:     "  MODEL1  ",
			name:    "Name",
			details: "",
			obj: Model{
				Key:     "model1",
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			key:     "",
			name:    "Name",
			details: "Details",
			errstr:  "Key: cannot be blank",
		},
		{
			key:     "model1",
			name:    "",
			details: "Details",
			errstr:  "Name: cannot be blank.",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		obj, err := NewModel(test.key, test.name, test.details)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
