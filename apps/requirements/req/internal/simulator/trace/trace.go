// Package trace provides serializable simulation trace output.
// It transforms the raw SimulationResult into clean view-model structs
// suitable for human-readable text and JSON output.
package trace

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/engine"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// SimulationTrace is the top-level serializable trace of a simulation run.
type SimulationTrace struct {
	StepsTaken        int         `json:"steps_taken"`
	TerminationReason string      `json:"termination_reason"`
	Steps             []TraceStep `json:"steps"`
	FinalState        *FinalState `json:"final_state,omitempty"`
}

// AssociationMaterializationTrace records endpoint classes linked by an association-class row.
type AssociationMaterializationTrace struct {
	AssociationName string `json:"association_name"`
	AssociationKey  string `json:"association_key"`
	FromClassName   string `json:"from_class_name"`
	FromClassKey    string `json:"from_class_key"`
	ToClassName     string `json:"to_class_name"`
	ToClassKey      string `json:"to_class_key"`
	FromInstanceID  uint64 `json:"from_instance_id"`
	ToInstanceID    uint64 `json:"to_instance_id"`
}

// TraceStep is a serializable view of one simulation step.
type TraceStep struct { //nolint:revive // public API name
	StepNumber                 int                              `json:"step_number"`
	Kind                       string                           `json:"kind"`
	ClassName                  string                           `json:"class_name"`
	ClassKey                   string                           `json:"class_key"`
	EventName                  string                           `json:"event_name,omitempty"`
	QueryName                  string                           `json:"query_name,omitempty"`
	DerivedAttributeName       string                           `json:"derived_attribute_name,omitempty"`
	DerivedReadValue           string                           `json:"derived_read_value,omitempty"`
	InstanceID                 uint64                           `json:"instance_id"`
	FromState                  string                           `json:"from_state,omitempty"`
	ToState                    string                           `json:"to_state,omitempty"`
	Parameters                 map[string]string                `json:"parameters,omitempty"`
	Assignments                map[string]string                `json:"assignments,omitempty"`
	AssociationMaterialization *AssociationMaterializationTrace `json:"association_materialization,omitempty"`
	CascadedSteps              []TraceStep                      `json:"cascaded_steps,omitempty"`
	Violations                 []string                         `json:"violations,omitempty"`
}

// FinalState is a snapshot of all instances at simulation end.
type FinalState struct {
	InstanceCount int             `json:"instance_count"`
	LinkCount     int             `json:"link_count"`
	Instances     []InstanceState `json:"instances"`
}

// InstanceState is one instance's final attribute values.
type InstanceState struct {
	InstanceID uint64                           `json:"instance_id"`
	ClassKey   string                           `json:"class_key"`
	Attributes map[string]string                `json:"attributes"`
	Endpoints  *AssociationMaterializationTrace `json:"association_endpoints,omitempty"`
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
		t.FinalState = buildFinalState(result.FinalState, result.Catalog)
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
			if inst.Endpoints != nil {
				mat := inst.Endpoints
				fmt.Fprintf(&b, " %s (%s#%d -> %s#%d)",
					mat.AssociationName,
					mat.FromClassName, mat.FromInstanceID,
					mat.ToClassName, mat.ToInstanceID,
				)
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
		StepNumber:           step.StepNumber,
		Kind:                 step.Kind.String(),
		ClassName:            step.ClassName,
		ClassKey:             step.ClassKey.String(),
		EventName:            step.EventName,
		QueryName:            step.QueryName,
		DerivedAttributeName: step.DerivedAttributeName,
		InstanceID:           uint64(step.InstanceID),
		FromState:            step.FromState,
		ToState:              step.ToState,
	}
	if step.DerivedReadValue != nil {
		ts.DerivedReadValue = step.DerivedReadValue.Inspect()
	}
	if step.QueryName != "" {
		ts.Kind = "query"
	}
	if step.DerivedAttributeName != "" {
		ts.Kind = "derived"
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

	if step.TransitionResult != nil && step.TransitionResult.AssociationMaterialization != nil {
		ts.AssociationMaterialization = convertAssociationMaterialization(step.TransitionResult.AssociationMaterialization)
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

func convertAssociationMaterialization(mat *actions.AssociationMaterialization) *AssociationMaterializationTrace {
	return &AssociationMaterializationTrace{
		AssociationName: mat.HostAssociationName,
		AssociationKey:  mat.HostAssociationKey.String(),
		FromClassName:   mat.FromClassName,
		FromClassKey:    mat.FromClassKey.String(),
		ToClassName:     mat.ToClassName,
		ToClassKey:      mat.ToClassKey.String(),
		FromInstanceID:  uint64(mat.FromInstanceID),
		ToInstanceID:    uint64(mat.ToInstanceID),
	}
}

// buildFinalState creates a FinalState snapshot from SimulationState.
func buildFinalState(simState *instance.State, catalog *engine.ClassCatalog) *FinalState {
	snap := simState.Snapshot()
	fs := &FinalState{
		InstanceCount: snap.InstanceCount,
		LinkCount:     snap.LinkCount,
		Instances:     make([]InstanceState, 0, len(snap.Instances)),
	}

	for _, inst := range snap.Instances {
		is := InstanceState{
			InstanceID: uint64(inst.ID),
			ClassKey:   inst.ClassKey.String(),
			Attributes: inst.Attributes,
			Endpoints:  associationEndpointsForSnapshot(inst, simState, catalog),
		}
		fs.Instances = append(fs.Instances, is)
	}

	return fs
}

func associationEndpointsForSnapshot(
	inst instance.SnapshotInstance,
	simState *instance.State,
	catalog *engine.ClassCatalog,
) *AssociationMaterializationTrace {
	if catalog == nil || !catalog.IsAssociationClass(inst.ClassKey) {
		return nil
	}
	link, ok := simState.AssociationLinkByInstance(inst.ID)
	if !ok {
		return nil
	}
	linkInfo := catalog.GetAssociationClassInfo(inst.ClassKey)
	if !linkInfo.Found {
		return nil
	}
	return &AssociationMaterializationTrace{
		AssociationName: linkInfo.HostAssociationName,
		AssociationKey:  linkInfo.HostAssocKey.String(),
		FromClassName:   linkInfo.FromClassName,
		FromClassKey:    linkInfo.FromClassKey.String(),
		ToClassName:     linkInfo.ToClassName,
		ToClassKey:      linkInfo.ToClassKey.String(),
		FromInstanceID:  uint64(link.FromEndpointID),
		ToInstanceID:    uint64(link.ToEndpointID),
	}
}

// extractAssignments gets the primed assignments for the primary instance.
func extractAssignments(instanceID instance.ID, assignments map[instance.ID]map[string]object.Object) map[string]string {
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
	case "destroy":
		fmt.Fprintf(b, "%s[%d] DESTROY %s#%d (%s ->)", indent, step.StepNumber, step.ClassName, step.InstanceID, step.FromState)
	case "query":
		fmt.Fprintf(b, "%s[%d] QUERY %s#%d: %s", indent, step.StepNumber, step.ClassName, step.InstanceID, step.QueryName)
	case "derived":
		fmt.Fprintf(b, "%s[%d] DERIVED %s#%d: %s = %s", indent, step.StepNumber, step.ClassName, step.InstanceID, step.DerivedAttributeName, step.DerivedReadValue)
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
	if step.AssociationMaterialization != nil {
		mat := step.AssociationMaterialization
		fmt.Fprintf(b, "%s  materializes: %s (%s#%d -> %s#%d)\n",
			indent,
			mat.AssociationName,
			mat.FromClassName, mat.FromInstanceID,
			mat.ToClassName, mat.ToInstanceID,
		)
	}
	// Nested work first so post-nesting world-state violations appear after the cascade.
	for _, cs := range step.CascadedSteps {
		writeStep(b, cs, indent+"  ")
	}
	for _, v := range step.Violations {
		fmt.Fprintf(b, "%s  VIOLATION: %s\n", indent, v)
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
