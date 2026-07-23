package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type StateTestSuite struct {
	suite.Suite
}

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

func (s *StateTestSuite) TestCreateInstance() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	inst := st.CreateInstance(classKey, attrs)

	s.NotNil(inst)
	s.Equal(ID(1), inst.ID)
	s.Equal(classKey, inst.ClassKey)
	s.Equal("pending", inst.GetAttribute("status").(*object.String).Value())
	s.Equal("100", inst.GetAttribute("total").(*object.Number).Inspect())
	s.NotNil(st.Schema())
}

func (s *StateTestSuite) TestSchemaSharedAcrossClone() {
	st := NewState(schema.New(schema.EmptyModel()))
	cloned := st.Clone()
	s.Same(st.Schema(), cloned.Schema())
}

func (s *StateTestSuite) TestCreateMultipleInstances() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs1 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(1),
	})
	attrs2 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(2),
	})

	inst1 := st.CreateInstance(classKey, attrs1)
	inst2 := st.CreateInstance(classKey, attrs2)

	s.Equal(ID(1), inst1.ID)
	s.Equal(ID(2), inst2.ID)
	s.Equal(2, st.InstanceCount())
}

func (s *StateTestSuite) TestGetInstance() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	created := st.CreateInstance(classKey, attrs)
	retrieved := st.GetInstance(created.ID)

	s.NotNil(retrieved)
	s.Equal(created.ID, retrieved.ID)

	notFound := st.GetInstance(ID(999))
	s.Nil(notFound)
}

func (s *StateTestSuite) TestUpdateInstance() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	inst := st.CreateInstance(classKey, attrs)

	newAttrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("shipped"),
	})
	err := st.UpdateInstance(inst.ID, newAttrs)
	s.Require().NoError(err)

	updated := st.GetInstance(inst.ID)
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestUpdateInstanceField() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	inst := st.CreateInstance(classKey, attrs)

	err := st.UpdateInstanceField(inst.ID, "status", object.NewString("shipped"))
	s.Require().NoError(err)

	updated := st.GetInstance(inst.ID)
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
	s.Equal("100", updated.GetAttribute("total").(*object.Number).Inspect())
}

func (s *StateTestSuite) TestDeleteInstance() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	inst := st.CreateInstance(classKey, attrs)
	s.Equal(1, st.InstanceCount())

	err := st.DeleteInstance(inst.ID)
	s.Require().NoError(err)
	s.Equal(0, st.InstanceCount())

	retrieved := st.GetInstance(inst.ID)
	s.Nil(retrieved)
}

func (s *StateTestSuite) TestInstancesByClass() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")

	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(lineKey, object.NewRecord())

	orders := st.InstancesByClass(orderKey)
	lines := st.InstancesByClass(lineKey)

	s.Len(orders, 2)
	s.Len(lines, 1)
}

func (s *StateTestSuite) TestForEachInstanceAndClassQueries() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")

	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(lineKey, object.NewRecord())

	var all int
	st.ForEachInstance(func(*Instance) { all++ })
	s.Equal(3, all)

	var orders int
	st.ForEachInstanceOfClass(orderKey, func(*Instance) { orders++ })
	s.Equal(2, orders)
	s.Equal(2, st.CountByClass(orderKey))
	s.True(st.HasInstanceOfClass(orderKey))
	s.False(st.HasInstanceOfClass(s.createClassKey("orders", "management", "missing")))
}

func (s *StateTestSuite) TestLookupIDByRecord() {
	st := NewState(schema.New(schema.EmptyModel()))
	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})
	inst := st.CreateInstance(classKey, attrs)

	id, ok := st.LookupIDByRecord(inst.Attributes)
	s.True(ok)
	s.Equal(inst.ID, id)

	extent := object.NewExtentElement(uint64(inst.ID), inst.Attributes)
	id, ok = st.LookupIDByRecord(extent)
	s.True(ok)
	s.Equal(inst.ID, id)
}

func (s *StateTestSuite) TestSnapshot() {
	st := NewState(schema.New(schema.EmptyModel()))
	classKey := s.createClassKey("orders", "management", "order")
	inst := st.CreateInstance(classKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("open"),
	}))

	snap := st.Snapshot()
	s.Equal(1, snap.InstanceCount)
	s.Equal(0, snap.LinkCount)
	s.Require().Len(snap.Instances, 1)
	s.Equal(inst.ID, snap.Instances[0].ID)
	s.Equal(classKey, snap.Instances[0].ClassKey)
	s.Equal(object.NewString("open").Inspect(), snap.Instances[0].Attributes["status"])
}

func (s *StateTestSuite) TestForEachBinaryLinkOfAssociation() {
	st := NewState(schema.New(schema.EmptyModel()))
	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())
	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))

	var pairs [][2]ID
	st.ForEachBinaryLinkOfAssociation(assocKey, func(fromID, toID ID) {
		pairs = append(pairs, [2]ID{fromID, toID})
	})
	s.Equal([][2]ID{{order.ID, line.ID}}, pairs)
}

func (s *StateTestSuite) TestAddLink() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))

	s.Equal(1, st.LinkCount())
}

func (s *StateTestSuite) TestAddLink_RejectsDuplicatePair() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))
	err := st.AddLink(assocKey, order.ID, line.ID)
	s.Require().Error(err)
	s.Equal(1, st.LinkCount())
}

func (s *StateTestSuite) TestRemoveLink() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))
	s.Equal(1, st.LinkCount())

	removed := st.RemoveLink(assocKey, order.ID, line.ID)
	s.True(removed)
	s.Equal(0, st.LinkCount())

	removed = st.RemoveLink(assocKey, order.ID, line.ID)
	s.False(removed)
}

func (s *StateTestSuite) TestGetLinkedForward() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line1 := st.CreateInstance(lineKey, object.NewRecord())
	line2 := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line1.ID))
	s.Require().NoError(st.AddLink(assocKey, order.ID, line2.ID))

	linked := st.GetLinkedForward(order.ID, assocKey)
	s.Len(linked, 2)
	s.Contains(linked, line1.ID)
	s.Contains(linked, line2.ID)
}

func (s *StateTestSuite) TestGetLinkedReverse() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))

	linked := st.GetLinkedReverse(line.ID, assocKey)
	s.Len(linked, 1)
	s.Equal(order.ID, linked[0])
}

func (s *StateTestSuite) TestDeleteInstanceRemovesLinks() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))
	s.Equal(1, st.LinkCount())

	err := st.DeleteInstance(order.ID)
	s.Require().NoError(err)
	s.Equal(0, st.LinkCount())
}

func (s *StateTestSuite) TestSetStateMachineState() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	inst := st.CreateInstance(classKey, object.NewRecord())

	err := st.SetStateMachineState(inst.ID, stateKey)
	s.Require().NoError(err)

	retrieved, ok := st.GetStateMachineState(inst.ID)
	s.True(ok)
	s.Equal(stateKey, retrieved)
}

func (s *StateTestSuite) TestClearStateMachineState() {
	st := NewState(schema.New(schema.EmptyModel()))

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	inst := st.CreateInstance(classKey, object.NewRecord())

	err := st.SetStateMachineState(inst.ID, stateKey)
	s.Require().NoError(err)
	st.ClearStateMachineState(inst.ID)

	_, ok := st.GetStateMachineState(inst.ID)
	s.False(ok)
}

func (s *StateTestSuite) TestClone() {
	st := NewState(schema.New(schema.EmptyModel()))

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	order := st.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	}))
	line := st.CreateInstance(lineKey, object.NewRecord())
	s.Require().NoError(st.AddLink(assocKey, order.ID, line.ID))
	err := st.SetStateMachineState(order.ID, stateKey)
	s.Require().NoError(err)

	cloned := st.Clone()

	s.Equal(st.InstanceCount(), cloned.InstanceCount())
	s.Equal(st.LinkCount(), cloned.LinkCount())

	clonedOrder := cloned.GetInstance(order.ID)
	s.NotNil(clonedOrder)
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())

	clonedState, ok := cloned.GetStateMachineState(order.ID)
	s.True(ok)
	s.Equal(stateKey, clonedState)

	err = st.UpdateInstanceField(order.ID, "status", object.NewString("shipped"))
	s.Require().NoError(err)
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestInstance_Clone() {
	classKey := s.createClassKey("orders", "management", "order")
	inst := &Instance{
		ID:       1,
		ClassKey: classKey,
		Attributes: object.NewRecordFromFields(map[string]object.Object{
			"status": object.NewString("pending"),
		}),
	}

	cloned := inst.Clone()

	s.Equal(inst.ID, cloned.ID)
	s.Equal(inst.ClassKey, cloned.ClassKey)
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())

	inst.SetAttribute("status", object.NewString("shipped"))
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestInstance_WithAttribute() {
	classKey := s.createClassKey("orders", "management", "order")
	inst := &Instance{
		ID:       1,
		ClassKey: classKey,
		Attributes: object.NewRecordFromFields(map[string]object.Object{
			"status": object.NewString("pending"),
			"total":  object.NewInteger(100),
		}),
	}

	updated := inst.WithAttribute("status", object.NewString("shipped"))

	s.Equal("pending", inst.GetAttribute("status").(*object.String).Value())
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
	s.Equal("100", updated.GetAttribute("total").(*object.Number).Inspect())
}

func (s *StateTestSuite) createClassKey(domain, subdomain, class string) identity.Key {
	domainKey, err := identity.NewDomainKey(domain)
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomain)
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, class)
	s.Require().NoError(err)
	return classKey
}

func (s *StateTestSuite) createStateKey(domain, subdomain, class, stateName string) identity.Key {
	classKey := s.createClassKey(domain, subdomain, class)
	stateKey, err := identity.NewStateKey(classKey, stateName)
	s.Require().NoError(err)
	return stateKey
}

func (s *StateTestSuite) createAssociationKey() identity.Key {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	fromClassKey, err := identity.NewClassKey(subdomainKey, "order")
	s.Require().NoError(err)
	toClassKey, err := identity.NewClassKey(subdomainKey, "line")
	s.Require().NoError(err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, "lines")
	s.Require().NoError(err)
	return assocKey
}
