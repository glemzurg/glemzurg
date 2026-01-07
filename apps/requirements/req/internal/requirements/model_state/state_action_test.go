package model_state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActionSuite(t *testing.T) {
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
}

func (suite *ActionSuite) TestNew() {
	tests := []struct {
		key        string
		name       string
		details    string
		requires   []string
		guarantees []string
		obj        Action
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			obj: Action{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				Requires:   []string{"Requires"},
				Guarantees: []string{"Guarantees"},
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			requires:   nil,
			guarantees: nil,
			obj: Action{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Requires:   nil,
				Guarantees: nil,
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewAction(test.key, test.name, test.details, test.requires, test.guarantees)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
