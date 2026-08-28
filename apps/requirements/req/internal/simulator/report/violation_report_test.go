package report

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"github.com/stretchr/testify/suite"
)

func TestViolationReportSuite(t *testing.T) {
	suite.Run(t, new(ViolationReportSuite))
}

type ViolationReportSuite struct {
	suite.Suite
}

func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

func (s *ViolationReportSuite) TestEmptyViolations() {
	report := FromViolations(nil)

	s.Equal(0, report.TotalCount)
	s.False(report.HasViolations())
	s.Empty(report.Categories)
	s.Equal("No violations found.", report.Summary)
}

func reportTestModel(class model_class.Class, classKey identity.Key) *core.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	return &model
}

// Violations are produced only via instance Check*/Package* APIs (no Report* constructors).

func tlaViolations() instance.ViolationErrors {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := helper.Must(identity.NewActionKey(classKey, "do"))
	return instance.PackageActionRequireFailures(
		actionKey, "Do",
		[]instance.AssessedFailure{{Index: 0, Spec: "self.x = 1", Message: "failed"}},
		1,
	)
}

func livenessViolations() instance.ViolationErrors {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/_new")
	attrKey := mustKey("domain/d/subdomain/s/class/order/attribute/amount")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: ""})
	attr := helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "amount", Details: ""}, "unconstrained", nil, false, model_class.AttributeAnnotations{}))
	class.SetAttributes([]model_class.Attribute{attr})
	class.SetStates(map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "Open", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventKey: model_state.NewEvent(eventKey, model_state.EventNameNew, "", nil),
	})

	sch := schema.New(reportTestModel(class, classKey), schema.RunScopeAll())
	st := instance.NewState(sch)
	return st.CheckLiveness(st.CollectLivenessHits(nil, nil))
}

func (s *ViolationReportSuite) TestTLAViolationsCategorized() {
	violations := tlaViolations()
	s.Require().NotEmpty(violations)

	report := FromViolations(violations)
	s.True(report.HasViolations())
	found := false
	for _, cat := range report.Categories {
		if cat.Name == "TLA+ Violations" {
			found = true
			s.GreaterOrEqual(cat.Count, 1)
		}
	}
	s.True(found, "expected TLA+ Violations category")
}

func (s *ViolationReportSuite) TestLivenessViolationsCategorized() {
	violations := livenessViolations()
	s.Require().NotEmpty(violations)

	report := FromViolations(violations)
	s.True(report.HasViolations())
	found := false
	for _, cat := range report.Categories {
		if cat.Name == "Liveness Violations" {
			found = true
			s.GreaterOrEqual(cat.Count, 1)
		}
	}
	s.True(found)
}

func (s *ViolationReportSuite) TestMixedViolations() {
	var violations instance.ViolationErrors
	violations = append(violations, tlaViolations()...)
	violations = append(violations, livenessViolations()...)
	s.Require().NotEmpty(violations)

	report := FromViolations(violations)
	s.True(report.HasViolations())
	s.GreaterOrEqual(len(report.Categories), 2)
}

func (s *ViolationReportSuite) TestJSONRoundTrip() {
	violations := livenessViolations()
	s.Require().NotEmpty(violations)
	report := FromViolations(violations)

	data, err := json.Marshal(report)
	s.Require().NoError(err)
	var decoded ViolationReport
	s.Require().NoError(json.Unmarshal(data, &decoded))
	s.Equal(report.TotalCount, decoded.TotalCount)
}

func (s *ViolationReportSuite) TestSummaryMentionsCount() {
	violations := livenessViolations()
	s.Require().NotEmpty(violations)
	report := FromViolations(violations)
	s.Contains(report.Summary, "violation")
}
