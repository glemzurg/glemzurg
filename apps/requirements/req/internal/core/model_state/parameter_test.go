package model_state

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestParameterSuite(t *testing.T) {
	suite.Run(t, new(ParameterSuite))
}

type ParameterSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Parameter.
func (suite *ParameterSuite) TestValidate() {
	tests := []struct {
		testName string
		param    Parameter
		errstr   string
	}{
		{
			testName: "valid parameter",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "Nat",
			},
		},
		{
			testName: "error blank name",
			param: Parameter{
				Name:          "",
				DataTypeRules: "Nat",
			},
			errstr: "Name",
		},
		{
			testName: "error blank data type rules",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "",
			},
			errstr: "DataTypeRules",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.param.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewParameter maps parameters correctly and calls Validate.
func (suite *ParameterSuite) TestNew() {
	// Test parameters are mapped correctly.
	param, err := NewParameter("amount", "Nat")
	suite.Require().NoError(err)
	suite.Equal("amount", param.Name)
	suite.Equal("Nat", param.DataTypeRules)
	// DataType may or may not be set depending on whether the parser is available

	// Test that Validate is called (invalid data should fail).
	_, err = NewParameter("", "Nat")
	suite.ErrorContains(err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *ParameterSuite) TestValidateWithParent() {
	// Test that Validate is called.
	param := Parameter{
		Name:          "",
		DataTypeRules: "Nat",
	}
	err := param.ValidateWithParent()
	suite.ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test valid case.
	param = Parameter{
		Name:          "amount",
		DataTypeRules: "Nat",
	}
	err = param.ValidateWithParent()
	suite.Require().NoError(err)
}
