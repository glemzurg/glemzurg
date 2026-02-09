package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestPreenSuite(t *testing.T) {
	suite.Run(t, new(PreenSuite))
}

type PreenSuite struct {
	suite.Suite
}

func (suite *PreenSuite) TestPreen() {
	tests := []struct {
		testName string
		key      string
		preened  string
		errstr   string
	}{
		// OK.
		{
			testName:    "ok no spaces",
			key:     "key",
			preened: "key",
		},
		{
			testName:    "ok with spaces",
			key:     " key ",
			preened: "key",
		},
		{
			testName:    "ok uppercase",
			key:     "KEY",
			preened: "key",
		},

		// Error states.
		{
			testName:   "error blank",
			key:    "   ",
			errstr: `cannot be blank`,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			preened, err := PreenKey(tt.key)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.preened, preened)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, preened)
			}
		})
		if !pass {
			break
		}
	}
}
