package req_model

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) TestNew() {
	tests := []struct {
		key     string
		name    string
		details string
		obj     Model
		errstr  string
	}{
		// OK.
		{
			key:     "model1",
			name:    "Name",
			details: "Details",
			obj: Model{
				Key:     "model1",
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			key:     "  MODEL1  ",
			name:    "Name",
			details: "",
			obj: Model{
				Key:     "model1",
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			key:     "",
			name:    "Name",
			details: "Details",
			errstr:  "Key: cannot be blank",
		},
		{
			key:     "model1",
			name:    "",
			details: "Details",
			errstr:  "Name: cannot be blank.",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		obj, err := NewModel(test.key, test.name, test.details)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

// TestValidateWithParent tests that ValidateWithParent calls both Validate() and ValidateParent().
// Individual Validate() and ValidateParent() methods are tested elsewhere.
// These tests just confirm both methods are invoked by ValidateWithParent.
func (suite *ModelSuite) TestValidateWithParent() {
	// Test 1: ValidateWithParent calls Model.Validate() - empty name should fail.
	model := Model{
		Key:     "model1",
		Name:    "", // Invalid - will fail Validate()
		Details: "Details",
	}
	err := model.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test 2: ValidateWithParent calls Actor.Validate() through the tree.
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Actors: []model_actor.Actor{
			{
				Key:  actorKey,
				Name: "", // Invalid - will fail Validate()
				Type: "person",
			},
		},
	}
	err = model.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call child Validate()")

	// Test 3: ValidateWithParent calls ValidateParent() - wrong parent key should fail.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	wrongParentSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Domains: []model_domain.Domain{
			{
				Key:     domainKey,
				Name:    "Domain Name",
				Details: "Details",
				Subdomains: []model_domain.Subdomain{
					{
						Key:     wrongParentSubdomainKey, // Parent is domain1, but attached to other_domain
						Name:    "Subdomain Name",
						Details: "Details",
					},
				},
			},
		},
	}
	// Manually set the wrong parent to test ValidateParent is called.
	model.Domains[0].Key = otherDomainKey
	err = model.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test 4: Valid model should pass.
	validSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Actors: []model_actor.Actor{
			{
				Key:     actorKey,
				Name:    "Actor Name",
				Type:    "person",
				Details: "Details",
			},
		},
		Domains: []model_domain.Domain{
			{
				Key:     domainKey,
				Name:    "Domain Name",
				Details: "Details",
				Subdomains: []model_domain.Subdomain{
					{
						Key:     validSubdomainKey,
						Name:    "Subdomain Name",
						Details: "Details",
					},
				},
			},
		},
	}
	err = model.ValidateWithParent()
	assert.NoError(suite.T(), err, "Valid model should pass ValidateWithParent()")
}
