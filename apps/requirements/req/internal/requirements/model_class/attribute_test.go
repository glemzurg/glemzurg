package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	attrParsedKey := helper.Must(identity.NewAttributeKey(classKey, "attrparsed"))

	tests := []struct {
		testName         string
		key              identity.Key
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
			testName:         "ok with all fields",
			key:              helper.Must(identity.NewAttributeKey(classKey, "attr1")),
			name:             "Name",
			details:          "Details",
			dataTypeRules:    "DataTypeRules",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			obj: Attribute{
				Key:              helper.Must(identity.NewAttributeKey(classKey, "attr1")),
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
			testName:         "ok with minimal fields",
			key:              helper.Must(identity.NewAttributeKey(classKey, "attr2")),
			name:             "Name",
			details:          "",
			dataTypeRules:    "",
			derivationPolicy: "",
			nullable:         false,
			umlComment:       "",
			indexNums:        nil,
			obj: Attribute{
				Key:              helper.Must(identity.NewAttributeKey(classKey, "attr2")),
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
			testName:         "ok with parsed data type",
			key:              attrParsedKey,
			name:             "NameParsed",
			details:          "Details",
			dataTypeRules:    "unconstrained",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			obj: Attribute{
				Key:              attrParsedKey,
				Name:             "NameParsed",
				Details:          "Details",
				DataTypeRules:    "unconstrained",
				DerivationPolicy: "DerivationPolicy",
				Nullable:         true,
				UmlComment:       "UmlComment",
				IndexNums:        []uint{1, 2},
				DataType: &model_data_type.DataType{
					Key:            attrParsedKey.String(),
					CollectionType: "atomic",
					Atomic: &model_data_type.Atomic{
						ConstraintType: "unconstrained",
					},
				},
			},
		},

		// Error states.
		{
			testName:         "error with blank key",
			key:              identity.Key{},
			name:             "Name",
			details:          "Details",
			dataTypeRules:    "DataTypeRules",
			derivationPolicy: "DerivationPolicy",
			nullable:         true,
			umlComment:       "UmlComment",
			indexNums:        []uint{1, 2},
			errstr:           `Key: key must be of type 'attribute', not ''`,
		},
		{
			testName:         "error with blank name",
			key:              helper.Must(identity.NewAttributeKey(classKey, "attr3")),
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
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewAttribute(tt.key, tt.name, tt.details, tt.dataTypeRules, tt.derivationPolicy, tt.nullable, tt.umlComment, tt.indexNums)
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
