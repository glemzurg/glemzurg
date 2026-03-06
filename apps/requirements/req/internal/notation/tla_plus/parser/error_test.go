package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) TestParseEmptyInput() {
	_, err := ParseExpression("")
	s.Error(err)
}

func (s *ErrorTestSuite) TestParseUnmatchedParen() {
	_, err := ParseExpression("(42")
	s.Error(err)
}

func (s *ErrorTestSuite) TestParseUnterminatedString() {
	_, err := ParseExpression(`"hello`)
	s.Error(err)
}

func (s *ErrorTestSuite) TestParseInvalidToken() {
	_, err := ParseExpression("@#$")
	s.Error(err)
}
