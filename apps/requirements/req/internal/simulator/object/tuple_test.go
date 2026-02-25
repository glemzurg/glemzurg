package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TupleSuite struct {
	suite.Suite
}

func TestTupleSuite(t *testing.T) {
	suite.Run(t, new(TupleSuite))
}

func (s *TupleSuite) TestNewTuple() {
	t := NewTuple()

	s.Equal(0, t.Len())
	s.Equal(TypeTuple, t.Type())
	s.Equal("<<>>", t.Inspect())
}

func (s *TupleSuite) TestNewTupleFromElements() {
	elements := []Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	}
	t := NewTupleFromElements(elements)

	s.Equal(3, t.Len())
	s.Equal("<<1, 2, 3>>", t.Inspect())
}

func (s *TupleSuite) TestAt() {
	t := NewTupleFromElements([]Object{
		NewString("a"),
		NewString("b"),
		NewString("c"),
	})

	// TLA+ uses 1-indexed access
	s.Equal("a", t.At(1).(*String).Value())
	s.Equal("b", t.At(2).(*String).Value())
	s.Equal("c", t.At(3).(*String).Value())

	// Out of bounds returns nil
	s.Nil(t.At(0))
	s.Nil(t.At(4))
}

func (s *TupleSuite) TestSet() {
	t := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	})

	err := t.Set(2, NewInteger(42))
	s.NoError(err)
	s.Equal("42", t.At(2).(*Number).Inspect())

	// Out of bounds returns error
	err = t.Set(0, NewInteger(99))
	s.Error(err)
	err = t.Set(4, NewInteger(99))
	s.Error(err)
}

func (s *TupleSuite) TestHead() {
	// Non-empty tuple
	t := NewTupleFromElements([]Object{
		NewInteger(10),
		NewInteger(20),
	})
	s.Equal("10", t.Head().(*Number).Inspect())

	// Empty tuple
	empty := NewTuple()
	s.Nil(empty.Head())
}

func (s *TupleSuite) TestTail() {
	t := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	})

	tail := t.Tail()
	s.Equal(2, tail.Len())
	s.Equal("<<2, 3>>", tail.Inspect())

	// Single element returns empty
	single := NewTupleFromElements([]Object{NewInteger(1)})
	s.Equal(0, single.Tail().Len())

	// Empty returns empty
	empty := NewTuple()
	s.Equal(0, empty.Tail().Len())
}

func (s *TupleSuite) TestAppend() {
	t := NewTupleFromElements([]Object{NewInteger(1)})

	result := t.Append(NewInteger(2))
	s.Equal("<<1, 2>>", result.Inspect())
	// Original unchanged
	s.Equal("<<1>>", t.Inspect())
}

func (s *TupleSuite) TestPrepend() {
	t := NewTupleFromElements([]Object{NewInteger(2)})

	result := t.Prepend(NewInteger(1))
	s.Equal("<<1, 2>>", result.Inspect())
	// Original unchanged
	s.Equal("<<2>>", t.Inspect())
}

func (s *TupleSuite) TestConcat() {
	t1 := NewTupleFromElements([]Object{NewInteger(1), NewInteger(2)})
	t2 := NewTupleFromElements([]Object{NewInteger(3), NewInteger(4)})

	result := t1.Concat(t2)
	s.Equal("<<1, 2, 3, 4>>", result.Inspect())
	// Originals unchanged
	s.Equal("<<1, 2>>", t1.Inspect())
	s.Equal("<<3, 4>>", t2.Inspect())
}

func (s *TupleSuite) TestSubSeq() {
	t := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
		NewInteger(4),
		NewInteger(5),
	})

	// Normal subsequence (1-indexed, inclusive)
	sub := t.SubSeq(2, 4)
	s.Equal("<<2, 3, 4>>", sub.Inspect())

	// Clamping
	sub = t.SubSeq(0, 10)
	s.Equal("<<1, 2, 3, 4, 5>>", sub.Inspect())

	// Invalid range returns empty
	sub = t.SubSeq(4, 2)
	s.Equal("<<>>", sub.Inspect())
}

func (s *TupleSuite) TestReverse() {
	t := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	})

	rev := t.Reverse()
	s.Equal("<<3, 2, 1>>", rev.Inspect())
	// Original unchanged
	s.Equal("<<1, 2, 3>>", t.Inspect())
}

func (s *TupleSuite) TestContains() {
	t := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	})

	s.True(t.Contains(NewInteger(2)))
	s.False(t.Contains(NewInteger(42)))
}

func (s *TupleSuite) TestStackOperations() {
	st := NewTuple()

	// Push elements
	st.Push(NewInteger(1))
	st.Push(NewInteger(2))
	st.Push(NewInteger(3))
	s.Equal("<<1, 2, 3>>", st.Inspect())

	// Pop from top (last element)
	popped := st.Pop()
	s.Equal("3", popped.(*Number).Inspect())
	s.Equal("<<1, 2>>", st.Inspect())

	// Pop remaining
	st.Pop()
	st.Pop()
	s.Nil(st.Pop()) // Empty returns nil
}

func (s *TupleSuite) TestQueueOperations() {
	q := NewTuple()

	// Enqueue elements
	q.Enqueue(NewInteger(1))
	q.Enqueue(NewInteger(2))
	q.Enqueue(NewInteger(3))
	s.Equal("<<1, 2, 3>>", q.Inspect())

	// Dequeue from front (first element)
	dequeued := q.Dequeue()
	s.Equal("1", dequeued.(*Number).Inspect())
	s.Equal("<<2, 3>>", q.Inspect())

	// Dequeue remaining
	q.Dequeue()
	q.Dequeue()
	s.Nil(q.Dequeue()) // Empty returns nil
}

func (s *TupleSuite) TestClone() {
	original := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
	})

	clone := original.Clone().(*Tuple)
	s.Equal(original.Inspect(), clone.Inspect())
	s.Equal(original.Type(), clone.Type())

	// Modify clone, original unchanged
	clone.Push(NewInteger(3))
	s.Equal("<<1, 2>>", original.Inspect())
	s.Equal("<<1, 2, 3>>", clone.Inspect())
}

func (s *TupleSuite) TestSetValue() {
	target := NewTuple()
	source := NewTupleFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
	})

	err := target.SetValue(source)
	s.NoError(err)
	s.Equal("<<1, 2>>", target.Inspect())
}

func (s *TupleSuite) TestSetValueIncompatibleType() {
	target := NewTuple()
	source := NewInteger(42)

	err := target.SetValue(source)
	s.Error(err)
}

func (s *TupleSuite) TestEquals() {
	t1 := NewTupleFromElements([]Object{NewInteger(1), NewInteger(2)})
	t2 := NewTupleFromElements([]Object{NewInteger(1), NewInteger(2)})
	t3 := NewTupleFromElements([]Object{NewInteger(1), NewInteger(3)})
	t4 := NewTupleFromElements([]Object{NewInteger(1)})

	s.True(t1.Equals(t2))
	s.False(t1.Equals(t3))
	s.False(t1.Equals(t4))
}

