package req_model

import (
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

// TestValidate tests all validation rules for Model.
func (suite *ModelSuite) TestValidate() {
	tests := []struct {
		testName string
		model    Model
		errstr   string
	}{
		{
			testName: "valid model",
			model: Model{
				Key:  "model1",
				Name: "Name",
			},
		},
		{
			testName: "error blank key",
			model: Model{
				Key:  "",
				Name: "Name",
			},
			errstr: "Key: cannot be blank",
		},
		{
			testName: "error blank name",
			model: Model{
				Key:  "model1",
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewModel maps parameters correctly and calls Validate.
func (suite *ModelSuite) TestNew() {
	// Test parameters are mapped correctly (key is normalized to lowercase and trimmed).
	model, err := NewModel("  MODEL1  ", "Name", "Details")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "model1", model.Key)
	assert.Equal(suite.T(), "Name", model.Name)
	assert.Equal(suite.T(), "Details", model.Details)

	// Test that Validate is called (invalid data should fail).
	_, err = NewModel("model1", "", "Details")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
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
