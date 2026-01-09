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
		name    string
		key     string
		preened string
		errstr  string
	}{
		// OK.
		{
			name:    "ok no spaces",
			key:     "key",
			preened: "key",
		},
		{
			name:    "ok with spaces",
			key:     " key ",
			preened: "key",
		},
		{
			name:    "ok uppercase",
			key:     "KEY",
			preened: "key",
		},

		// Error states.
		{
			name:   "error blank",
			key:    "   ",
			errstr: `cannot be blank`,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.name, func(t *testing.T) {
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
