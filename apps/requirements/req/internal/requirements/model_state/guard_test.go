package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGuardSuite(t *testing.T) {
	suite.Run(t, new(GuardSuite))
}

type GuardSuite struct {
	suite.Suite
}

func (suite *GuardSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName string
		key      identity.Key
		name     string
		details  string
		obj      Guard
		errstr   string
	}{
		// OK.
		{
			testName: "ok with all fields",
			key:      helper.Must(identity.NewGuardKey(classKey, "guard1")),
			name:     "Name",
			details:  "Details",
			obj: Guard{
				Key:     helper.Must(identity.NewGuardKey(classKey, "guard1")),
				Name:    "Name",
				Details: "Details",
			},
		},

		// Error states.
		{
			testName: "error empty key",
			key:      identity.Key{},
			name:     "Name",
			details:  "Details",
			errstr:   "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			key:      helper.Must(identity.NewDomainKey("domain1")),
			name:     "Name",
			details:  "Details",
			errstr:   "Key: invalid key type 'domain' for guard",
		},
		{
			testName: "error with blank name",
			key:      helper.Must(identity.NewGuardKey(classKey, "guard2")),
			name:     "",
			details:  "Details",
			errstr:   `Name: cannot be blank`,
		},
		{
			testName: "error with blank details",
			key:      helper.Must(identity.NewGuardKey(classKey, "guard3")),
			name:     "Name",
			details:  "",
			errstr:   `Details: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewGuard(tt.key, tt.name, tt.details)
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
