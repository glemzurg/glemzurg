package model_state

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.param.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewParameter maps parameters correctly and calls Validate.
func (suite *ParameterSuite) TestNew() {
	// Test parameters are mapped correctly.
	param, err := NewParameter("amount", "Nat")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "amount", param.Name)
	assert.Equal(suite.T(), "Nat", param.DataTypeRules)
	// DataType may or may not be set depending on whether the parser is available

	// Test that Validate is called (invalid data should fail).
	_, err = NewParameter("", "Nat")
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *ParameterSuite) TestValidateWithParent() {
	// Test that Validate is called.
	param := Parameter{
		Name:          "",
		DataTypeRules: "Nat",
	}
	err := param.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test valid case.
	param = Parameter{
		Name:          "amount",
		DataTypeRules: "Nat",
	}
	err = param.ValidateWithParent()
	assert.NoError(suite.T(), err)
}

// TestSetSortOrder tests that setSortOrder assigns the slice index to each parameter's SortOrder.
func (suite *ParameterSuite) TestSetSortOrder() {
	// Empty slice.
	params := []Parameter{}
	setSortOrder(params)
	assert.Empty(suite.T(), params)

	// Single parameter.
	params = []Parameter{
		{Name: "a", DataTypeRules: "Nat"},
	}
	setSortOrder(params)
	assert.Equal(suite.T(), 0, params[0].SortOrder)

	// Multiple parameters get ascending sort order.
	params = []Parameter{
		{Name: "a", DataTypeRules: "Nat"},
		{Name: "b", DataTypeRules: "Int"},
		{Name: "c", DataTypeRules: "Bool"},
	}
	setSortOrder(params)
	assert.Equal(suite.T(), 0, params[0].SortOrder)
	assert.Equal(suite.T(), 1, params[1].SortOrder)
	assert.Equal(suite.T(), 2, params[2].SortOrder)

	// Overwrites any existing SortOrder values.
	params = []Parameter{
		{Name: "a", SortOrder: 99, DataTypeRules: "Nat"},
		{Name: "b", SortOrder: 42, DataTypeRules: "Int"},
	}
	setSortOrder(params)
	assert.Equal(suite.T(), 0, params[0].SortOrder)
	assert.Equal(suite.T(), 1, params[1].SortOrder)
}
