package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(GeneralizationSuite))
}

type GeneralizationSuite struct {
	suite.Suite
}

func (suite *GeneralizationSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		isComplete bool
		isStatic   bool
		umlComment string
		obj        Generalization
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
			key:        helper.Must(identity.NewClassGeneralizationKey(subdomainKey, "gen1")),
			name:       "Name",
			details:    "Details",
			isComplete: true,
			isStatic:   false,
			umlComment: "UmlComment",
			obj: Generalization{
				Key:        helper.Must(identity.NewClassGeneralizationKey(subdomainKey, "gen1")),
				Name:       "Name",
				IsComplete: true,
				IsStatic:   false,
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok with minimal fields",
			key:        helper.Must(identity.NewClassGeneralizationKey(subdomainKey, "gen2")),
			name:       "Name",
			details:    "",
			isComplete: false,
			isStatic:   true,
			umlComment: "",
			obj: Generalization{
				Key:        helper.Must(identity.NewClassGeneralizationKey(subdomainKey, "gen2")),
				Name:       "Name",
				Details:    "",
				IsComplete: false,
				IsStatic:   true,
				UmlComment: "",
			},
		},

		// Error states.
		{
			testName:   "error with blank key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			isComplete: true,
			isStatic:   true,
			umlComment: "UmlComment",
			errstr:     `Key: key must be of type 'generalization', not ''`,
		},
		{
			testName:   "error with blank name",
			key:        helper.Must(identity.NewClassGeneralizationKey(subdomainKey, "gen3")),
			name:       "",
			details:    "Details",
			isComplete: true,
			isStatic:   true,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewGeneralization(tt.key, tt.name, tt.details, tt.isComplete, tt.isStatic, tt.umlComment)
			if tt.errstr == "" {
				assert.Nil(t, err)
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
