package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/suite"
)

// OutOfScopeProtocolSuite covers model-agnostic surface boundary defaults:
// empty class extents, object params as {}, and link-to-OOS no-op.
type OutOfScopeProtocolSuite struct {
	suite.Suite
}

func TestOutOfScopeProtocolSuite(t *testing.T) {
	suite.Run(t, new(OutOfScopeProtocolSuite))
}

func (s *OutOfScopeProtocolSuite) orderItemAssocModel() (
	orderClass model_class.Class,
	orderKey identity.Key,
	itemClass model_class.Class,
	itemKey identity.Key,
	assocKey identity.Key,
	assoc model_class.Association,
) {
	orderClass, orderKey = testOrderClass()
	itemClass, itemKey = testItemClass()
	assocKey = testAssocKey(orderKey, itemKey, "Lines")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc = model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Lines", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult},
		model_class.AssociationOptions{},
	)
	return orderClass, orderKey, itemClass, itemKey, assocKey, assoc
}

func (s *OutOfScopeProtocolSuite) TestRegisterOutOfScopeMetadata_EmptyExtentNames() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	full := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	active := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(active)
	s.NotContains(catalog.ClassNameMap(), itemKey)

	catalog.RegisterOutOfScopeMetadata(full)
	names := catalog.ClassNameMap()
	s.Equal("Order", names[orderKey])
	s.Equal("Item", names[itemKey])
	s.False(catalog.IsClassInScope(itemKey))
	s.True(catalog.IsClassInScope(orderKey))
}

func (s *OutOfScopeProtocolSuite) TestRegisterOutOfScopeMetadata_BoundaryAssociationNavigable() {
	orderClass, orderKey, itemClass, itemKey, assocKey, assoc := s.orderItemAssocModel()
	full := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	full.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	active := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(active)
	_, _, foundBefore := catalog.OutgoingAssociationByTLAField(orderKey, "Lines")
	s.False(foundBefore)

	catalog.RegisterOutOfScopeMetadata(full)
	gotKey, gotAssoc, found := catalog.OutgoingAssociationByTLAField(orderKey, "Lines")
	s.True(found)
	s.Equal(assocKey, gotKey)
	s.Equal(itemKey, gotAssoc.ToClassKey)
	_, peerOK := catalog.PeerClass(itemKey)
	s.False(peerOK)
}

func (s *OutOfScopeProtocolSuite) TestClassExtentBinding_OutOfScopeIsEmptySet() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	full := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	active := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(active)
	catalog.RegisterOutOfScopeMetadata(full)

	simState := instance.NewState(schema.New(schema.EmptyModel()))
	bb := state.NewBindingsBuilder(simState)
	_ = simState.CreateInstance(orderKey, object.NewRecord())

	bindings := bb.BuildWithClassInstances(catalog.ClassNameMap())
	orderVal, ok := bindings.GetValue("Order")
	s.True(ok)
	orderSet, ok := orderVal.(*object.Set)
	s.True(ok)
	s.Equal(1, orderSet.Size())

	itemVal, ok := bindings.GetValue("Item")
	s.True(ok, "out-of-scope class must bind as empty set, not be missing")
	itemSet, ok := itemVal.(*object.Set)
	s.True(ok)
	s.Equal(0, itemSet.Size())
}

func (s *OutOfScopeProtocolSuite) TestObjectParameterSamplesEmptySet() {
	binder := actions.NewParameterBinder()
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := helper.Must(identity.NewActionKey(classKey, "init"))
	objKey := "item"
	dt := &model_data_type.DataType{
		CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.CONSTRAINT_TYPE_OBJECT,
			ObjectClassKey: &objKey,
		},
	}
	param, err := model_state.NewParameter(actionKey, "Peer", "object of item", false)
	s.Require().NoError(err)
	param.DataType = dt

	rng := rand.New(rand.NewSource(1)) //nolint:gosec
	values := binder.GenerateRandomParameters([]model_state.Parameter{param}, rng)
	s.Require().Contains(values, "Peer")
	// EMPTY_SET is a shared empty *object.Set.
	got := values["Peer"]
	set, ok := got.(*object.Set)
	s.True(ok, "expected empty set, got %T", got)
	s.Equal(0, set.Size())
	s.True(got == evaluator.EMPTY_SET || set.Size() == 0)
}

func (s *OutOfScopeProtocolSuite) TestSetAddToOutOfScopePeerIsNoOp() {
	orderClass, orderKey, itemClass, itemKey, assocKey, assoc := s.orderItemAssocModel()
	full := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	full.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	// Item uses create event name "create" (testItemClass), not _new — set-add matches PeerCreationEvent.
	// Use EventNameNew only if peer has that event. testItemClass has "create" as creation event.
	// MatchAssociationSetAddExpr needs EventCall with peer creation event key.
	createEvent, ok := NewClassCatalog(testModel(classEntry(itemClass, itemKey))).GetCreationEvent(itemKey)
	s.Require().True(ok)
	expr := &me.SetOp{
		Op:   me.SetUnion,
		Left: &me.AssociationRef{AssociationKey: assocKey},
		Right: &me.SetLiteral{
			Elements: []me.Expression{&me.EventCall{EventKey: createEvent.Key}},
		},
	}
	actionKey := helper.Must(identity.NewActionKey(orderKey, "add_line"))
	guarKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	logic := model_logic.NewLogic(
		guarKey,
		model_logic.LogicTypeStateChange,
		"create peer",
		"Lines",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Lines \\union {_new()}"},
		nil,
	)
	logic.Spec.Expression = expr
	action := model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "AddLine", Details: ""},
		nil,
		[]model_logic.Logic{logic},
		nil,
		nil,
	)

	active := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(active)
	catalog.RegisterOutOfScopeMetadata(full)

	simState := instance.NewState(schema.New(schema.EmptyModel()))
	bb := state.NewBindingsBuilder(simState)
	registerCatalogAssociations(catalog, bb)
	orderInst := simState.CreateInstance(orderKey, object.NewRecord())

	ae := actions.NewActionExecutor(
		bb,
		actions.InvariantRuntimeCheckers{},
		nil,
		actions.NewGuardEvaluator(bb),
		catalog,
		rand.New(rand.NewSource(1)), //nolint:gosec
	)
	result, err := ae.ExecuteAction(action, orderInst, nil)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Empty(simState.InstancesByClass(itemKey))
	s.Empty(simState.GetLinkedForward(orderInst.ID, assocKey))
}

func (s *OutOfScopeProtocolSuite) TestReverseStateChangeToOutOfScopePeerIsNoOp() {
	// Peer (from) defines Item (to). Item is in scope; Peer is out of scope.
	// Reverse link on Item: _Defines = {peerParam} with empty peer param → no-op.
	peerClass, peerKey := simpleCreateClass("peer_def", "PeerDef")
	itemClass, itemKey := testItemClass()
	assocKey := testAssocKey(peerKey, itemKey, "Defines")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Defines", Details: ""},
		model_class.AssociationEnd{ClassKey: peerKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult},
		model_class.AssociationOptions{},
	)
	full := testModel(classEntry(peerClass, peerKey), classEntry(itemClass, itemKey))
	full.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	// Guarantee: _Defines = {} (empty set) — reverse navigable, peer OOS.
	actionKey := helper.Must(identity.NewActionKey(itemKey, "init"))
	guarKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	expr := &me.SetLiteral{Elements: nil}
	logic := model_logic.NewLogic(
		guarKey,
		model_logic.LogicTypeStateChange,
		"link parent",
		model_class.ReverseAssociationTLAFieldName("Defines"),
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "{}"},
		nil,
	)
	logic.Spec.Expression = expr
	action := model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "Initialize", Details: ""},
		nil,
		[]model_logic.Logic{logic},
		nil,
		nil,
	)

	active := testModel(classEntry(itemClass, itemKey))
	catalog := NewClassCatalog(active)
	catalog.RegisterOutOfScopeMetadata(full)

	simState := instance.NewState(schema.New(schema.EmptyModel()))
	bb := state.NewBindingsBuilder(simState)
	registerCatalogAssociations(catalog, bb)
	itemInst := simState.CreateInstance(itemKey, object.NewRecord())

	ae := actions.NewActionExecutor(
		bb,
		actions.InvariantRuntimeCheckers{},
		nil,
		actions.NewGuardEvaluator(bb),
		catalog,
		rand.New(rand.NewSource(1)), //nolint:gosec
	)
	result, err := ae.ExecuteAction(action, itemInst, nil)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Empty(simState.GetLinkedReverse(itemInst.ID, assocKey))
	s.Empty(simState.InstancesByClass(peerKey))
}

func (s *OutOfScopeProtocolSuite) TestEngineWithSurface_RunsWithOutOfScopePeerClass() {
	// Use Item (no attribute guarantees) so the run only exercises scope wiring.
	itemClass, itemKey := testItemClass()
	peerClass, peerKey := simpleCreateClass("peer_def", "PeerDef")
	assocKey := testAssocKey(peerKey, itemKey, "Defines")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Defines", Details: ""},
		model_class.AssociationEnd{ClassKey: peerKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult},
		model_class.AssociationOptions{},
	)
	full := testModel(classEntry(itemClass, itemKey), classEntry(peerClass, peerKey))
	full.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	eng, err := NewSimulationEngine(full, SimulationConfig{
		MaxSteps:   5,
		RandomSeed: 42,
		Surface: &surface.SurfaceSpecification{
			IncludeClasses: []identity.Key{itemKey},
		},
	})
	s.Require().NoError(err)
	s.NotNil(eng)
	s.True(eng.catalog.IsClassInScope(itemKey))
	s.False(eng.catalog.IsClassInScope(peerKey))
	s.Contains(eng.catalog.ClassNameMap(), peerKey)

	result, err := eng.Run()
	s.Require().NoError(err)
	s.NotNil(result)
	s.False(result.FinalState.HasInstanceOfClass(peerKey))
}
