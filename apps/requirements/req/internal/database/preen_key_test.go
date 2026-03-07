package database

import (
	"testing"

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
			testName: "ok no spaces",
			key:      "key",
			preened:  "key",
		},
		{
			testName: "ok with spaces",
			key:      " key ",
			preened:  "key",
		},
		{
			testName: "ok uppercase",
			key:      "KEY",
			preened:  "key",
		},
		{
			testName: "ok spaces inside",
			key:      "K  E	\nY",
			preened:  "k-e-y",
		},

		// Error states.
		{
			testName: "error blank",
			key:      "   ",
			errstr:   `cannot be blank`,
		},
	}
	for _, tt := range tests {
		pass := suite.Run(tt.testName, func() {
			preened, err := preenKey(tt.key)
			if tt.errstr == "" {
				suite.Require().NoError(err)
				suite.Equal(tt.preened, preened)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
				suite.Empty(preened)
			}
		})
		if !pass {
			break
		}
	}
}
