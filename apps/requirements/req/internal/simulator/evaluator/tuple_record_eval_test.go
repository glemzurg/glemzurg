package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/parser"
	"github.com/stretchr/testify/suite"
)

func TestTupleRecordEvalSuite(t *testing.T) {
	suite.Run(t, new(TupleRecordEvalSuite))
}

type TupleRecordEvalSuite struct {
	suite.Suite
}

// =============================================================================
// Tuple Literal Evaluation
// =============================================================================

func (s *TupleRecordEvalSuite) TestTupleLiteral_Empty() {
	expr, err := parser.ParseExpression("<<>>")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(0, tuple.Len())
}

func (s *TupleRecordEvalSuite) TestTupleLiteral_Integers() {
	expr, err := parser.ParseExpression("<<1, 2, 3>>")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(3, tuple.Len())
}

func (s *TupleRecordEvalSuite) TestTupleLiteral_WithVariables() {
	expr, err := parser.ParseExpression("<<x, y, z>>")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(10), NamespaceGlobal)
	bindings.Set("y", object.NewInteger(20), NamespaceGlobal)
	bindings.Set("z", object.NewInteger(30), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(3, tuple.Len())

	// Check values (1-indexed)
	s.Equal("10", tuple.At(1).Inspect())
	s.Equal("20", tuple.At(2).Inspect())
	s.Equal("30", tuple.At(3).Inspect())
}

func (s *TupleRecordEvalSuite) TestTupleLiteral_WithExpressions() {
	expr, err := parser.ParseExpression("<<1 + 2, 3 * 4>>")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(2, tuple.Len())
	s.Equal("3", tuple.At(1).Inspect())  // 1 + 2 = 3
	s.Equal("12", tuple.At(2).Inspect()) // 3 * 4 = 12
}

func (s *TupleRecordEvalSuite) TestTupleLiteral_Nested() {
	expr, err := parser.ParseExpression("<<1, <<2, 3>>, 4>>")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(3, tuple.Len())

	// Second element should be a tuple
	innerTuple, ok := tuple.At(2).(*object.Tuple)
	s.True(ok, "inner element should be *object.Tuple, got %T", tuple.At(2))
	s.Equal(2, innerTuple.Len())
}

// =============================================================================
// Tuple Indexing Evaluation
// =============================================================================

func (s *TupleRecordEvalSuite) TestTupleIndex_First() {
	expr, err := parser.ParseExpression("tuple[1]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewInteger(10),
		object.NewInteger(20),
		object.NewInteger(30),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("10", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestTupleIndex_Middle() {
	expr, err := parser.ParseExpression("tuple[2]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewInteger(10),
		object.NewInteger(20),
		object.NewInteger(30),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("20", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestTupleIndex_Last() {
	expr, err := parser.ParseExpression("tuple[3]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewInteger(10),
		object.NewInteger(20),
		object.NewInteger(30),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("30", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestTupleIndex_OutOfBounds() {
	expr, err := parser.ParseExpression("tuple[10]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewInteger(10),
		object.NewInteger(20),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "out of bounds")
}

func (s *TupleRecordEvalSuite) TestTupleIndex_LiteralTuple() {
	expr, err := parser.ParseExpression("<<10, 20, 30>>[2]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("20", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestTupleIndex_WithExpressionIndex() {
	expr, err := parser.ParseExpression("tuple[i + 1]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewInteger(10),
		object.NewInteger(20),
		object.NewInteger(30),
	}), NamespaceGlobal)
	bindings.Set("i", object.NewInteger(1), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("20", num.Inspect()) // tuple[1+1] = tuple[2] = 20
}

func (s *TupleRecordEvalSuite) TestTupleIndex_Chained() {
	// matrix[1][2] where matrix is <<⟨10, 20⟩, ⟨30, 40⟩>>
	expr, err := parser.ParseExpression("matrix[1][2]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("matrix", object.NewTupleFromElements([]object.Object{
		object.NewTupleFromElements([]object.Object{
			object.NewInteger(10),
			object.NewInteger(20),
		}),
		object.NewTupleFromElements([]object.Object{
			object.NewInteger(30),
			object.NewInteger(40),
		}),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("20", num.Inspect()) // matrix[1][2] = first row, second column = 20
}

// =============================================================================
// Record Literal Evaluation
// =============================================================================

func (s *TupleRecordEvalSuite) TestRecordInstance_Single() {
	expr, err := parser.ParseExpression("[name |-> \"Alice\"]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	nameVal := record.Get("name")
	s.NotNil(nameVal)
	str, ok := nameVal.(*object.String)
	s.True(ok, "expected *object.String, got %T", nameVal)
	s.Equal("Alice", str.Value())
}

func (s *TupleRecordEvalSuite) TestRecordInstance_Multiple() {
	expr, err := parser.ParseExpression("[name |-> \"Alice\", age |-> 30]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	nameVal := record.Get("name")
	s.NotNil(nameVal)
	str, ok := nameVal.(*object.String)
	s.True(ok, "expected *object.String, got %T", nameVal)
	s.Equal("Alice", str.Value())

	ageVal := record.Get("age")
	s.NotNil(ageVal)
	num, ok := ageVal.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", ageVal)
	s.Equal("30", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestRecordInstance_WithExpressions() {
	expr, err := parser.ParseExpression("[x |-> 1 + 2, y |-> 3 * 4]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	xVal := record.Get("x")
	s.Equal("3", xVal.Inspect()) // 1 + 2

	yVal := record.Get("y")
	s.Equal("12", yVal.Inspect()) // 3 * 4
}

func (s *TupleRecordEvalSuite) TestRecordInstance_WithVariables() {
	expr, err := parser.ParseExpression("[a |-> x, b |-> y]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(100), NamespaceGlobal)
	bindings.Set("y", object.NewInteger(200), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	s.Equal("100", record.Get("a").Inspect())
	s.Equal("200", record.Get("b").Inspect())
}

func (s *TupleRecordEvalSuite) TestRecordInstance_Nested() {
	expr, err := parser.ParseExpression("[person |-> [name |-> \"Alice\", age |-> 30]]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	personVal := record.Get("person")
	innerRecord, ok := personVal.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", personVal)

	s.Equal("Alice", innerRecord.Get("name").(*object.String).Value())
	s.Equal("30", innerRecord.Get("age").Inspect())
}

// =============================================================================
// Record Field Access Evaluation
// =============================================================================

func (s *TupleRecordEvalSuite) TestRecordFieldAccess() {
	expr, err := parser.ParseExpression("record.name")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("record", object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
		"age":  object.NewInteger(30),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	str, ok := result.Value.(*object.String)
	s.True(ok, "expected *object.String, got %T", result.Value)
	s.Equal("Alice", str.Value())
}

// =============================================================================
// Record EXCEPT Evaluation
// =============================================================================

func (s *TupleRecordEvalSuite) TestRecordAltered_Simple() {
	expr, err := parser.ParseExpression("[r EXCEPT !.count = 10]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewInteger(5),
		"name":  object.NewString("test"),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	// count should be updated
	s.Equal("10", record.Get("count").Inspect())
	// name should be unchanged
	s.Equal("test", record.Get("name").(*object.String).Value())
}

func (s *TupleRecordEvalSuite) TestRecordAltered_Multiple() {
	expr, err := parser.ParseExpression("[r EXCEPT !.x = 100, !.y = 200]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"x": object.NewInteger(1),
		"y": object.NewInteger(2),
		"z": object.NewInteger(3),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	s.Equal("100", record.Get("x").Inspect())
	s.Equal("200", record.Get("y").Inspect())
	s.Equal("3", record.Get("z").Inspect()) // unchanged
}

func (s *TupleRecordEvalSuite) TestRecordAltered_WithAt() {
	// @ references the current value of the field
	expr, err := parser.ParseExpression("[r EXCEPT !.count = @ + 1]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewInteger(41),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	s.Equal("42", record.Get("count").Inspect()) // 41 + 1 = 42
}

func (s *TupleRecordEvalSuite) TestRecordAltered_MultipleWithAt() {
	expr, err := parser.ParseExpression("[r EXCEPT !.count = @ + 1, !.total = @ * 2]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewInteger(10),
		"total": object.NewInteger(50),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	s.Equal("11", record.Get("count").Inspect())  // 10 + 1 = 11
	s.Equal("100", record.Get("total").Inspect()) // 50 * 2 = 100
}

func (s *TupleRecordEvalSuite) TestRecordAltered_DoesNotMutateOriginal() {
	// EXCEPT should create a new record, not mutate the original
	expr, err := parser.ParseExpression("[r EXCEPT !.x = 999]")
	s.NoError(err)

	originalRecord := object.NewRecordFromFields(map[string]object.Object{
		"x": object.NewInteger(1),
	})

	bindings := NewBindings()
	bindings.Set("r", originalRecord, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	newRecord, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	// New record should have updated value
	s.Equal("999", newRecord.Get("x").Inspect())

	// Original record should be unchanged
	s.Equal("1", originalRecord.Get("x").Inspect())
}

// =============================================================================
// Combined Tests
// =============================================================================

func (s *TupleRecordEvalSuite) TestCombined_RecordWithTuple() {
	expr, err := parser.ParseExpression("[point |-> <<1, 2, 3>>]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record, ok := result.Value.(*object.Record)
	s.True(ok, "expected *object.Record, got %T", result.Value)

	pointVal := record.Get("point")
	tuple, ok := pointVal.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", pointVal)
	s.Equal(3, tuple.Len())
}

func (s *TupleRecordEvalSuite) TestCombined_TupleWithRecords() {
	expr, err := parser.ParseExpression("<<[x |-> 1], [x |-> 2]>>")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok, "expected *object.Tuple, got %T", result.Value)
	s.Equal(2, tuple.Len())

	// Check each element is a record with x field
	for i := 1; i <= 2; i++ {
		rec, ok := tuple.At(i).(*object.Record)
		s.True(ok, "element %d should be *object.Record", i)
		s.Equal(string(rune('0'+i)), rec.Get("x").Inspect())
	}
}

func (s *TupleRecordEvalSuite) TestCombined_AccessTupleInRecord() {
	// Access element of tuple stored in record
	expr, err := parser.ParseExpression("record.points[1]")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("record", object.NewRecordFromFields(map[string]object.Object{
		"points": object.NewTupleFromElements([]object.Object{
			object.NewInteger(10),
			object.NewInteger(20),
			object.NewInteger(30),
		}),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num, ok := result.Value.(*object.Number)
	s.True(ok, "expected *object.Number, got %T", result.Value)
	s.Equal("10", num.Inspect())
}

func (s *TupleRecordEvalSuite) TestCombined_AccessRecordInTuple() {
	// Access field of record stored in tuple
	expr, err := parser.ParseExpression("tuple[1].name")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("tuple", object.NewTupleFromElements([]object.Object{
		object.NewRecordFromFields(map[string]object.Object{
			"name": object.NewString("Alice"),
		}),
		object.NewRecordFromFields(map[string]object.Object{
			"name": object.NewString("Bob"),
		}),
	}), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	str, ok := result.Value.(*object.String)
	s.True(ok, "expected *object.String, got %T", result.Value)
	s.Equal("Alice", str.Value())
}
