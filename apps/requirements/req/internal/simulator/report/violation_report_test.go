package report

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/simulator/invariants"
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

func (s *ViolationReportSuite) TestTLAViolationsCategorized() {
	violations := invariants.ViolationList{
		invariants.NewModelInvariantViolation(0, "x > 0", "evaluated to FALSE"),
		invariants.NewActionGuaranteeViolation(
			mustKey("domain/d/subdomain/s/class/c/action/a"),
			"DoSomething", 0, "self.x' = 1", 1, "guarantee failed",
		),
	}

	report := FromViolations(violations)

	s.Equal(2, report.TotalCount)
	s.True(report.HasViolations())
	s.Require().Len(report.Categories, 1)
	s.Equal("TLA+ Violations", report.Categories[0].Name)
	s.Equal(2, report.Categories[0].Count)
}

func (s *ViolationReportSuite) TestDataTypeViolationsCategorized() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	violations := invariants.ViolationList{
		invariants.NewRequiredAttributeViolation(1, classKey, "name"),
		invariants.NewSpanConstraintViolation(1, classKey, "amount", "150", "[0, 100]"),
	}

	report := FromViolations(violations)

	s.Equal(2, report.TotalCount)
	s.Require().Len(report.Categories, 1)
	s.Equal("Data Type Violations", report.Categories[0].Name)
	s.Equal(2, report.Categories[0].Count)

	// Verify entries have attribute info.
	s.Equal("name", report.Categories[0].Violations[0].Attribute)
}

func (s *ViolationReportSuite) TestLivenessViolationsCategorized() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	violations := invariants.ViolationList{
		invariants.NewLivenessClassNotInstantiatedViolation(classKey, "Order"),
		invariants.NewLivenessAttributeNotWrittenViolation(classKey, "Order", "amount"),
	}

	report := FromViolations(violations)

	s.Equal(2, report.TotalCount)
	s.Require().Len(report.Categories, 1)
	s.Equal("Liveness Violations", report.Categories[0].Name)
	s.Equal(2, report.Categories[0].Count)
}

func (s *ViolationReportSuite) TestMixedViolations() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	violations := invariants.ViolationList{
		invariants.NewModelInvariantViolation(0, "TRUE", "failed"),
		invariants.NewRequiredAttributeViolation(1, classKey, "name"),
		invariants.NewLivenessClassNotInstantiatedViolation(classKey, "Order"),
		invariants.NewMultiplicityViolation(1, classKey, "items", "to", 0, 1, 10, "too few"),
	}

	report := FromViolations(violations)

	s.Equal(4, report.TotalCount)
	s.Len(report.Categories, 4)

	// Verify category names.
	names := make(map[string]int)
	for _, cat := range report.Categories {
		names[cat.Name] = cat.Count
	}
	s.Equal(1, names["TLA+ Violations"])
	s.Equal(1, names["Data Type Violations"])
	s.Equal(1, names["Liveness Violations"])
	s.Equal(1, names["Other Violations"])

	s.Contains(report.Summary, "4 violations found")
	s.Contains(report.Summary, "1 TLA+")
	s.Contains(report.Summary, "1 data type")
	s.Contains(report.Summary, "1 liveness")
	s.Contains(report.Summary, "1 other")
}

func (s *ViolationReportSuite) TestFormatTextEmpty() {
	report := FromViolations(nil)
	text := report.FormatText()
	s.Equal("No violations found.\n", text)
}

func (s *ViolationReportSuite) TestFormatTextWithViolations() {
	violations := invariants.ViolationList{
		invariants.NewModelInvariantViolation(0, "x > 0", "failed"),
	}

	report := FromViolations(violations)
	text := report.FormatText()

	s.Contains(text, "1 violations found: 1 TLA+")
	s.Contains(text, "TLA+ Violations (1)")
	s.Contains(text, "[model_invariant]")
}

func (s *ViolationReportSuite) TestFormatJSONRoundTrip() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	violations := invariants.ViolationList{
		invariants.NewRequiredAttributeViolation(1, classKey, "name"),
	}

	report := FromViolations(violations)
	data, err := report.FormatJSON()
	s.Require().NoError(err)

	var decoded ViolationReport
	err = json.Unmarshal(data, &decoded)
	s.Require().NoError(err)

	s.Equal(1, decoded.TotalCount)
	s.Len(decoded.Categories, 1)
	s.Equal("Data Type Violations", decoded.Categories[0].Name)
}
