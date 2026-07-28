package actions

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"

// violationsByType filters violations for assertions (test helper; production filters via TLAViolations).
func violationsByType(vs invariants.ViolationErrors, t invariants.ViolationType) invariants.ViolationErrors {
	var out invariants.ViolationErrors
	for _, v := range vs {
		if v != nil && v.Type == t {
			out = append(out, v)
		}
	}
	return out
}
