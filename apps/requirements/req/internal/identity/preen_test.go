package identity

import (
	"fmt"
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
		key     string
		preened string
		errstr  string
	}{
		// OK.
		{
			key:     "key",
			preened: "key",
		},
		{
			key:     " key ",
			preened: "key",
		},
		{
			key:     "KEY",
			preened: "key",
		},

		// Error states.
		{
			key:    "   ",
			errstr: `cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		preened, err := PreenKey(test.key)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.preened, preened, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), preened, testName)
		}
	}
}
