package engine

import siminst "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"

// violationsByType filters violations for assertions (test helper; production filters via TLAViolations).
func violationsByType(vs siminst.ViolationErrors, t siminst.ViolationType) siminst.ViolationErrors {
	var out siminst.ViolationErrors
	for _, v := range vs {
		if v != nil && v.Type == t {
			out = append(out, v)
		}
	}
	return out
}
