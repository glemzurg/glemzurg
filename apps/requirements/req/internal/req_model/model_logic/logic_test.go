package model_logic

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogicTestSuite struct {
	suite.Suite
}

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicTestSuite))
}

// validSpec returns a valid ExpressionSpec for testing.
func validSpec() model_spec.ExpressionSpec {
	return model_spec.ExpressionSpec{Notation: NotationTLAPlus}
}

// validSpecWithBody returns a valid ExpressionSpec with a specification body.
func validSpecWithBody(body string) model_spec.ExpressionSpec {
	return model_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: body}
}

// TestValidate tests all validation rules for Logic.
func (s *LogicTestSuite) TestValidate() {
	validKey := helper.Must(identity.NewInvariantKey("0"))
	validKey2 := helper.Must(identity.NewInvariantKey("1"))
	validKey3 := helper.Must(identity.NewInvariantKey("2"))

	tests := []struct {
		testName string
		logic    Logic
		errstr   string
	}{
		{
			testName: "valid minimal",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "All orders must have at least one item.",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid with specification",
			logic: Logic{
				Key:         validKey2,
				Type:        LogicTypeAssessment,
				Description: "Stock is never negative.",
				Spec:        validSpecWithBody("\\A p \\in Products : p.stock >= 0"),
			},
		},
		{
			testName: "valid with empty specification",
			logic: Logic{
				Key:         validKey3,
				Type:        LogicTypeAssessment,
				Description: "Placeholder invariant.",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid state_change kind",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeStateChange,
				Description: "Some state change.",
				Target:      "shipping",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid query kind",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeQuery,
				Description: "Some query.",
				Target:      "result",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid query kind with mixed case target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeQuery,
				Description: "Some query.",
				Target:      "TotalAmount",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid safety_rule kind",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeSafetyRule,
				Description: "Some safety rule.",
				Spec:        validSpec(),
			},
		},
		{
			testName: "valid value kind",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeValue,
				Description: "Some value.",
				Spec:        validSpec(),
			},
		},
		{
			testName: "error missing key",
			logic: Logic{
				Key:         identity.Key{},
				Type:        LogicTypeAssessment,
				Description: "Some description.",
				Spec:        validSpec(),
			},
			errstr: "KeyType",
		},
		{
			testName: "error missing kind",
			logic: Logic{
				Key:         validKey,
				Type:        "",
				Description: "Some description.",
				Spec:        validSpec(),
			},
			errstr: "Type",
		},
		{
			testName: "error invalid kind",
			logic: Logic{
				Key:         validKey,
				Type:        "unknown",
				Description: "Some description.",
				Spec:        validSpec(),
			},
			errstr: "Type",
		},
		{
			testName: "error missing description",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "",
				Spec:        validSpec(),
			},
			errstr: "Description",
		},
		{
			testName: "error missing notation in spec",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "Some description.",
				Spec:        model_spec.ExpressionSpec{},
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid notation in spec",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "Some description.",
				Spec:        model_spec.ExpressionSpec{Notation: "Z"},
			},
			errstr: "Notation",
		},
		{
			testName: "error missing key and description",
			logic: Logic{
				Key:         identity.Key{},
				Type:        LogicTypeAssessment,
				Description: "",
				Spec:        validSpec(),
			},
			errstr: "KeyType",
		},
		{
			testName: "error missing key with specification set",
			logic: Logic{
				Key:         identity.Key{},
				Type:        LogicTypeAssessment,
				Description: "Some description.",
				Spec:        validSpecWithBody("TRUE"),
			},
			errstr: "KeyType",
		},
		{
			testName: "error invalid notation with specification set",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "Some description.",
				Spec:        model_spec.ExpressionSpec{Notation: "Alloy", Specification: "some spec"},
			},
			errstr: "Notation",
		},
		// Target validation.
		{
			testName: "error state_change missing target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeStateChange,
				Description: "Some state change.",
				Spec:        validSpec(),
			},
			errstr: "requires a non-empty target",
		},
		{
			testName: "error query missing target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeQuery,
				Description: "Some query.",
				Spec:        validSpec(),
			},
			errstr: "requires a non-empty target",
		},
		{
			testName: "error query target starts with underscore",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeQuery,
				Description: "Some query.",
				Target:      "_hidden",
				Spec:        validSpec(),
			},
			errstr: "starting with '_'",
		},
		{
			testName: "error assessment with target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeAssessment,
				Description: "Some assessment.",
				Target:      "shipping",
				Spec:        validSpec(),
			},
			errstr: "must not have a target",
		},
		{
			testName: "error safety_rule with target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeSafetyRule,
				Description: "Some safety rule.",
				Target:      "shipping",
				Spec:        validSpec(),
			},
			errstr: "must not have a target",
		},
		{
			testName: "error value with target",
			logic: Logic{
				Key:         validKey,
				Type:        LogicTypeValue,
				Description: "Some value.",
				Target:      "shipping",
				Spec:        validSpec(),
			},
			errstr: "must not have a target",
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
	validKey := helper.Must(identity.NewInvariantKey("0"))
	validKey2 := helper.Must(identity.NewInvariantKey("1"))

	// Test all parameters are mapped correctly (assessment — no target).
	spec := validSpecWithBody("\\A p \\in Products : p.stock >= 0")
	logic, err := NewLogic(validKey, LogicTypeAssessment, "Stock is never negative.", "", spec, nil)
	s.NoError(err)
	s.Equal(Logic{
		Key:         validKey,
		Type:        LogicTypeAssessment,
		Description: "Stock is never negative.",
		Spec:        spec,
	}, logic)

	// Test with empty specification (optional).
	logic, err = NewLogic(validKey2, LogicTypeAssessment, "Placeholder.", "", validSpec(), nil)
	s.NoError(err)
	s.Equal(Logic{
		Key:         validKey2,
		Type:        LogicTypeAssessment,
		Description: "Placeholder.",
		Spec:        validSpec(),
	}, logic)

	// Test state_change with target.
	logic, err = NewLogic(validKey, LogicTypeStateChange, "Set shipping.", "shipping", validSpecWithBody("address"), nil)
	s.NoError(err)
	s.Equal("shipping", logic.Target)

	// Test query with target.
	logic, err = NewLogic(validKey, LogicTypeQuery, "Return result.", "result", validSpecWithBody("expr"), nil)
	s.NoError(err)
	s.Equal("result", logic.Target)

	// Test that Validate is called (invalid data should fail).
	_, err = NewLogic(identity.Key{}, LogicTypeAssessment, "Some description.", "", validSpec(), nil)
	s.Error(err)
	s.Contains(err.Error(), "KeyType")

	// Test that invalid notation fails.
	_, err = NewLogic(validKey, LogicTypeAssessment, "Some description.", "", model_spec.ExpressionSpec{Notation: "Z"}, nil)
	s.Error(err)
	s.Contains(err.Error(), "Notation")

	// Test that invalid kind fails.
	_, err = NewLogic(validKey, "bogus", "Some description.", "", validSpec(), nil)
	s.Error(err)
	s.Contains(err.Error(), "Type")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (s *LogicTestSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewInvariantKey("0"))

	// Test valid case - invariant keys have nil parent.
	logic := Logic{
		Key:         validKey,
		Type:        LogicTypeAssessment,
		Description: "Some description.",
		Spec:        validSpec(),
	}
	err := logic.ValidateWithParent(nil)
	s.NoError(err)

	// Test that Validate is called.
	logic = Logic{
		Key:         validKey,
		Type:        LogicTypeAssessment,
		Description: "", // Invalid
		Spec:        validSpec(),
	}
	err = logic.ValidateWithParent(nil)
	s.ErrorContains(err, "Description")

	// Test that ValidateParent is called - invariant key should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	logic = Logic{
		Key:         validKey,
		Type:        LogicTypeAssessment,
		Description: "Some description.",
		Spec:        validSpec(),
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
		Type:        LogicTypeAssessment,
		Description: "Precondition.",
		Spec:        validSpec(),
	}
	err = logic.ValidateWithParent(&actionKey)
	s.NoError(err)

	// Test wrong parent for action require key.
	err = logic.ValidateWithParent(&classKey)
	s.Error(err)
}
