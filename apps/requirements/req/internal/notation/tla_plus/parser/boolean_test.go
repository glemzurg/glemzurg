package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

type BooleanTestSuite struct {
	suite.Suite
}

func TestBooleanSuite(t *testing.T) {
	suite.Run(t, new(BooleanTestSuite))
}

func (s *BooleanTestSuite) TestParseTrue() {
	expr, err := ParseExpression("TRUE")
	s.NoError(err)
	s.Equal(&ast.BooleanLiteral{Value: true}, expr)
}

func (s *BooleanTestSuite) TestParseFalse() {
	expr, err := ParseExpression("FALSE")
	s.NoError(err)
	s.Equal(&ast.BooleanLiteral{Value: false}, expr)
}

func (s *BooleanTestSuite) TestParseWithWhitespace() {
	expr, err := ParseExpression("  TRUE  ")
	s.NoError(err)
	s.Equal(&ast.BooleanLiteral{Value: true}, expr)
}
