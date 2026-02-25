package state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type StateTestSuite struct {
	suite.Suite
}

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

// =============================================================================
// Instance Creation and Management
// =============================================================================

func (s *StateTestSuite) TestCreateInstance() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	instance := state.CreateInstance(classKey, attrs)

	s.NotNil(instance)
	s.Equal(InstanceID(1), instance.ID)
	s.Equal(classKey, instance.ClassKey)
	s.Equal("pending", instance.GetAttribute("status").(*object.String).Value())
	s.Equal("100", instance.GetAttribute("total").(*object.Number).Inspect())
}

func (s *StateTestSuite) TestCreateMultipleInstances() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs1 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(1),
	})
	attrs2 := object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(2),
	})

	instance1 := state.CreateInstance(classKey, attrs1)
	instance2 := state.CreateInstance(classKey, attrs2)

	s.Equal(InstanceID(1), instance1.ID)
	s.Equal(InstanceID(2), instance2.ID)
	s.Equal(2, state.InstanceCount())
}

func (s *StateTestSuite) TestGetInstance() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	created := state.CreateInstance(classKey, attrs)
	retrieved := state.GetInstance(created.ID)

	s.NotNil(retrieved)
	s.Equal(created.ID, retrieved.ID)

	// Non-existent instance
	notFound := state.GetInstance(InstanceID(999))
	s.Nil(notFound)
}

func (s *StateTestSuite) TestUpdateInstance() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	instance := state.CreateInstance(classKey, attrs)

	// Update the entire attributes
	newAttrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("shipped"),
	})
	err := state.UpdateInstance(instance.ID, newAttrs)
	s.NoError(err)

	updated := state.GetInstance(instance.ID)
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestUpdateInstanceField() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	})

	instance := state.CreateInstance(classKey, attrs)

	// Update a single field
	err := state.UpdateInstanceField(instance.ID, "status", object.NewString("shipped"))
	s.NoError(err)

	updated := state.GetInstance(instance.ID)
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
	s.Equal("100", updated.GetAttribute("total").(*object.Number).Inspect())
}

func (s *StateTestSuite) TestDeleteInstance() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	instance := state.CreateInstance(classKey, attrs)
	s.Equal(1, state.InstanceCount())

	err := state.DeleteInstance(instance.ID)
	s.NoError(err)
	s.Equal(0, state.InstanceCount())

	// Instance should no longer exist
	retrieved := state.GetInstance(instance.ID)
	s.Nil(retrieved)
}

func (s *StateTestSuite) TestInstancesByClass() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")

	state.CreateInstance(orderKey, object.NewRecord())
	state.CreateInstance(orderKey, object.NewRecord())
	state.CreateInstance(lineKey, object.NewRecord())

	orders := state.InstancesByClass(orderKey)
	lines := state.InstancesByClass(lineKey)

	s.Len(orders, 2)
	s.Len(lines, 1)
}

// =============================================================================
// Association Links
// =============================================================================

func (s *StateTestSuite) TestAddLink() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	order := state.CreateInstance(orderKey, object.NewRecord())
	line := state.CreateInstance(lineKey, object.NewRecord())

	state.AddLink(assocKey, order.ID, line.ID)

	s.Equal(1, state.LinkCount())
}

func (s *StateTestSuite) TestRemoveLink() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	order := state.CreateInstance(orderKey, object.NewRecord())
	line := state.CreateInstance(lineKey, object.NewRecord())

	state.AddLink(assocKey, order.ID, line.ID)
	s.Equal(1, state.LinkCount())

	removed := state.RemoveLink(assocKey, order.ID, line.ID)
	s.True(removed)
	s.Equal(0, state.LinkCount())

	// Removing again should return false
	removed = state.RemoveLink(assocKey, order.ID, line.ID)
	s.False(removed)
}

func (s *StateTestSuite) TestGetLinkedForward() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	order := state.CreateInstance(orderKey, object.NewRecord())
	line1 := state.CreateInstance(lineKey, object.NewRecord())
	line2 := state.CreateInstance(lineKey, object.NewRecord())

	state.AddLink(assocKey, order.ID, line1.ID)
	state.AddLink(assocKey, order.ID, line2.ID)

	linked := state.GetLinkedForward(order.ID, assocKey)
	s.Len(linked, 2)
	s.Contains(linked, line1.ID)
	s.Contains(linked, line2.ID)
}

func (s *StateTestSuite) TestGetLinkedReverse() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	order := state.CreateInstance(orderKey, object.NewRecord())
	line := state.CreateInstance(lineKey, object.NewRecord())

	state.AddLink(assocKey, order.ID, line.ID)

	linked := state.GetLinkedReverse(line.ID, assocKey)
	s.Len(linked, 1)
	s.Equal(order.ID, linked[0])
}

func (s *StateTestSuite) TestDeleteInstanceRemovesLinks() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	order := state.CreateInstance(orderKey, object.NewRecord())
	line := state.CreateInstance(lineKey, object.NewRecord())

	state.AddLink(assocKey, order.ID, line.ID)
	s.Equal(1, state.LinkCount())

	// Delete order - should remove links
	err := state.DeleteInstance(order.ID)
	s.NoError(err)
	s.Equal(0, state.LinkCount())
}

// =============================================================================
// State Machine States
// =============================================================================

func (s *StateTestSuite) TestSetStateMachineState() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	instance := state.CreateInstance(classKey, object.NewRecord())

	err := state.SetStateMachineState(instance.ID, stateKey)
	s.NoError(err)

	retrieved, ok := state.GetStateMachineState(instance.ID)
	s.True(ok)
	s.Equal(stateKey, retrieved)
}

func (s *StateTestSuite) TestClearStateMachineState() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	instance := state.CreateInstance(classKey, object.NewRecord())

	state.SetStateMachineState(instance.ID, stateKey)
	state.ClearStateMachineState(instance.ID)

	_, ok := state.GetStateMachineState(instance.ID)
	s.False(ok)
}

// =============================================================================
// Clone
// =============================================================================

func (s *StateTestSuite) TestClone() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")
	stateKey := s.createStateKey("orders", "management", "order", "pending")

	order := state.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	}))
	line := state.CreateInstance(lineKey, object.NewRecord())
	state.AddLink(assocKey, order.ID, line.ID)
	state.SetStateMachineState(order.ID, stateKey)

	// Clone the state
	cloned := state.Clone()

	// Verify clone has same data
	s.Equal(state.InstanceCount(), cloned.InstanceCount())
	s.Equal(state.LinkCount(), cloned.LinkCount())

	clonedOrder := cloned.GetInstance(order.ID)
	s.NotNil(clonedOrder)
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())

	clonedState, ok := cloned.GetStateMachineState(order.ID)
	s.True(ok)
	s.Equal(stateKey, clonedState)

	// Verify independence - modify original
	state.UpdateInstanceField(order.ID, "status", object.NewString("shipped"))
	s.Equal("pending", clonedOrder.GetAttribute("status").(*object.String).Value())
}

// =============================================================================
// BindingsBuilder Tests
// =============================================================================

func (s *StateTestSuite) TestBindingsBuilder_BuildGlobal() {
	state := NewSimulationState()
	builder := NewBindingsBuilder(state)

	bindings := builder.BuildGlobal()

	s.NotNil(bindings)
	s.NotNil(bindings.RelationContext())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildForInstance() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	instance := state.CreateInstance(classKey, attrs)
	builder := NewBindingsBuilder(state)

	bindings := builder.BuildForInstance(instance)

	s.NotNil(bindings)
	s.NotNil(bindings.Self())
	s.Equal("pending", bindings.Self().Get("status").(*object.String).Value())
	s.Equal(classKey.String(), bindings.SelfClassKey())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildForInstanceWithVariables() {
	state := NewSimulationState()

	classKey := s.createClassKey("orders", "management", "order")
	instance := state.CreateInstance(classKey, object.NewRecord())
	builder := NewBindingsBuilder(state)

	variables := map[string]object.Object{
		"quantity": object.NewInteger(5),
		"price":    object.NewInteger(100),
	}

	bindings := builder.BuildForInstanceWithVariables(instance, variables)

	qty, found := bindings.GetValue("quantity")
	s.True(found)
	s.Equal("5", qty.(*object.Number).Inspect())

	price, found := bindings.GetValue("price")
	s.True(found)
	s.Equal("100", price.(*object.Number).Inspect())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildWithClassInstances() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")

	state.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(1),
	}))
	state.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"id": object.NewInteger(2),
	}))

	builder := NewBindingsBuilder(state)
	classNameMap := map[identity.Key]string{
		orderKey: "Orders",
	}

	bindings := builder.BuildWithClassInstances(classNameMap)

	ordersSet, found := bindings.GetValue("Orders")
	s.True(found)
	s.NotNil(ordersSet)

	set := ordersSet.(*object.Set)
	s.Equal(2, set.Size())
}

func (s *StateTestSuite) TestBindingsBuilder_AddAssociation() {
	state := NewSimulationState()

	orderKey := s.createClassKey("orders", "management", "order")
	lineKey := s.createClassKey("orders", "management", "line")
	assocKey := s.createAssociationKey("orders", "management", "order", "line", "lines")

	builder := NewBindingsBuilder(state)
	builder.AddAssociation(
		assocKey,
		"Lines",
		orderKey,
		lineKey,
		evaluator.Multiplicity{LowerBound: 1, HigherBound: 1},
		evaluator.Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	// Verify association was added
	relCtx := builder.RelationContext()
	forwardRel := relCtx.GetForwardRelation(orderKey.String(), "Lines")
	s.NotNil(forwardRel)
	s.Equal("Lines", forwardRel.Name)
	s.Equal(lineKey.String(), forwardRel.TargetClassKey)
}

// =============================================================================
// ClassInstance Tests
// =============================================================================

func (s *StateTestSuite) TestClassInstance_Clone() {
	classKey := s.createClassKey("orders", "management", "order")
	instance := &ClassInstance{
		ID:       1,
		ClassKey: classKey,
		Attributes: object.NewRecordFromFields(map[string]object.Object{
			"status": object.NewString("pending"),
		}),
	}

	cloned := instance.Clone()

	s.Equal(instance.ID, cloned.ID)
	s.Equal(instance.ClassKey, cloned.ClassKey)
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())

	// Verify independence
	instance.SetAttribute("status", object.NewString("shipped"))
	s.Equal("pending", cloned.GetAttribute("status").(*object.String).Value())
}

func (s *StateTestSuite) TestClassInstance_WithAttribute() {
	classKey := s.createClassKey("orders", "management", "order")
	instance := &ClassInstance{
		ID:       1,
		ClassKey: classKey,
		Attributes: object.NewRecordFromFields(map[string]object.Object{
			"status": object.NewString("pending"),
			"total":  object.NewInteger(100),
		}),
	}

	updated := instance.WithAttribute("status", object.NewString("shipped"))

	// Original unchanged
	s.Equal("pending", instance.GetAttribute("status").(*object.String).Value())

	// New instance has updated value
	s.Equal("shipped", updated.GetAttribute("status").(*object.String).Value())
	s.Equal("100", updated.GetAttribute("total").(*object.Number).Inspect())
}

// =============================================================================
// Helper Methods
// =============================================================================

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

func (s *StateTestSuite) createAssociationKey(domain, subdomain, fromClass, toClass, name string) identity.Key {
	domainKey, err := identity.NewDomainKey(domain)
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomain)
	s.Require().NoError(err)
	fromClassKey, err := identity.NewClassKey(subdomainKey, fromClass)
	s.Require().NoError(err)
	toClassKey, err := identity.NewClassKey(subdomainKey, toClass)
	s.Require().NoError(err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, name)
	s.Require().NoError(err)
	return assocKey
}
