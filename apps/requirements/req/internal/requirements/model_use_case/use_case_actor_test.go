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
