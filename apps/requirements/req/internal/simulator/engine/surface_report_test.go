package engine

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	catalog := NewClassCatalog(model)
	report := BuildSurfaceReport(catalog)

	s.Require().Len(report.Classes, 2)

	orderEntry := findSurfaceClass(report, orderKey.String())
	s.Require().NotNil(orderEntry)
	s.Equal("simulatable", orderEntry.Role)
	s.Require().Len(orderEntry.CreationEvents, 1)
	s.Equal("create", orderEntry.CreationEvents[0].EventName)
	s.Require().Len(orderEntry.States, 1)
	s.Equal("Open", orderEntry.States[0].StateName)
	s.Require().Len(orderEntry.States[0].Events, 1)
	s.Equal("close", orderEntry.States[0].Events[0].EventName)
	s.Equal("DoClose", orderEntry.States[0].Events[0].ActionName)
	s.Require().Len(orderEntry.Queries, 1)
	s.Equal("balance", orderEntry.Queries[0].QueryName)

	itemEntry := findSurfaceClass(report, itemKey.String())
	s.Require().NotNil(itemEntry)
	s.Empty(itemEntry.CreationEvents, "mandatory association target excludes external creation")
}

func (s *SurfaceReportSuite) TestBuildSurfaceReportIncludesLivenessOnlyClasses() {
	classKey := mustKey("domain/d/subdomain/s/class/simple")
	simpleClass := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Simple", Details: "", UnfinishedNotes: "", UmlComment: ""})
	simpleClass.SetAttributes(nil)
	simpleClass.SetStates(map[identity.Key]model_state.State{})
	simpleClass.SetEvents(map[identity.Key]model_state.Event{})
	simpleClass.SetGuards(map[identity.Key]model_state.Guard{})
	simpleClass.SetActions(map[identity.Key]model_state.Action{})
	simpleClass.SetQueries(map[identity.Key]model_state.Query{})
	simpleClass.SetTransitions(map[identity.Key]model_state.Transition{})

	catalog := NewClassCatalog(testModel(classEntry(simpleClass, classKey)))
	report := BuildSurfaceReport(catalog)

	s.Require().Len(report.Classes, 1)
	s.Equal("liveness_only", report.Classes[0].Role)
}

func (s *SurfaceReportSuite) TestFormatTextIncludesClassAndSurfaceEntries() {
	orderClass, orderKey := testOrderClass()
	catalog := NewClassCatalog(testModel(classEntry(orderClass, orderKey)))
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
	if !strings.Contains(text, "(empty)") {
		t.Fatalf("expected empty surface marker, got: %q", text)
	}
}
