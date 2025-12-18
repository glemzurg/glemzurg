package requirements

import (
	"fmt"
	"testing"

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
	tests := []struct {
		key        string
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
			key:        "Key",
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			obj: UseCase{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				Level:      "sea",
				ReadOnly:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			level:      "sky",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Level:      "sky",
				ReadOnly:   false,
				UmlComment: "",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			level:      "mud",
			readOnly:   false,
			umlComment: "",
			obj: UseCase{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Level:      "mud",
				ReadOnly:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			level:      "sea",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			level:      "",
			readOnly:   true,
			umlComment: "UmlComment",
			errstr:     `Level: cannot be blank.`,
		},
		{
			key:        "Key",
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
