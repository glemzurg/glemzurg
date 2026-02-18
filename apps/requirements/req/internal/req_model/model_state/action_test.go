package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActionSuite(t *testing.T) {
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Action.
func (suite *ActionSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewActionKey(classKey, "action1"))

	tests := []struct {
		testName string
		action   Action
		errstr   string
	}{
		{
			testName: "valid action minimal",
			action: Action{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "valid action with all optional fields",
			action: Action{
				Key:     validKey,
				Name:    "Name",
				Details: "Details",
				Requires: []model_logic.Logic{
					{Key: "req_1", Description: "Precondition 1.", Notation: model_logic.NotationTLAPlus, Specification: "req1"},
				},
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Postcondition 1.", Notation: model_logic.NotationTLAPlus, Specification: "guar1"},
				},
				SafetyRules: []model_logic.Logic{
					{Key: "safety_1", Description: "Safety rule 1.", Notation: model_logic.NotationTLAPlus, Specification: "safety1"},
				},
			},
		},
		{
			testName: "valid action with requires only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					{Key: "req_1", Description: "x must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "x > 0"},
				},
			},
		},
		{
			testName: "valid action with guarantees only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Set x to 1.", Notation: model_logic.NotationTLAPlus, Specification: "self.x' = 1"},
				},
			},
		},
		{
			testName: "valid action with safety rules only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					{Key: "safety_1", Description: "x must stay positive.", Notation: model_logic.NotationTLAPlus, Specification: "self.x' > 0"},
				},
			},
		},
		{
			testName: "error empty key",
			action: Action{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			action: Action{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for action",
		},
		{
			testName: "error blank name",
			action: Action{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
		{
			testName: "error blank name with logic fields set",
			action: Action{
				Key:  validKey,
				Name: "",
				Requires: []model_logic.Logic{
					{Key: "req_1", Description: "x must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "x > 0"},
				},
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Set x to 1.", Notation: model_logic.NotationTLAPlus, Specification: "self.x' = 1"},
				},
			},
			errstr: "Name",
		},
		{
			testName: "error invalid requires logic missing key",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					{Key: "", Description: "x must be positive.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "requires 0",
		},
		{
			testName: "error invalid guarantee logic missing key",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					{Key: "", Description: "Set x to 1.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "guarantee 0",
		},
		{
			testName: "error invalid safety rule logic missing key",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					{Key: "", Description: "x must stay positive.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "safety rule 0",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewAction maps parameters correctly and calls Validate.
func (suite *ActionSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewActionKey(classKey, "action1"))

	requires := []model_logic.Logic{
		{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "tla_req"},
	}
	guarantees := []model_logic.Logic{
		{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "tla_guar"},
	}
	safetyRules := []model_logic.Logic{
		{Key: "safety_1", Description: "Safety rule.", Notation: model_logic.NotationTLAPlus, Specification: "tla_safety"},
	}

	// Test all parameters are mapped correctly.
	action, err := NewAction(key, "Name", "Details",
		requires, guarantees, safetyRules, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Action{
		Key:         key,
		Name:        "Name",
		Details:     "Details",
		Requires:    requires,
		Guarantees:  guarantees,
		SafetyRules: safetyRules,
	}, action)

	// Test with nil optional fields (all Logic slice fields are optional).
	action, err = NewAction(key, "Name", "Details",
		nil, nil, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Action{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, action)

	// Test that Validate is called (invalid data should fail).
	_, err = NewAction(key, "", "Details", nil, nil, nil, nil)
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ActionSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	action := Action{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := action.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - action key has class1 as parent, but we pass other_class.
	action = Action{
		Key:  validKey,
		Name: "Name",
	}
	err = action.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = action.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)

	// Test child Parameter validation propagates error.
	action = Action{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			{Name: "", DataTypeRules: "Nat"}, // Invalid: blank name
		},
	}
	err = action.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should validate child Parameters")

	// Test valid with child Parameters.
	action = Action{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			{Name: "param1", DataTypeRules: "Nat"},
		},
	}
	err = action.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
