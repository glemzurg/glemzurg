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
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
		},
		{
			testName: "valid guard with specification",
			guard: Guard{
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "Balance must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "self.balance > 0",
				},
			},
		},
		{
			testName: "error empty key",
			guard: Guard{
				Key:  identity.Key{},
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			guard: Guard{
				Key:  domainKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "Key: invalid key type 'domain' for guard",
		},
		{
			testName: "error blank name",
			guard: Guard{
				Key:  validKey,
				Name: "",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error invalid logic missing key",
			guard: Guard{
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "logic",
		},
		{
			testName: "error invalid logic missing description",
			guard: Guard{
				Key:  validKey,
				Name: "Name",
				Logic: model_logic.Logic{
					Key: "guard_logic_1", Description: "", Notation: model_logic.NotationTLAPlus,
				},
			},
			errstr: "logic",
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

	logic := model_logic.Logic{
		Key: "guard_logic_1", Description: "Balance check.", Notation: model_logic.NotationTLAPlus, Specification: "self.x > 0",
	}

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
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GuardSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	validLogic := model_logic.Logic{
		Key: "guard_logic_1", Description: "Guard condition.", Notation: model_logic.NotationTLAPlus,
	}

	// Test that Validate is called.
	guard := Guard{
		Key:   validKey,
		Name:  "", // Invalid
		Logic: validLogic,
	}
	err := guard.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

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
}
