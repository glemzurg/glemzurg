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
// BindingsBuilder Tests
// =============================================================================

func (s *StateTestSuite) TestBindingsBuilder_BuildGlobal() {
	simState := NewSimulationState()
	builder := NewBindingsBuilder(simState)

	bindings := builder.BuildGlobal()

	s.NotNil(bindings)
	s.NotNil(bindings.RelationContext())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildForInstance() {
	simState := NewSimulationState()

	classKey := s.createClassKey("order")
	attrs := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
	})

	inst := simState.CreateInstance(classKey, attrs)
	builder := NewBindingsBuilder(simState)

	bindings := builder.BuildForInstance(inst)

	s.NotNil(bindings)
	s.NotNil(bindings.Self())
	s.Equal("pending", bindings.Self().Get("status").(*object.String).Value())
	s.Equal(classKey.String(), bindings.SelfClassKey())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildForInstanceWithVariables() {
	simState := NewSimulationState()

	classKey := s.createClassKey("order")
	inst := simState.CreateInstance(classKey, object.NewRecord())
	builder := NewBindingsBuilder(simState)

	variables := map[string]object.Object{
		"quantity": object.NewInteger(5),
		"price":    object.NewInteger(100),
	}

	bindings := builder.BuildForInstanceWithVariables(inst, variables)

	qty, found := bindings.GetValue("quantity")
	s.True(found)
	s.Equal("5", qty.(*object.Number).Inspect())

	price, found := bindings.GetValue("price")
	s.True(found)
	s.Equal("100", price.(*object.Number).Inspect())
}

func (s *StateTestSuite) TestBindingsBuilder_BuildWithClassInstances() {
	simState := NewSimulationState()

	orderKey := s.createClassKey("order")

	// Identical attribute data: extent must still have two elements (distinct ids).
	simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Open"),
	}))
	simState.CreateInstance(orderKey, object.NewRecordFromFields(map[string]object.Object{
		"_state": object.NewString("Open"),
	}))

	builder := NewBindingsBuilder(simState)
	classNameMap := map[identity.Key]string{
		orderKey: "Orders",
	}

	bindings := builder.BuildWithClassInstances(classNameMap)

	ordersSet, found := bindings.GetValue("Orders")
	s.True(found)
	s.NotNil(ordersSet)

	set := ordersSet.(*object.Set)
	s.Equal(2, set.Size())

	for _, elem := range set.Elements() {
		rec, ok := elem.(*object.Record)
		s.Require().True(ok)
		s.True(rec.Has(ClassExtentIDField))
		s.True(rec.Has(ClassExtentDataField))
		data, ok := rec.Get(ClassExtentDataField).(*object.Record)
		s.Require().True(ok)
		s.Equal("Open", data.Get("_state").(*object.String).Value())
		// data must not be polluted with a synthetic id field for identity
		s.False(data.Has(ClassExtentIDField))
	}
}

func (s *StateTestSuite) TestBindingsBuilder_AddAssociation() {
	simState := NewSimulationState()

	orderKey := s.createClassKey("order")
	lineKey := s.createClassKey("line")
	assocKey := s.createAssociationKey()

	builder := NewBindingsBuilder(simState)
	builder.AddAssociation(
		assocKey,
		"Lines",
		orderKey,
		lineKey,
		evaluator.Multiplicity{LowerBound: 1, HigherBound: 1},
		evaluator.Multiplicity{LowerBound: 0, HigherBound: 0},
	)

	relCtx := builder.RelationContext()
	forwardRel := relCtx.GetForwardRelation(orderKey.String(), "Lines")
	s.NotNil(forwardRel)
	s.Equal("Lines", forwardRel.Name)
	s.Equal(lineKey.String(), forwardRel.TargetClassKey)
}

func (s *StateTestSuite) createClassKey(class string) identity.Key {
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	classKey, err := identity.NewClassKey(subdomainKey, class)
	s.Require().NoError(err)
	return classKey
}

func (s *StateTestSuite) createAssociationKey() identity.Key {
	orderKey := s.createClassKey("order")
	lineKey := s.createClassKey("line")
	domainKey, err := identity.NewDomainKey("orders")
	s.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "management")
	s.Require().NoError(err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, orderKey, lineKey, "lines")
	s.Require().NoError(err)
	return assocKey
}
