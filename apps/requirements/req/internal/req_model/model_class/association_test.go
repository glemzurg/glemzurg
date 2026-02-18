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

// TestValidate tests all validation rules for Association.
func (suite *AssociationSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	validKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))

	tests := []struct {
		testName    string
		association Association
		errstr      string
	}{
		{
			testName: "valid association",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   toClassKey,
			},
		},
		{
			testName: "error empty key",
			association: Association{
				Key:          identity.Key{},
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   toClassKey,
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			association: Association{
				Key:          domainKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   toClassKey,
			},
			errstr: "Key: invalid key type 'domain' for association.",
		},
		{
			testName: "error blank name",
			association: Association{
				Key:          validKey,
				Name:         "",
				FromClassKey: fromClassKey,
				ToClassKey:   toClassKey,
			},
			errstr: "Name",
		},
		{
			testName: "error empty from class key",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: identity.Key{},
				ToClassKey:   toClassKey,
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong from class key type",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: domainKey,
				ToClassKey:   toClassKey,
			},
			errstr: "FromClassKey: invalid key type 'domain' for from class.",
		},
		{
			testName: "error empty to class key",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   identity.Key{},
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong to class key type",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   domainKey,
			},
			errstr: "ToClassKey: invalid key type 'domain' for to class.",
		},
		{
			testName: "error AssociationClassKey same as FromClassKey",
			association: Association{
				Key:                 validKey,
				Name:                "Name",
				FromClassKey:        fromClassKey,
				ToClassKey:          toClassKey,
				AssociationClassKey: &fromClassKey,
			},
			errstr: "AssociationClassKey cannot be the same as FromClassKey",
		},
		{
			testName: "error AssociationClassKey same as ToClassKey",
			association: Association{
				Key:                 validKey,
				Name:                "Name",
				FromClassKey:        fromClassKey,
				ToClassKey:          toClassKey,
				AssociationClassKey: &toClassKey,
			},
			errstr: "AssociationClassKey cannot be the same as ToClassKey",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.association.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewAssociation maps parameters correctly and calls Validate.
func (suite *AssociationSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	assocClassKey := helper.Must(identity.NewClassKey(subdomainKey, "assocclass"))
	key := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))
	multiplicity, err := NewMultiplicity("2..3")
	assert.NoError(suite.T(), err)

	// Test parameters are mapped correctly.
	assoc, err := NewAssociation(key, "Name", "Details", fromClassKey, multiplicity, toClassKey, multiplicity, &assocClassKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Association{
		Key:                 key,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        fromClassKey,
		FromMultiplicity:    multiplicity,
		ToClassKey:          toClassKey,
		ToMultiplicity:      multiplicity,
		AssociationClassKey: &assocClassKey,
		UmlComment:          "UmlComment",
	}, assoc)

	// Test that Validate is called (invalid data should fail).
	_, err = NewAssociation(key, "", "Details", fromClassKey, multiplicity, toClassKey, multiplicity, &assocClassKey, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *AssociationSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	validKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	assoc := Association{
		Key:          validKey,
		Name:         "", // Invalid
		FromClassKey: fromClassKey,
		ToClassKey:   toClassKey,
	}
	err := assoc.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - association key has subdomain1 as parent, but we pass other_subdomain.
	assoc = Association{
		Key:          validKey,
		Name:         "Name",
		FromClassKey: fromClassKey,
		ToClassKey:   toClassKey,
	}
	err = assoc.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = assoc.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)
}

// TestValidateReferences tests that ValidateReferences validates class references correctly.
func (suite *AssociationSuite) TestValidateReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	assocClassKey := helper.Must(identity.NewClassKey(subdomainKey, "assocclass"))
	nonExistentClassKey := helper.Must(identity.NewClassKey(subdomainKey, "nonexistent"))
	validKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))

	// Build lookup map with all valid classes.
	classes := map[identity.Key]bool{
		fromClassKey:  true,
		toClassKey:    true,
		assocClassKey: true,
	}

	tests := []struct {
		testName    string
		association Association
		classes     map[identity.Key]bool
		errstr      string
	}{
		{
			testName: "valid association with all classes existing",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   toClassKey,
			},
			classes: classes,
		},
		{
			testName: "valid association with AssociationClassKey",
			association: Association{
				Key:                 validKey,
				Name:                "Name",
				FromClassKey:        fromClassKey,
				ToClassKey:          toClassKey,
				AssociationClassKey: &assocClassKey,
			},
			classes: classes,
		},
		{
			testName: "error FromClassKey references non-existent class",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: nonExistentClassKey,
				ToClassKey:   toClassKey,
			},
			classes: classes,
			errstr:  "references non-existent from class",
		},
		{
			testName: "error ToClassKey references non-existent class",
			association: Association{
				Key:          validKey,
				Name:         "Name",
				FromClassKey: fromClassKey,
				ToClassKey:   nonExistentClassKey,
			},
			classes: classes,
			errstr:  "references non-existent to class",
		},
		{
			testName: "error AssociationClassKey references non-existent class",
			association: Association{
				Key:                 validKey,
				Name:                "Name",
				FromClassKey:        fromClassKey,
				ToClassKey:          toClassKey,
				AssociationClassKey: &nonExistentClassKey,
			},
			classes: classes,
			errstr:  "references non-existent association class",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.association.ValidateReferences(tt.classes)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestIncludes tests that Includes returns true for FromClassKey, ToClassKey, and AssociationClassKey.
func (suite *AssociationSuite) TestIncludes() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	assocClassKey := helper.Must(identity.NewClassKey(subdomainKey, "assocclass"))
	unknownClassKey := helper.Must(identity.NewClassKey(subdomainKey, "unknown"))
	validKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))

	// Without AssociationClassKey.
	assoc := Association{
		Key:          validKey,
		Name:         "Name",
		FromClassKey: fromClassKey,
		ToClassKey:   toClassKey,
	}
	assert.True(suite.T(), assoc.Includes(fromClassKey), "Should include FromClassKey")
	assert.True(suite.T(), assoc.Includes(toClassKey), "Should include ToClassKey")
	assert.False(suite.T(), assoc.Includes(unknownClassKey), "Should not include unknown key")

	// With AssociationClassKey.
	assoc.AssociationClassKey = &assocClassKey
	assert.True(suite.T(), assoc.Includes(assocClassKey), "Should include AssociationClassKey")
	assert.False(suite.T(), assoc.Includes(unknownClassKey), "Should not include unknown key")
}

func (suite *AssociationSuite) TestOther() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	multiplicity, err := NewMultiplicity("2..3")
	assert.NoError(suite.T(), err)

	fromClassKey := helper.Must(identity.NewClassKey(subdomainKey, "from"))
	toClassKey := helper.Must(identity.NewClassKey(subdomainKey, "to"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "test association"))
	unknownClassKey := helper.Must(identity.NewClassKey(subdomainKey, "unknown"))

	tests := []struct {
		testName string
		obj      Association
		classKey identity.Key
		otherKey identity.Key
		errstr   string
	}{
		{
			testName: "other of from is to",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
			},
			classKey: fromClassKey,
			otherKey: toClassKey,
		},
		{
			testName: "other of to is from",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
			},
			classKey: toClassKey,
			otherKey: fromClassKey,
		},
		{
			testName: "error with unknown class key",
			obj: Association{
				Key:              assocKey,
				Name:             "Name",
				FromClassKey:     fromClassKey,
				FromMultiplicity: multiplicity,
				ToClassKey:       toClassKey,
				ToMultiplicity:   multiplicity,
			},
			classKey: unknownClassKey,
			errstr:   `association does not include class:`,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
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
