package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	reqKey := helper.Must(identity.NewActionRequireKey(validKey, "req_1"))
	guarKey := helper.Must(identity.NewActionGuaranteeKey(validKey, "guar_1"))
	safetyKey := helper.Must(identity.NewActionSafetyKey(validKey, "safety_1"))

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
					model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition 1.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "req1"}, nil),
				},
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Postcondition 1.", "shipping", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "guar1"}, nil),
				},
				SafetyRules: []model_logic.Logic{
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeSafetyRule, "Safety rule 1.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "safety1"}, nil),
				},
			},
		},
		{
			testName: "valid action with requires only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "x must be positive.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "x > 0"}, nil),
				},
			},
		},
		{
			testName: "valid action with guarantees only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x to 1.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
				},
			},
		},
		{
			testName: "valid action with safety rules only",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeSafetyRule, "x must stay positive.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.x' > 0"}, nil),
				},
			},
		},
		{
			testName: "error empty key",
			action: Action{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "key type is required",
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
					model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "x must be positive.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "x > 0"}, nil),
				},
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x to 1.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
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
					{Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Spec: model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}},
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
					{Key: identity.Key{}, Type: model_logic.LogicTypeStateChange, Description: "Set x to 1.", Spec: model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}},
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
					{Key: identity.Key{}, Type: model_logic.LogicTypeSafetyRule, Description: "x must stay positive.", Spec: model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}},
				},
			},
			errstr: "safety rule 0",
		},
		{
			testName: "error requires wrong kind",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey, model_logic.LogicTypeStateChange, "x must be positive.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
				},
			},
			errstr: "requires 0: logic kind must be 'assessment' or 'let'",
		},
		{
			testName: "error guarantee wrong kind",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeAssessment, "Set x to 1.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
				},
			},
			errstr: "guarantee 0: logic kind must be 'state_change' or 'let'",
		},
		{
			testName: "error safety rule wrong kind",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeAssessment, "x must stay positive.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
				},
			},
			errstr: "safety rule 0: logic kind must be 'safety_rule' or 'let'",
		},
		{
			testName: "error duplicate guarantee target",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x again.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate target",
		},
		// Let in logic lists.
		{
			testName: "valid action with let in requires",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey, model_logic.LogicTypeLet, "Local total.", "total", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1 + 2"}, nil),
				},
			},
		},
		{
			testName: "valid action with let in guarantees",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeLet, "Local value.", "localVar", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1 + 2"}, nil),
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
				},
			},
		},
		{
			testName: "valid action with let in safety rules",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeLet, "Local value.", "localVar", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1 + 2"}, nil),
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeSafetyRule, "x must stay positive.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.x' > 0"}, nil),
				},
			},
		},
		{
			testName: "error duplicate let target in requires",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey, model_logic.LogicTypeLet, "Local a.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(reqKey, model_logic.LogicTypeLet, "Local a again.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate let target \"a\"",
		},
		{
			testName: "error duplicate let target in guarantees",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeLet, "Local a.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(guarKey, model_logic.LogicTypeLet, "Local a again.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate let target \"a\"",
		},
		{
			testName: "error let target collides with state_change target in guarantees",
			action: Action{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Set x.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(guarKey, model_logic.LogicTypeLet, "Local x.", "x", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate let target \"x\"",
		},
		{
			testName: "error duplicate let target in safety rules",
			action: Action{
				Key:  validKey,
				Name: "Name",
				SafetyRules: []model_logic.Logic{
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeLet, "Local a.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(safetyKey, model_logic.LogicTypeLet, "Local a again.", "a", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate let target \"a\"",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.action.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
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
	reqKey := helper.Must(identity.NewActionRequireKey(key, "req_1"))
	guarKey := helper.Must(identity.NewActionGuaranteeKey(key, "guar_1"))
	safetyKey := helper.Must(identity.NewActionSafetyKey(key, "safety_1"))

	requires := []model_logic.Logic{
		model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "tla_req"}, nil),
	}
	guarantees := []model_logic.Logic{
		model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Postcondition.", "shipping", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "tla_guar"}, nil),
	}
	safetyRules := []model_logic.Logic{
		model_logic.NewLogic(safetyKey, model_logic.LogicTypeSafetyRule, "Safety rule.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "tla_safety"}, nil),
	}

	// Test all parameters are mapped correctly.
	params := []Parameter{
		helper.Must(NewParameter("ParamA", "Nat")),
		helper.Must(NewParameter("ParamB", "Int")),
	}

	action := NewAction(key, "Name", "Details",
		requires, guarantees, safetyRules, params)
	suite.Equal(Action{
		Key:         key,
		Name:        "Name",
		Details:     "Details",
		Requires:    requires,
		Guarantees:  guarantees,
		SafetyRules: safetyRules,
		Parameters: []Parameter{
			helper.Must(NewParameter("ParamA", "Nat")),
			helper.Must(NewParameter("ParamB", "Int")),
		},
	}, action)

	// Test with nil optional fields (all Logic slice fields are optional).

	action = NewAction(key, "Name", "Details",
		nil, nil, nil, nil)
	suite.Equal(Action{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, action)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ActionSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))
	reqKey := helper.Must(identity.NewActionRequireKey(validKey, "req_1"))
	guarKey := helper.Must(identity.NewActionGuaranteeKey(validKey, "guar_1"))
	safetyKey := helper.Must(identity.NewActionSafetyKey(validKey, "safety_1"))

	// Test that Validate is called.
	action := Action{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := action.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - action key has class1 as parent, but we pass other_class.
	action = Action{
		Key:  validKey,
		Name: "Name",
	}
	err = action.ValidateWithParent(&otherClassKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = action.ValidateWithParent(&classKey)
	suite.Require().NoError(err)

	// Test valid with logic children.
	action = Action{
		Key:  validKey,
		Name: "Name",
		Requires: []model_logic.Logic{
			model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
		Guarantees: []model_logic.Logic{
			model_logic.NewLogic(guarKey, model_logic.LogicTypeStateChange, "Postcondition.", "shipping", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
		SafetyRules: []model_logic.Logic{
			model_logic.NewLogic(safetyKey, model_logic.LogicTypeSafetyRule, "Safety rule.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
	}
	err = action.ValidateWithParent(&classKey)
	suite.Require().NoError(err)

	// Test logic key validation - require with wrong parent should fail.
	otherActionKey := helper.Must(identity.NewActionKey(classKey, "other_action"))
	wrongReqKey := helper.Must(identity.NewActionRequireKey(otherActionKey, "req_1"))
	action = Action{
		Key:  validKey,
		Name: "Name",
		Requires: []model_logic.Logic{
			model_logic.NewLogic(wrongReqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
	}
	err = action.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "requires 0", "ValidateWithParent should validate logic key parent")

	// Test child Parameter validation propagates error.
	action = Action{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			{Name: "", DataTypeRules: "Nat"}, // Invalid: blank name
		},
	}
	err = action.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should validate child Parameters")

	// Test valid with child Parameters.
	action = Action{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			helper.Must(NewParameter("param1", "Nat")),
		},
	}
	err = action.ValidateWithParent(&classKey)
	suite.Require().NoError(err)
}
