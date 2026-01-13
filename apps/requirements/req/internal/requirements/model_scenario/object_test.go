package model_scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestObjectSuite(t *testing.T) {
	suite.Run(t, new(ObjectSuite))
}

type ObjectSuite struct {
	suite.Suite
}

func (suite *ObjectSuite) TestNew() {
	tests := []struct {
		testName     string
		key          string
		objectNumber uint
		name         string
		nameStyle    string
		classKey     string
		multi        bool
		umlComment   string
		obj          Object
		errstr       string
	}{
		// OK.
		{
			testName:     "ok with name style",
			key:          "Key",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     "ClassKey",
			multi:        true,
			umlComment:   "UmlComment",
			obj: Object{
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
			testName:     "ok with id style",
			key:          "Key",
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "id",
			classKey:     "ClassKey",
			multi:        true,
			umlComment:   "UmlComment",
			obj: Object{
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
			testName:     "ok with unnamed style",
			key:          "Key",
			objectNumber: 0,
			name:         "",
			nameStyle:    "unnamed",
			classKey:     "ClassKey",
			multi:        false,
			umlComment:   "",
			obj: Object{
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
			testName:     "error with blank key",
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
			testName:     "error with blank name for name style",
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
			testName:     "error with blank name for id style",
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
			testName:     "error with name for unnamed style",
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
			testName:     "error with blank class key",
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
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewObject(tt.key, tt.objectNumber, tt.name, tt.nameStyle, tt.classKey, tt.multi, tt.umlComment)
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
