package model_state

import (
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
		testName   string
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
			testName:   "ok with all fields",
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
			testName:   "ok with minimal fields",
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
			testName:   "error with blank key",
			key:        "",
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     `Key: cannot be blank`,
		},
		{
			testName:   "error with blank name",
			key:        "Key",
			name:       "",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewAction(tt.key, tt.name, tt.details, tt.requires, tt.guarantees)
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
