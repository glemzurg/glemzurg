package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/stretchr/testify/suite"
)

func TestTupleRecordSuite(t *testing.T) {
	suite.Run(t, new(TupleRecordSuite))
}

type TupleRecordSuite struct {
	suite.Suite
}

// =============================================================================
// Tuple Literals
// =============================================================================

func (s *TupleRecordSuite) TestTupleLiteral_Empty_ASCII() {
	expr, err := ParseExpression("<<>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(0, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_Empty_Unicode() {
	expr, err := ParseExpression("⟨⟩")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(0, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_Single() {
	expr, err := ParseExpression("<<1>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(1, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_Multiple_ASCII() {
	expr, err := ParseExpression("<<1, 2, 3>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(3, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_Multiple_Unicode() {
	expr, err := ParseExpression("⟨1, 2, 3⟩")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(3, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_WithVariables() {
	expr, err := ParseExpression("<<a, b, c>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(3, len(tuple.Elements))

	// Check that elements are identifiers
	for i, elem := range tuple.Elements {
		_, ok := elem.(*ast.Identifier)
		s.True(ok, "element %d should be *ast.Identifier, got %T", i, elem)
	}
}

func (s *TupleRecordSuite) TestTupleLiteral_WithExpressions() {
	expr, err := ParseExpression("<<1 + 2, x * y, TRUE>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(3, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_Nested() {
	expr, err := ParseExpression("<<1, <<2, 3>>, 4>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(3, len(tuple.Elements))

	// Second element should be a tuple
	innerTuple, ok := tuple.Elements[1].(*ast.TupleLiteral)
	s.True(ok, "inner element should be *ast.TupleLiteral, got %T", tuple.Elements[1])
	s.Equal(2, len(innerTuple.Elements))
}

func (s *TupleRecordSuite) TestTupleLiteral_String() {
	expr, err := ParseExpression("<<1, 2>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal("⟨1, 2⟩", tuple.String())
	s.Equal("<<1, 2>>", tuple.Ascii())
}

// =============================================================================
// Tuple Indexing
// =============================================================================

func (s *TupleRecordSuite) TestTupleIndex_Simple() {
	expr, err := ParseExpression("tuple[1]")
	s.NoError(err)

	idx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected *ast.TupleIndex, got %T", expr)

	ident, ok := idx.Tuple.(*ast.Identifier)
	s.True(ok, "tuple should be *ast.Identifier, got %T", idx.Tuple)
	s.Equal("tuple", ident.Value)
}

func (s *TupleRecordSuite) TestTupleIndex_Literal() {
	expr, err := ParseExpression("<<1, 2, 3>>[2]")
	s.NoError(err)

	idx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected *ast.TupleIndex, got %T", expr)

	_, ok = idx.Tuple.(*ast.TupleLiteral)
	s.True(ok, "tuple should be *ast.TupleLiteral, got %T", idx.Tuple)
}

func (s *TupleRecordSuite) TestTupleIndex_WithExpression() {
	expr, err := ParseExpression("tuple[i + 1]")
	s.NoError(err)

	idx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected *ast.TupleIndex, got %T", expr)

	_, ok = idx.Index.(*ast.RealInfixExpression)
	s.True(ok, "index should be *ast.RealInfixExpression, got %T", idx.Index)
}

func (s *TupleRecordSuite) TestTupleIndex_Chained() {
	// tuple[1][2] = (tuple[1])[2]
	expr, err := ParseExpression("matrix[1][2]")
	s.NoError(err)

	outerIdx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected outer *ast.TupleIndex, got %T", expr)

	innerIdx, ok := outerIdx.Tuple.(*ast.TupleIndex)
	s.True(ok, "inner tuple should be *ast.TupleIndex, got %T", outerIdx.Tuple)

	ident, ok := innerIdx.Tuple.(*ast.Identifier)
	s.True(ok, "innermost should be *ast.Identifier, got %T", innerIdx.Tuple)
	s.Equal("matrix", ident.Value)
}

func (s *TupleRecordSuite) TestTupleIndex_MixedWithFieldAccess() {
	// record.tuple[1] = (record.tuple)[1]
	expr, err := ParseExpression("record.tuple[1]")
	s.NoError(err)

	idx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected *ast.TupleIndex, got %T", expr)

	field, ok := idx.Tuple.(*ast.FieldAccess)
	s.True(ok, "tuple should be *ast.FieldAccess, got %T", idx.Tuple)
	s.Equal("tuple", field.Member)
}

// =============================================================================
// Record Literals
// =============================================================================

func (s *TupleRecordSuite) TestRecordInstance_Single_ASCII() {
	expr, err := ParseExpression("[name |-> \"Alice\"]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(1, len(record.Bindings))
	s.Equal("name", record.Bindings[0].Field.Value)
}

func (s *TupleRecordSuite) TestRecordInstance_Single_Unicode() {
	expr, err := ParseExpression("[name ↦ \"Alice\"]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(1, len(record.Bindings))
	s.Equal("name", record.Bindings[0].Field.Value)
}

func (s *TupleRecordSuite) TestRecordInstance_Multiple() {
	expr, err := ParseExpression("[name |-> \"Alice\", age |-> 30]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(2, len(record.Bindings))
	s.Equal("name", record.Bindings[0].Field.Value)
	s.Equal("age", record.Bindings[1].Field.Value)
}

func (s *TupleRecordSuite) TestRecordInstance_WithExpressions() {
	expr, err := ParseExpression("[x |-> 1 + 2, y |-> a * b]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(2, len(record.Bindings))

	// First binding value should be addition
	_, ok = record.Bindings[0].Expression.(*ast.RealInfixExpression)
	s.True(ok, "expected *ast.RealInfixExpression, got %T", record.Bindings[0].Expression)
}

func (s *TupleRecordSuite) TestRecordInstance_Nested() {
	expr, err := ParseExpression("[person |-> [name |-> \"Alice\", age |-> 30]]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(1, len(record.Bindings))

	innerRecord, ok := record.Bindings[0].Expression.(*ast.RecordInstance)
	s.True(ok, "inner value should be *ast.RecordInstance, got %T", record.Bindings[0].Expression)
	s.Equal(2, len(innerRecord.Bindings))
}

func (s *TupleRecordSuite) TestRecordInstance_String() {
	expr, err := ParseExpression("[x |-> 1, y |-> 2]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal("[x ↦ 1, y ↦ 2]", record.String())
	s.Equal("[x |-> 1, y |-> 2]", record.Ascii())
}

// =============================================================================
// Record EXCEPT
// =============================================================================

func (s *TupleRecordSuite) TestRecordAltered_Single() {
	expr, err := ParseExpression("[r EXCEPT !.count = 10]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal("r", altered.Identifier.Value)
	s.Equal(1, len(altered.Alterations))
	s.Equal("count", altered.Alterations[0].Field.Member)
}

func (s *TupleRecordSuite) TestRecordAltered_Multiple() {
	expr, err := ParseExpression("[record EXCEPT !.field1 = 1, !.field2 = 2, !.field3 = 3]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal("record", altered.Identifier.Value)
	s.Equal(3, len(altered.Alterations))
	s.Equal("field1", altered.Alterations[0].Field.Member)
	s.Equal("field2", altered.Alterations[1].Field.Member)
	s.Equal("field3", altered.Alterations[2].Field.Member)
}

func (s *TupleRecordSuite) TestRecordAltered_WithAt() {
	// @ references the current value of the field
	expr, err := ParseExpression("[r EXCEPT !.count = @ + 1]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal(1, len(altered.Alterations))

	// Expression should be @ + 1 (RealInfixExpression with ExistingValue on left)
	addExpr, ok := altered.Alterations[0].Expression.(*ast.RealInfixExpression)
	s.True(ok, "expected *ast.RealInfixExpression, got %T", altered.Alterations[0].Expression)
	s.Equal("+", addExpr.Operator)

	_, ok = addExpr.Left.(*ast.ExistingValue)
	s.True(ok, "left should be *ast.ExistingValue, got %T", addExpr.Left)
}

func (s *TupleRecordSuite) TestRecordAltered_MultipleWithAt() {
	expr, err := ParseExpression("[r EXCEPT !.count = @ + 1, !.total = @ * 2]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal(2, len(altered.Alterations))
	s.Equal("count", altered.Alterations[0].Field.Member)
	s.Equal("total", altered.Alterations[1].Field.Member)
}

func (s *TupleRecordSuite) TestRecordAltered_WithExpression() {
	expr, err := ParseExpression("[r EXCEPT !.value = x + y * 2]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal(1, len(altered.Alterations))
}

func (s *TupleRecordSuite) TestRecordAltered_String() {
	expr, err := ParseExpression("[r EXCEPT !.x = 1]")
	s.NoError(err)

	altered, ok := expr.(*ast.RecordAltered)
	s.True(ok, "expected *ast.RecordAltered, got %T", expr)
	s.Equal("[r EXCEPT !.x = 1]", altered.String())
	s.Equal("[r EXCEPT !.x = 1]", altered.Ascii())
}

// =============================================================================
// Mixed Tests
// =============================================================================

func (s *TupleRecordSuite) TestMixed_RecordWithTuple() {
	expr, err := ParseExpression("[point |-> <<1, 2, 3>>]")
	s.NoError(err)

	record, ok := expr.(*ast.RecordInstance)
	s.True(ok, "expected *ast.RecordInstance, got %T", expr)
	s.Equal(1, len(record.Bindings))

	tuple, ok := record.Bindings[0].Expression.(*ast.TupleLiteral)
	s.True(ok, "value should be *ast.TupleLiteral, got %T", record.Bindings[0].Expression)
	s.Equal(3, len(tuple.Elements))
}

func (s *TupleRecordSuite) TestMixed_TupleWithRecord() {
	expr, err := ParseExpression("<<[x |-> 1], [x |-> 2]>>")
	s.NoError(err)

	tuple, ok := expr.(*ast.TupleLiteral)
	s.True(ok, "expected *ast.TupleLiteral, got %T", expr)
	s.Equal(2, len(tuple.Elements))

	for i, elem := range tuple.Elements {
		_, ok := elem.(*ast.RecordInstance)
		s.True(ok, "element %d should be *ast.RecordInstance, got %T", i, elem)
	}
}

func (s *TupleRecordSuite) TestMixed_TupleIndexOnRecord() {
	// Access a tuple field on a record, then index it
	expr, err := ParseExpression("person.positions[1]")
	s.NoError(err)

	idx, ok := expr.(*ast.TupleIndex)
	s.True(ok, "expected *ast.TupleIndex, got %T", expr)

	field, ok := idx.Tuple.(*ast.FieldAccess)
	s.True(ok, "tuple should be *ast.FieldAccess, got %T", idx.Tuple)
	s.Equal("positions", field.Member)
}
