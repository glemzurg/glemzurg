package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestParameterSuite(t *testing.T) {
	suite.Run(t, new(ParameterSuite))
}

type ParameterSuite struct {
	suite.Suite
}

// testActionKey returns a deterministic action key used as the parent of test parameters.
func testActionKey() identity.Key {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	return helper.Must(identity.NewActionKey(classKey, "action1"))
}

// testEventKey returns a deterministic event key (different KEY_TYPE) for parent-mismatch cases.
func testEventKey() identity.Key {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	return helper.Must(identity.NewEventKey(classKey, "event1"))
}

func testParamKey(name string) identity.Key {
	return helper.Must(identity.NewParameterKey(testActionKey(), name))
}

// TestValidate tests all validation rules for Parameter.
func (suite *ParameterSuite) TestValidate() {
	validKey := testParamKey("amount")
	otherParamKey := helper.Must(identity.NewParameterKey(testActionKey(), "other"))
	validDtKey := helper.Must(identity.NewDataTypeKey(validKey, ""))
	mismatchedDtKey := helper.Must(identity.NewDataTypeKey(otherParamKey, ""))

	tests := []struct {
		testName string
		param    Parameter
		errstr   string
	}{
		{
			testName: "valid parameter",
			param: Parameter{
				Key:           validKey,
				Name:          "amount",
				DataTypeRules: "Nat",
			},
		},
		{
			testName: "error missing key",
			param: Parameter{
				Name:          "amount",
				DataTypeRules: "Nat",
			},
			errstr: "Key",
		},
		{
			testName: "error wrong key type",
			param: Parameter{
				Key:           testActionKey(), // not KEY_TYPE_PARAMETER
				Name:          "amount",
				DataTypeRules: "Nat",
			},
			errstr: "invalid key type",
		},
		{
			testName: "error blank name",
			param: Parameter{
				Key:           validKey,
				Name:          "",
				DataTypeRules: "Nat",
			},
			errstr: "Name",
		},
		{
			testName: "error blank data type rules",
			param: Parameter{
				Key:           validKey,
				Name:          "amount",
				DataTypeRules: "",
			},
			errstr: "DataTypeRules",
		},
		{
			testName: "valid parameter with data type key parented by the parameter key",
			param: Parameter{
				Key:           validKey,
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            validDtKey,
					CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
				},
			},
		},
		{
			testName: "error data type key parent is not the parameter key",
			param: Parameter{
				Key:           validKey,
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            mismatchedDtKey,
					CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
				},
			},
			errstr: "DataType.Key",
		},
		{
			testName: "error data type key has wrong KeyType",
			param: Parameter{
				Key:           validKey,
				Name:          "amount",
				DataTypeRules: "Nat",
				DataType: &model_data_type.DataType{
					Key:            otherParamKey, // KEY_TYPE_PARAMETER, not KEY_TYPE_DATA_TYPE
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

// TestNew tests that NewParameter maps parameters correctly and sets the identity.Key.
func (suite *ParameterSuite) TestNew() {
	actionKey := testActionKey()
	param, err := NewParameter(actionKey, "amount", "unconstrained", false)
	suite.Require().NoError(err)
	suite.Equal("amount", param.Name)
	suite.Equal("unconstrained", param.DataTypeRules)
	suite.False(param.Nullable)
	suite.Equal(identity.KEY_TYPE_PARAMETER, param.Key.KeyType)
	suite.Equal("amount", param.Key.SubKey)
	suite.Equal(actionKey.String(), param.Key.ParentKey)
}

func (suite *ParameterSuite) TestNewStoresNullable() {
	actionKey := testActionKey()
	param, err := NewParameter(actionKey, "country_code", "unconstrained", true)
	suite.Require().NoError(err)
	suite.True(param.Nullable)
}

// TestNewRejectsBadParent: NewParameter requires its parent to be action or query.
func (suite *ParameterSuite) TestNewRejectsBadParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	_, err := NewParameter(classKey, "amount", "unconstrained", false)
	suite.Require().Error(err)
	suite.ErrorContains(err, "parent key cannot be of type 'class' for 'parameter' key")
}

// TestNewSetsDataTypeKey verifies that NewParameter sets the parsed DataType.Key to
// the canonical identity.Key string parented by the Parameter's own identity.Key.
func (suite *ParameterSuite) TestNewSetsDataTypeKey() {
	actionKey := testActionKey()
	param, err := NewParameter(actionKey, "amount", "unconstrained", false)
	suite.Require().NoError(err)
	suite.Require().NotNil(param.DataType, "NewParameter should parse DataTypeRules into a DataType")

	expectedKey := helper.Must(identity.NewDataTypeKey(param.Key, ""))
	suite.Equal(expectedKey, param.DataType.Key, "DataType.Key must be the typed identity.Key parented by Parameter.Key")

	ctx := coreerr.NewContext("test", "")
	suite.Require().NoError(param.Validate(ctx))
}

// TestValidateWithParent tests that ValidateWithParent checks both the base validation
// and the parent relationship of the key.
func (suite *ParameterSuite) TestValidateWithParent() {
	ctx := coreerr.NewContext("test", "")
	actionKey := testActionKey()
	param, err := NewParameter(actionKey, "amount", "unconstrained", false)
	suite.Require().NoError(err)

	// Correct parent.
	suite.Require().NoError(param.ValidateWithParent(ctx, &actionKey))

	// Wrong parent.
	eventKey := testEventKey()
	suite.Require().ErrorContains(param.ValidateWithParent(ctx, &eventKey), "does not match expected parent")

	// Underlying Validate failure still surfaces.
	badParam := Parameter{Name: "amount", DataTypeRules: "Nat"} // missing key
	suite.Require().ErrorContains(badParam.ValidateWithParent(ctx, &actionKey), "Key")
}
