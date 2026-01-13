package model_state

import (
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
		testName   string
		key        string
		name       string
		details    string
		umlComment string
		obj        State
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
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
			testName:   "ok with minimal fields",
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
			testName:   "error with blank key",
			key:        "",
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     `Key: cannot be blank`,
		},
		{
			testName:   "error with blank name",
			key:        "Key",
			name:       "",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewState(tt.key, tt.name, tt.details, tt.umlComment)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
