package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

func (suite *ClassSuite) TestNew() {
	tests := []struct {
		key             string
		name            string
		details         string
		actorKey        string
		superclassOfKey string
		subclassOfKey   string
		umlComment      string
		obj             Class
		errstr          string
	}{
		// OK.
		{
			key:             "Key",
			name:            "Name",
			details:         "Details",
			actorKey:        "ActorKey",
			superclassOfKey: "SuperclassOfKey",
			subclassOfKey:   "SubclassOfKey",
			umlComment:      "UmlComment",
			obj: Class{
				Key:             "Key",
				Name:            "Name",
				Details:         "Details",
				ActorKey:        "ActorKey",
				SuperclassOfKey: "SuperclassOfKey",
				SubclassOfKey:   "SubclassOfKey",
				UmlComment:      "UmlComment",
			},
		},
		{
			key:             "Key",
			name:            "Name",
			details:         "",
			actorKey:        "",
			superclassOfKey: "",
			subclassOfKey:   "",
			umlComment:      "",
			obj: Class{
				Key:             "Key",
				Name:            "Name",
				Details:         "",
				ActorKey:        "",
				SuperclassOfKey: "",
				SubclassOfKey:   "",
				UmlComment:      "",
			},
		},

		// Error states.
		{
			key:             "",
			name:            "Name",
			details:         "Details",
			actorKey:        "ActorKey",
			superclassOfKey: "SuperclassOfKey",
			subclassOfKey:   "SubclassOfKey",
			umlComment:      "UmlComment",
			errstr:          `Key: cannot be blank`,
		},
		{
			key:             "Key",
			name:            "",
			details:         "Details",
			actorKey:        "ActorKey",
			superclassOfKey: "SuperclassOfKey",
			subclassOfKey:   "SubclassOfKey",
			umlComment:      "UmlComment",
			errstr:          `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewClass(test.key, test.name, test.details, test.actorKey, test.superclassOfKey, test.subclassOfKey, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
