package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
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
		{
			testName: "valid parameter with data type key matching name",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            "amount",
					CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
				},
			},
		},
		{
			testName: "error data type key does not match parameter name",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            "different",
					CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
				},
			},
			errstr: "DataType.Key",
		},
		{
			testName: "error data type key empty when parameter has name",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            "",
					CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
				},
			},
			errstr: "DataType.Key",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.param.Validate(ctx)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewParameter maps parameters correctly.
func (suite *ParameterSuite) TestNew() {
	// Test parameters are mapped correctly.
	param, err := NewParameter("amount", "Nat")
	suite.Require().NoError(err)
	suite.Equal("amount", param.Name)
	suite.Equal("Nat", param.DataTypeRules)
}

// TestNewSetsDataTypeKeyToName verifies that NewParameter sets the parsed
// DataType.Key to match the Parameter.Name. This invariant is what allows
// the parameter and its data type to share a key in the database.
func (suite *ParameterSuite) TestNewSetsDataTypeKeyToName() {
	param, err := NewParameter("amount", "unconstrained")
	suite.Require().NoError(err)
	suite.Require().NotNil(param.DataType, "NewParameter should parse DataTypeRules into a DataType")
	suite.Equal(param.Name, param.DataType.Key, "DataType.Key must equal Parameter.Name")

	// Also confirm Validate accepts the constructed parameter.
	ctx := coreerr.NewContext("test", "")
	suite.Require().NoError(param.Validate(ctx))
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *ParameterSuite) TestValidateWithParent() {
	ctx := coreerr.NewContext("test", "")
	// Test that Validate is called.
	param := Parameter{
		Name:          "",
		DataTypeRules: "Nat",
	}
	err := param.ValidateWithParent(ctx)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test valid case.
	param = Parameter{
		Name:          "amount",
		DataTypeRules: "Nat",
	}
	err = param.ValidateWithParent(ctx)
	suite.Require().NoError(err)
}
