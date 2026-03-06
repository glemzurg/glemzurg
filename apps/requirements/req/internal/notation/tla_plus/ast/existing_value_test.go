package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(suite.T(), `@`, e.String())
}

func (suite *ExistingValueSuite) TestAscii() {
	e := &ExistingValue{}
	assert.Equal(suite.T(), `@`, e.Ascii())
}

func (suite *ExistingValueSuite) TestValidate() {
	e := &ExistingValue{}
	err := e.Validate()
	assert.NoError(suite.T(), err)
}

func (suite *ExistingValueSuite) TestExpressionNode() {
	// Verify that ExistingValue implements the expressionNode interface method.
	e := &ExistingValue{}
	// This should compile and not panic.
	e.expressionNode()
}
