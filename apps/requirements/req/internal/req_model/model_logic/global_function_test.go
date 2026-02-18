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
	specKey1 := helper.Must(identity.NewInvariantKey("spec_1"))
	specKey2 := helper.Must(identity.NewInvariantKey("spec_2"))
	specKey3 := helper.Must(identity.NewInvariantKey("spec_3"))

	validSpec := Logic{
		Key:         specKey1,
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
				Name:          "_Max",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
		},
		{
			testName: "valid function with comment",
			gf: GlobalFunction{
				Name:          "_Max",
				Comment:       "Returns the maximum of two values.",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
		},
		{
			testName: "valid set definition no parameters",
			gf: GlobalFunction{
				Name: "_ValidStatuses",
				Specification: Logic{
					Key:           specKey2,
					Description:   "Set of valid statuses.",
					Notation:      NotationTLAPlus,
					Specification: `{"pending", "active", "complete"}`,
				},
			},
		},
		{
			testName: "valid predicate with nil parameters",
			gf: GlobalFunction{
				Name:       "_Constant",
				Parameters: nil,
				Specification: Logic{
					Key:           specKey3,
					Description:   "A constant value.",
					Notation:      NotationTLAPlus,
					Specification: "42",
				},
			},
		},
		{
			testName: "error missing name",
			gf: GlobalFunction{
				Name:          "",
				Parameters:    []string{"x"},
				Specification: validSpec,
			},
			errstr: "Name",
		},
		{
			testName: "error name missing underscore",
			gf: GlobalFunction{
				Name:          "Max",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
			errstr: "must start with underscore",
		},
		{
			testName: "error missing specification key",
			gf: GlobalFunction{
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: Logic{
					Key:         identity.Key{},
					Description: "Some desc.",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "KeyType",
		},
		{
			testName: "error missing specification description",
			gf: GlobalFunction{
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: Logic{
					Key:         specKey1,
					Description: "",
					Notation:    NotationTLAPlus,
				},
			},
			errstr: "Description",
		},
		{
			testName: "error missing specification notation",
			gf: GlobalFunction{
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: Logic{
					Key:         specKey1,
					Description: "Some desc.",
					Notation:    "",
				},
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid specification notation",
			gf: GlobalFunction{
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: Logic{
					Key:         specKey1,
					Description: "Some desc.",
					Notation:    "Z",
				},
			},
			errstr: "Notation",
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
	specKey1 := helper.Must(identity.NewInvariantKey("spec_1"))
	specKey2 := helper.Must(identity.NewInvariantKey("spec_2"))

	spec := Logic{
		Key:           specKey1,
		Description:   "Returns the maximum of two values.",
		Notation:      NotationTLAPlus,
		Specification: "IF x > y THEN x ELSE y",
	}

	// Test all parameters are mapped correctly.
	gf, err := NewGlobalFunction("_Max", "Max function", []string{"x", "y"}, spec)
	s.NoError(err)
	s.Equal(GlobalFunction{
		Name:          "_Max",
		Comment:       "Max function",
		Parameters:    []string{"x", "y"},
		Specification: spec,
	}, gf)

	// Test with nil optional fields (Comment and Parameters are optional).
	gf, err = NewGlobalFunction("_Constant", "", nil, Logic{
		Key:           specKey2,
		Description:   "A constant.",
		Notation:      NotationTLAPlus,
		Specification: "42",
	})
	s.NoError(err)
	s.Equal("_Constant", gf.Name)
	s.Equal("", gf.Comment)
	s.Nil(gf.Parameters)

	// Test that Validate is called (invalid name should fail).
	_, err = NewGlobalFunction("Max", "", []string{"x"}, spec)
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")

	// Test that invalid specification fails.
	_, err = NewGlobalFunction("_Max", "", []string{"x"}, Logic{
		Key:         identity.Key{},
		Description: "Some desc.",
		Notation:    NotationTLAPlus,
	})
	s.Error(err)
	s.Contains(err.Error(), "KeyType")
}

// TestValidateWithParent tests that ValidateWithParent validates the specification logic's parent relationship.
func (s *GlobalFunctionTestSuite) TestValidateWithParent() {
	specKey := helper.Must(identity.NewInvariantKey("spec_1"))

	validSpec := Logic{
		Key:         specKey,
		Description: "Max of two values.",
		Notation:    NotationTLAPlus,
	}

	// Test valid case - spec key is root-level (nil parent).
	gf := GlobalFunction{
		Name:          "_Max",
		Parameters:    []string{"x", "y"},
		Specification: validSpec,
	}
	err := gf.ValidateWithParent()
	s.NoError(err)

	// Test that Validate is called.
	gf = GlobalFunction{
		Name:          "Max", // Invalid: no underscore
		Parameters:    []string{"x", "y"},
		Specification: validSpec,
	}
	err = gf.ValidateWithParent()
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")

	// Test that spec key parent validation is called - invariant key should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	wrongSpecKey := helper.Must(identity.NewActionRequireKey(actionKey, "req_1"))

	gf = GlobalFunction{
		Name:       "_Max",
		Parameters: []string{"x", "y"},
		Specification: Logic{
			Key:         wrongSpecKey,
			Description: "Max of two values.",
			Notation:    NotationTLAPlus,
		},
	}
	err = gf.ValidateWithParent()
	s.Error(err)
	s.Contains(err.Error(), "specification")
}
