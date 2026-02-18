package model_use_case

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.obj.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewUseCaseShared maps parameters correctly and calls Validate.
func (suite *UseCaseSharedSuite) TestNew() {
	// Test parameters are mapped correctly.
	obj, err := NewUseCaseShared("include", "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}, obj)

	// Test that Validate is called (invalid data should fail).
	_, err = NewUseCaseShared("", "UmlComment")
	assert.ErrorContains(suite.T(), err, "ShareType")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *UseCaseSharedSuite) TestValidateWithParent() {
	// Test that Validate is called.
	obj := UseCaseShared{
		ShareType:  "", // Invalid
		UmlComment: "UmlComment",
	}
	err := obj.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "ShareType", "ValidateWithParent should call Validate()")

	// Test valid case.
	obj = UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}
	err = obj.ValidateWithParent()
	assert.NoError(suite.T(), err)
}
