package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RecordSuite struct {
	suite.Suite
}

func TestRecordSuite(t *testing.T) {
	suite.Run(t, new(RecordSuite))
}

func (s *RecordSuite) TestNewRecord() {
	r := NewRecord()

	s.Equal(0, r.Len())
	s.Equal(TypeRecord, r.Type())
	s.Equal("[]", r.Inspect())
}

func (s *RecordSuite) TestNewRecordFromFields() {
	fields := map[string]Object{
		"name": NewString("Alice"),
		"age":  NewInteger(30),
	}
	r := NewRecordFromFields(fields)

	s.Equal(2, r.Len())
	s.Equal("Alice", r.Get("name").(*String).Value())
	s.Equal("30", r.Get("age").(*Number).Inspect())
}

func (s *RecordSuite) TestInspect() {
	fields := map[string]Object{
		"name": NewString("Bob"),
		"age":  NewInteger(25),
	}
	r := NewRecordFromFields(fields)

	// Fields should be sorted alphabetically
	s.Equal(`[age |-> 25, name |-> "Bob"]`, r.Inspect())
}

func (s *RecordSuite) TestGetAndSet() {
	r := NewRecord()

	// Get on empty record returns nil
	s.Nil(r.Get("name"))

	// Set a field
	r.Set("name", NewString("Charlie"))
	s.Equal("Charlie", r.Get("name").(*String).Value())

	// Update a field
	r.Set("name", NewString("David"))
	s.Equal("David", r.Get("name").(*String).Value())
}

func (s *RecordSuite) TestHas() {
	fields := map[string]Object{
		"name": NewString("Eve"),
	}
	r := NewRecordFromFields(fields)

	s.True(r.Has("name"))
	s.False(r.Has("age"))
}

func (s *RecordSuite) TestFieldNames() {
	fields := map[string]Object{
		"z": NewInteger(1),
		"a": NewString("x"),
		"m": NewBoolean(true),
	}
	r := NewRecordFromFields(fields)

	// Should be sorted alphabetically
	s.Equal([]string{"a", "m", "z"}, r.FieldNames())
}

func (s *RecordSuite) TestClone() {
	original := NewRecordFromFields(map[string]Object{
		"count": NewInteger(42),
	})

	clone := original.Clone().(*Record)
	s.Equal(original.Inspect(), clone.Inspect())
	s.Equal(original.Type(), clone.Type())

	// Modify clone, original unchanged
	clone.Set("count", NewInteger(100))
	s.Equal("42", original.Get("count").(*Number).Inspect())
	s.Equal("100", clone.Get("count").(*Number).Inspect())
}

func (s *RecordSuite) TestSetValue() {
	target := NewRecord()
	source := NewRecordFromFields(map[string]Object{
		"value": NewInteger(99),
	})

	err := target.SetValue(source)
	s.NoError(err)
	s.Equal("99", target.Get("value").(*Number).Inspect())
}

func (s *RecordSuite) TestSetValueIncompatibleType() {
	target := NewRecord()
	source := NewInteger(42)

	err := target.SetValue(source)
	s.Error(err)
}

func (s *RecordSuite) TestEquals() {
	r1 := NewRecordFromFields(map[string]Object{
		"x": NewInteger(1),
		"y": NewInteger(2),
	})

	r2 := NewRecordFromFields(map[string]Object{
		"x": NewInteger(1),
		"y": NewInteger(2),
	})

	r3 := NewRecordFromFields(map[string]Object{
		"x": NewInteger(1),
		"y": NewInteger(3),
	})

	r4 := NewRecordFromFields(map[string]Object{
		"x": NewInteger(1),
	})

	s.True(r1.Equals(r2))
	s.False(r1.Equals(r3))
	s.False(r1.Equals(r4))
}


func (s *RecordSuite) TestWithField() {
	original := NewRecordFromFields(map[string]Object{
		"x": NewInteger(1),
	})

	updated := original.WithField("x", NewInteger(2))

	// Original unchanged
	s.Equal("1", original.Get("x").(*Number).Inspect())
	// Updated has new value
	s.Equal("2", updated.Get("x").(*Number).Inspect())
}

func (s *RecordSuite) TestWithout() {
	original := NewRecordFromFields(map[string]Object{
		"a": NewInteger(1),
		"b": NewInteger(2),
	})

	without := original.Without("a")

	// Original unchanged
	s.Equal(2, original.Len())
	// Without has one less field
	s.Equal(1, without.Len())
	s.False(without.Has("a"))
	s.True(without.Has("b"))
}
