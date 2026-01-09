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
		keyType   string
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK cases.
		{
			parentKey: "domain1",
			keyType:   "class",
			subKey:    "thing1",
			expected:  Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},
		{
			parentKey: "domain1",
			keyType:   "association",
			subKey:    "1",
			expected:  Key{parentKey: "domain1", keyType: "association", subKey: "1"},
		},
		{
			parentKey: " PARENT ",
			keyType:   "class",
			subKey:    " KEY ",
			expected:  Key{parentKey: "parent", keyType: "class", subKey: "key"},
		},
		{
			parentKey: "",
			keyType:   "use_case",
			subKey:    "rootkey",
			expected:  Key{parentKey: "", keyType: "use_case", subKey: "rootkey"},
		},

		// Error cases: verify that validate is being called.
		{
			parentKey: "domain1",
			keyType:   "",
			subKey:    "thing1",
			errstr:    "keyType: cannot be blank.",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewKey(test.parentKey, test.keyType, test.subKey)
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
			input:    "domain/domain1",
			expected: Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			input:    "domain/domain1/subdomain/subdomain1",
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			input:    "domain/domain1/subdomain/subdomain1/class/thing1",
			expected: Key{parentKey: "domain/domain1/subdomain/subdomain1", keyType: "class", subKey: "thing1"},
		},
		{
			input:    " DOMAIN / DOMAIN1  /  SUBDOMAIN  /  SUBDOMAIN1  ", // with spaces
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},

		// Error cases: invalid format.
		{
			input:  "", // empty string
			errstr: "invalid key format",
		},
		{
			input:  "domain/domain1/subdomain/subdomain1//thing1", // empty keyType
			errstr: "keyType: cannot be blank.",
		},
		{
			input:  "domain/domain1/subdomain/subdomain1/unknown/thing1", // unknown keyType
			errstr: "keyType: must be a valid value.",
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
			key:      Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
			expected: "domain1/class/thing1",
		},
		{
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
			expected: "domain/domain1",
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
			key: Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			key: Key{parentKey: "", keyType: "use_case", subKey: "usecase1"},
		},
		{
			key: Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},

		// Error cases.
		{
			key:    Key{parentKey: "domain1", keyType: "class", subKey: ""},
			errstr: "cannot be blank",
		},
		{
			key:    Key{parentKey: "domain1", keyType: "", subKey: "thing1"},
			errstr: "cannot be blank",
		},
		{
			key:    Key{parentKey: "domain1", keyType: "unknown", subKey: "thing1"},
			errstr: "keyType: must be a valid value.",
		},

		// Error cases: parentKey issues.
		{
			key:    Key{parentKey: "notallowed", keyType: "domain", subKey: "domain1"},
			errstr: "parentKey: parentKey must be blank for domain, use_case keys.",
		},
		{
			key:    Key{parentKey: "notallowed", keyType: "use_case", subKey: "domain1"},
			errstr: "parentKey: parentKey must be blank for domain, use_case keys.",
		},
		{
			key:    Key{parentKey: "", keyType: "class", subKey: "thing1"},
			errstr: "parentKey: parentKey must be non-blank for non-domain, non-use_case keys.",
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
