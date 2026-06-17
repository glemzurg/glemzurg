package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type AssociationClassSuite struct {
	suite.Suite
}

func TestAssociationClassSuite(t *testing.T) {
	suite.Run(t, new(AssociationClassSuite))
}

type acTestModel struct {
	model           *core.Model
	partnerKey      identity.Key
	jurisdictionKey identity.Key
	linkDefKey      identity.Key
	hostAssocKey    identity.Key
}

func buildAssociationClassTestModel() *acTestModel {
	partnerClass, partnerKey := testPartnerClass()
	jurisdictionClass, jurisdictionKey := testJurisdictionClass()
	linkDefClass, linkDefKey := testLinkDefClass()

	hostAssocKey := testAssocKey(partnerKey, jurisdictionKey, "Configures")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	hostAssoc := model_class.NewAssociation(
		hostAssocKey,
		"Configures",
		"",
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: toMult},
		&linkDefKey,
		"",
	)

	m := testModel(
		classEntry(partnerClass, partnerKey),
		classEntry(jurisdictionClass, jurisdictionKey),
		classEntry(linkDefClass, linkDefKey),
	)
	m.ClassAssociations = map[identity.Key]model_class.Association{
		hostAssocKey: hostAssoc,
	}

	return &acTestModel{
		model:           m,
		partnerKey:      partnerKey,
		jurisdictionKey: jurisdictionKey,
		linkDefKey:      linkDefKey,
		hostAssocKey:    hostAssocKey,
	}
}

func testPartnerClass() (model_class.Class, identity.Key) {
	return simpleCreateClass("partner", "Partner")
}

func testJurisdictionClass() (model_class.Class, identity.Key) {
	return simpleCreateClass("jurisdiction", "Jurisdiction")
}

func testLinkDefClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/link_def")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/link_def/state/active")
	stateDeletedKey := mustKey("domain/d/subdomain/s/class/link_def/state/deleted")
	eventAddKey := mustKey("domain/d/subdomain/s/class/link_def/event/add")
	eventUpdateKey := mustKey("domain/d/subdomain/s/class/link_def/event/update")
	eventDeleteKey := mustKey("domain/d/subdomain/s/class/link_def/event/delete")
	transAddKey := mustKey("domain/d/subdomain/s/class/link_def/transition/add")
	transUpdateKey := mustKey("domain/d/subdomain/s/class/link_def/transition/update")
	transDeleteKey := mustKey("domain/d/subdomain/s/class/link_def/transition/delete")

	eventAdd := model_state.NewEvent(eventAddKey, "Add", "", nil)
	eventUpdate := model_state.NewEvent(eventUpdateKey, "Update", "", nil)
	eventDelete := model_state.NewEvent(eventDeleteKey, "Delete", "", nil)

	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	stateDeleted := model_state.NewState(stateDeletedKey, "Deleted", "", "")

	transAdd := model_state.NewTransition(transAddKey, nil, eventAddKey, nil, nil, &stateActiveKey, "")
	transUpdate := model_state.NewTransition(transUpdateKey, &stateActiveKey, eventUpdateKey, nil, nil, &stateActiveKey, "")
	transDelete := model_state.NewTransition(transDeleteKey, &stateActiveKey, eventDeleteKey, nil, nil, &stateDeletedKey, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "LinkDef", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey:  stateActive,
		stateDeletedKey: stateDeleted,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventAddKey:    eventAdd,
		eventUpdateKey: eventUpdate,
		eventDeleteKey: eventDelete,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transAddKey:    transAdd,
		transUpdateKey: transUpdate,
		transDeleteKey: transDelete,
	})

	return class, classKey
}

func simpleCreateClass(subKey, name string) (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/" + subKey)
	stateActiveKey := mustKey("domain/d/subdomain/s/class/" + subKey + "/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/" + subKey + "/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/" + subKey + "/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: name, Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})

	return class, classKey
}

func (s *AssociationClassSuite) TestCatalogIndexesAssociationClass() {
	tcm := buildAssociationClassTestModel()
	catalog := NewClassCatalog(tcm.model)

	s.True(catalog.IsAssociationClass(tcm.linkDefKey))
	s.True(catalog.IsAssociationClassHost(tcm.hostAssocKey))

	acInfo := catalog.LookupAssociationClass(tcm.linkDefKey)
	s.Require().NotNil(acInfo)
	s.Equal(tcm.partnerKey, acInfo.FromClassKey)
	s.Equal(tcm.jurisdictionKey, acInfo.ToClassKey)
	s.Equal(tcm.hostAssocKey, acInfo.HostAssociation.Key)
	s.True(acInfo.InactiveStates["Deleted"])

	assocs := catalog.AllAssociations()
	s.Len(assocs, 1)
	s.Equal(tcm.hostAssocKey, assocs[0].Association.Key)
	s.Require().NotNil(assocs[0].Association.AssociationClassKey)

	s.Empty(catalog.ExternalCreationEvents(tcm.linkDefKey))
}

func (s *AssociationClassSuite) TestAssociationClassAddCreatesNativeHostLink() {
	tcm := buildAssociationClassTestModel()
	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	catalog := NewClassCatalog(tcm.model)
	registerCatalogAssociations(catalog, bb)

	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed
	ae := actions.NewActionExecutor(bb, nil, nil, nil, ge, catalog, rng)

	partnerClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.partnerKey]
	jurisdictionClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.jurisdictionKey]
	linkDefClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.linkDefKey]

	partnerEvent := partnerClass.Events[mustKey("domain/d/subdomain/s/class/partner/event/create")]
	jurisdictionEvent := jurisdictionClass.Events[mustKey("domain/d/subdomain/s/class/jurisdiction/event/create")]
	addEvent := linkDefClass.Events[mustKey("domain/d/subdomain/s/class/link_def/event/add")]

	partnerResult, err := ae.ExecuteTransition(partnerClass, partnerEvent, nil, nil, nil, nil, nil)
	s.Require().NoError(err)
	jurisdictionResult, err := ae.ExecuteTransition(jurisdictionClass, jurisdictionEvent, nil, nil, nil, nil, nil)
	s.Require().NoError(err)

	acInfo := catalog.LookupAssociationClass(tcm.linkDefKey)
	hostAssocKey := acInfo.HostAssociation.Key
	result, err := ae.ExecuteTransition(
		linkDefClass,
		addEvent,
		nil,
		nil,
		&hostAssocKey,
		&partnerResult.InstanceID,
		&jurisdictionResult.InstanceID,
	)
	s.Require().NoError(err)
	s.True(result.WasCreation)

	links := simState.AssociationLinksFromEndpoint(hostAssocKey, partnerResult.InstanceID)
	s.Len(links, 1)
	s.Equal(result.InstanceID, links[0].LinkInstanceID)
	s.Equal(partnerResult.InstanceID, links[0].FromEndpointID)
	s.Equal(jurisdictionResult.InstanceID, links[0].ToEndpointID)
	s.Equal(hostAssocKey, links[0].HostAssocKey)

	s.Empty(simState.GetLinkedForward(partnerResult.InstanceID, hostAssocKey))
}

func (s *AssociationClassSuite) TestAssociationClassAddRequiresEndpoints() {
	tcm := buildAssociationClassTestModel()
	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	catalog := NewClassCatalog(tcm.model)
	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed
	ae := actions.NewActionExecutor(bb, nil, nil, nil, ge, catalog, rng)

	linkDefClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.linkDefKey]
	addEvent := linkDefClass.Events[mustKey("domain/d/subdomain/s/class/link_def/event/add")]

	_, err := ae.ExecuteTransition(linkDefClass, addEvent, nil, nil, nil, nil, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "requires both endpoint instances")
}

func (s *AssociationClassSuite) TestDeleteSoftDeleteExcludesFromActiveCount() {
	tcm := buildAssociationClassTestModel()
	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	catalog := NewClassCatalog(tcm.model)
	registerCatalogAssociations(catalog, bb)

	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed
	ae := actions.NewActionExecutor(bb, nil, nil, nil, ge, catalog, rng)
	multChecker := NewMultiplicityChecker(catalog)

	linkDefClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.linkDefKey]
	partnerClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.partnerKey]
	jurisdictionClass := tcm.model.Domains[mustKey("domain/d")].Subdomains[testSubdomainKey()].Classes[tcm.jurisdictionKey]

	partnerResult, err := ae.ExecuteTransition(
		partnerClass,
		partnerClass.Events[mustKey("domain/d/subdomain/s/class/partner/event/create")],
		nil, nil, nil, nil, nil,
	)
	s.Require().NoError(err)
	jurisdictionResult, err := ae.ExecuteTransition(
		jurisdictionClass,
		jurisdictionClass.Events[mustKey("domain/d/subdomain/s/class/jurisdiction/event/create")],
		nil, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	acInfo := catalog.LookupAssociationClass(tcm.linkDefKey)
	hostAssocKey := acInfo.HostAssociation.Key
	addResult, err := ae.ExecuteTransition(
		linkDefClass,
		linkDefClass.Events[mustKey("domain/d/subdomain/s/class/link_def/event/add")],
		nil, nil, &hostAssocKey, &partnerResult.InstanceID, &jurisdictionResult.InstanceID,
	)
	s.Require().NoError(err)

	partnerInst := simState.GetInstance(partnerResult.InstanceID)
	violations := multChecker.CheckInstance(partnerInst, simState)
	s.Empty(violations)

	acInstance := simState.GetInstance(addResult.InstanceID)
	deleteEvent := linkDefClass.Events[mustKey("domain/d/subdomain/s/class/link_def/event/delete")]
	_, err = ae.ExecuteTransition(linkDefClass, deleteEvent, acInstance, nil, nil, nil, nil)
	s.Require().NoError(err)

	violations = multChecker.CheckInstance(partnerInst, simState)
	s.Require().Len(violations, 1)
	s.Contains(violations[0].Message, "expected at least 1 links, got 0")
}

func (s *AssociationClassSuite) TestSimulationRunsAssociationClassScenario() {
	tcm := buildAssociationClassTestModel()

	config := SimulationConfig{
		MaxSteps:   50,
		RandomSeed: 7,
	}

	engine, err := NewSimulationEngine(tcm.model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)
	s.Positive(result.StepsTaken)

	foundAdd := false
	foundDelete := false
	var walkSteps func(steps []*SimulationStep)
	walkSteps = func(steps []*SimulationStep) {
		for _, step := range steps {
			if step.ClassKey == tcm.linkDefKey && step.EventName == "Add" {
				foundAdd = true
			}
			if step.ClassKey == tcm.linkDefKey && step.EventName == "Delete" {
				foundDelete = true
			}
			if len(step.CascadedSteps) > 0 {
				walkSteps(step.CascadedSteps)
			}
		}
	}
	walkSteps(result.Steps)
	s.True(foundAdd, "simulation should exercise AC Add with bound endpoints")

	acInfo := NewClassCatalog(tcm.model).LookupAssociationClass(tcm.linkDefKey)
	s.Require().NotNil(acInfo)
	linkedHosts := result.FinalState.AssociationLinks().AllHostAssociationKeys()
	s.True(linkedHosts[evaluator.AssociationKey(acInfo.HostAssociation.Key.String())])

	_ = foundDelete
}
