package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

func TestControlFlowSuite(t *testing.T) {
	suite.Run(t, new(ControlFlowSuite))
}

type ControlFlowSuite struct {
	suite.Suite
}

// =============================================================================
// IF-THEN-ELSE
// =============================================================================

func (s *ControlFlowSuite) TestIfThenElse_Simple() {
	expr, err := ParseExpression("IF x > 0 THEN x ELSE -x")
	s.Require().NoError(err)

	ite, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected *ast.IfThenElse, got %T", expr)

	// Condition should be a comparison
	_, ok = ite.Condition.(*ast.BinaryComparison)
	s.True(ok, "condition should be *ast.BinaryComparison, got %T", ite.Condition)

	// Then should be identifier
	thenIdent, ok := ite.Then.(*ast.Identifier)
	s.True(ok, "then should be *ast.Identifier, got %T", ite.Then)
	s.Equal("x", thenIdent.Value)
}

func (s *ControlFlowSuite) TestIfThenElse_Boolean() {
	expr, err := ParseExpression("IF flag THEN TRUE ELSE FALSE")
	s.Require().NoError(err)

	ite, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected *ast.IfThenElse, got %T", expr)

	_, ok = ite.Then.(*ast.BooleanLiteral)
	s.True(ok, "then should be *ast.BooleanLiteral, got %T", ite.Then)

	_, ok = ite.Else.(*ast.BooleanLiteral)
	s.True(ok, "else should be *ast.BooleanLiteral, got %T", ite.Else)
}

func (s *ControlFlowSuite) TestIfThenElse_Nested() {
	expr, err := ParseExpression("IF a THEN IF b THEN 1 ELSE 2 ELSE 3")
	s.Require().NoError(err)

	outer, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected outer *ast.IfThenElse, got %T", expr)

	inner, ok := outer.Then.(*ast.IfThenElse)
	s.True(ok, "then should be nested *ast.IfThenElse, got %T", outer.Then)
	s.NotNil(inner)
}

func (s *ControlFlowSuite) TestIfThenElse_WithArithmetic() {
	expr, err := ParseExpression("IF n > 0 THEN n * 2 ELSE n + 1")
	s.Require().NoError(err)

	ite, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected *ast.IfThenElse, got %T", expr)

	_, ok = ite.Then.(*ast.BinaryArithmetic)
	s.True(ok, "then should be *ast.BinaryArithmetic, got %T", ite.Then)

	_, ok = ite.Else.(*ast.BinaryArithmetic)
	s.True(ok, "else should be *ast.BinaryArithmetic, got %T", ite.Else)
}

func (s *ControlFlowSuite) TestIfThenElse_String() {
	expr, err := ParseExpression("IF x > 0 THEN 1 ELSE 0")
	s.Require().NoError(err)

	ite, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected *ast.IfThenElse, got %T", expr)
	s.Equal("IF x > 0 THEN 1 ELSE 0", ite.String())
	s.Equal("IF x > 0 THEN 1 ELSE 0", ite.ASCII())
}

// =============================================================================
// CASE Expressions
// =============================================================================

func (s *ControlFlowSuite) TestCaseExpr_SingleBranch() {
	expr, err := ParseExpression("CASE x > 0 -> 1")
	s.Require().NoError(err)

	caseExpr, ok := expr.(*ast.CaseExpr)
	s.True(ok, "expected *ast.CaseExpr, got %T", expr)
	s.Len(caseExpr.Branches, 1)
	s.Nil(caseExpr.Other)
}

func (s *ControlFlowSuite) TestCaseExpr_MultipleBranches() {
	expr, err := ParseExpression("CASE x > 0 -> 1 [] x < 0 -> 2 [] x = 0 -> 0")
	s.Require().NoError(err)

	caseExpr, ok := expr.(*ast.CaseExpr)
	s.True(ok, "expected *ast.CaseExpr, got %T", expr)
	s.Len(caseExpr.Branches, 3)
	s.Nil(caseExpr.Other)
}

func (s *ControlFlowSuite) TestCaseExpr_WithOther() {
	expr, err := ParseExpression("CASE x > 0 -> 1 [] OTHER -> 0")
	s.Require().NoError(err)

	caseExpr, ok := expr.(*ast.CaseExpr)
	s.True(ok, "expected *ast.CaseExpr, got %T", expr)
	s.Len(caseExpr.Branches, 1)
	s.NotNil(caseExpr.Other)
}

func (s *ControlFlowSuite) TestCaseExpr_UnicodeArrow() {
	expr, err := ParseExpression("CASE x > 0 → 1 □ OTHER → 0")
	s.Require().NoError(err)

	caseExpr, ok := expr.(*ast.CaseExpr)
	s.True(ok, "expected *ast.CaseExpr, got %T", expr)
	s.Len(caseExpr.Branches, 1)
	s.NotNil(caseExpr.Other)
}

func (s *ControlFlowSuite) TestCaseExpr_WithExpressions() {
	expr, err := ParseExpression("CASE n >= 0 -> n * 2 [] n < 0 -> -n")
	s.Require().NoError(err)

	caseExpr, ok := expr.(*ast.CaseExpr)
	s.True(ok, "expected *ast.CaseExpr, got %T", expr)
	s.Len(caseExpr.Branches, 2)

	// Check first branch result is multiplication
	_, ok = caseExpr.Branches[0].Result.(*ast.BinaryArithmetic)
	s.True(ok, "first branch result should be *ast.BinaryArithmetic")
}

// =============================================================================
// Function Calls
// =============================================================================

func (s *ControlFlowSuite) TestFunctionCall_NoArgs() {
	expr, err := ParseExpression("Func()")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Empty(call.ScopePath)
	s.Equal("Func", call.Name.Value)
	s.Empty(call.Args)
}

func (s *ControlFlowSuite) TestFunctionCall_SingleArg() {
	expr, err := ParseExpression("Len(seq)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Empty(call.ScopePath)
	s.Equal("Len", call.Name.Value)
	s.Len(call.Args, 1)
}

func (s *ControlFlowSuite) TestFunctionCall_MultipleArgs() {
	expr, err := ParseExpression("SubSeq(seq, 1, 5)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Equal("SubSeq", call.Name.Value)
	s.Len(call.Args, 3)
}

func (s *ControlFlowSuite) TestFunctionCall_WithModule() {
	expr, err := ParseExpression("_Seq!Len(seq)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Len(call.ScopePath, 1)
	s.Equal("_Seq", call.ScopePath[0].Value)
	s.Equal("Len", call.Name.Value)
	s.Len(call.Args, 1)
	s.True(call.IsGlobalOrBuiltin())
}

func (s *ControlFlowSuite) TestFunctionCall_MultiLevelScope() {
	expr, err := ParseExpression("Domain!Subdomain!Class!Action(x, y)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Len(call.ScopePath, 3)
	s.Equal("Domain", call.ScopePath[0].Value)
	s.Equal("Subdomain", call.ScopePath[1].Value)
	s.Equal("Class", call.ScopePath[2].Value)
	s.Equal("Action", call.Name.Value)
	s.Len(call.Args, 2)
	s.False(call.IsGlobalOrBuiltin())
	s.Equal("Domain!Subdomain!Class!Action", call.FullName())
}

func (s *ControlFlowSuite) TestFunctionCall_TwoLevelScope() {
	// Class!Action pattern (from subdomain scope)
	expr, err := ParseExpression("Class!Action()")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Len(call.ScopePath, 1)
	s.Equal("Class", call.ScopePath[0].Value)
	s.Equal("Action", call.Name.Value)
	s.Equal("Class!Action", call.FullName())
}

func (s *ControlFlowSuite) TestFunctionCall_BuiltinModule() {
	expr, err := ParseExpression("_Bags!SetToBag(s)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Len(call.ScopePath, 1)
	s.Equal("_Bags", call.ScopePath[0].Value)
	s.Equal("SetToBag", call.Name.Value)
	s.True(call.IsGlobalOrBuiltin())
	s.Equal("_Bags!SetToBag", call.FullName())
}

func (s *ControlFlowSuite) TestFunctionCall_String() {
	expr, err := ParseExpression("_Seq!Len(seq)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Equal("_Seq!Len(seq)", call.String())
	s.Equal("_Seq!Len(seq)", call.ASCII())
}

func (s *ControlFlowSuite) TestFunctionCall_Cardinality() {
	expr, err := ParseExpression("Cardinality({1, 2, 3})")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Equal("Cardinality", call.Name.Value)
	s.Len(call.Args, 1)

	// Argument should be a set literal
	_, ok = call.Args[0].(*ast.SetLiteral)
	s.True(ok, "argument should be *ast.SetLiteral, got %T", call.Args[0])
}

func (s *ControlFlowSuite) TestFunctionCall_WithExpressionArgs() {
	expr, err := ParseExpression("Max(a + b, c * d)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Equal("Max", call.Name.Value)
	s.Len(call.Args, 2)

	// Both args should be arithmetic expressions
	_, ok = call.Args[0].(*ast.BinaryArithmetic)
	s.True(ok, "first arg should be *ast.BinaryArithmetic")
	_, ok = call.Args[1].(*ast.BinaryArithmetic)
	s.True(ok, "second arg should be *ast.BinaryArithmetic")
}

func (s *ControlFlowSuite) TestFunctionCall_Nested() {
	expr, err := ParseExpression("Len(Tail(seq))")
	s.Require().NoError(err)

	outer, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected outer *ast.FunctionCall, got %T", expr)
	s.Equal("Len", outer.Name.Value)
	s.Len(outer.Args, 1)

	inner, ok := outer.Args[0].(*ast.FunctionCall)
	s.True(ok, "inner arg should be *ast.FunctionCall, got %T", outer.Args[0])
	s.Equal("Tail", inner.Name.Value)
}

func (s *ControlFlowSuite) TestFunctionCall_String_NoScope() {
	expr, err := ParseExpression("Len(seq)")
	s.Require().NoError(err)

	call, ok := expr.(*ast.FunctionCall)
	s.True(ok, "expected *ast.FunctionCall, got %T", expr)
	s.Equal("Len(seq)", call.String())
	s.Equal("Len", call.FullName())
}

// =============================================================================
// Combined Tests
// =============================================================================

func (s *ControlFlowSuite) TestCombined_IfWithFunctionCall() {
	expr, err := ParseExpression("IF Len(seq) > 0 THEN Head(seq) ELSE 0")
	s.Require().NoError(err)

	ite, ok := expr.(*ast.IfThenElse)
	s.True(ok, "expected *ast.IfThenElse, got %T", expr)

	// Then should be a function call
	_, ok = ite.Then.(*ast.FunctionCall)
	s.True(ok, "then should be *ast.FunctionCall, got %T", ite.Then)
}

func (s *ControlFlowSuite) TestCombined_FunctionCallInSet() {
	expr, err := ParseExpression("{Len(a), Len(b), Len(c)}")
	s.Require().NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 3)

	// Each element should be a function call
	for i, elem := range set.Elements {
		_, ok := elem.(*ast.FunctionCall)
		s.True(ok, "element %d should be *ast.FunctionCall, got %T", i, elem)
	}
}
