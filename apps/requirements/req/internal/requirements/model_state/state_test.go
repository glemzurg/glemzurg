package model_state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}

type StateSuite struct {
	suite.Suite
}

func (suite *StateSuite) TestNew() {
	tests := []struct {
		key        string
		name       string
		details    string
		umlComment string
		obj        State
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: State{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: State{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewState(test.key, test.name, test.details, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
