package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
	domainKey identity.Key
}

func (suite *SubdomainSuite) SetupTest() {
	suite.domainKey = helper.Must(identity.NewDomainKey("domain1"))
}

func (suite *SubdomainSuite) TestNew() {

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		umlComment string
		obj        Subdomain
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with details",
			key:        helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1")),
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: Subdomain{
				Key:        helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1")),
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok minimal",
			key:        helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1")),
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: Subdomain{
				Key:        helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1")),
				Name:       "Name",
				Details:    "",
				UmlComment: "",
			},
		},

		// Errors.
		{
			testName: "error empty key",
			key:      identity.Key{},
			name:     "Name",
			details:  "Details",
			errstr:   "keyType: cannot be blank",
		},
		{
			testName: "error empty key",
			key:      helper.Must(identity.NewActorKey("actor1")),
			name:     "Name",
			details:  "Details",
			errstr:   "Key: invalid key type 'actor' for subdomain.",
		},
		{
			testName: "error blank name",
			key:      helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1")),
			name:     "",
			details:  "Details",
			errstr:   "Name: cannot be blank.",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewSubdomain(tt.key, tt.name, tt.details, tt.umlComment)
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
