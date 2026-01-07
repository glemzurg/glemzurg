package model_use_case

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseActorSuite(t *testing.T) {
	suite.Run(t, new(UseCaseActorSuite))
}

type UseCaseActorSuite struct {
	suite.Suite
}

func (suite *UseCaseActorSuite) TestNew() {
	tests := []struct {
		umlComment string
		obj        UseCaseActor
		errstr     string
	}{
		// OK.
		{
			umlComment: "UmlComment",
			obj: UseCaseActor{
				UmlComment: "UmlComment",
			},
		},
		{
			umlComment: "",
			obj: UseCaseActor{
				UmlComment: "",
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewUseCaseActor(test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
