package identity

import (
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
		testName  string
		parentKey string
		keyType   string
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK cases.
		{
			testName:  "ok basic",
			parentKey: "domain1",
			keyType:   "class",
			subKey:    "thing1",
			expected:  Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},
		{
			testName:  "ok domain association",
			parentKey: "domain1",
			keyType:   "dassociation",
			subKey:    "1",
			expected:  Key{parentKey: "domain1", keyType: "dassociation", subKey: "1"},
		},
		{
			testName:  "ok with spaces",
			parentKey: " PARENT ",
			keyType:   "class",
			subKey:    " KEY ",
			expected:  Key{parentKey: "parent", keyType: "class", subKey: "key"},
		},
		{
			testName:  "ok root",
			parentKey: "",
			keyType:   "actor",
			subKey:    "rootkey",
			expected:  Key{parentKey: "", keyType: "actor", subKey: "rootkey"},
		},

		// Error cases: verify that validate is being called.
		{
			testName:  "validate being called",
			parentKey: "domain1",
			keyType:   "", // Trigger validation error.
			subKey:    "thing1",
			errstr:    "keyType: cannot be blank.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := newKey(tt.parentKey, tt.keyType, tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestParseKey() {
	tests := []struct {
		testName string
		input    string
		expected Key
		errstr   string
	}{
		// OK cases.
		{
			testName: "ok simple",
			input:    "domain/domain1",
			expected: Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			testName: "ok nested",
			input:    "domain/domain1/subdomain/subdomain1",
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			testName: "ok deep",
			input:    "domain/domain1/subdomain/subdomain1/class/thing1",
			expected: Key{parentKey: "domain/domain1/subdomain/subdomain1", keyType: "class", subKey: "thing1"},
		},
		{
			testName: "ok with spaces",
			input:    " DOMAIN / DOMAIN1  /  SUBDOMAIN  /  SUBDOMAIN1  ", // with spaces
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},

		// Error cases: invalid format.
		{
			testName: "error empty",
			input:    "", // empty string
			errstr:   "invalid key format",
		},
		{
			testName: "error empty keyType",
			input:    "domain/domain1/subdomain/subdomain1//thing1", // empty keyType
			errstr:   "keyType: cannot be blank.",
		},
		{
			testName: "error unknown keyType",
			input:    "domain/domain1/subdomain/subdomain1/unknown/thing1", // unknown keyType
			errstr:   "keyType: must be a valid value.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := ParseKey(tt.input)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestString() {
	tests := []struct {
		testName string
		key      Key
		expected string
	}{
		{
			testName: "with parent",
			key:      Key{parentKey: "domain/domain1", keyType: "class", subKey: "thing1"},
			expected: "domain/domain1/class/thing1",
		},
		{
			testName: "root",
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
			expected: "domain/domain1",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.key.String())
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestValidate() {
	tests := []struct {
		testName string
		key      Key
		errstr   string
	}{
		// OK cases.
		{
			testName: "ok domain",
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			testName: "ok actor",
			key:      Key{parentKey: "", keyType: "actor", subKey: "actor1"},
		},
		{
			testName: "ok class",
			key:      Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},

		// Error cases.
		{
			testName: "error blank subKey",
			key:      Key{parentKey: "domain1", keyType: "class", subKey: ""},
			errstr:   "cannot be blank",
		},
		{
			testName: "error blank keyType",
			key:      Key{parentKey: "domain1", keyType: "", subKey: "thing1"},
			errstr:   "cannot be blank",
		},
		{
			testName: "error invalid keyType",
			key:      Key{parentKey: "domain1", keyType: "unknown", subKey: "thing1"},
			errstr:   "keyType: must be a valid value.",
		},

		// Error cases: parentKey issues.
		{
			testName: "error parentKey for domain",
			key:      Key{parentKey: "notallowed", keyType: "domain", subKey: "domain1"},
			errstr:   "parentKey: parentKey must be blank for 'domain' keys.",
		},
		{
			testName: "error parentKey for actor",
			key:      Key{parentKey: "notallowed", keyType: "actor", subKey: "domain1"},
			errstr:   "parentKey: parentKey must be blank for 'actor' keys.",
		},
		{
			testName: "error blank parentKey for class",
			key:      Key{parentKey: "", keyType: "class", subKey: "thing1"},
			errstr:   "parentKey: parentKey must be non-blank for 'class' keys.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
		if !pass {
			break
		}
	}
}
