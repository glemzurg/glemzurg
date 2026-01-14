package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}

type StateSuite struct {
	suite.Suite
}

func (suite *StateSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		umlComment string
		obj        State
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
			key:        helper.Must(identity.NewStateKey(classKey, "state1")),
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: State{
				Key:        helper.Must(identity.NewStateKey(classKey, "state1")),
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok with minimal fields",
			key:        helper.Must(identity.NewStateKey(classKey, "state2")),
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: State{
				Key:        helper.Must(identity.NewStateKey(classKey, "state2")),
				Name:       "Name",
				Details:    "",
				UmlComment: "",
			},
		},

		// Error states.
		{
			testName:   "error empty key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     "keyType: cannot be blank",
		},
		{
			testName:   "error wrong key type",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     "Key: invalid key type 'domain' for state",
		},
		{
			testName:   "error with blank name",
			key:        helper.Must(identity.NewStateKey(classKey, "state3")),
			name:       "",
			details:    "Details",
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewState(tt.key, tt.name, tt.details, tt.umlComment)
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
