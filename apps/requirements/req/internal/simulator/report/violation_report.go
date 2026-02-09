// Package report provides violation categorization and formatting
// for simulation results.
package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/go-tlaplus/internal/simulator/invariants"
)

// ViolationReport categorizes and summarizes violations.
type ViolationReport struct {
	TotalCount int                 `json:"total_count"`
	Categories []ViolationCategory `json:"categories"`
	Summary    string              `json:"summary"`
}

// ViolationCategory groups violations of one logical kind.
type ViolationCategory struct {
	Name       string           `json:"name"`
	Count      int              `json:"count"`
	Violations []ViolationEntry `json:"violations"`
}

// ViolationEntry is one formatted violation.
type ViolationEntry struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	InstanceID uint64 `json:"instance_id,omitempty"`
	ClassKey   string `json:"class_key,omitempty"`
	Attribute  string `json:"attribute,omitempty"`
	Expression string `json:"expression,omitempty"`
}

// FromViolations builds a ViolationReport from a ViolationList.
func FromViolations(violations invariants.ViolationList) *ViolationReport {
	r := &ViolationReport{
		TotalCount: len(violations),
	}

	// Categorize using existing filter methods.
	tla := violations.TLAViolations()
	dataType := violations.DataTypeViolations()
	liveness := violations.LivenessViolations()

	// Collect remaining violations (multiplicity, safety rules, unparsed data type).
	categorized := make(map[*invariants.Violation]bool)
	for _, v := range tla {
		categorized[v] = true
	}
	for _, v := range dataType {
		categorized[v] = true
	}
	for _, v := range liveness {
		categorized[v] = true
	}
	var other invariants.ViolationList
	for _, v := range violations {
		if !categorized[v] {
			other = append(other, v)
		}
	}

	if len(tla) > 0 {
		r.Categories = append(r.Categories, buildCategory("TLA+ Violations", tla))
	}
	if len(dataType) > 0 {
		r.Categories = append(r.Categories, buildCategory("Data Type Violations", dataType))
	}
	if len(liveness) > 0 {
		r.Categories = append(r.Categories, buildCategory("Liveness Violations", liveness))
	}
	if len(other) > 0 {
		r.Categories = append(r.Categories, buildCategory("Other Violations", other))
	}

	r.Summary = buildSummary(r.TotalCount, len(tla), len(dataType), len(liveness), len(other))

	return r
}

// HasViolations returns true if there are any violations.
func (r *ViolationReport) HasViolations() bool {
	return r.TotalCount > 0
}

// FormatText renders the report as a human-readable multi-line string.
func (r *ViolationReport) FormatText() string {
	if r.TotalCount == 0 {
		return "No violations found.\n"
	}

	var b strings.Builder
	fmt.Fprintln(&b, r.Summary)
	fmt.Fprintln(&b)

	for _, cat := range r.Categories {
		fmt.Fprintf(&b, "%s (%d):\n", cat.Name, cat.Count)
		for _, v := range cat.Violations {
			fmt.Fprintf(&b, "  - [%s] %s\n", v.Type, v.Message)
		}
		fmt.Fprintln(&b)
	}

	return b.String()
}

// FormatJSON renders the report as indented JSON bytes.
func (r *ViolationReport) FormatJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// buildCategory creates a ViolationCategory from a ViolationList.
func buildCategory(name string, violations invariants.ViolationList) ViolationCategory {
	cat := ViolationCategory{
		Name:  name,
		Count: len(violations),
	}
	for _, v := range violations {
		entry := ViolationEntry{
			Type:       v.Type.String(),
			Message:    v.Message,
			InstanceID: uint64(v.InstanceID),
			ClassKey:   v.ClassKey.String(),
			Attribute:  v.AttributeName,
			Expression: v.Expression,
		}
		cat.Violations = append(cat.Violations, entry)
	}
	return cat
}

// buildSummary creates the summary line.
func buildSummary(total, tla, dataType, liveness, other int) string {
	if total == 0 {
		return "No violations found."
	}

	parts := make([]string, 0, 4)
	if tla > 0 {
		parts = append(parts, fmt.Sprintf("%d TLA+", tla))
	}
	if dataType > 0 {
		parts = append(parts, fmt.Sprintf("%d data type", dataType))
	}
	if liveness > 0 {
		parts = append(parts, fmt.Sprintf("%d liveness", liveness))
	}
	if other > 0 {
		parts = append(parts, fmt.Sprintf("%d other", other))
	}

	return fmt.Sprintf("%d violations found: %s", total, strings.Join(parts, ", "))
}
