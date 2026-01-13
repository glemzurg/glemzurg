package model_use_case

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSuite))
}

type UseCaseSuite struct {
	suite.Suite
}

func (suite *UseCaseSuite) SetupTest() {
}

func (suite *UseCaseSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		level      string
		readOnly   bool
		umlComment string
		obj        UseCase
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1")),
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			obj: UseCase{
				Key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1")),
				Name:       "Name",
				Details:    "Details",
				Level:      "sea",
				ReadOnly:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok sky level",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase2")),
			name:       "Name",
			details:    "",
			level:      "sky",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase2")),
				Name:       "Name",
				Details:    "",
				Level:      "sky",
				ReadOnly:   false,
				UmlComment: "",
			},
		},
		{
			testName:   "ok mud level",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase3")),
			name:       "Name",
			details:    "",
			level:      "mud",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase3")),
				Name:       "Name",
				Details:    "",
				Level:      "mud",
				ReadOnly:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			testName:   "error empty key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Key:`,
		},
		{
			testName:   "error blank name",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase4")),
			name:       "",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
		{
			testName:   "error blank level",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase5")),
			name:       "Name",
			details:    "Details",
			level:      "",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Level: cannot be blank`,
		},
		{
			testName:   "error invalid level",
			key:        helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase6")),
			name:       "Name",
			details:    "Details",
			level:      "unknown",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Level: must be a valid value`,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewUseCase(tt.key, tt.name, tt.details, tt.level, tt.readOnly, tt.umlComment)
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
