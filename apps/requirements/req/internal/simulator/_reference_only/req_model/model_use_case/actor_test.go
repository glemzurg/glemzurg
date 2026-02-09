package model_use_case

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
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
	assert.NoError(suite.T(), err)
}
