package evaluator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LinkTableTestSuite struct {
	suite.Suite
}

func TestLinkTableSuite(t *testing.T) {
	suite.Run(t, new(LinkTableTestSuite))
}

func (s *LinkTableTestSuite) TestNewLinkTable() {
	table := NewLinkTable()
	s.NotNil(table)
	s.Equal(0, table.Count())
}

func (s *LinkTableTestSuite) TestAddLink_QueryForward() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))

	forward := table.GetForward(ObjectID(1), assocKey)
	s.Len(forward, 1)
	s.Contains(forward, ObjectID(2))
}

func (s *LinkTableTestSuite) TestAddLink_QueryReverse() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))

	reverse := table.GetReverse(ObjectID(2), assocKey)
	s.Len(reverse, 1)
	s.Contains(reverse, ObjectID(1))
}

func (s *LinkTableTestSuite) TestAddLink_PreventsDuplicates() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))
	table.AddLink(assocKey, ObjectID(1), ObjectID(2)) // Duplicate

	s.Equal(1, table.Count())

	forward := table.GetForward(ObjectID(1), assocKey)
	s.Len(forward, 1)
}

func (s *LinkTableTestSuite) TestRemoveLink() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))
	s.Equal(1, table.Count())

	removed := table.RemoveLink(assocKey, ObjectID(1), ObjectID(2))
	s.True(removed)
	s.Equal(0, table.Count())

	forward := table.GetForward(ObjectID(1), assocKey)
	s.Len(forward, 0)

	reverse := table.GetReverse(ObjectID(2), assocKey)
	s.Len(reverse, 0)
}

func (s *LinkTableTestSuite) TestRemoveLink_NonExistent() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	removed := table.RemoveLink(assocKey, ObjectID(1), ObjectID(2))
	s.False(removed)
}

func (s *LinkTableTestSuite) TestMultipleLinksFromSameObject() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))
	table.AddLink(assocKey, ObjectID(1), ObjectID(3))
	table.AddLink(assocKey, ObjectID(1), ObjectID(4))

	forward := table.GetForward(ObjectID(1), assocKey)
	s.Len(forward, 3)
	s.Contains(forward, ObjectID(2))
	s.Contains(forward, ObjectID(3))
	s.Contains(forward, ObjectID(4))
}

func (s *LinkTableTestSuite) TestMultipleLinksToSameObject() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(10))
	table.AddLink(assocKey, ObjectID(2), ObjectID(10))
	table.AddLink(assocKey, ObjectID(3), ObjectID(10))

	reverse := table.GetReverse(ObjectID(10), assocKey)
	s.Len(reverse, 3)
	s.Contains(reverse, ObjectID(1))
	s.Contains(reverse, ObjectID(2))
	s.Contains(reverse, ObjectID(3))
}

func (s *LinkTableTestSuite) TestDifferentAssociations() {
	table := NewLinkTable()
	assoc1 := AssociationKey("assoc/one")
	assoc2 := AssociationKey("assoc/two")

	table.AddLink(assoc1, ObjectID(1), ObjectID(2))
	table.AddLink(assoc2, ObjectID(1), ObjectID(3))

	forward1 := table.GetForward(ObjectID(1), assoc1)
	s.Len(forward1, 1)
	s.Contains(forward1, ObjectID(2))

	forward2 := table.GetForward(ObjectID(1), assoc2)
	s.Len(forward2, 1)
	s.Contains(forward2, ObjectID(3))
}

func (s *LinkTableTestSuite) TestGetAllForward() {
	table := NewLinkTable()
	assoc1 := AssociationKey("assoc/one")
	assoc2 := AssociationKey("assoc/two")

	table.AddLink(assoc1, ObjectID(1), ObjectID(2))
	table.AddLink(assoc2, ObjectID(1), ObjectID(3))

	all := table.GetAllForward(ObjectID(1))
	s.Len(all, 2)
}

func (s *LinkTableTestSuite) TestGetAllReverse() {
	table := NewLinkTable()
	assoc1 := AssociationKey("assoc/one")
	assoc2 := AssociationKey("assoc/two")

	table.AddLink(assoc1, ObjectID(1), ObjectID(10))
	table.AddLink(assoc2, ObjectID(2), ObjectID(10))

	all := table.GetAllReverse(ObjectID(10))
	s.Len(all, 2)
}

func (s *LinkTableTestSuite) TestClear() {
	table := NewLinkTable()
	assocKey := AssociationKey("test/association")

	table.AddLink(assocKey, ObjectID(1), ObjectID(2))
	table.AddLink(assocKey, ObjectID(3), ObjectID(4))

	s.Equal(2, table.Count())

	table.Clear()

	s.Equal(0, table.Count())
}
