package model_bridge

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/stretchr/testify/suite"
)

type ContainsPrimedSuite struct {
	suite.Suite
}

func TestContainsPrimedSuite(t *testing.T) {
	suite.Run(t, new(ContainsPrimedSuite))
}

func (s *ContainsPrimedSuite) TestNilExpression() {
	s.False(ContainsAnyPrimed(nil))
}

func (s *ContainsPrimedSuite) TestSimpleIdentifier() {
	expr, err := parser.ParseExpression("x")
	s.Require().NoError(err)
	s.False(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestSimpleNumber() {
	expr, err := parser.ParseExpression("42")
	s.Require().NoError(err)
	s.False(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestPrimedVariable() {
	expr, err := parser.ParseExpression("x'")
	s.Require().NoError(err)
	s.True(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestPrimedFieldAccess() {
	expr, err := parser.ParseExpression("self.count'")
	s.Require().NoError(err)
	s.True(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestBinaryArithmeticNoPrime() {
	expr, err := parser.ParseExpression("x + y")
	s.Require().NoError(err)
	s.False(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestBinaryArithmeticWithPrime() {
	expr, err := parser.ParseExpression("x' + y")
	s.Require().NoError(err)
	s.True(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestComparisonWithPrime() {
	expr, err := parser.ParseExpression("x' > 0")
	s.Require().NoError(err)
	s.True(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestComparisonNoPrime() {
	expr, err := parser.ParseExpression("x > 0")
	s.Require().NoError(err)
	s.False(ContainsAnyPrimed(expr))
}

func (s *ContainsPrimedSuite) TestNestedDeepPrime() {
	expr, err := parser.ParseExpression("(a + b) > (c' + d)")
	s.Require().NoError(err)
	s.True(ContainsAnyPrimed(expr))
}
