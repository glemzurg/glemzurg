package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BagSuite struct {
	suite.Suite
}

func TestBagSuite(t *testing.T) {
	suite.Run(t, new(BagSuite))
}

func (s *BagSuite) TestNewBag() {
	bag := NewBag()
	s.Equal(0, bag.Size())
	s.Equal(0, bag.UniqueCount())
	s.Equal(TypeBag, bag.Type())
}

func (s *BagSuite) TestInspect() {
	bag := NewBag()
	s.Equal("()", bag.Inspect())

	bag.Add(NewInteger(1), 2)
	bag.Add(NewInteger(2), 1)
	s.Equal("(1, 1, 2)", bag.Inspect())
}

func (s *BagSuite) TestAddAndSize() {
	bag := NewBag()

	bag.Add(NewInteger(1), 2)
	bag.Add(NewInteger(2), 3)
	s.Equal(5, bag.Size())
	s.Equal(2, bag.UniqueCount())
}

func (s *BagSuite) TestAddIncreasesCount() {
	bag := NewBag()

	bag.Add(NewInteger(1), 2)
	bag.Add(NewInteger(1), 3)
	s.Equal(5, bag.CopiesIn(NewInteger(1)))
}

func (s *BagSuite) TestRemove() {
	bag := NewBag()

	bag.Add(NewInteger(1), 5)
	bag.Remove(NewInteger(1), 2)
	s.Equal(3, bag.CopiesIn(NewInteger(1)))

	bag.Remove(NewInteger(1), 10) // Remove more than exists
	s.Equal(0, bag.CopiesIn(NewInteger(1)))
	s.False(bag.Contains(NewInteger(1)))
}

func (s *BagSuite) TestCopiesIn() {
	bag := NewBag()

	bag.Add(NewInteger(1), 3)
	bag.Add(NewInteger(2), 1)

	s.Equal(3, bag.CopiesIn(NewInteger(1)))
	s.Equal(1, bag.CopiesIn(NewInteger(2)))
	s.Equal(0, bag.CopiesIn(NewInteger(3)))
}

func (s *BagSuite) TestSetValue() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 1)

	b2 := NewBag()
	b2.Add(NewInteger(10), 2)
	b2.Add(NewInteger(20), 3)

	err := b1.SetValue(b2)
	s.NoError(err)
	s.Equal(5, b1.Size())
	s.Equal(2, b1.CopiesIn(NewInteger(10)))
}

func (s *BagSuite) TestClone() {
	original := NewBag()
	original.Add(NewInteger(1), 2)
	original.Add(NewInteger(2), 3)

	clone := original.Clone().(*Bag)
	s.Equal(original.Size(), clone.Size())

	clone.Add(NewInteger(3), 1)
	s.Equal(5, original.Size())
	s.Equal(6, clone.Size())
}

func (s *BagSuite) TestUnion() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 2)
	b1.Add(NewInteger(2), 3)

	b2 := NewBag()
	b2.Add(NewInteger(1), 4)
	b2.Add(NewInteger(3), 1)

	union := b1.Union(b2)
	s.Equal(4, union.CopiesIn(NewInteger(1))) // max(2, 4)
	s.Equal(3, union.CopiesIn(NewInteger(2))) // only in b1
	s.Equal(1, union.CopiesIn(NewInteger(3))) // only in b2
}

func (s *BagSuite) TestSum() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 2)

	b2 := NewBag()
	b2.Add(NewInteger(1), 3)

	sum := b1.Sum(b2)
	s.Equal(5, sum.CopiesIn(NewInteger(1))) // 2 + 3
}

func (s *BagSuite) TestDifference() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 5)
	b1.Add(NewInteger(2), 3)

	b2 := NewBag()
	b2.Add(NewInteger(1), 2)
	b2.Add(NewInteger(2), 5)

	diff := b1.Difference(b2)
	s.Equal(3, diff.CopiesIn(NewInteger(1))) // 5 - 2
	s.Equal(0, diff.CopiesIn(NewInteger(2))) // 3 - 5 <= 0
}

func (s *BagSuite) TestIsSubBagOf() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 2)

	b2 := NewBag()
	b2.Add(NewInteger(1), 3)
	b2.Add(NewInteger(2), 1)

	s.True(b1.IsSubBagOf(b2))
	s.False(b2.IsSubBagOf(b1))
}

func (s *BagSuite) TestEquals() {
	b1 := NewBag()
	b1.Add(NewInteger(1), 2)

	b2 := NewBag()
	b2.Add(NewInteger(1), 2)

	b3 := NewBag()
	b3.Add(NewInteger(1), 3)

	s.True(b1.Equals(b2))
	s.False(b1.Equals(b3))
}

