package model_domain

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
	problemDomainKey  identity.Key
	solutionDomainKey identity.Key
}

func (suite *AssociationSuite) SetupTest() {
	suite.problemDomainKey = helper.Must(identity.NewDomainKey("domain1"))
	suite.solutionDomainKey = helper.Must(identity.NewDomainKey("domain2"))
}

// TestValidate tests all validation rules for Association.
func (suite *AssociationSuite) TestValidate() {
	validKey := helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey))

	tests := []struct {
		testName    string
		association Association
		errstr      string
	}{
		{
			testName: "valid association",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
			},
		},
		{
			testName: "error empty key",
			association: Association{
				Key:               identity.Key{},
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			association: Association{
				Key:               helper.Must(identity.NewActorKey("actor1")),
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
			},
			errstr: "Key: invalid key type 'actor' for domain association",
		},
		{
			testName: "error empty problem key",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  identity.Key{},
				SolutionDomainKey: suite.solutionDomainKey,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong problem key type",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  helper.Must(identity.NewActorKey("actor1")),
				SolutionDomainKey: suite.solutionDomainKey,
			},
			errstr: "ProblemDomainKey: invalid key type 'actor' for domain",
		},
		{
			testName: "error empty solution key",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: identity.Key{},
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong solution key type",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: helper.Must(identity.NewActorKey("actor1")),
			},
			errstr: "SolutionDomainKey: invalid key type 'actor' for domain",
		},
		{
			testName: "error ProblemDomainKey and SolutionDomainKey are the same",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.problemDomainKey,
			},
			errstr: "ProblemDomainKey and SolutionDomainKey cannot be the same",
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
	key := helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey))

	// Test parameters are mapped correctly.
	assoc, err := NewAssociation(key, suite.problemDomainKey, suite.solutionDomainKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Association{
		Key:               key,
		ProblemDomainKey:  suite.problemDomainKey,
		SolutionDomainKey: suite.solutionDomainKey,
		UmlComment:        "UmlComment",
	}, assoc)

	// Test that Validate is called (invalid data should fail).
	_, err = NewAssociation(identity.Key{}, suite.problemDomainKey, suite.solutionDomainKey, "UmlComment")
	assert.ErrorContains(suite.T(), err, "keyType: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *AssociationSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	assoc := Association{
		Key:               identity.Key{}, // Invalid
		ProblemDomainKey:  suite.problemDomainKey,
		SolutionDomainKey: suite.solutionDomainKey,
	}
	err := assoc.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "keyType: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - domain association is a root key, so it should not have a parent.
	assoc = Association{
		Key:               validKey,
		ProblemDomainKey:  suite.problemDomainKey,
		SolutionDomainKey: suite.solutionDomainKey,
	}
	err = assoc.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case - domain association key has no parent (root-level entity).
	err = assoc.ValidateWithParent(nil)
	assert.NoError(suite.T(), err)
}

// TestValidateReferences tests that ValidateReferences validates domain references correctly.
func (suite *AssociationSuite) TestValidateReferences() {
	validKey := helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey))
	nonExistentDomainKey := helper.Must(identity.NewDomainKey("nonexistent"))

	// Build lookup map with all valid domains.
	domains := map[identity.Key]bool{
		suite.problemDomainKey:  true,
		suite.solutionDomainKey: true,
	}

	tests := []struct {
		testName    string
		association Association
		domains     map[identity.Key]bool
		errstr      string
	}{
		{
			testName: "valid association with all domains existing",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
			},
			domains: domains,
		},
		{
			testName: "error ProblemDomainKey references non-existent domain",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  nonExistentDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
			},
			domains: domains,
			errstr:  "references non-existent problem domain",
		},
		{
			testName: "error SolutionDomainKey references non-existent domain",
			association: Association{
				Key:               validKey,
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: nonExistentDomainKey,
			},
			domains: domains,
			errstr:  "references non-existent solution domain",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.association.ValidateReferences(tt.domains)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}
