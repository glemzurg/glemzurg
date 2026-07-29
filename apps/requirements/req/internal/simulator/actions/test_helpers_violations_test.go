package actions

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"

// violationsByType filters violations for assertions (test helper; production filters via TLAViolations).
func violationsByType(vs instance.ViolationErrors, t instance.ViolationType) instance.ViolationErrors {
	var out instance.ViolationErrors
	for _, v := range vs {
		if v != nil && v.Type == t {
			out = append(out, v)
		}
	}
	return out
}
