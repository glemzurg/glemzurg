package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActionSuite(t *testing.T) {
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
}

func (suite *ActionSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		requires   []string
		guarantees []string
		obj        Action
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
			key:        helper.Must(identity.NewActionKey(classKey, "action1")),
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			obj: Action{
				Key:        helper.Must(identity.NewActionKey(classKey, "action1")),
				Name:       "Name",
				Details:    "Details",
				Requires:   []string{"Requires"},
				Guarantees: []string{"Guarantees"},
			},
		},
		{
			testName:   "ok with minimal fields",
			key:        helper.Must(identity.NewActionKey(classKey, "action2")),
			name:       "Name",
			details:    "",
			requires:   nil,
			guarantees: nil,
			obj: Action{
				Key:        helper.Must(identity.NewActionKey(classKey, "action2")),
				Name:       "Name",
				Details:    "",
				Requires:   nil,
				Guarantees: nil,
			},
		},

		// Error states.
		{
			testName:   "error empty key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     "keyType: cannot be blank",
		},
		{
			testName:   "error wrong key type",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     "Key: invalid key type 'domain' for action",
		},
		{
			testName:   "error with blank name",
			key:        helper.Must(identity.NewActionKey(classKey, "action3")),
			name:       "",
			details:    "Details",
			requires:   []string{"Requires"},
			guarantees: []string{"Guarantees"},
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewAction(tt.key, tt.name, tt.details, tt.requires, tt.guarantees)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
