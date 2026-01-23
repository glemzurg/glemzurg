package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Class.
func (suite *ClassSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	tests := []struct {
		testName string
		class    Class
		errstr   string
	}{
		{
			testName: "valid class",
			class: Class{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			class: Class{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			class: Class{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for class.",
		},
		{
			testName: "error blank name",
			class: Class{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error SuperclassOfKey and SubclassOfKey are the same",
			class: func() Class {
				genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
				return Class{
					Key:             validKey,
					Name:            "Name",
					SuperclassOfKey: &genKey,
					SubclassOfKey:   &genKey,
				}
			}(),
			errstr: "SuperclassOfKey and SubclassOfKey cannot be the same",
		},
		{
			testName: "valid class with SuperclassOfKey referencing a generalization",
			class: Class{
				Key:             validKey,
				Name:            "Name",
				SuperclassOfKey: &genKey,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.class.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewClass maps parameters correctly and calls Validate.
func (suite *ClassSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	superclassOfKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	subclassOfKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen2"))

	// Test parameters are mapped correctly.
	class, err := NewClass(key, "Name", "Details", &actorKey, &superclassOfKey, &subclassOfKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Class{
		Key:             key,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &actorKey,
		SuperclassOfKey: &superclassOfKey,
		SubclassOfKey:   &subclassOfKey,
		UmlComment:      "UmlComment",
	}, class)

	// Test that Validate is called (invalid data should fail).
	_, err = NewClass(key, "", "Details", nil, nil, nil, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ClassSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	class := Class{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - class key has subdomain1 as parent, but we pass other_subdomain.
	class = Class{
		Key:  validKey,
		Name: "Name",
	}
	err = class.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = class.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)
}

// TestValidateReferences tests that ValidateReferences validates cross-references correctly.
func (suite *ClassSuite) TestValidateReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	genInOtherSubdomain := helper.Must(identity.NewGeneralizationKey(otherSubdomainKey, "gen2"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	nonExistentActorKey := helper.Must(identity.NewActorKey("nonexistent"))
	nonExistentGenKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "nonexistent"))

	// Build lookup maps
	actors := map[identity.Key]bool{
		actorKey: true,
	}
	generalizations := map[identity.Key]bool{
		genKey: true,
	}

	tests := []struct {
		testName        string
		class           Class
		actors          map[identity.Key]bool
		generalizations map[identity.Key]bool
		errstr          string
	}{
		{
			testName: "valid class with no references",
			class: Class{
				Key:  classKey,
				Name: "Name",
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "valid class with ActorKey reference",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &actorKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error ActorKey references non-existent actor",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &nonExistentActorKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent actor",
		},
		{
			testName: "valid class with SuperclassOfKey reference",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error SuperclassOfKey references non-existent generalization",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &nonExistentGenKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent generalization",
		},
		{
			testName: "error SuperclassOfKey references generalization in different subdomain",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genInOtherSubdomain,
			},
			actors: actors,
			generalizations: map[identity.Key]bool{
				genKey:              true,
				genInOtherSubdomain: true,
			},
			errstr: "must be in the same subdomain",
		},
		{
			testName: "valid class with SubclassOfKey reference",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error SubclassOfKey references non-existent generalization",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &nonExistentGenKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent generalization",
		},
		{
			testName: "error SubclassOfKey references generalization in different subdomain",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genInOtherSubdomain,
			},
			actors: actors,
			generalizations: map[identity.Key]bool{
				genKey:              true,
				genInOtherSubdomain: true,
			},
			errstr: "must be in the same subdomain",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.class.ValidateReferences(tt.actors, tt.generalizations)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}
