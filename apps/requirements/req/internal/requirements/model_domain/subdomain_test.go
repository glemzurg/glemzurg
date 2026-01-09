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
}

func (suite *SubdomainSuite) TestNew() {

	domainKey := helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1"))

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
			key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: Subdomain{
				Key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok with blank values",
			key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: Subdomain{
				Key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
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
			testName: "error blank name",
			key:      helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:     "",
			details:  "Details",
			errstr:   "Name: cannot be blank.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewSubdomain(tt.key, tt.name, tt.details, tt.umlComment)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
		if !pass {
			break
		}
	}
}
