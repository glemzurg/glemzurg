package engine

import (
	"fmt"
	"maps"
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AssociationPeerEffectsSuite struct {
	suite.Suite
}

func TestAssociationPeerEffectsSuite(t *testing.T) {
	suite.Run(t, new(AssociationPeerEffectsSuite))
}

func (s *AssociationPeerEffectsSuite) TestInlineDestroyGuaranteeRemovesPlainAssociationLink() {
	fix := buildPlainAssocPeerFixture(true)
	simState, ae := s.buildPeerEffectExecutor(fix.model)

	orderInst := s.createPeerEffectInstance(simState, fix.orderKey, "Open")
	itemInst := s.createPeerEffectInstance(simState, fix.itemKey, "Active")
	simState.AddLink(fix.assocKey, orderInst.ID, itemInst.ID)

	action := peerInlineDestroyGuaranteeAction(fix.orderKey, fix.assocKey, fix.itemKey, "OrderItem")
	result, err := ae.ExecuteAction(action, orderInst, nil)
	s.Require().NoError(err)
	s.Empty(result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable))
	s.Require().Len(result.PeerTransitions, 1)
	s.Equal(model_state.EventNameDestroy, result.PeerTransitions[0].EventName)
	s.True(result.PeerTransitions[0].Result.WasDestroy)
	s.Nil(simState.GetInstance(itemInst.ID))
	s.Empty(simState.GetLinkedForward(orderInst.ID, fix.assocKey))
}

func (s *AssociationPeerEffectsSuite) TestSetMapDeleteRemovesPlainAssociationLink() {
	fix := buildPlainAssocPeerFixture(true)
	simState, ae := s.buildPeerEffectExecutor(fix.model)

	orderInst := s.createPeerEffectInstance(simState, fix.orderKey, "Open")
	itemInst := s.createPeerEffectInstance(simState, fix.itemKey, "Active")
	simState.AddLink(fix.assocKey, orderInst.ID, itemInst.ID)

	action := peerDestroyGuaranteeAction(fix.orderKey, fix.assocKey, fix.itemKey, "OrderItem")
	result, err := ae.ExecuteAction(action, orderInst, nil)
	s.Require().NoError(err)
	s.Empty(result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable))
	s.Require().Len(result.PeerTransitions, 1)
	s.Equal(model_state.EventNameDestroy, result.PeerTransitions[0].EventName)
	s.Nil(simState.GetInstance(itemInst.ID))
	s.Empty(simState.GetLinkedForward(orderInst.ID, fix.assocKey))
}

func (s *AssociationPeerEffectsSuite) TestSetMapDeleteViolationWhenPeerLacksDelete() {
	fix := buildPlainAssocPeerFixture(false)
	simState, ae := s.buildPeerEffectExecutor(fix.model)

	orderInst := s.createPeerEffectInstance(simState, fix.orderKey, "Open")
	itemInst := s.createPeerEffectInstance(simState, fix.itemKey, "Active")
	simState.AddLink(fix.assocKey, orderInst.ID, itemInst.ID)

	action := peerDestroyGuaranteeAction(fix.orderKey, fix.assocKey, fix.itemKey, "OrderItem")
	result, err := ae.ExecuteAction(action, orderInst, nil)
	s.Require().NoError(err)
	s.False(result.Success)
	s.Require().Len(result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable), 1)
	s.Equal(fmt.Sprintf("%d", itemInst.ID), result.Violations[0].ActualValue)
	s.NotNil(simState.GetInstance(itemInst.ID))
	s.Len(simState.GetLinkedForward(orderInst.ID, fix.assocKey), 1)
}

func (s *AssociationPeerEffectsSuite) TestSetAddCreatesAssociationClassRow() {
	tcm := buildAssociationClassPeerEffectModel()
	simState, ae := s.buildPeerEffectExecutor(tcm.model)
	registerCatalogAssociations(NewClassCatalog(tcm.model), state.NewBindingsBuilder(simState))

	partnerInst := s.createPeerEffectInstance(simState, tcm.partnerKey, "Active")
	action := peerNewSetAddAction(tcm.partnerKey, tcm.hostAssocKey, tcm.jurisdictionKey, "Configures")
	result, err := ae.ExecuteAction(action, partnerInst, nil)
	s.Require().NoError(err)
	s.Empty(result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable))
	s.Require().Len(result.PeerTransitions, 2)
	s.Equal("create", result.PeerTransitions[0].EventName)
	s.Equal("Add", result.PeerTransitions[1].EventName)

	links := simState.AssociationLinksFromEndpoint(tcm.hostAssocKey, partnerInst.ID)
	s.Len(links, 1)
	s.NotNil(simState.GetInstance(links[0].ToEndpointID))
	s.NotNil(simState.GetInstance(links[0].LinkInstanceID))
}

func (s *AssociationPeerEffectsSuite) TestSetMapDeleteRemovesAssociationClassRow() {
	tcm := buildAssociationClassPeerEffectModel()
	simState, ae := s.buildPeerEffectExecutor(tcm.model)
	registerCatalogAssociations(NewClassCatalog(tcm.model), state.NewBindingsBuilder(simState))

	partnerInst := s.createPeerEffectInstance(simState, tcm.partnerKey, "Active")
	jurisdictionInst := s.createPeerEffectInstance(simState, tcm.jurisdictionKey, "Active")

	linkDefClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.linkDefKey]
	addEvent := linkDefClass.Events[mustKey("domain/d/subdomain/s/class/link_def/event/add")]
	addResult, err := ae.ExecuteTransition(
		linkDefClass, addEvent, nil, nil,
		actions.CreationLinkSource{SourceAssocKey: &tcm.hostAssocKey, SourceID: &partnerInst.ID},
		&jurisdictionInst.ID,
	)
	s.Require().NoError(err)

	action := peerDestroyGuaranteeAction(tcm.partnerKey, tcm.hostAssocKey, tcm.jurisdictionKey, "Configures")
	result, err := ae.ExecuteAction(action, partnerInst, nil)
	s.Require().NoError(err)
	s.Empty(result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable))
	s.Require().GreaterOrEqual(len(result.PeerTransitions), 2)

	s.Nil(simState.GetInstance(jurisdictionInst.ID))
	s.Empty(simState.AssociationLinksFromEndpoint(tcm.hostAssocKey, partnerInst.ID))
	s.Nil(simState.GetInstance(addResult.InstanceID))
}

func (s *AssociationPeerEffectsSuite) buildPeerEffectExecutor(model *core.Model) (*state.SimulationState, *actions.ActionExecutor) {
	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	catalog := NewClassCatalog(model)
	registerCatalogAssociations(catalog, bb)
	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed
	ae := actions.NewActionExecutor(bb, actions.InvariantRuntimeCheckers{Checker: nil, DataType: nil}, nil, ge, catalog, rng)
	return simState, ae
}

func (s *AssociationPeerEffectsSuite) createPeerEffectInstance(
	simState *state.SimulationState,
	classKey identity.Key,
	stateName string,
) *state.ClassInstance {
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString(stateName))
	return simState.CreateInstance(classKey, attrs)
}

type plainAssocPeerFixture struct {
	model    *core.Model
	orderKey identity.Key
	itemKey  identity.Key
	assocKey identity.Key
}

func buildPlainAssocPeerFixture(itemHasDelete bool) plainAssocPeerFixture {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClassWithOptionalDelete(itemHasDelete)
	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "OrderItem", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult},
		model_class.Multiplicity{},
		model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	return plainAssocPeerFixture{model: model, orderKey: orderKey, itemKey: itemKey, assocKey: assocKey}
}

func testItemClassWithOptionalDelete(withDelete bool) (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/item/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(
		transCreateKey, eventCreateKey,
		model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey},
		model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "",
	)

	events := map[identity.Key]model_state.Event{eventCreateKey: eventCreate}
	states := map[identity.Key]model_state.State{stateActiveKey: stateActive}
	transitions := map[identity.Key]model_state.Transition{transCreateKey: transCreate}

	if withDelete {
		eventDeleteKey := helper.Must(identity.NewEventKey(classKey, model_state.EventNameDestroy))
		transDeleteKey := helper.Must(identity.NewTransitionKey(classKey, "active", model_state.EventNameDestroy, "", "", ""))
		eventDelete := model_state.NewEvent(eventDeleteKey, model_state.EventNameDestroy, "", nil)
		transDelete := model_state.NewTransition(
			transDeleteKey, eventDeleteKey,
			model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: nil},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "",
		)
		events[eventDeleteKey] = eventDelete
		transitions[transDeleteKey] = transDelete
	}

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Item", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(states)
	class.SetEvents(events)
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(transitions)
	return class, classKey
}

type associationClassPeerEffectModel struct {
	model           *core.Model
	partnerKey      identity.Key
	jurisdictionKey identity.Key
	linkDefKey      identity.Key
	hostAssocKey    identity.Key
}

func buildAssociationClassPeerEffectModel() associationClassPeerEffectModel {
	partnerClass, partnerKey := simpleCreateClass("partner", "Partner")
	jurisdictionClass, jurisdictionKey := testJurisdictionClassWithDelete()
	linkDefClass, linkDefKey := testLinkDefClassWithFinalDelete()

	hostAssocKey := testAssocKey(partnerKey, jurisdictionKey, "Configures")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	hostAssoc := model_class.NewAssociation(
		hostAssocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: toMult},
		model_class.Multiplicity{},
		model_class.AssociationOptions{AssociationClassKey: &linkDefKey, UmlComment: ""},
	)

	model := testModel(
		classEntry(partnerClass, partnerKey),
		classEntry(jurisdictionClass, jurisdictionKey),
		classEntry(linkDefClass, linkDefKey),
	)
	model.ClassAssociations = map[identity.Key]model_class.Association{hostAssocKey: hostAssoc}

	return associationClassPeerEffectModel{
		model: model, partnerKey: partnerKey, jurisdictionKey: jurisdictionKey,
		linkDefKey: linkDefKey, hostAssocKey: hostAssocKey,
	}
}

func testJurisdictionClassWithDelete() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/jurisdiction")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/jurisdiction/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/jurisdiction/event/create")
	eventDeleteKey := helper.Must(identity.NewEventKey(classKey, model_state.EventNameDestroy))
	transCreateKey := mustKey("domain/d/subdomain/s/class/jurisdiction/transition/create")
	transDeleteKey := helper.Must(identity.NewTransitionKey(classKey, "active", model_state.EventNameDestroy, "", "", ""))

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	eventDelete := model_state.NewEvent(eventDeleteKey, model_state.EventNameDestroy, "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")
	transDelete := model_state.NewTransition(transDeleteKey, eventDeleteKey, model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: nil}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Jurisdiction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate, eventDeleteKey: eventDelete})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate, transDeleteKey: transDelete})
	return class, classKey
}

func testLinkDefClassWithFinalDelete() (model_class.Class, identity.Key) {
	class, classKey := testLinkDefClass()
	eventDeleteKey := helper.Must(identity.NewEventKey(classKey, model_state.EventNameDestroy))
	transDeleteKey := helper.Must(identity.NewTransitionKey(classKey, "active", model_state.EventNameDestroy, "", "", ""))
	stateActiveKey := mustKey("domain/d/subdomain/s/class/link_def/state/active")
	eventDelete := model_state.NewEvent(eventDeleteKey, model_state.EventNameDestroy, "", nil)
	transDelete := model_state.NewTransition(
		transDeleteKey, eventDeleteKey,
		model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: nil},
		model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "",
	)
	class.SetEvents(appendEvent(class.Events, eventDeleteKey, eventDelete))
	class.SetTransitions(appendTransition(class.Transitions, transDeleteKey, transDelete))
	return class, classKey
}

func appendEvent(events map[identity.Key]model_state.Event, key identity.Key, event model_state.Event) map[identity.Key]model_state.Event {
	out := make(map[identity.Key]model_state.Event, len(events)+1)
	maps.Copy(out, events)
	out[key] = event
	return out
}

func appendTransition(transitions map[identity.Key]model_state.Transition, key identity.Key, transition model_state.Transition) map[identity.Key]model_state.Transition {
	out := make(map[identity.Key]model_state.Transition, len(transitions)+1)
	maps.Copy(out, transitions)
	out[key] = transition
	return out
}

func TestSetMapUpdateViolationWhenEventParamsOmitted(t *testing.T) {
	fix := buildPlainAssocPeerFixtureWithUpdate()
	suite := new(AssociationPeerEffectsSuite)
	simState, ae := suite.buildPeerEffectExecutor(fix.model)

	orderInst := suite.createPeerEffectInstance(simState, fix.orderKey, "Open")
	itemInst := suite.createPeerEffectInstance(simState, fix.itemKey, "Active")
	simState.AddLink(fix.assocKey, orderInst.ID, itemInst.ID)

	action := peerUpdateSetMapAction(fix.orderKey, fix.assocKey, fix.itemKey, "OrderItem")
	result, err := ae.ExecuteAction(action, orderInst, nil)
	require.NoError(t, err)
	require.Len(t, result.Violations.ByType(invariants.ViolationTypePeerEventUnavailable), 1)
	require.Contains(t, result.Violations[0].Message, "parameter binding failed")
}

func TestSetMapUpdateRespectsArgumentOrder(t *testing.T) {
	fix := buildPlainAssocPeerFixtureWithUpdateAction()
	suite := new(AssociationPeerEffectsSuite)
	simState, ae := suite.buildPeerEffectExecutor(fix.model)

	orderInst := suite.createPeerEffectInstance(simState, fix.orderKey, "Open")
	itemInst := suite.createPeerEffectInstance(simState, fix.itemKey, "Active")
	simState.AddLink(fix.assocKey, orderInst.ID, itemInst.ID)

	action := peerUpdateSetMapActionWithArgOrder(fix.orderKey, fix.assocKey, fix.itemKey, "OrderItem", []string{"TopoffBalance", "MinimumBalance"})
	ownerParams := map[string]object.Object{
		"MinimumBalance": object.NewInteger(100),
		"TopoffBalance":  object.NewInteger(200),
	}
	result, err := ae.ExecuteAction(action, orderInst, ownerParams)
	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Len(t, result.PeerTransitions, 1)

	updated := simState.GetInstance(itemInst.ID)
	require.Equal(t, "200", updated.GetAttribute("minimum_balance").Inspect())
	require.Equal(t, "100", updated.GetAttribute("topoff_balance").Inspect())
}

func buildPlainAssocPeerFixtureWithUpdate() plainAssocPeerFixture {
	fix := buildPlainAssocPeerFixture(false)
	itemClass, itemKey := testItemClassWithUpdateEvent(false)
	fix.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[itemKey] = itemClass
	fix.itemKey = itemKey
	return fix
}

func buildPlainAssocPeerFixtureWithUpdateAction() plainAssocPeerFixture {
	fix := buildPlainAssocPeerFixtureWithUpdate()
	itemClass, _ := testItemClassWithUpdateEvent(true)
	fix.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[fix.itemKey] = itemClass
	return fix
}

func testItemClassWithUpdateEvent(withUpdateAction bool) (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	eventUpdateKey := mustKey("domain/d/subdomain/s/class/item/event/update")
	transCreateKey := mustKey("domain/d/subdomain/s/class/item/transition/create")
	transUpdateKey := mustKey("domain/d/subdomain/s/class/item/transition/update")
	attrMinKey := helper.Must(identity.NewAttributeKey(classKey, "minimum_balance"))
	attrTopKey := helper.Must(identity.NewAttributeKey(classKey, "topoff_balance"))

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	eventUpdate := model_state.NewEvent(eventUpdateKey, "Update", "", []string{"MinimumBalance", "TopoffBalance"})
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")

	transitions := map[identity.Key]model_state.Transition{
		transCreateKey: model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{}, ""),
	}
	actions := map[identity.Key]model_state.Action{}
	if withUpdateAction {
		actionUpdateKey := helper.Must(identity.NewActionKey(classKey, "update"))
		guarMinKey := helper.Must(identity.NewActionGuaranteeKey(actionUpdateKey, "0"))
		guarTopKey := helper.Must(identity.NewActionGuaranteeKey(actionUpdateKey, "1"))
		guarMin := model_logic.NewLogic(guarMinKey, model_logic.LogicTypeStateChange, "", "minimum_balance", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "MinimumBalance"}, nil)
		guarMin.Spec.Expression = &me.LocalVar{Name: "MinimumBalance"}
		guarTop := model_logic.NewLogic(guarTopKey, model_logic.LogicTypeStateChange, "", "topoff_balance", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TopoffBalance"}, nil)
		guarTop.Spec.Expression = &me.LocalVar{Name: "TopoffBalance"}
		updateAction := model_state.NewAction(actionUpdateKey, model_state.ActionDetails{Name: "Update", Details: ""}, nil, []model_logic.Logic{guarMin, guarTop}, nil, nil)
		actions[actionUpdateKey] = updateAction
		transitions[transUpdateKey] = model_state.NewTransition(
			transUpdateKey, eventUpdateKey,
			model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: &stateActiveKey},
			model_state.TransitionLogicKeys{ActionKey: &actionUpdateKey}, "",
		)
	} else {
		transitions[transUpdateKey] = model_state.NewTransition(
			transUpdateKey, eventUpdateKey,
			model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: &stateActiveKey},
			model_state.TransitionLogicKeys{}, "",
		)
	}

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item", Details: ""})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(attrMinKey, model_class.AttributeDetails{Name: "minimum_balance", Details: ""}, "Nat", nil, false, model_class.AttributeAnnotations{})),
		helper.Must(model_class.NewAttribute(attrTopKey, model_class.AttributeDetails{Name: "topoff_balance", Details: ""}, "Nat", nil, false, model_class.AttributeAnnotations{})),
	})
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate, eventUpdateKey: eventUpdate})
	class.SetActions(actions)
	class.SetTransitions(transitions)
	return class, classKey
}

func peerUpdateSetMapAction(ownerKey, assocKey, peerClassKey identity.Key, assocTLAField string) model_state.Action {
	return peerUpdateSetMapActionWithArgOrder(ownerKey, assocKey, peerClassKey, assocTLAField, nil)
}

func peerUpdateSetMapActionWithArgOrder(
	ownerKey, assocKey, peerClassKey identity.Key,
	assocTLAField string,
	argOrder []string,
) model_state.Action {
	updateEventKey := helper.Must(identity.NewEventKey(peerClassKey, "Update"))
	args := []me.Expression{&me.LocalVar{Name: "r"}}
	for _, name := range argOrder {
		args = append(args, &me.LocalVar{Name: name})
	}
	expr := &me.SetMap{
		Variable: "r",
		Set:      &me.AssociationRef{AssociationKey: assocKey},
		Transform: &me.EventCall{
			EventKey: updateEventKey,
			Args:     args,
		},
	}
	return peerEffectAction(ownerKey, assocTLAField, expr)
}

func peerInlineDestroyGuaranteeAction(ownerKey, assocKey, peerClassKey identity.Key, assocTLAField string) model_state.Action {
	actionKey := helper.Must(identity.NewActionKey(ownerKey, "peer_inline_destroy"))
	deleteKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	deleteEventKey := helper.Must(identity.NewEventKey(peerClassKey, model_state.EventNameDestroy))

	selection := &me.SetFilter{
		Variable:  "b",
		Set:       &me.AssociationRef{AssociationKey: assocKey},
		Predicate: &me.BoolLiteral{Value: true},
	}
	inlineExpr := &me.SetOp{
		Op:    me.SetDifference,
		Left:  &me.AssociationRef{AssociationKey: assocKey},
		Right: selection,
	}
	deleteGuar := model_logic.NewLogic(
		deleteKey,
		model_logic.LogicTypeDestroy,
		"",
		assocTLAField,
		logic_spec.ExpressionSpec{Expression: inlineExpr},
		nil,
	)
	deleteGuar.DestroyEventSpec.Expression = &me.EventCall{
		EventKey: deleteEventKey,
		Args:     []me.Expression{&me.LocalVar{Name: "item"}},
	}

	return model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "PeerInlineDestroy", Details: ""},
		nil,
		[]model_logic.Logic{deleteGuar},
		nil,
		nil,
	)
}

func peerDestroyGuaranteeAction(ownerKey, assocKey, peerClassKey identity.Key, assocTLAField string) model_state.Action {
	actionKey := helper.Must(identity.NewActionKey(ownerKey, "peer_destroy"))
	letKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	stateKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "1"))
	deleteKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "2"))
	deleteEventKey := helper.Must(identity.NewEventKey(peerClassKey, model_state.EventNameDestroy))

	toDelete := &me.SetFilter{
		Variable:  "r",
		Set:       &me.AssociationRef{AssociationKey: assocKey},
		Predicate: &me.BoolLiteral{Value: true},
	}
	letGuar := model_logic.NewLogic(letKey, model_logic.LogicTypeLet, "", "ToDelete", logic_spec.ExpressionSpec{}, nil)
	letGuar.Spec.Expression = toDelete

	stateExpr := &me.SetOp{
		Op:    me.SetDifference,
		Left:  &me.AssociationRef{AssociationKey: assocKey},
		Right: &me.LocalVar{Name: "ToDelete"},
	}
	stateGuar := model_logic.NewLogic(stateKey, model_logic.LogicTypeStateChange, "", assocTLAField, logic_spec.ExpressionSpec{}, nil)
	stateGuar.Spec.Expression = stateExpr

	deleteSelection := &me.SetFilter{
		Variable:  "b",
		Set:       &me.AssociationRef{AssociationKey: assocKey},
		Predicate: &me.BoolLiteral{Value: true},
	}
	deleteGuar := model_logic.NewLogic(
		deleteKey,
		model_logic.LogicTypeDestroy,
		"",
		assocTLAField,
		logic_spec.ExpressionSpec{Expression: deleteSelection},
		nil,
	)
	deleteGuar.DestroyEventSpec.Expression = &me.EventCall{
		EventKey: deleteEventKey,
		Args:     []me.Expression{&me.LocalVar{Name: "b"}},
	}

	return model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "PeerDelete", Details: ""},
		nil,
		[]model_logic.Logic{letGuar, stateGuar, deleteGuar},
		nil,
		nil,
	)
}

func peerNewSetAddAction(ownerKey, assocKey, peerClassKey identity.Key, assocTLAField string) model_state.Action {
	newEventKey := helper.Must(identity.NewEventKey(peerClassKey, model_state.EventNameNew))
	expr := &me.SetOp{
		Op:   me.SetUnion,
		Left: &me.AssociationRef{AssociationKey: assocKey},
		Right: &me.SetLiteral{
			Elements: []me.Expression{&me.EventCall{EventKey: newEventKey}},
		},
	}
	return peerEffectAction(ownerKey, assocTLAField, expr)
}

func peerEffectParamWithNatTypeSpec(actionKey identity.Key, name string) model_state.Parameter {
	param := helper.Must(model_state.NewParameter(actionKey, name, "Nat", false))
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	if param.DataType == nil {
		param.DataType = &model_data_type.DataType{
			CollectionType: "atomic",
			Atomic:         &model_data_type.Atomic{ConstraintType: model_data_type.CONSTRAINT_TYPE_UNCONSTRAINED},
		}
	}
	param.DataType.TypeSpec = &natTypeSpec
	return param
}

func peerEffectAction(ownerKey identity.Key, target string, expr me.Expression) model_state.Action {
	actionKey := helper.Must(identity.NewActionKey(ownerKey, "peer_effect"))
	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))
	logic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "", target, logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}, nil)
	logic.Spec.Expression = expr
	minBal := peerEffectParamWithNatTypeSpec(actionKey, "MinimumBalance")
	topoff := peerEffectParamWithNatTypeSpec(actionKey, "TopoffBalance")
	return model_state.NewAction(actionKey, model_state.ActionDetails{Name: "PeerEffect", Details: ""}, nil, []model_logic.Logic{logic}, nil, []model_state.Parameter{minBal, topoff})
}
