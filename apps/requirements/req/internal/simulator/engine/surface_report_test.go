package engine

import (
	"math/big"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/suite"
)

type SurfaceReportSuite struct {
	suite.Suite
}

func TestSurfaceReportSuite(t *testing.T) {
	suite.Run(t, new(SurfaceReportSuite))
}

func (s *SurfaceReportSuite) TestBuildSurfaceReportListsClassesEventsQueriesAndDoActions() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()

	queryKey := helper.Must(identity.NewQueryKey(orderKey, "balance"))
	query := model_state.NewQuery(queryKey, "balance", "", nil, nil, nil)
	orderClass.SetQueries(map[identity.Key]model_state.Query{queryKey: query})

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	catalog := NewClassCatalog(schema.New(model))
	report := BuildSurfaceReport(catalog)

	// Item has no external drivers (mandatory peer of Order); surface lists drivers only.
	s.Require().Len(report.Classes, 1)

	orderEntry := findSurfaceClass(report, orderKey.String())
	s.Require().NotNil(orderEntry)
	s.Require().Len(orderEntry.CreationEvents, 1)
	s.Equal("create", orderEntry.CreationEvents[0].EventName)
	s.Require().Len(orderEntry.States, 1)
	s.Equal("Open", orderEntry.States[0].StateName)
	s.Require().Len(orderEntry.States[0].Events, 1)
	s.Equal("close", orderEntry.States[0].Events[0].EventName)
	s.Equal("DoClose", orderEntry.States[0].Events[0].ActionName)
	s.Require().Len(orderEntry.Queries, 1)
	s.Equal("balance", orderEntry.Queries[0].QueryName)

	s.Nil(findSurfaceClass(report, itemKey.String()), "peer-only class must not appear on the surface report")
}

func (s *SurfaceReportSuite) TestBuildSurfaceReportOmitsClassesWithNoExternalDrivers() {
	classKey := mustKey("domain/d/subdomain/s/class/simple")
	simpleClass := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Simple", Details: "", UnfinishedNotes: "", UmlComment: ""})
	simpleClass.SetAttributes(nil)
	simpleClass.SetStates(map[identity.Key]model_state.State{})
	simpleClass.SetEvents(map[identity.Key]model_state.Event{})
	simpleClass.SetGuards(map[identity.Key]model_state.Guard{})
	simpleClass.SetActions(map[identity.Key]model_state.Action{})
	simpleClass.SetQueries(map[identity.Key]model_state.Query{})
	simpleClass.SetTransitions(map[identity.Key]model_state.Transition{})

	catalog := NewClassCatalog(schema.New(testModel(classEntry(simpleClass, classKey))))
	report := BuildSurfaceReport(catalog)

	s.Empty(report.Classes, "liveness-only / peer-only classes are not surface drivers")
	s.Contains(report.FormatText(), "(empty)")
}

func (s *SurfaceReportSuite) TestBuildSurfaceReportListsExternalDerivedAttributes() {
	accountKey := mustKey("domain/finance/subdomain/wallet/class/account")
	balanceAttrKey := helper.Must(identity.NewAttributeKey(accountKey, "balance"))
	balanceDeriv := model_logic.NewLogic(
		mustKey("invariant/11"),
		model_logic.LogicTypeValue,
		"Constant.",
		"",
		logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: "0",
			Expression:    &me.IntLiteral{Value: big.NewInt(0)},
		},
		nil,
	)
	balanceAttr := helper.Must(model_class.NewAttribute(
		balanceAttrKey,
		model_class.AttributeDetails{Name: "balance", Details: ""},
		"",
		&balanceDeriv,
		false,
		model_class.AttributeAnnotations{},
	))

	stateKey := helper.Must(identity.NewStateKey(accountKey, "open"))
	createEventKey := helper.Must(identity.NewEventKey(accountKey, "create"))
	transKey := helper.Must(identity.NewTransitionKey(accountKey, "", "create", "", "", "open"))
	accountClass := model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"})
	accountClass.SetAttributes([]model_class.Attribute{balanceAttr})
	accountClass.SetStates(map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "Open", "", ""),
	})
	accountClass.SetEvents(map[identity.Key]model_state.Event{
		createEventKey: model_state.NewEvent(createEventKey, "create", "", nil),
	})
	accountClass.SetGuards(map[identity.Key]model_state.Guard{})
	accountClass.SetActions(map[identity.Key]model_state.Action{})
	accountClass.SetQueries(map[identity.Key]model_state.Query{})
	accountClass.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: model_state.NewTransition(
			transKey,
			createEventKey,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateKey},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil},
			"",
		),
	})

	model := testModel(classEntry(accountClass, accountKey))
	catalog := NewClassCatalog(schema.New(model))
	PopulateDerivedAttributeCallersFromSchema(schema.New(model), catalog)
	report := BuildSurfaceReport(catalog)

	entry := findSurfaceClass(report, accountKey.String())
	s.Require().NotNil(entry)
	s.Require().Len(entry.DerivedAttributes, 1)
	s.Equal("balance", entry.DerivedAttributes[0].AttributeName)
	s.Contains(BuildSurfaceReport(catalog).FormatText(), "derived: balance")
}

func (s *SurfaceReportSuite) TestFormatTextIncludesClassAndSurfaceEntries() {
	orderClass, orderKey := testOrderClass()
	catalog := NewClassCatalog(schema.New(testModel(classEntry(orderClass, orderKey))))
	text := BuildSurfaceReport(catalog).FormatText()

	s.Contains(text, "Simulation surface")
	s.Contains(text, orderKey.String())
	s.Contains(text, "Order")
	s.Contains(text, "creation: event create")
	s.Contains(text, "state Open:")
	s.Contains(text, "transition: event close (action DoClose)")
}

func findSurfaceClass(report *SurfaceReport, classKey string) *SurfaceClassReport {
	for i := range report.Classes {
		if report.Classes[i].ClassKey == classKey {
			return &report.Classes[i]
		}
	}
	return nil
}

func TestFormatTextEmptySurface(t *testing.T) {
	text := (&SurfaceReport{}).FormatText()
	if !strings.Contains(text, "Simulation scope") || !strings.Contains(text, "Simulation surface") {
		t.Fatalf("expected scope and surface sections, got: %q", text)
	}
	if !strings.Contains(text, "(empty)") {
		t.Fatalf("expected empty markers, got: %q", text)
	}
}

func (s *SurfaceReportSuite) TestFormatTextIncludesScopeBeforeSurface() {
	orderClass, orderKey := testOrderClass()
	catalog := NewClassCatalog(schema.New(testModel(classEntry(orderClass, orderKey))))
	report := BuildSurfaceReport(catalog)
	report.Scope = []surface.ScopeEntry{
		{Kind: surface.ScopeClass, Path: "d/s/order"},
	}
	text := report.FormatText()
	s.Contains(text, "Simulation scope")
	s.Contains(text, "class d/s/order")
	s.Contains(text, "Simulation surface")
	s.Greater(strings.Index(text, "Simulation surface"), strings.Index(text, "Simulation scope"))
}
