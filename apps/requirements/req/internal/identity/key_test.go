package identity

import (
	"fmt"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestKeySuite(t *testing.T) {
	suite.Run(t, new(KeySuite))
}

type KeySuite struct {
	suite.Suite
}

func (suite *KeySuite) TestPreen() {
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

func (suite *KeySuite) TestHasPrefix() {
	tests := []struct {
		parent    string
		childType string
		value     interface{}
		errstr    string
	}{
		// OK cases.
		{
			parent:    "01_order_fulfillment",
			childType: "association",
			value:     "01_order_fulfillment/association/1",
		},
		{
			parent:    "domain1",
			childType: "class",
			value:     "domain1/class/thing1",
		},
		{
			parent:    "parent",
			childType: "child",
			value:     "parent/child/extrastuff",
		},

		// Error cases: wrong prefix.
		{
			parent:    "01_order_fulfillment",
			childType: "association",
			value:     "wrong/association/1",
			errstr:    `must have prefix 01_order_fulfillment/association/`,
		},
		{
			parent:    "domain1",
			childType: "class",
			value:     "domain1/wrong/5",
			errstr:    `must have prefix domain1/class/`,
		},
		{
			parent:    "parent",
			childType: "child",
			value:     "parent/child/extra/stuff",
			errstr:    `must not contain '/' after prefix parent/child/`,
		},

		// Error cases: blank parent.
		{
			parent:    "",
			childType: "association",
			value:     "anything",
			errstr:    "parent cannot be blank",
		},

		// Error cases: blank childType.
		{
			parent:    "01_order_fulfillment",
			childType: "",
			value:     "anything",
			errstr:    "childType cannot be blank",
		},

		// Error cases: non-string value.
		{
			parent:    "01_order_fulfillment",
			childType: "association",
			value:     123, // non-string
			errstr:    "must be a string",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		err := validation.Validate(test.value, HasPrefix(test.parent, test.childType))
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
		}
	}
}
