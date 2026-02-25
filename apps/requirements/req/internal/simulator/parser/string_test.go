package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

type StringTestSuite struct {
	suite.Suite
}

func TestStringSuite(t *testing.T) {
	suite.Run(t, new(StringTestSuite))
}

func (s *StringTestSuite) TestParseSimple() {
	expr, err := ParseExpression(`"hello"`)
	s.NoError(err)
	s.Equal(&ast.StringLiteral{Value: "hello"}, expr)
}

func (s *StringTestSuite) TestParseWithSpaces() {
	expr, err := ParseExpression(`"hello world"`)
	s.NoError(err)
	s.Equal(&ast.StringLiteral{Value: "hello world"}, expr)
}

func (s *StringTestSuite) TestParseEmpty() {
	expr, err := ParseExpression(`""`)
	s.NoError(err)
	s.Equal(&ast.StringLiteral{Value: ""}, expr)
}

func (s *StringTestSuite) TestParseWithEscapes() {
	expr, err := ParseExpression(`"line1\nline2\ttab\r\f"`)
	s.NoError(err)
	s.Equal(&ast.StringLiteral{Value: "line1\nline2\ttab\r\f"}, expr)
}
