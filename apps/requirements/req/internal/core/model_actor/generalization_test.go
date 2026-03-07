package model_actor

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(GeneralizationSuite))
}

type GeneralizationSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Generalization.
func (suite *GeneralizationSuite) TestValidate() {
	validKey := helper.Must(identity.NewActorGeneralizationKey("gen1"))

	tests := []struct {
		testName       string
		generalization Generalization
		errstr         string
	}{
		{
			testName: "valid generalization",
			generalization: Generalization{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			generalization: Generalization{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			generalization: Generalization{
				Key:  helper.Must(identity.NewDomainKey("domain1")),
				Name: "Name",
			},
			errstr: "key: invalid key type 'domain' for actor generalization",
		},
		{
			testName: "error blank name",
			generalization: Generalization{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.generalization.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGeneralization maps parameters correctly and calls Validate.
func (suite *GeneralizationSuite) TestNew() {
	key := helper.Must(identity.NewActorGeneralizationKey("gen1"))

	// Test parameters are mapped correctly.
	gen, err := NewGeneralization(key, "Name", "Details", true, false, "UmlComment")
	suite.Require().NoError(err)
	suite.Equal(Generalization{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, gen)

	// Test that Validate is called (invalid data should fail).
	_, err = NewGeneralization(key, "", "Details", true, false, "UmlComment")
	suite.ErrorContains(err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GeneralizationSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewActorGeneralizationKey("gen1"))

	// Test that Validate is called.
	gen := Generalization{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := gen.ValidateWithParent(nil)
	suite.ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - actor generalizations should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	gen = Generalization{
		Key:  validKey,
		Name: "Name",
	}
	err = gen.ValidateWithParent(&domainKey)
	suite.ErrorContains(err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = gen.ValidateWithParent(nil)
	suite.Require().NoError(err)
}
