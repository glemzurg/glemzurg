package model_class

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAssociationSuite(t *testing.T) {
	suite.Run(t, new(AssociationSuite))
}

type AssociationSuite struct {
	suite.Suite
}

func (suite *AssociationSuite) TestNew() {

	multiplicity, err := NewMultiplicity("2..3")
	assert.Nil(suite.T(), err)

	tests := []struct {
		key                 string
		name                string
		details             string
		fromClassKey        string
		fromMultiplicity    Multiplicity
		toClassKey          string
		toMultiplicity      Multiplicity
		associationClassKey string
		umlComment          string
		obj                 Association
		errstr              string
	}{
		// OK.
		{
			key:                 "Key",
			name:                "Name",
			details:             "Details",
			fromClassKey:        "FromClassKey",
			fromMultiplicity:    multiplicity,
			toClassKey:          "ToClassKey",
			toMultiplicity:      multiplicity,
			associationClassKey: "AssociationClassKey",
			umlComment:          "UmlComment",
			obj: Association{
				Key:                 "Key",
				Name:                "Name",
				Details:             "Details",
				FromClassKey:        "FromClassKey",
				FromMultiplicity:    multiplicity,
				ToClassKey:          "ToClassKey",
				ToMultiplicity:      multiplicity,
				AssociationClassKey: "AssociationClassKey",
				UmlComment:          "UmlComment",
			},
		},
		{
			key:                 "Key",
			name:                "Name",
			details:             "",
			fromClassKey:        "FromClassKey",
			fromMultiplicity:    multiplicity,
			toClassKey:          "ToClassKey",
			toMultiplicity:      multiplicity,
			associationClassKey: "",
			umlComment:          "",
			obj: Association{
				Key:                 "Key",
				Name:                "Name",
				Details:             "",
				FromClassKey:        "FromClassKey",
				FromMultiplicity:    multiplicity,
				ToClassKey:          "ToClassKey",
				ToMultiplicity:      multiplicity,
				AssociationClassKey: "",
				UmlComment:          "",
			},
		},

		// Error states.
		{
			key:                 "",
			name:                "Name",
			details:             "Details",
			fromClassKey:        "FromClassKey",
			fromMultiplicity:    multiplicity,
			toClassKey:          "ToClassKey",
			toMultiplicity:      multiplicity,
			associationClassKey: "AssociationClassKey",
			umlComment:          "UmlComment",
			errstr:              `Key: cannot be blank`,
		},
		{
			key:                 "Key",
			name:                "",
			details:             "Details",
			fromClassKey:        "FromClassKey",
			fromMultiplicity:    multiplicity,
			toClassKey:          "ToClassKey",
			toMultiplicity:      multiplicity,
			associationClassKey: "AssociationClassKey",
			umlComment:          "UmlComment",
			errstr:              `Name: cannot be blank`,
		},
		{
			key:                 "Key",
			name:                "Name",
			details:             "Details",
			fromClassKey:        "",
			fromMultiplicity:    multiplicity,
			toClassKey:          "ToClassKey",
			toMultiplicity:      multiplicity,
			associationClassKey: "AssociationClassKey",
			umlComment:          "UmlComment",
			errstr:              `FromClassKey: cannot be blank`,
		},
		{
			key:                 "Key",
			name:                "Name",
			details:             "Details",
			fromClassKey:        "FromClassKey",
			fromMultiplicity:    multiplicity,
			toClassKey:          "",
			toMultiplicity:      multiplicity,
			associationClassKey: "AssociationClassKey",
			umlComment:          "UmlComment",
			errstr:              `ToClassKey: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewAssociation(test.key, test.name, test.details, test.fromClassKey, test.fromMultiplicity, test.toClassKey, test.toMultiplicity, test.associationClassKey, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

func (suite *AssociationSuite) TestOther() {

	multiplicity, err := NewMultiplicity("2..3")
	assert.Nil(suite.T(), err)

	tests := []struct {
		obj      Association
		classKey string
		otherKey string
		errstr   string
	}{
		// OK.
		{
			obj: Association{
				Key:              "Key",
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     "FromClassKey",
				FromMultiplicity: multiplicity,
				ToClassKey:       "ToClassKey",
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: "FromClassKey",
			otherKey: "ToClassKey",
		},
		{
			obj: Association{
				Key:              "Key",
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     "FromClassKey",
				FromMultiplicity: multiplicity,
				ToClassKey:       "ToClassKey",
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: "ToClassKey",
			otherKey: "FromClassKey",
		},

		// Error states.
		{
			obj: Association{
				Key:              "Key",
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     "FromClassKey",
				FromMultiplicity: multiplicity,
				ToClassKey:       "ToClassKey",
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: "UnknownClassKey",
			errstr:   `association does not include class: 'UnknownClassKey'`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		otherKey, err := test.obj.Other(test.classKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.otherKey, otherKey, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), otherKey, testName)
		}
	}
}
