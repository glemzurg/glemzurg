package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
	domainKey identity.Key
}

func (suite *SubdomainSuite) SetupTest() {
	suite.domainKey = helper.Must(identity.NewDomainKey("domain1"))
}

// TestValidate tests all validation rules for Subdomain.
func (suite *SubdomainSuite) TestValidate() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	tests := []struct {
		testName  string
		subdomain Subdomain
		errstr    string
	}{
		{
			testName: "valid subdomain",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			subdomain: Subdomain{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			subdomain: Subdomain{
				Key:  helper.Must(identity.NewActorKey("actor1")),
				Name: "Name",
			},
			errstr: "Key: invalid key type 'actor' for subdomain.",
		},
		{
			testName: "error blank name",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.subdomain.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewSubdomain maps parameters correctly and calls Validate.
func (suite *SubdomainSuite) TestNew() {
	key := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	// Test parameters are mapped correctly.
	subdomain, err := NewSubdomain(key, "Name", "Details", "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Subdomain{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)

	// Test that Validate is called (invalid data should fail).
	_, err = NewSubdomain(key, "", "Details", "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *SubdomainSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	subdomain := Subdomain{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := subdomain.ValidateWithParent(&suite.domainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - subdomain key has domain1 as parent, but we pass other_domain.
	subdomain = Subdomain{
		Key:  validKey,
		Name: "Name",
	}
	err = subdomain.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = subdomain.ValidateWithParent(&suite.domainKey)
	assert.NoError(suite.T(), err)
}

// TestSetClassAssociations tests that SetClassAssociations validates parent relationships.
func (suite *SubdomainSuite) TestSetClassAssociations() {
	subdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "other_subdomain"))
	classKey1 := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomainKey, "class2"))
	otherClassKey1 := helper.Must(identity.NewClassKey(otherSubdomainKey, "class1"))
	otherClassKey2 := helper.Must(identity.NewClassKey(otherSubdomainKey, "class2"))

	// Create a subdomain.
	subdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
	}

	// Test: valid association with subdomain as parent.
	validAssocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, classKey1, classKey2))
	validAssoc := model_class.Association{
		Key:              validAssocKey,
		Name:             "Association",
		FromClassKey:     classKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       classKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err := subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		validAssocKey: validAssoc,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(subdomain.ClassAssociations))

	// Test: error when association has no parent (model-level association).
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	otherDomainSubdomainKey := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain1"))
	crossDomainClassKey := helper.Must(identity.NewClassKey(otherDomainSubdomainKey, "class1"))
	modelLevelAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, classKey1, crossDomainClassKey))
	modelLevelAssoc := model_class.Association{
		Key:              modelLevelAssocKey,
		Name:             "Model Level Association",
		FromClassKey:     classKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       crossDomainClassKey,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err = subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		modelLevelAssocKey: modelLevelAssoc,
	})
	assert.ErrorContains(suite.T(), err, "has no parent")

	// Test: error when association parent is different subdomain.
	wrongParentAssocKey := helper.Must(identity.NewClassAssociationKey(otherSubdomainKey, otherClassKey1, otherClassKey2))
	wrongParentAssoc := model_class.Association{
		Key:              wrongParentAssocKey,
		Name:             "Wrong Parent Association",
		FromClassKey:     otherClassKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       otherClassKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err = subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		wrongParentAssocKey: wrongParentAssoc,
	})
	assert.ErrorContains(suite.T(), err, "parent does not match subdomain")
}
