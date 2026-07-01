package engine

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// SimulationCoverageTracker records which simulator-authored parameter specs were exercised.
type SimulationCoverageTracker struct {
	UsedSimulationParams map[identity.Key]bool
}

// NewSimulationCoverageTracker creates an empty coverage tracker.
func NewSimulationCoverageTracker() *SimulationCoverageTracker {
	return &SimulationCoverageTracker{
		UsedSimulationParams: make(map[identity.Key]bool),
	}
}

// MarkSimulationParamUsed records that a parameter simulation specification produced a value.
func (t *SimulationCoverageTracker) MarkSimulationParamUsed(paramKey identity.Key) {
	if t == nil {
		return
	}
	if t.UsedSimulationParams == nil {
		t.UsedSimulationParams = make(map[identity.Key]bool)
	}
	t.UsedSimulationParams[paramKey] = true
}
