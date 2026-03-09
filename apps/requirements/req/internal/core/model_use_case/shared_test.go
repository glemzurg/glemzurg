package model_use_case

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUseCaseSharedSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSharedSuite))
}

type UseCaseSharedSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for UseCaseShared.
func (suite *UseCaseSharedSuite) TestValidate() {
	tests := []struct {
		testName string
		obj      UseCaseShared
		errstr   string
	}{
		{
			testName: "valid include",
			obj: UseCaseShared{
				ShareType:  "include",
				UmlComment: "UmlComment",
			},
		},
		{
			testName: "valid extend",
			obj: UseCaseShared{
				ShareType:  "extend",
				UmlComment: "",
			},
		},
		{
			testName: "error empty share type",
			obj: UseCaseShared{
				ShareType:  "",
				UmlComment: "UmlComment",
			},
			errstr: "ShareType",
		},
		{
			testName: "error invalid share type",
			obj: UseCaseShared{
				ShareType:  "unknown",
				UmlComment: "UmlComment",
			},
			errstr: "ShareType",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.obj.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewUseCaseShared maps parameters correctly and calls Validate.
func (suite *UseCaseSharedSuite) TestNew() {
	// Test parameters are mapped correctly.

	obj := NewUseCaseShared("include", "UmlComment")
	suite.Equal(UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}, obj)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *UseCaseSharedSuite) TestValidateWithParent() {
	// Test that Validate is called.
	obj := UseCaseShared{
		ShareType:  "", // Invalid
		UmlComment: "UmlComment",
	}
	err := obj.ValidateWithParent()
	suite.Require().ErrorContains(err, "ShareType", "ValidateWithParent should call Validate()")

	// Test valid case.
	obj = UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}
	err = obj.ValidateWithParent()
	suite.Require().NoError(err)
}
