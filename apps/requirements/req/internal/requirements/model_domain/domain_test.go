package model_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

func (suite *DomainSuite) TestNew() {
	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		realized   bool
		umlComment string
		obj        Domain
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with details",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "Details",
			realized:   true,
			umlComment: "UmlComment",
			obj: Domain{
				Key:        helper.Must(identity.NewDomainKey("domain1")),
				Name:       "Name",
				Details:    "Details",
				Realized:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok with blank values",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "",
			realized:   false,
			umlComment: "",
			obj: Domain{
				Key:        helper.Must(identity.NewDomainKey("domain1")),
				Name:       "Name",
				Details:    "",
				Realized:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			testName: "error empty key",
			key:      identity.Key{},
			name:     "Name",
			errstr:   "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			key:      helper.Must(identity.NewActorKey("actor1")),
			name:     "Name",
			errstr:   "Key: invalid key type 'actor' for domain.",
		},
		{
			testName: "error blank name",
			key:      helper.Must(identity.NewDomainKey("domain1")),
			name:     "",
			errstr:   `Name: cannot be blank.`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewDomain(tt.key, tt.name, tt.details, tt.realized, tt.umlComment)
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
