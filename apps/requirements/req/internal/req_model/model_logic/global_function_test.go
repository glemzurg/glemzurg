package model_logic

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GlobalFunctionTestSuite struct {
	suite.Suite
}

func TestGlobalFunctionSuite(t *testing.T) {
	suite.Run(t, new(GlobalFunctionTestSuite))
}

// TestValidate tests all validation rules for GlobalFunction.
func (s *GlobalFunctionTestSuite) TestValidate() {
	gfKey1 := helper.Must(identity.NewGlobalFunctionKey("_max"))
	gfKey2 := helper.Must(identity.NewGlobalFunctionKey("_valid_statuses"))
	gfKey3 := helper.Must(identity.NewGlobalFunctionKey("_constant"))

	validSpec := Logic{
		Key:         gfKey1,
		Type:        LogicTypeValue,
		Description: "Max of two values.",
		Notation:    NotationTLAPlus,
	}

	tests := []struct {
		testName string
		gf       GlobalFunction
		errstr   string
	}{
		{
			testName: "valid function with parameters",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic:      validSpec,
			},
		},
		{
			testName: "valid function with comment",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic:      validSpec,
			},
		},
		{
			testName: "valid set definition no parameters",
			gf: GlobalFunction{
				Key:  gfKey2,
				Name: "_ValidStatuses",
				Logic: Logic{
					Key:           gfKey2,
					Type:          LogicTypeValue,
					Description:   "Set of valid statuses.",
					Notation:      NotationTLAPlus,
					Specification: `{"pending", "active", "complete"}`,
				},
			},
		},
		{
			testName: "valid predicate with nil parameters",
			gf: GlobalFunction{
				Key:        gfKey3,
				Name:       "_Constant",
				Parameters: nil,
				Logic: Logic{
					Key:           gfKey3,
					Type:          LogicTypeValue,
					Description:   "A constant value.",
					Notation:      NotationTLAPlus,
					Specification: "42",
				},
			},
		},
		{
			testName: "error missing key",
			gf: GlobalFunction{
				Key:        identity.Key{},
				Name:       "_Max",
				Parameters: []string{"x"},
				Logic:      validSpec,
			},
			errstr: "KeyType",
		},
		{
			testName: "error wrong key type",
			gf: GlobalFunction{
				Key:        helper.Must(identity.NewInvariantKey("0")),
				Name:       "_Max",
				Parameters: []string{"x"},
				Logic:      validSpec,
			},
			errstr: "invalid key type",
		},
		{
			testName: "error missing name",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "",
				Parameters: []string{"x"},
				Logic:      validSpec,
			},
			errstr: "Name",
		},
		{
			testName: "error name missing underscore",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "Max",
				Parameters: []string{"x", "y"},
				Logic:      validSpec,
			},
			errstr: "must start with underscore",
		},
		{
			testName: "error missing specification key",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         identity.Key{},
					Type:        LogicTypeValue,
					Description: "Some desc.",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "KeyType",
		},
		{
			testName: "error specification key mismatch",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         gfKey2, // Different key than gfKey1
					Type:        LogicTypeValue,
					Description: "Some desc.",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "does not match global function key",
		},
		{
			testName: "error missing specification description",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         gfKey1,
					Type:        LogicTypeValue,
					Description: "",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "Description",
		},
		{
			testName: "error missing specification notation",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         gfKey1,
					Type:        LogicTypeValue,
					Description: "Some desc.",
					Notation:    "",
				},
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid specification notation",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         gfKey1,
					Type:        LogicTypeValue,
					Description: "Some desc.",
					Notation:    "Z",
				},
			},
			errstr: "Notation",
		},
		{
			testName: "error wrong logic kind",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic: Logic{
					Key:         gfKey1,
					Type:        LogicTypeAssessment, // Wrong kind for global function
					Description: "Some desc.",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "logic kind must be 'value'",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.gf.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGlobalFunction maps parameters correctly and calls Validate.
func (s *GlobalFunctionTestSuite) TestNew() {
	gfKey1 := helper.Must(identity.NewGlobalFunctionKey("_max"))
	gfKey2 := helper.Must(identity.NewGlobalFunctionKey("_constant"))

	spec := Logic{
		Key:           gfKey1,
		Type:          LogicTypeValue,
		Description:   "Returns the maximum of two values.",
		Notation:      NotationTLAPlus,
		Specification: "IF x > y THEN x ELSE y",
	}

	// Test all parameters are mapped correctly.
	gf, err := NewGlobalFunction(gfKey1, "_Max", []string{"x", "y"}, spec)
	s.NoError(err)
	s.Equal(GlobalFunction{
		Key:        gfKey1,
		Name:       "_Max",
		Parameters: []string{"x", "y"},
		Logic:      spec,
	}, gf)

	// Test with nil optional fields (Comment and Parameters are optional).
	gf, err = NewGlobalFunction(gfKey2, "_Constant", nil, Logic{
		Key:           gfKey2,
		Type:          LogicTypeValue,
		Description:   "A constant.",
		Notation:      NotationTLAPlus,
		Specification: "42",
	})
	s.NoError(err)
	s.Equal("_Constant", gf.Name)
	s.Nil(gf.Parameters)

	// Test that Validate is called (invalid name should fail).
	_, err = NewGlobalFunction(gfKey1, "Max", []string{"x"}, spec)
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")

	// Test that invalid specification fails.
	_, err = NewGlobalFunction(gfKey1, "_Max", []string{"x"}, Logic{
		Key:         identity.Key{},
		Type:        LogicTypeValue,
		Description: "Some desc.",
		Notation:    NotationTLAPlus,
	})
	s.Error(err)
	s.Contains(err.Error(), "KeyType")
}

// TestValidateWithParent tests that ValidateWithParent validates the key's parent relationship.
func (s *GlobalFunctionTestSuite) TestValidateWithParent() {
	gfKey := helper.Must(identity.NewGlobalFunctionKey("_max"))

	validSpec := Logic{
		Key:         gfKey,
		Type:        LogicTypeValue,
		Description: "Max of two values.",
		Notation:    NotationTLAPlus,
	}

	// Test valid case - gfunc key is root-level (nil parent).
	gf := GlobalFunction{
		Key:        gfKey,
		Name:       "_Max",
		Parameters: []string{"x", "y"},
		Logic:      validSpec,
	}
	err := gf.ValidateWithParent()
	s.NoError(err)

	// Test that Validate is called.
	gf = GlobalFunction{
		Key:        gfKey,
		Name:       "Max", // Invalid: no underscore
		Parameters: []string{"x", "y"},
		Logic:      validSpec,
	}
	err = gf.ValidateWithParent()
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")

	// Test that Specification.ValidateWithParent is called - invalid spec description should fail.
	gf = GlobalFunction{
		Key:        gfKey,
		Name:       "_Max",
		Parameters: []string{"x", "y"},
		Logic: Logic{
			Key:         gfKey,
			Type:        LogicTypeValue,
			Description: "", // Invalid: missing description
			Notation:    NotationTLAPlus,
		},
	}
	err = gf.ValidateWithParent()
	s.Error(err)
	s.Contains(err.Error(), "specification")
	s.Contains(err.Error(), "Description")
}
