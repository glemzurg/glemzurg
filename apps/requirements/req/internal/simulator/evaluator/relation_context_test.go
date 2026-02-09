package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type RelationContextTestSuite struct {
	suite.Suite
}

func TestRelationContextSuite(t *testing.T) {
	suite.Run(t, new(RelationContextTestSuite))
}

func (s *RelationContextTestSuite) TestNewRelationContext() {
	ctx := NewRelationContext()
	s.NotNil(ctx)
	s.NotNil(ctx.ForwardRelations)
	s.NotNil(ctx.ReverseRelations)
	s.NotNil(ctx.Identities())
	s.NotNil(ctx.Links())
}

func (s *RelationContextTestSuite) TestAddAssociation_ForwardRelation() {
	ctx := NewRelationContext()

	ctx.AddAssociation(
		AssociationKey("domain/sub/cassociation/Order/LineItem/lines"),
		"Lines",
		"domain/sub/class/Order",
		"domain/sub/class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0}, // 0..* means unlimited
	)

	// Forward: Order.Lines
	info := ctx.GetForwardRelation("domain/sub/class/Order", "Lines")
	s.NotNil(info)
	s.Equal("Lines", info.Name)
	s.Equal("domain/sub/class/LineItem", info.TargetClassKey)
	s.False(info.Reverse)
}

func (s *RelationContextTestSuite) TestAddAssociation_ReverseRelation() {
	ctx := NewRelationContext()

	ctx.AddAssociation(
		AssociationKey("domain/sub/cassociation/Order/LineItem/lines"),
		"Lines",
		"domain/sub/class/Order",
		"domain/sub/class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	// Reverse: LineItem._Lines
	info := ctx.GetReverseRelation("domain/sub/class/LineItem", "_Lines")
	s.NotNil(info)
	s.Equal("Lines", info.Name)
	s.Equal("domain/sub/class/Order", info.TargetClassKey)
	s.True(info.Reverse)
}

func (s *RelationContextTestSuite) TestGetRelation_Forward() {
	ctx := NewRelationContext()

	ctx.AddAssociation(
		AssociationKey("test/assoc"),
		"Items",
		"class/Parent",
		"class/Child",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	// GetRelation should find forward relation
	info := ctx.GetRelation("class/Parent", "Items")
	s.NotNil(info)
	s.False(info.Reverse)
}

func (s *RelationContextTestSuite) TestGetRelation_Reverse() {
	ctx := NewRelationContext()

	ctx.AddAssociation(
		AssociationKey("test/assoc"),
		"Items",
		"class/Parent",
		"class/Child",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	// GetRelation should find reverse relation with underscore prefix
	info := ctx.GetRelation("class/Child", "_Items")
	s.NotNil(info)
	s.True(info.Reverse)
}

func (s *RelationContextTestSuite) TestGetRelation_NotFound() {
	ctx := NewRelationContext()

	info := ctx.GetRelation("class/Something", "NotExisting")
	s.Nil(info)
}

func (s *RelationContextTestSuite) TestCreateLink() {
	ctx := NewRelationContext()

	ctx.AddAssociation(
		AssociationKey("test/assoc"),
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	lineItem := object.NewRecord()
	lineItem.Set("id", object.NewNatural(100))

	ctx.CreateLink(AssociationKey("test/assoc"), order, lineItem)

	// Verify link was created
	s.Equal(1, ctx.Links().Count())

	// Verify both records have IDs
	orderId, ok := ctx.GetObjectID(order)
	s.True(ok)
	s.NotEqual(ObjectID(0), orderId)

	lineId, ok := ctx.GetObjectID(lineItem)
	s.True(ok)
	s.NotEqual(ObjectID(0), lineId)
}

func (s *RelationContextTestSuite) TestGetRelatedRecords_Forward() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")
	ctx.AddAssociation(
		assocKey,
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	lineItem1 := object.NewRecord()
	lineItem1.Set("id", object.NewNatural(101))

	lineItem2 := object.NewRecord()
	lineItem2.Set("id", object.NewNatural(102))

	// Create links
	ctx.CreateLink(assocKey, order, lineItem1)
	ctx.CreateLink(assocKey, order, lineItem2)

	// Forward traversal: order.Lines
	related := ctx.GetRelatedRecords(order, assocKey, false)
	s.Len(related, 2)
	s.Contains(related, lineItem1)
	s.Contains(related, lineItem2)
}

func (s *RelationContextTestSuite) TestGetRelatedRecords_Reverse() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")
	ctx.AddAssociation(
		assocKey,
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	lineItem := object.NewRecord()
	lineItem.Set("id", object.NewNatural(101))

	// Create link
	ctx.CreateLink(assocKey, order, lineItem)

	// Reverse traversal: lineItem._Lines
	related := ctx.GetRelatedRecords(lineItem, assocKey, true)
	s.Len(related, 1)
	s.Contains(related, order)
}

func (s *RelationContextTestSuite) TestGetRelatedRecords_Empty() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")
	ctx.AddAssociation(
		assocKey,
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	// Order with no line items
	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))
	ctx.RegisterRecord(order) // Register but don't create links

	related := ctx.GetRelatedRecords(order, assocKey, false)
	s.Len(related, 0)
}

func (s *RelationContextTestSuite) TestGetRelatedRecords_UnregisteredRecord() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")

	// Unregistered record
	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	// Should return nil for unregistered record
	related := ctx.GetRelatedRecords(order, assocKey, false)
	s.Nil(related)
}

func (s *RelationContextTestSuite) TestRemoveLink() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")
	ctx.AddAssociation(
		assocKey,
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	lineItem := object.NewRecord()
	lineItem.Set("id", object.NewNatural(101))

	// Create and then remove link
	ctx.CreateLink(assocKey, order, lineItem)
	s.Equal(1, ctx.Links().Count())

	removed := ctx.RemoveLink(assocKey, order, lineItem)
	s.True(removed)
	s.Equal(0, ctx.Links().Count())

	// Verify no related records
	related := ctx.GetRelatedRecords(order, assocKey, false)
	s.Len(related, 0)
}

func (s *RelationContextTestSuite) TestClear() {
	ctx := NewRelationContext()

	assocKey := AssociationKey("test/lines")
	ctx.AddAssociation(
		assocKey,
		"Lines",
		"class/Order",
		"class/LineItem",
		Multiplicity{LowerBound: 1, HigherBound: 1},
		Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	order := object.NewRecord()
	order.Set("id", object.NewNatural(1))

	lineItem := object.NewRecord()
	lineItem.Set("id", object.NewNatural(101))

	ctx.CreateLink(assocKey, order, lineItem)

	// Clear runtime state
	ctx.Clear()

	// Links and identities should be cleared
	s.Equal(0, ctx.Links().Count())
	s.Equal(0, ctx.Identities().Count())

	// But association metadata should remain
	info := ctx.GetForwardRelation("class/Order", "Lines")
	s.NotNil(info)
}
