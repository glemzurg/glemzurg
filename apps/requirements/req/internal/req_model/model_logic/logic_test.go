package model_logic

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogicTestSuite struct {
	suite.Suite
}

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicTestSuite))
}

// TestValidate tests all validation rules for Logic.
func (s *LogicTestSuite) TestValidate() {
	validKey := helper.Must(identity.NewInvariantKey("inv_1"))
	validKey2 := helper.Must(identity.NewInvariantKey("inv_2"))
	validKey3 := helper.Must(identity.NewInvariantKey("inv_3"))

	tests := []struct {
		testName string
		logic    Logic
		errstr   string
	}{
		{
			testName: "valid minimal",
			logic: Logic{
				Key:         validKey,
				Description: "All orders must have at least one item.",
				Notation:    NotationTLAPlus,
			},
		},
		{
			testName: "valid with specification",
			logic: Logic{
				Key:           validKey2,
				Description:   "Stock is never negative.",
				Notation:      NotationTLAPlus,
				Specification: "\\A p \\in Products : p.stock >= 0",
			},
		},
		{
			testName: "valid with empty specification",
			logic: Logic{
				Key:           validKey3,
				Description:   "Placeholder invariant.",
				Notation:      NotationTLAPlus,
				Specification: "",
			},
		},
		{
			testName: "error missing key",
			logic: Logic{
				Key:         identity.Key{},
				Description: "Some description.",
				Notation:    NotationTLAPlus,
			},
			errstr: "KeyType",
		},
		{
			testName: "error missing description",
			logic: Logic{
				Key:         validKey,
				Description: "",
				Notation:    NotationTLAPlus,
			},
			errstr: "Description",
		},
		{
			testName: "error missing notation",
			logic: Logic{
				Key:         validKey,
				Description: "Some description.",
				Notation:    "",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid notation",
			logic: Logic{
				Key:         validKey,
				Description: "Some description.",
				Notation:    "Z",
			},
			errstr: "Notation",
		},
		{
			testName: "error missing key and description",
			logic: Logic{
				Key:         identity.Key{},
				Description: "",
				Notation:    NotationTLAPlus,
			},
			errstr: "KeyType",
		},
		{
			testName: "error missing key with specification set",
			logic: Logic{
				Key:           identity.Key{},
				Description:   "Some description.",
				Notation:      NotationTLAPlus,
				Specification: "TRUE",
			},
			errstr: "KeyType",
		},
		{
			testName: "error invalid notation with specification set",
			logic: Logic{
				Key:           validKey,
				Description:   "Some description.",
				Notation:      "Alloy",
				Specification: "some spec",
			},
			errstr: "Notation",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.logic.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

// TestNew tests that NewLogic maps parameters correctly and calls Validate.
func (s *LogicTestSuite) TestNew() {
	validKey := helper.Must(identity.NewInvariantKey("inv_1"))
	validKey2 := helper.Must(identity.NewInvariantKey("inv_2"))

	// Test all parameters are mapped correctly.
	logic, err := NewLogic(validKey, "Stock is never negative.", NotationTLAPlus, "\\A p \\in Products : p.stock >= 0")
	s.NoError(err)
	s.Equal(Logic{
		Key:           validKey,
		Description:   "Stock is never negative.",
		Notation:      NotationTLAPlus,
		Specification: "\\A p \\in Products : p.stock >= 0",
	}, logic)

	// Test with empty specification (optional).
	logic, err = NewLogic(validKey2, "Placeholder.", NotationTLAPlus, "")
	s.NoError(err)
	s.Equal(Logic{
		Key:         validKey2,
		Description: "Placeholder.",
		Notation:    NotationTLAPlus,
	}, logic)

	// Test that Validate is called (invalid data should fail).
	_, err = NewLogic(identity.Key{}, "Some description.", NotationTLAPlus, "")
	s.Error(err)
	s.Contains(err.Error(), "KeyType")

	// Test that invalid notation fails.
	_, err = NewLogic(validKey, "Some description.", "Z", "")
	s.Error(err)
	s.Contains(err.Error(), "Notation")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (s *LogicTestSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewInvariantKey("inv_1"))

	// Test valid case - invariant keys have nil parent.
	logic := Logic{
		Key:         validKey,
		Description: "Some description.",
		Notation:    NotationTLAPlus,
	}
	err := logic.ValidateWithParent(nil)
	s.NoError(err)

	// Test that Validate is called.
	logic = Logic{
		Key:         validKey,
		Description: "", // Invalid
		Notation:    NotationTLAPlus,
	}
	err = logic.ValidateWithParent(nil)
	s.ErrorContains(err, "Description")

	// Test that ValidateParent is called - invariant key should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	logic = Logic{
		Key:         validKey,
		Description: "Some description.",
		Notation:    NotationTLAPlus,
	}
	err = logic.ValidateWithParent(&domainKey)
	s.ErrorContains(err, "should not have a parent")

	// Test with action require key and action parent.
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	requireKey := helper.Must(identity.NewActionRequireKey(actionKey, "req1"))

	logic = Logic{
		Key:         requireKey,
		Description: "Precondition.",
		Notation:    NotationTLAPlus,
	}
	err = logic.ValidateWithParent(&actionKey)
	s.NoError(err)

	// Test wrong parent for action require key.
	err = logic.ValidateWithParent(&classKey)
	s.Error(err)
}
