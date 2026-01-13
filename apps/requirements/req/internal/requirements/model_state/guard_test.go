package model_state

import (
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
		testName string
		key      string
		name     string
		details  string
		obj      Guard
		errstr   string
	}{
		// OK.
		{
			testName: "ok with all fields",
			key:      "Key",
			name:     "Name",
			details:  "Details",
			obj: Guard{
				Key:     "Key",
				Name:    "Name",
				Details: "Details",
			},
		},

		// Error states.
		{
			testName: "error with blank key",
			key:      "",
			name:     "Name",
			details:  "Details",
			errstr:   `Key: cannot be blank`,
		},
		{
			testName: "error with blank name",
			key:      "Key",
			name:     "",
			details:  "Details",
			errstr:   `Name: cannot be blank`,
		},
		{
			testName: "error with blank details",
			key:      "Key",
			name:     "Name",
			details:  "",
			errstr:   `Details: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewGuard(tt.key, tt.name, tt.details)
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
