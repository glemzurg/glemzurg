package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	multiplicity, err := NewMultiplicity("2..3")
	assert.Nil(suite.T(), err)

	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	assocClassKey := helper.Must(identity.NewClassKey(subdomainKey, "assocclass"))

	tests := []struct {
		testName            string
		key                 identity.Key
		name                string
		details             string
		fromClassKey        identity.Key
		fromMultiplicity    Multiplicity
		toClassKey          identity.Key
		toMultiplicity      Multiplicity
		associationClassKey identity.Key
		umlComment          string
		obj                 Association
		errstr              string
	}{
		// OK.
		{
			testName:            "ok with all fields",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc1")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			obj: Association{
				Key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc1")),
				Name:                "Name",
				Details:             "Details",
				FromClassKey:        fromClassKey,
				FromMultiplicity:    multiplicity,
				ToClassKey:          toClassKey,
				ToMultiplicity:      multiplicity,
				AssociationClassKey: assocClassKey,
				UmlComment:          "UmlComment",
			},
		},
		{
			testName:            "ok with minimal fields",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc2")),
			name:                "Name",
			details:             "",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: identity.Key{},
			umlComment:          "",
			obj: Association{
				Key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc2")),
				Name:                "Name",
				Details:             "",
				FromClassKey:        fromClassKey,
				FromMultiplicity:    multiplicity,
				ToClassKey:          toClassKey,
				ToMultiplicity:      multiplicity,
				AssociationClassKey: identity.Key{},
				UmlComment:          "",
			},
		},

		// Error states.
		{
			testName:            "error empty key",
			key:                 identity.Key{},
			name:                "Name",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "keyType: cannot be blank",
		},
		{
			testName:            "error wrong key type",
			key:                 helper.Must(identity.NewDomainKey("domain1")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "Key: invalid key type 'domain' for association.",
		},
		{
			testName:            "error with blank name",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc3")),
			name:                "",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              `Name: cannot be blank`,
		},
		{
			testName:            "error empty from class key",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc4")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        identity.Key{},
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "keyType: cannot be blank",
		},
		{
			testName:            "error wrong from class key type",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc5")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        domainKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          toClassKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "FromClassKey: invalid key type 'domain' for from class.",
		},
		{
			testName:            "error empty to class key",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc6")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          identity.Key{},
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "keyType: cannot be blank",
		},
		{
			testName:            "error wrong to class key type",
			key:                 helper.Must(identity.NewClassAssociationKey(domainKey, "assoc7")),
			name:                "Name",
			details:             "Details",
			fromClassKey:        fromClassKey,
			fromMultiplicity:    multiplicity,
			toClassKey:          domainKey,
			toMultiplicity:      multiplicity,
			associationClassKey: assocClassKey,
			umlComment:          "UmlComment",
			errstr:              "ToClassKey: invalid key type 'domain' for to class.",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewAssociation(tt.key, tt.name, tt.details, tt.fromClassKey, tt.fromMultiplicity, tt.toClassKey, tt.toMultiplicity, tt.associationClassKey, tt.umlComment)
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

func (suite *AssociationSuite) TestOther() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	multiplicity, err := NewMultiplicity("2..3")
	assert.Nil(suite.T(), err)

	assocKey := helper.Must(identity.NewClassAssociationKey(domainKey, "assoc1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	unknownClassKey := helper.Must(identity.NewClassKey(subdomainKey, "unknown"))

	tests := []struct {
		testName string
		obj      Association
		classKey identity.Key
		otherKey identity.Key
		errstr   string
	}{
		// OK.
		{
			testName: "other of from is to",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: fromClassKey,
			otherKey: toClassKey,
		},
		{
			testName: "other of to is from",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: toClassKey,
			otherKey: fromClassKey,
		},

		// Error states.
		{
			testName: "error with unknown class key",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				Details:          "Details",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
				UmlComment:       "UmlComment",
			},
			classKey: unknownClassKey,
			errstr:   `association does not include class:`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			otherKey, err := tt.obj.Other(tt.classKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.otherKey, otherKey)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, otherKey)
			}
		})
	}
}
