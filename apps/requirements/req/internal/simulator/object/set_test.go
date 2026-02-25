package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SetSuite struct {
	suite.Suite
}

func TestSetSuite(t *testing.T) {
	suite.Run(t, new(SetSuite))
}

func (s *SetSuite) TestNewSet() {
	set := NewSet()
	s.Equal(0, set.Size())
	s.Equal(TypeSet, set.Type())
}

func (s *SetSuite) TestNewSetFromElements() {
	set := NewSetFromElements([]Object{
		NewInteger(1),
		NewInteger(2),
		NewInteger(3),
	})
	s.Equal(3, set.Size())
}

func (s *SetSuite) TestInspect() {
	set := NewSet()
	s.Equal("{}", set.Inspect())

	set.Add(NewInteger(1))
	set.Add(NewInteger(2))
	s.Equal("{1, 2}", set.Inspect())
}

func (s *SetSuite) TestAddAndContains() {
	set := NewSet()

	set.Add(NewInteger(1))
	set.Add(NewInteger(2))
	s.Equal(2, set.Size())

	// Duplicate should not increase size
	set.Add(NewInteger(1))
	s.Equal(2, set.Size())

	s.True(set.Contains(NewInteger(1)))
	s.True(set.Contains(NewInteger(2)))
	s.False(set.Contains(NewInteger(3)))
}

func (s *SetSuite) TestRemove() {
	set := NewSetFromElements([]Object{
		NewInteger(1), NewInteger(2), NewInteger(3),
	})

	set.Remove(NewInteger(2))
	s.Equal(2, set.Size())
	s.False(set.Contains(NewInteger(2)))
}

func (s *SetSuite) TestSetValue() {
	s1 := NewSet()
	s1.Add(NewInteger(1))

	s2 := NewSetFromElements([]Object{
		NewInteger(10), NewInteger(20),
	})

	err := s1.SetValue(s2)
	s.NoError(err)
	s.Equal(2, s1.Size())
	s.True(s1.Contains(NewInteger(10)))
}

func (s *SetSuite) TestClone() {
	original := NewSetFromElements([]Object{
		NewInteger(1), NewInteger(2),
	})

	clone := original.Clone().(*Set)
	s.Equal(original.Size(), clone.Size())

	clone.Add(NewInteger(3))
	s.Equal(2, original.Size())
	s.Equal(3, clone.Size())
}

func (s *SetSuite) TestUnion() {
	s1 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2)})
	s2 := NewSetFromElements([]Object{NewInteger(2), NewInteger(3)})

	union := s1.Union(s2)
	s.Equal(3, union.Size())
	s.True(union.Contains(NewInteger(1)))
	s.True(union.Contains(NewInteger(2)))
	s.True(union.Contains(NewInteger(3)))
}

func (s *SetSuite) TestIntersection() {
	s1 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2), NewInteger(3)})
	s2 := NewSetFromElements([]Object{NewInteger(2), NewInteger(3), NewInteger(4)})

	intersection := s1.Intersection(s2)
	s.Equal(2, intersection.Size())
	s.True(intersection.Contains(NewInteger(2)))
	s.True(intersection.Contains(NewInteger(3)))
}

func (s *SetSuite) TestDifference() {
	s1 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2), NewInteger(3)})
	s2 := NewSetFromElements([]Object{NewInteger(2), NewInteger(3)})

	diff := s1.Difference(s2)
	s.Equal(1, diff.Size())
	s.True(diff.Contains(NewInteger(1)))
}

func (s *SetSuite) TestIsSubsetOf() {
	s1 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2)})
	s2 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2), NewInteger(3)})

	s.True(s1.IsSubsetOf(s2))
	s.False(s2.IsSubsetOf(s1))
}

func (s *SetSuite) TestEquals() {
	s1 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2)})
	s2 := NewSetFromElements([]Object{NewInteger(1), NewInteger(2)})
	s3 := NewSetFromElements([]Object{NewInteger(1), NewInteger(3)})

	s.True(s1.Equals(s2))
	s.False(s1.Equals(s3))
}

