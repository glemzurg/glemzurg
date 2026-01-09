package model_class

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_data_type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAttributeSuite(t *testing.T) {
	suite.Run(t, new(AttributeSuite))
}

type AttributeSuite struct {
	suite.Suite
}

func (suite *AttributeSuite) TestNew() {
	tests := []struct {
		key              string
		name             string
		details          string
		dataTypeRules    string
		derivationPolicy string
		nullable         bool
		umlComment       string
		indexNums        []uint
		obj              Attribute
		errstr           string
	}{
		// OK.
		{
			key:              "Key",
			name:             "Name",
			details:          "Details",
			dataTypeRules:    "DataTypeRules",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			obj: Attribute{
				Key:              "Key",
				Name:             "Name",
				Details:          "Details",
				DataTypeRules:    "DataTypeRules",
				DerivationPolicy: "DerivationPolicy",
				Nullable:         true,
				UmlComment:       "UmlComment",
				IndexNums:        []uint{1, 2},
			},
		},
		{
			key:              "Key",
			name:             "Name",
			details:          "",
			dataTypeRules:    "",
			derivationPolicy: "",
			nullable:         false,
			umlComment:       "",
			indexNums:        nil,
			obj: Attribute{
				Key:              "Key",
				Name:             "Name",
				Details:          "",
				DataTypeRules:    "",
				DerivationPolicy: "",
				Nullable:         false,
				UmlComment:       "",
				IndexNums:        nil,
			},
		},
		{
			key:              "KeyParsed",
			name:             "NameParsed",
			details:          "Details",
			dataTypeRules:    "unconstrained",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			obj: Attribute{
				Key:              "KeyParsed",
				Name:             "NameParsed",
				Details:          "Details",
				DataTypeRules:    "unconstrained",
				DerivationPolicy: "DerivationPolicy",
				Nullable:         true,
				UmlComment:       "UmlComment",
				IndexNums:        []uint{1, 2},
				DataType: &model_data_type.DataType{
					Key:            "KeyParsed",
					CollectionType: "atomic",
					Atomic: &model_data_type.Atomic{
						ConstraintType: "unconstrained",
					},
				},
			},
		},

		// Error states.
		{
			key:              "",
			name:             "Name",
			details:          "Details",
			dataTypeRules:    "DataTypeRules",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			errstr:           `Key: cannot be blank`,
		},
		{
			key:              "Key",
			name:             "",
			details:          "Details",
			dataTypeRules:    "DataTypeRules",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			errstr:           `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewAttribute(test.key, test.name, test.details, test.dataTypeRules, test.derivationPolicy, test.nullable, test.umlComment, test.indexNums)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
