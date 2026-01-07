package model_scenario

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioObjectSuite(t *testing.T) {
	suite.Run(t, new(ScenarioObjectSuite))
}

type ScenarioObjectSuite struct {
	suite.Suite
}

func (suite *ScenarioObjectSuite) TestNew() {
	tests := []struct {
		key          string
		objectNumber uint
		name         string
		nameStyle    string
		classKey     string
		multi        bool
		umlComment   string
		obj          ScenarioObject
		errstr       string
	}{
		// OK.
		{
			key:          "Key",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     "ClassKey",
			multi:        true,
			umlComment:   "UmlComment",
			obj: ScenarioObject{
				Key:          "Key",
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "name",
				ClassKey:     "ClassKey",
				Multi:        true,
				UmlComment:   "UmlComment",
			},
		},
		{
			key:          "Key",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "id",
			classKey:     "ClassKey",
			multi:        true,
			umlComment:   "UmlComment",
			obj: ScenarioObject{
				Key:          "Key",
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "id",
				ClassKey:     "ClassKey",
				Multi:        true,
				UmlComment:   "UmlComment",
			},
		},
		{
			key:          "Key",
			objectNumber: 0,
			name:         "",
			nameStyle:    "unnamed",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "",
			obj: ScenarioObject{
				Key:          "Key",
				ObjectNumber: 0,
				Name:         "",
				NameStyle:    "unnamed",
				ClassKey:     "ClassKey",
				Multi:        false,
				UmlComment:   "",
			},
		},

		// Error states.
		{
			key:          "",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Key: cannot be blank`,
		},
		{
			key:          "Key",
			objectNumber: 1,
			name:         "",
			nameStyle:    "name",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name cannot be blank`,
		},
		{
			key:          "Key",
			objectNumber: 1,
			name:         "",
			nameStyle:    "id",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name cannot be blank`,
		},
		{
			key:          "Key",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "unnamed",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name must be blank for unnamed style`,
		},
		{
			key:          "",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     "",
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `ClassKey: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewScenarioObject(test.key, test.objectNumber, test.name, test.nameStyle, test.classKey, test.multi, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
