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
		name      string
		parentKey string
		keyType   string
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK cases.
		{
			name:      "ok basic",
			parentKey: "domain1",
			keyType:   "class",
			subKey:    "thing1",
			expected:  Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},
		{
			name:      "ok association",
			parentKey: "domain1",
			keyType:   "association",
			subKey:    "1",
			expected:  Key{parentKey: "domain1", keyType: "association", subKey: "1"},
		},
		{
			name:      "ok with spaces",
			parentKey: " PARENT ",
			keyType:   "class",
			subKey:    " KEY ",
			expected:  Key{parentKey: "parent", keyType: "class", subKey: "key"},
		},
		{
			name:      "ok root",
			parentKey: "",
			keyType:   "actor",
			subKey:    "rootkey",
			expected:  Key{parentKey: "", keyType: "actor", subKey: "rootkey"},
		},

		// Error cases: verify that validate is being called.
		{
			name:      "validate being called",
			parentKey: "domain1",
			keyType:   "", // Trigger validation error.
			subKey:    "thing1",
			errstr:    "keyType: cannot be blank.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.name, func(t *testing.T) {
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
		name     string
		input    string
		expected Key
		errstr   string
	}{
		// OK cases.
		{
			name:     "ok simple",
			input:    "domain/domain1",
			expected: Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			name:     "ok nested",
			input:    "domain/domain1/subdomain/subdomain1",
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			name:     "ok deep",
			input:    "domain/domain1/subdomain/subdomain1/class/thing1",
			expected: Key{parentKey: "domain/domain1/subdomain/subdomain1", keyType: "class", subKey: "thing1"},
		},
		{
			name:     "ok with spaces",
			input:    " DOMAIN / DOMAIN1  /  SUBDOMAIN  /  SUBDOMAIN1  ", // with spaces
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},

		// Error cases: invalid format.
		{
			name:   "error empty",
			input:  "", // empty string
			errstr: "invalid key format",
		},
		{
			name:   "error empty keyType",
			input:  "domain/domain1/subdomain/subdomain1//thing1", // empty keyType
			errstr: "keyType: cannot be blank.",
		},
		{
			name:   "error unknown keyType",
			input:  "domain/domain1/subdomain/subdomain1/unknown/thing1", // unknown keyType
			errstr: "keyType: must be a valid value.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.name, func(t *testing.T) {
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
		name     string
		key      Key
		expected string
	}{
		{
			name:     "with parent",
			key:      Key{parentKey: "domain/domain1", keyType: "class", subKey: "thing1"},
			expected: "domain/domain1/class/thing1",
		},
		{
			name:     "root",
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
			expected: "domain/domain1",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.key.String())
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestValidate() {
	tests := []struct {
		name   string
		key    Key
		errstr string
	}{
		// OK cases.
		{
			name: "ok domain",
			key:  Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			name: "ok actor",
			key:  Key{parentKey: "", keyType: "actor", subKey: "actor1"},
		},
		{
			name: "ok class",
			key:  Key{parentKey: "domain1", keyType: "class", subKey: "thing1"},
		},

		// Error cases.
		{
			name:   "error blank subKey",
			key:    Key{parentKey: "domain1", keyType: "class", subKey: ""},
			errstr: "cannot be blank",
		},
		{
			name:   "error blank keyType",
			key:    Key{parentKey: "domain1", keyType: "", subKey: "thing1"},
			errstr: "cannot be blank",
		},
		{
			name:   "error invalid keyType",
			key:    Key{parentKey: "domain1", keyType: "unknown", subKey: "thing1"},
			errstr: "keyType: must be a valid value.",
		},

		// Error cases: parentKey issues.
		{
			name:   "error parentKey for domain",
			key:    Key{parentKey: "notallowed", keyType: "domain", subKey: "domain1"},
			errstr: "parentKey: parentKey must be blank for 'domain' keys.",
		},
		{
			name:   "error parentKey for actor",
			key:    Key{parentKey: "notallowed", keyType: "actor", subKey: "domain1"},
			errstr: "parentKey: parentKey must be blank for 'actor' keys.",
		},
		{
			name:   "error blank parentKey for class",
			key:    Key{parentKey: "", keyType: "class", subKey: "thing1"},
			errstr: "parentKey: parentKey must be non-blank for 'class' keys.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.name, func(t *testing.T) {
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
