package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	siminst "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// LivenessChecker adapts a completed SimulationResult into instance liveness hits.
// Obligations live on FinalState (installed at NewState); check does not query schema.
type LivenessChecker struct{}

// NewLivenessChecker creates a liveness adapter. The catalog argument is ignored
// (kept for call-site compatibility during wiring); obligations come from sim state.
func NewLivenessChecker(_ *schema.Schema) *LivenessChecker {
	return &LivenessChecker{}
}

// Check performs all liveness checks against a completed simulation result.
func (lc *LivenessChecker) Check(result *SimulationResult) siminst.ViolationErrors {
	if result == nil {
		return nil
	}
	st := result.FinalState
	if st == nil {
		// Tests may omit FinalState; still evaluate step-based obligations via Schema.
		if result.Schema == nil {
			return nil
		}
		st = siminst.NewState(result.Schema)
	}
	var usedParams map[identity.Key]bool
	if result.SimulationCoverage != nil {
		usedParams = result.SimulationCoverage.UsedSimulationParams
	}
	hits := st.CollectLivenessHits(convertStepsToLiveness(result.Steps), usedParams)
	// When FinalState is nil, association link hits are empty (no links observed).
	if result.FinalState == nil {
		hits.LinkedAssocs = map[string]bool{}
	}
	return st.CheckLiveness(hits)
}

func convertStepsToLiveness(steps []*SimulationStep) []siminst.LivenessStep {
	if len(steps) == 0 {
		return nil
	}
	out := make([]siminst.LivenessStep, 0, len(steps))
	for _, step := range steps {
		if step == nil {
			continue
		}
		rec := siminst.LivenessStep{
			IsCreation:           step.Kind == StepKindCreation,
			ClassKey:             step.ClassKey,
			EventKey:             step.EventKey,
			EventName:            step.EventName,
			QueryKey:             step.QueryKey,
			QueryName:            step.QueryName,
			DerivedAttributeKey:  step.DerivedAttributeKey,
			DerivedAttributeName: step.DerivedAttributeName,
			ExecutedActionKeys:   append([]identity.Key(nil), step.ExecutedActionKeys...),
		}
		if step.TransitionResult != nil {
			rec.TransitionKey = step.TransitionResult.TransitionKey
			if step.TransitionResult.ActionResult != nil {
				rec.HasTransitionAction = true
				rec.PrimedAttrSubKeys = primedSubKeys(step.TransitionResult.ActionResult.PrimedAssignments)
			}
		}
		if step.DoActionResult != nil {
			rec.PrimedAttrSubKeys = append(rec.PrimedAttrSubKeys, primedSubKeys(step.DoActionResult.PrimedAssignments)...)
		}
		if len(step.CascadedSteps) > 0 {
			rec.Cascaded = convertStepsToLiveness(step.CascadedSteps)
		}
		out = append(out, rec)
	}
	return out
}

func primedSubKeys(assignments map[siminst.ID]map[string]object.Object) []string {
	var keys []string
	for _, fields := range assignments {
		for fieldName := range fields {
			keys = append(keys, fieldName)
		}
	}
	return keys
}
