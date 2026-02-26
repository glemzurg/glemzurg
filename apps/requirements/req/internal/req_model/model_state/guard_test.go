package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGuardSuite(t *testing.T) {
	suite.Run(t, new(GuardSuite))
}

type GuardSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Guard.
func (suite *GuardSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))

	tests := []struct {
		testName string
		guard    Guard
		errstr   string
	}{
		{
			testName: "valid guard minimal",
			guard: Guard{
				Key:   validKey,
				Name:  "Name",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
			},
		},
		{
			testName: "valid guard with specification",
			guard: Guard{
				Key:   validKey,
				Name:  "Name",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Balance must be positive.", "", model_logic.NotationTLAPlus, "self.balance > 0")),
			},
		},
		{
			testName: "error empty key",
			guard: Guard{
				Key:   identity.Key{},
				Name:  "Name",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			guard: Guard{
				Key:   domainKey,
				Name:  "Name",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
			},
			errstr: "Key: invalid key type 'domain' for guard",
		},
		{
			testName: "error blank name",
			guard: Guard{
				Key:   validKey,
				Name:  "",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
			},
			errstr: "Name",
		},
		{
			testName: "error invalid logic missing key",
			guard: Guard{
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "KeyType",
		},
		{
			testName: "error invalid logic missing description",
			guard: Guard{
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: validKey, Type: model_logic.LogicTypeAssessment, Description: "", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "Description",
		},
		{
			testName: "error logic wrong kind",
			guard: Guard{
				Key:   validKey,
				Name:  "Name",
				Logic: helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeStateChange, "Guard condition.", "x", model_logic.NotationTLAPlus, "")),
			},
			errstr: "logic kind must be 'assessment'",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.guard.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGuard maps parameters correctly and calls Validate.
func (suite *GuardSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewGuardKey(classKey, "guard1"))

	logic := helper.Must(model_logic.NewLogic(key, model_logic.LogicTypeAssessment, "Balance check.", "", model_logic.NotationTLAPlus, "self.x > 0"))

	// Test all parameters are mapped correctly.
	guard, err := NewGuard(key, "Name", logic)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Guard{
		Key:   key,
		Name:  "Name",
		Logic: logic,
	}, guard)

	// Test that Validate is called (invalid data should fail).
	_, err = NewGuard(key, "", logic)
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GuardSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	validLogic := helper.Must(model_logic.NewLogic(validKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, ""))

	// Test that Validate is called.
	guard := Guard{
		Key:   validKey,
		Name:  "", // Invalid
		Logic: validLogic,
	}
	err := guard.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - guard key has class1 as parent, but we pass other_class.
	guard = Guard{
		Key:   validKey,
		Name:  "Name",
		Logic: validLogic,
	}
	err = guard.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = guard.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)

	// Test logic key equality - logic key must match the guard's own key.
	differentGuardKey := helper.Must(identity.NewGuardKey(classKey, "other_guard"))
	guard = Guard{
		Key:   validKey,
		Name:  "Name",
		Logic: helper.Must(model_logic.NewLogic(differentGuardKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
	}
	err = guard.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "does not match guard key", "ValidateWithParent should enforce logic key == guard key")

	// Test logic ValidateWithParent is called - wrong parent should fail.
	otherClassKey2 := helper.Must(identity.NewClassKey(subdomainKey, "wrong_class"))
	wrongParentGuardKey := helper.Must(identity.NewGuardKey(otherClassKey2, "guard1"))
	guard = Guard{
		Key:   wrongParentGuardKey,
		Name:  "Name",
		Logic: helper.Must(model_logic.NewLogic(wrongParentGuardKey, model_logic.LogicTypeAssessment, "Guard condition.", "", model_logic.NotationTLAPlus, "")),
	}
	// The guard key has otherClassKey2 as parent, but we pass otherClassKey as the parent.
	err = guard.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should validate logic key parent")
}
