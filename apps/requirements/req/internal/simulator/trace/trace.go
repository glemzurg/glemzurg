// Package trace provides serializable simulation trace output.
// It transforms the raw SimulationResult into clean view-model structs
// suitable for human-readable text and JSON output.
package trace

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/engine"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// SimulationTrace is the top-level serializable trace of a simulation run.
type SimulationTrace struct {
	StepsTaken        int         `json:"steps_taken"`
	TerminationReason string      `json:"termination_reason"`
	Steps             []TraceStep `json:"steps"`
	FinalState        *FinalState `json:"final_state,omitempty"`
}

// TraceStep is a serializable view of one simulation step.
type TraceStep struct {
	StepNumber    int               `json:"step_number"`
	Kind          string            `json:"kind"`
	ClassName     string            `json:"class_name"`
	ClassKey      string            `json:"class_key"`
	EventName     string            `json:"event_name,omitempty"`
	InstanceID    uint64            `json:"instance_id"`
	FromState     string            `json:"from_state,omitempty"`
	ToState       string            `json:"to_state,omitempty"`
	Parameters    map[string]string `json:"parameters,omitempty"`
	Assignments   map[string]string `json:"assignments,omitempty"`
	CascadedSteps []TraceStep       `json:"cascaded_steps,omitempty"`
	Violations    []string          `json:"violations,omitempty"`
}

// FinalState is a snapshot of all instances at simulation end.
type FinalState struct {
	InstanceCount int             `json:"instance_count"`
	LinkCount     int             `json:"link_count"`
	Instances     []InstanceState `json:"instances"`
}

// InstanceState is one instance's final attribute values.
type InstanceState struct {
	InstanceID uint64            `json:"instance_id"`
	ClassKey   string            `json:"class_key"`
	Attributes map[string]string `json:"attributes"`
}

// FromResult builds a SimulationTrace from a SimulationResult.
func FromResult(result *engine.SimulationResult) *SimulationTrace {
	t := &SimulationTrace{
		StepsTaken:        result.StepsTaken,
		TerminationReason: result.TerminationReason,
	}

	for _, step := range result.Steps {
		t.Steps = append(t.Steps, convertStep(step))
	}

	if result.FinalState != nil {
		t.FinalState = buildFinalState(result.FinalState)
	}

	return t
}

// FormatText renders the trace as a human-readable multi-line string.
func (t *SimulationTrace) FormatText() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Simulation: %d steps, terminated: %s\n", t.StepsTaken, t.TerminationReason)

	for _, step := range t.Steps {
		writeStep(&b, step, "  ")
	}

	if t.FinalState != nil {
		fmt.Fprintf(&b, "\nFinal State: %d instances, %d links\n", t.FinalState.InstanceCount, t.FinalState.LinkCount)
		for _, inst := range t.FinalState.Instances {
			fmt.Fprintf(&b, "  %s#%d", inst.ClassKey, inst.InstanceID)
			if len(inst.Attributes) > 0 {
				attrs := sortedMapEntries(inst.Attributes)
				fmt.Fprintf(&b, " {%s}", attrs)
			}
			fmt.Fprintln(&b)
		}
	}

	return b.String()
}

// FormatJSON renders the trace as indented JSON bytes.
func (t *SimulationTrace) FormatJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// convertStep transforms a SimulationStep into a TraceStep.
func convertStep(step *engine.SimulationStep) TraceStep {
	ts := TraceStep{
		StepNumber: step.StepNumber,
		Kind:       step.Kind.String(),
		ClassName:  step.ClassName,
		ClassKey:   step.ClassKey.String(),
		EventName:  step.EventName,
		InstanceID: uint64(step.InstanceID),
		FromState:  step.FromState,
		ToState:    step.ToState,
	}

	// Convert parameters.
	if len(step.Parameters) > 0 {
		ts.Parameters = inspectMap(step.Parameters)
	}

	// Extract assignments from transition action result.
	if step.TransitionResult != nil && step.TransitionResult.ActionResult != nil {
		ts.Assignments = extractAssignments(step.InstanceID, step.TransitionResult.ActionResult.PrimedAssignments)
	}

	// Extract assignments from do action result.
	if step.DoActionResult != nil {
		ts.Assignments = extractAssignments(step.InstanceID, step.DoActionResult.PrimedAssignments)
	}

	// Convert cascaded steps.
	for _, cs := range step.CascadedSteps {
		ts.CascadedSteps = append(ts.CascadedSteps, convertStep(cs))
	}

	// Convert step-level violations.
	for _, v := range step.Violations {
		ts.Violations = append(ts.Violations, v.Message)
	}

	return ts
}

// buildFinalState creates a FinalState snapshot from SimulationState.
func buildFinalState(simState *state.SimulationState) *FinalState {
	instances := simState.AllInstances()

	// Sort by ID for deterministic output.
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].ID < instances[j].ID
	})

	fs := &FinalState{
		InstanceCount: len(instances),
		LinkCount:     simState.LinkCount(),
	}

	for _, inst := range instances {
		is := InstanceState{
			InstanceID: uint64(inst.ID),
			ClassKey:   inst.ClassKey.String(),
			Attributes: make(map[string]string),
		}
		for _, name := range inst.AttributeNames() {
			val := inst.GetAttribute(name)
			if val != nil {
				is.Attributes[name] = val.Inspect()
			}
		}
		fs.Instances = append(fs.Instances, is)
	}

	return fs
}

// extractAssignments gets the primed assignments for the primary instance.
func extractAssignments(instanceID state.InstanceID, assignments map[state.InstanceID]map[string]object.Object) map[string]string {
	fields, ok := assignments[instanceID]
	if !ok || len(fields) == 0 {
		return nil
	}
	result := make(map[string]string, len(fields))
	for name, val := range fields {
		result[name] = val.Inspect()
	}
	return result
}

// inspectMap converts a map of object.Object values to strings.
func inspectMap(m map[string]object.Object) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v.Inspect()
	}
	return result
}

// writeStep writes a single trace step at the given indent level.
func writeStep(b *strings.Builder, step TraceStep, indent string) {
	switch step.Kind {
	case "creation":
		fmt.Fprintf(b, "%s[%d] CREATE %s#%d -> %s", indent, step.StepNumber, step.ClassName, step.InstanceID, step.ToState)
	case "deletion":
		fmt.Fprintf(b, "%s[%d] DELETE %s#%d (%s ->)", indent, step.StepNumber, step.ClassName, step.InstanceID, step.FromState)
	default:
		fmt.Fprintf(b, "%s[%d] %s#%d: %s -> %s", indent, step.StepNumber, step.ClassName, step.InstanceID, step.FromState, step.ToState)
	}

	if step.EventName != "" {
		fmt.Fprintf(b, " (event: %s)", step.EventName)
	}
	fmt.Fprintln(b)

	if len(step.Parameters) > 0 {
		fmt.Fprintf(b, "%s  params: %s\n", indent, sortedMapEntries(step.Parameters))
	}
	if len(step.Assignments) > 0 {
		fmt.Fprintf(b, "%s  assigns: %s\n", indent, sortedMapEntries(step.Assignments))
	}
	for _, v := range step.Violations {
		fmt.Fprintf(b, "%s  VIOLATION: %s\n", indent, v)
	}
	for _, cs := range step.CascadedSteps {
		writeStep(b, cs, indent+"  ")
	}
}

// sortedMapEntries formats a string map as "k1=v1, k2=v2" with sorted keys.
func sortedMapEntries(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(parts, ", ")
}
