package instance

import (
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
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	inst := st.CreateInstance(classKey, attrs)

	s.NotNil(inst)
	s.Equal(ID(1), inst.GetID())
	s.Equal(classKey, inst.GetClassKey())
	s.Equal("pending", inst.GetAttribute("status").(*object.String).Value())
	s.Equal("100", inst.GetAttribute("total").(*object.Number).Inspect())
	s.NotNil(st.schema())
}

func (s *StateTestSuite) TestSchemaSharedAcrossClone() {
	st := NewState(emptySchema())
	cloned := st.Clone()
	s.Same(st.schema(), cloned.schema())
}

func (s *StateTestSuite) TestCreateMultipleInstances() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs1 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(1),
	})
	attrs2 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(2),
	})

	inst1 := st.CreateInstance(classKey, attrs1)
	inst2 := st.CreateInstance(classKey, attrs2)

	s.Equal(ID(1), inst1.GetID())
	s.Equal(ID(2), inst2.GetID())
	s.Equal(2, st.instanceCount())
}

func (s *StateTestSuite) TestGetInstance() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	created := st.CreateInstance(classKey, attrs)
	retrieved := st.GetInstance(created.GetID())

	s.NotNil(retrieved)
	s.Equal(created.GetID(), retrieved.GetID())

	notFound := st.GetInstance(ID(999))
	s.Nil(notFound)
}

func (s *StateTestSuite) TestUpdateInstance() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	inst := st.CreateInstance(classKey, attrs)

	newAttrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("shipped"),
	})
	err := st.updateInstance(inst.GetID(), newAttrs)
	s.Require().NoError(err)

	updated := st.GetInstance(inst.GetID())
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestUpdateInstanceField() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	inst := st.CreateInstance(classKey, attrs)

	err := st.UpdateInstanceField(inst.GetID(), "status", object.NewString("shipped"))
	s.Require().NoError(err)

	updated := st.GetInstance(inst.GetID())
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
	s.Equal("100", updated.GetAttribute("total").(*object.Number).Inspect())
}

func (s *StateTestSuite) TestDeleteInstance() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	inst := st.CreateInstance(classKey, attrs)
	s.Equal(1, st.instanceCount())

	err := st.DeleteInstance(inst.GetID())
	s.Require().NoError(err)
	s.Equal(0, st.instanceCount())

	retrieved := st.GetInstance(inst.GetID())
	s.Nil(retrieved)
}

func (s *StateTestSuite) TestInstancesByClass() {
	st := NewState(emptySchema())

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
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")

	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(orderKey, object.NewRecord())
	st.CreateInstance(lineKey, object.NewRecord())

	var all int
	st.ForEachInstance(func(*Instance) { all++ })
	s.Equal(3, all)

	var orders int
	st.forEachInstanceOfClass(orderKey, func(*Instance) { orders++ })
	s.Equal(2, orders)
	s.Equal(2, st.countByClass(orderKey))
	s.True(st.hasInstanceOfClass(orderKey))
	s.False(st.hasInstanceOfClass(s.createClassKey("orders", "management", "missing")))
}

func (s *StateTestSuite) TestLookupIDByRecord() {
	st := NewState(emptySchema())
	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})
	inst := st.CreateInstance(classKey, attrs)

	id, ok := st.LookupIDByRecord(inst.GetAttributes())
	s.True(ok)
	s.Equal(inst.GetID(), id)

	extent := object.NewExtentElement(uint64(inst.GetID()), inst.GetAttributes())
	id, ok = st.LookupIDByRecord(extent)
	s.True(ok)
	s.Equal(inst.GetID(), id)
}

func (s *StateTestSuite) TestSnapshot() {
	st := NewState(emptySchema())
	classKey := s.createClassKey("orders", "management", "order")
	inst := st.CreateInstance(classKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("open"),
	}))

	snap := st.Snapshot()
	s.Equal(1, snap.InstanceCount)
	s.Equal(0, snap.LinkCount)
	s.Require().Len(snap.Instances, 1)
	s.Equal(inst.GetID(), snap.Instances[0].ID)
	s.Equal(classKey, snap.Instances[0].ClassKey)
	s.Equal(object.NewString("open").Inspect(), snap.Instances[0].Attributes["status"])
}

func (s *StateTestSuite) TestForEachBinaryLinkOfAssociation() {
	st := NewState(emptySchema())
	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())
	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))

	var pairs [][2]ID
	st.ForEachBinaryLinkOfAssociation(assocKey, func(fromID, toID ID) {
		pairs = append(pairs, [2]ID{fromID, toID})
	})
	s.Equal([][2]ID{{order.GetID(), line.GetID()}}, pairs)
}

func (s *StateTestSuite) TestAddLink() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))

	s.Equal(1, st.linkCount())
}

func (s *StateTestSuite) TestAddLink_RejectsDuplicatePair() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))
	err := st.AddLink(assocKey, order.GetID(), line.GetID())
	s.Require().Error(err)
	s.Equal(1, st.linkCount())
}

func (s *StateTestSuite) TestRemoveLink() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))
	s.Equal(1, st.linkCount())

	removed := st.RemoveLink(assocKey, order.GetID(), line.GetID())
	s.True(removed)
	s.Equal(0, st.linkCount())

	removed = st.RemoveLink(assocKey, order.GetID(), line.GetID())
	s.False(removed)
}

func (s *StateTestSuite) TestGetLinkedForward() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line1 := st.CreateInstance(lineKey, object.NewRecord())
	line2 := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line1.GetID()))
	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line2.GetID()))

	linked := st.GetLinkedForward(order.GetID(), assocKey)
	s.Len(linked, 2)
	s.Contains(linked, line1.GetID())
	s.Contains(linked, line2.GetID())
}

func (s *StateTestSuite) TestGetLinkedReverse() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))

	linked := st.GetLinkedReverse(line.GetID(), assocKey)
	s.Len(linked, 1)
	s.Equal(order.GetID(), linked[0])
}

func (s *StateTestSuite) TestDeleteInstanceRemovesLinks() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()

	order := st.CreateInstance(orderKey, object.NewRecord())
	line := st.CreateInstance(lineKey, object.NewRecord())

	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))
	s.Equal(1, st.linkCount())

	err := st.DeleteInstance(order.GetID())
	s.Require().NoError(err)
	s.Equal(0, st.linkCount())
}

func (s *StateTestSuite) TestSetStateMachineState() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	inst := st.CreateInstance(classKey, object.NewRecord())

	err := st.SetStateMachineState(inst.GetID(), stateKey)
	s.Require().NoError(err)

	retrieved, ok := st.getStateMachineState(inst.GetID())
	s.True(ok)
	s.Equal(stateKey, retrieved)
}

func (s *StateTestSuite) TestClearStateMachineState() {
	st := NewState(emptySchema())

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	inst := st.CreateInstance(classKey, object.NewRecord())

	err := st.SetStateMachineState(inst.GetID(), stateKey)
	s.Require().NoError(err)
	st.clearStateMachineState(inst.GetID())

	_, ok := st.getStateMachineState(inst.GetID())
	s.False(ok)
}

func (s *StateTestSuite) TestClone() {
	st := NewState(emptySchema())

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey()
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	order := st.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	}))
	line := st.CreateInstance(lineKey, object.NewRecord())
	s.Require().NoError(st.AddLink(assocKey, order.GetID(), line.GetID()))
	err := st.SetStateMachineState(order.GetID(), stateKey)
	s.Require().NoError(err)

	cloned := st.Clone()

	s.Equal(st.instanceCount(), cloned.instanceCount())
	s.Equal(st.linkCount(), cloned.linkCount())

	clonedOrder := cloned.GetInstance(order.GetID())
	s.NotNil(clonedOrder)
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())

	clonedState, ok := cloned.getStateMachineState(order.GetID())
	s.True(ok)
	s.Equal(stateKey, clonedState)

	err = st.UpdateInstanceField(order.GetID(), "status", object.NewString("shipped"))
	s.Require().NoError(err)
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestInstance_Clone() {
	classKey := s.createClassKey("orders", "management", "order")
	inst := NewInstance(1, classKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	}))

	cloned := inst.Clone()

	s.Equal(inst.GetID(), cloned.GetID())
	s.Equal(inst.GetClassKey(), cloned.GetClassKey())
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())

	inst.SetAttribute("status", object.NewString("shipped"))
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestInstance_WithAttribute() {
	classKey := s.createClassKey("orders", "management", "order")
	inst := NewInstance(1, classKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	}))

	updated := inst.withAttribute("status", object.NewString("shipped"))

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
