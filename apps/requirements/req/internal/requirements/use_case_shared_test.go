package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSharedSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSharedSuite))
}

type UseCaseSharedSuite struct {
	suite.Suite
}

func (suite *UseCaseSharedSuite) TestNew() {
	tests := []struct {
		shareType  string
		umlComment string
		obj        UseCaseShared
		errstr     string
	}{
		// OK.
		{
			shareType:  "include",
			umlComment: "UmlComment",
			obj: UseCaseShared{
				ShareType:  "include",
				UmlComment: "UmlComment",
			},
		},
		{
			shareType:  "extend",
			umlComment: "",
			obj: UseCaseShared{
				ShareType:  "extend",
				UmlComment: "",
			},
		},

		// Error states.
		{
			shareType:  "",
			umlComment: "UmlComment",
			errstr:     `ShareType: cannot be blank.`,
		},
		{
			shareType:  "unknown",
			umlComment: "UmlComment",
			errstr:     `ShareType: must be a valid value.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewUseCaseShared(test.shareType, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
