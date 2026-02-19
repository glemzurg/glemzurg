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
				Key:           gfKey1,
				Name:          "_Max",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
		},
		{
			testName: "valid function with comment",
			gf: GlobalFunction{
				Key:           gfKey1,
				Name:          "_Max",
				Comment:       "Returns the maximum of two values.",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
		},
		{
			testName: "valid set definition no parameters",
			gf: GlobalFunction{
				Key:  gfKey2,
				Name: "_ValidStatuses",
				Specification: Logic{
					Key:           gfKey2,
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
				Specification: Logic{
					Key:           gfKey3,
					Description:   "A constant value.",
					Notation:      NotationTLAPlus,
					Specification: "42",
				},
			},
		},
		{
			testName: "error missing key",
			gf: GlobalFunction{
				Key:           identity.Key{},
				Name:          "_Max",
				Parameters:    []string{"x"},
				Specification: validSpec,
			},
			errstr: "KeyType",
		},
		{
			testName: "error wrong key type",
			gf: GlobalFunction{
				Key:           helper.Must(identity.NewInvariantKey("inv_1")),
				Name:          "_Max",
				Parameters:    []string{"x"},
				Specification: validSpec,
			},
			errstr: "invalid key type",
		},
		{
			testName: "error missing name",
			gf: GlobalFunction{
				Key:           gfKey1,
				Name:          "",
				Parameters:    []string{"x"},
				Specification: validSpec,
			},
			errstr: "Name",
		},
		{
			testName: "error name missing underscore",
			gf: GlobalFunction{
				Key:           gfKey1,
				Name:          "Max",
				Parameters:    []string{"x", "y"},
				Specification: validSpec,
			},
			errstr: "must start with underscore",
		},
		{
			testName: "error missing specification key",
			gf: GlobalFunction{
				Key:        gfKey1,
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
			testName: "error specification key mismatch",
			gf: GlobalFunction{
				Key:        gfKey1,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Specification: Logic{
					Key:         gfKey2, // Different key than gfKey1
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
				Specification: Logic{
					Key:         gfKey1,
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
				Specification: Logic{
					Key:         gfKey1,
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
				Specification: Logic{
					Key:         gfKey1,
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
	gfKey1 := helper.Must(identity.NewGlobalFunctionKey("_max"))
	gfKey2 := helper.Must(identity.NewGlobalFunctionKey("_constant"))

	spec := Logic{
		Key:           gfKey1,
		Description:   "Returns the maximum of two values.",
		Notation:      NotationTLAPlus,
		Specification: "IF x > y THEN x ELSE y",
	}

	// Test all parameters are mapped correctly.
	gf, err := NewGlobalFunction(gfKey1, "_Max", "Max function", []string{"x", "y"}, spec)
	s.NoError(err)
	s.Equal(GlobalFunction{
		Key:           gfKey1,
		Name:          "_Max",
		Comment:       "Max function",
		Parameters:    []string{"x", "y"},
		Specification: spec,
	}, gf)

	// Test with nil optional fields (Comment and Parameters are optional).
	gf, err = NewGlobalFunction(gfKey2, "_Constant", "", nil, Logic{
		Key:           gfKey2,
		Description:   "A constant.",
		Notation:      NotationTLAPlus,
		Specification: "42",
	})
	s.NoError(err)
	s.Equal("_Constant", gf.Name)
	s.Equal("", gf.Comment)
	s.Nil(gf.Parameters)

	// Test that Validate is called (invalid name should fail).
	_, err = NewGlobalFunction(gfKey1, "Max", "", []string{"x"}, spec)
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")

	// Test that invalid specification fails.
	_, err = NewGlobalFunction(gfKey1, "_Max", "", []string{"x"}, Logic{
		Key:         identity.Key{},
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
		Description: "Max of two values.",
		Notation:    NotationTLAPlus,
	}

	// Test valid case - gfunc key is root-level (nil parent).
	gf := GlobalFunction{
		Key:           gfKey,
		Name:          "_Max",
		Parameters:    []string{"x", "y"},
		Specification: validSpec,
	}
	err := gf.ValidateWithParent()
	s.NoError(err)

	// Test that Validate is called.
	gf = GlobalFunction{
		Key:           gfKey,
		Name:          "Max", // Invalid: no underscore
		Parameters:    []string{"x", "y"},
		Specification: validSpec,
	}
	err = gf.ValidateWithParent()
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")
}
