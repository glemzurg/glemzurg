package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestExistingValueSuite(t *testing.T) {
	suite.Run(t, new(ExistingValueSuite))
}

type ExistingValueSuite struct {
	suite.Suite
}

func (suite *ExistingValueSuite) TestString() {
	e := &ExistingValue{}
	suite.Equal(`@`, e.String())
}

func (suite *ExistingValueSuite) TestASCII() {
	e := &ExistingValue{}
	suite.Equal(`@`, e.ASCII())
}

func (suite *ExistingValueSuite) TestValidate() {
	e := &ExistingValue{}
	err := e.Validate()
	suite.Require().NoError(err)
}

func (suite *ExistingValueSuite) TestExpressionNode() {
	// Verify that ExistingValue implements the expressionNode interface method.
	e := &ExistingValue{}
	// This should compile and not panic.
	e.expressionNode()
}
