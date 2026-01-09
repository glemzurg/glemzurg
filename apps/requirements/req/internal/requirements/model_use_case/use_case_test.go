package model_use_case

import (
	"fmt"
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

func (suite *UseCaseSuite) TestNew() {
	domainKey := helper.Must(identity.NewRootKey("domain1"))

	tests := []struct {
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
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase1")),
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			obj: UseCase{
				Key:        helper.Must(NewUseCaseKey(domainKey, "usecase1")),
				Name:       "Name",
				Details:    "Details",
				Level:      "sea",
				ReadOnly:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase2")),
			name:       "Name",
			details:    "",
			level:      "sky",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        helper.Must(NewUseCaseKey(domainKey, "usecase2")),
				Name:       "Name",
				Details:    "",
				Level:      "sky",
				ReadOnly:   false,
				UmlComment: "",
			},
		},
		{
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase3")),
			name:       "Name",
			details:    "",
			level:      "mud",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        helper.Must(NewUseCaseKey(domainKey, "usecase3")),
				Name:       "Name",
				Details:    "",
				Level:      "mud",
				ReadOnly:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Key: (subKey: cannot be blank.).`,
		},
		{
			key:        helper.Must(identity.NewKey(domainKey.String(), "unknown", "usecase1")),
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Key: invalid child type for use_case.`,
		},
		{
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase4")),
			name:       "",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank.`,
		},
		{
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase5")),
			name:       "Name",
			details:    "Details",
			level:      "",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Level: cannot be blank.`,
		},
		{
			key:        helper.Must(NewUseCaseKey(domainKey, "usecase6")),
			name:       "Name",
			details:    "Details",
			level:      "unknown",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Level: must be a valid value.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewUseCase(test.key, test.name, test.details, test.level, test.readOnly, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
func (suite *UseCaseSuite) TestNewUseCaseKey() {
	domainKey := helper.Must(identity.NewRootKey("domain1"))

	tests := []struct {
		domainKey identity.Key
		subKey    string
		expected  identity.Key
		errstr    string
	}{
		{
			domainKey: domainKey,
			subKey:    "usecase1",
			expected:  helper.Must(identity.NewKey(domainKey.String(), "use_case", "usecase1")),
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewUseCaseKey(test.domainKey, test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
