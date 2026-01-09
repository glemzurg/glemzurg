package identity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestKeySuite(t *testing.T) {
	suite.Run(t, new(KeySuite))
}

type KeySuite struct {
	suite.Suite
}

func (suite *KeySuite) TestNewKey() {
	tests := []struct {
		parentKey string
		childType string
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK cases.
		{
			parentKey: "domain1",
			childType: "class",
			subKey:    "thing1",
			expected:  Key{ParentKey: "domain1", ChildType: "class", SubKey: "thing1"},
		},
		{
			parentKey: "01_order_fulfillment",
			childType: "association",
			subKey:    "1",
			expected:  Key{ParentKey: "01_order_fulfillment", ChildType: "association", SubKey: "1"},
		},
		{
			parentKey: " PARENT ",
			childType: "child",
			subKey:    " KEY ",
			expected:  Key{ParentKey: "parent", ChildType: "child", SubKey: "key"},
		},
		{
			parentKey: "",
			childType: "",
			subKey:    "rootkey",
			expected:  Key{ParentKey: "", ChildType: "", SubKey: "rootkey"},
		},

		// Error cases: blank subKey.
		{
			parentKey: "domain1",
			childType: "class",
			subKey:    "",
			errstr:    "cannot be blank",
		},

		// Error cases: only one of parentKey or childType set.
		{
			parentKey: "domain1",
			childType: "",
			subKey:    "thing1",
			errstr:    "ParentKey and ChildType must both be set or both be blank",
		},
		{
			parentKey: "",
			childType: "class",
			subKey:    "thing1",
			errstr:    "ParentKey and ChildType must both be set or both be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewKey(test.parentKey, test.childType, test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}

func (suite *KeySuite) TestParseKey() {
	tests := []struct {
		input    string
		expected Key
		errstr   string
	}{
		// OK cases.
		{
			input:    "domain1/class/thing1",
			expected: Key{ParentKey: "domain1", ChildType: "class", SubKey: "thing1"},
		},
		{
			input:    "rootkey",
			expected: Key{ParentKey: "", ChildType: "", SubKey: "rootkey"},
		},
		{
			input:    "  DOMAIN1  /  CLASS  /  THING1  ", // with spaces
			expected: Key{ParentKey: "domain1", ChildType: "class", SubKey: "thing1"},
		},

		// Error cases: invalid format.
		{
			input:  "domain1/class",
			errstr: "invalid key format",
		},
		{
			input:  "domain1/class/thing1/extra",
			errstr: "invalid key format",
		},
		{
			input:  "",
			errstr: "invalid key format",
		},
		{
			input:  "domain1//thing1", // empty childType
			errstr: "ParentKey and ChildType must both be set or both be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := ParseKey(test.input)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}

func (suite *KeySuite) TestString() {
	tests := []struct {
		key      Key
		expected string
	}{
		{
			key:      Key{ParentKey: "domain1", ChildType: "class", SubKey: "thing1"},
			expected: "domain1/class/thing1",
		},
		{
			key:      Key{ParentKey: "", ChildType: "", SubKey: "rootkey"},
			expected: "rootkey",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		assert.Equal(suite.T(), test.expected, test.key.String(), testName)
	}
}

func (suite *KeySuite) TestValidate() {
	tests := []struct {
		key    Key
		errstr string
	}{
		// OK cases.
		{
			key: Key{ParentKey: "domain1", ChildType: "class", SubKey: "thing1"},
		},
		{
			key: Key{ParentKey: "", ChildType: "", SubKey: "rootkey"},
		},

		// Error cases: blank SubKey.
		{
			key:    Key{ParentKey: "domain1", ChildType: "class", SubKey: ""},
			errstr: "cannot be blank",
		},

		// Error cases: only ParentKey set.
		{
			key:    Key{ParentKey: "domain1", ChildType: "", SubKey: "thing1"},
			errstr: "ParentKey and ChildType must both be set or both be blank",
		},

		// Error cases: only ChildType set.
		{
			key:    Key{ParentKey: "", ChildType: "class", SubKey: "thing1"},
			errstr: "ParentKey and ChildType must both be set or both be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		err := test.key.Validate()
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
		}
	}
}
