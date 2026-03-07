package model_use_case

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestActorSuite(t *testing.T) {
	suite.Run(t, new(ActorSuite))
}

type ActorSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Actor.
func (suite *ActorSuite) TestValidate() {
	tests := []struct {
		testName string
		obj      Actor
		errstr   string
	}{
		{
			testName: "valid actor with comment",
			obj: Actor{
				UmlComment: "UmlComment",
			},
		},
		{
			testName: "valid actor without comment",
			obj: Actor{
				UmlComment: "",
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.obj.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewActor maps parameters correctly and calls Validate.
func (suite *ActorSuite) TestNew() {
	tests := []struct {
		umlComment string
		obj        Actor
		errstr     string
	}{
		// OK.
		{
			umlComment: "UmlComment",
			obj: Actor{
				UmlComment: "UmlComment",
			},
		},
		{
			umlComment: "",
			obj: Actor{
				UmlComment: "",
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewActor(test.umlComment)
		if test.errstr == "" {
			suite.Require().NoError(err, testName)
			suite.Equal(test.obj, obj, testName)
		} else {
			suite.ErrorContains(err, test.errstr, testName)
			suite.Empty(obj, testName)
		}
	}
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *ActorSuite) TestValidateWithParent() {
	// Test valid case - Actor.Validate() always returns nil.
	obj := Actor{
		UmlComment: "UmlComment",
	}
	err := obj.ValidateWithParent()
	suite.Require().NoError(err)
}
