package model_class

import (
	"fmt"
	"testing"

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
	tests := []struct {
		key        string
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
			key:        "Key",
			name:       "Name",
			details:    "Details",
			isComplete: true,
			isStatic:   false,
			umlComment: "UmlComment",
			obj: Generalization{
				Key:        "Key",
				Name:       "Name",
				IsComplete: true,
				IsStatic:   false,
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			isComplete: false,
			isStatic:   true,
			umlComment: "",
			obj: Generalization{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				IsComplete: false,
				IsStatic:   true,
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			isComplete: true,
			isStatic:   true,
			umlComment: "UmlComment",
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			isComplete: true,
			isStatic:   true,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewGeneralization(test.key, test.name, test.details, test.isComplete, test.isStatic, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
