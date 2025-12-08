package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGuardSuite(t *testing.T) {
	suite.Run(t, new(GuardSuite))
}

type GuardSuite struct {
	suite.Suite
}

func (suite *GuardSuite) TestNew() {
	tests := []struct {
		key     string
		name    string
		details string
		obj     Guard
		errstr  string
	}{
		// OK.
		{
			key:     "Key",
			name:    "Name",
			details: "Details",
			obj: Guard{
				Key:     "Key",
				Name:    "Name",
				Details: "Details",
			},
		},

		// Error states.
		{
			key:     "",
			name:    "Name",
			details: "Details",
			errstr:  `Key: cannot be blank`,
		},
		{
			key:     "Key",
			name:    "",
			details: "Details",
			errstr:  `Name: cannot be blank`,
		},
		{
			key:     "Key",
			name:    "Name",
			details: "",
			errstr:  `Details: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewGuard(test.key, test.name, test.details)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
