package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// SurfaceReport describes every class and surface-eligible action/query before a run.
type SurfaceReport struct {
	Classes []SurfaceClassReport `json:"classes"`
}

// SurfaceClassReport is one scoped class and its surface-level simulation entries.
type SurfaceClassReport struct {
	ClassKey          string                          `json:"class_key"`
	ClassName         string                          `json:"class_name"`
	Role              string                          `json:"role"`
	CreationEvents    []SurfaceEventReport            `json:"creation_events,omitempty"`
	States            []SurfaceStateReport            `json:"states,omitempty"`
	Queries           []SurfaceQueryReport            `json:"queries,omitempty"`
	DerivedAttributes []SurfaceDerivedAttributeReport `json:"derived_attributes,omitempty"`
	AssociationCreate *SurfaceAssocCreateNote         `json:"association_create,omitempty"`
}

// SurfaceEventReport is an external creation or state-transition event on the surface.
type SurfaceEventReport struct {
	EventName  string `json:"event_name"`
	ActionName string `json:"action_name,omitempty"`
}

// SurfaceStateReport lists surface events and do-actions for one state.
type SurfaceStateReport struct {
	StateName string                `json:"state_name"`
	Events    []SurfaceEventReport  `json:"events,omitempty"`
	DoActions []SurfaceActionReport `json:"do_actions,omitempty"`
}

// SurfaceQueryReport is an external query on the surface.
type SurfaceQueryReport struct {
	QueryName string `json:"query_name"`
}

// SurfaceDerivedAttributeReport is an external derived attribute on the surface.
type SurfaceDerivedAttributeReport struct {
	AttributeName string `json:"attribute_name"`
}

// SurfaceActionReport is a surface do-action.
type SurfaceActionReport struct {
	ActionName string `json:"action_name"`
}

// SurfaceAssocCreateNote documents association-class creation that binds host endpoints.
type SurfaceAssocCreateNote struct {
	EventName       string `json:"event_name"`
	ActionName      string `json:"action_name,omitempty"`
	HostAssociation string `json:"host_association"`
	FromClassKey    string `json:"from_class_key"`
	ToClassKey      string `json:"to_class_key"`
}

// BuildSurfaceReport enumerates scoped classes and surface-eligible events, actions, and queries.
// Association-class roles are omitted: they are not independently selected for external
// creation; when listed on the surface they only materialize via host association guarantees.
func BuildSurfaceReport(catalog *ClassCatalog) *SurfaceReport {
	report := &SurfaceReport{
		Classes: make([]SurfaceClassReport, 0, len(catalog.classes)),
	}

	for _, classInfo := range catalog.AllScopedClasses() {
		if catalog.IsAssociationClass(classInfo.ClassKey) {
			continue
		}
		report.Classes = append(report.Classes, buildSurfaceClassReport(catalog, classInfo))
	}

	return report
}

func buildSurfaceClassReport(catalog *ClassCatalog, classInfo *ClassInfo) SurfaceClassReport {
	entry := SurfaceClassReport{
		ClassKey:  classInfo.ClassKey.String(),
		ClassName: classInfo.Class.Name,
		Role:      surfaceClassRole(catalog, classInfo),
	}

	if !classInfo.HasStates {
		return entry
	}

	for _, ev := range catalog.ExternalCreationEvents(classInfo.ClassKey) {
		entry.CreationEvents = append(entry.CreationEvents, surfaceEventReport(catalog, classInfo.ClassKey, ev, ""))
	}

	if acNote := associationClassCreateNote(catalog, classInfo); acNote != nil {
		entry.AssociationCreate = acNote
	}

	stateNames := sortedStateNames(classInfo)
	for _, stateName := range stateNames {
		stateEntry := SurfaceStateReport{StateName: stateName}

		for _, eventInfo := range catalog.ExternalStateEvents(classInfo.ClassKey, stateName) {
			stateEntry.Events = append(stateEntry.Events, surfaceEventReport(
				catalog,
				classInfo.ClassKey,
				eventInfo.Event,
				stateName,
			))
		}

		for _, action := range catalog.SurfaceDoActions(classInfo.ClassKey, stateName) {
			stateEntry.DoActions = append(stateEntry.DoActions, SurfaceActionReport{ActionName: action.Name})
		}

		if len(stateEntry.Events) > 0 || len(stateEntry.DoActions) > 0 {
			entry.States = append(entry.States, stateEntry)
		}
	}

	appendSurfaceReadEntries(catalog, classInfo.ClassKey, &entry)

	return entry
}

func appendSurfaceReadEntries(catalog *ClassCatalog, classKey identity.Key, entry *SurfaceClassReport) {
	for _, query := range catalog.ExternalQueries(classKey) {
		entry.Queries = append(entry.Queries, SurfaceQueryReport{QueryName: query.Name})
	}
	for _, attr := range catalog.ExternalDerivedAttributes(classKey) {
		entry.DerivedAttributes = append(entry.DerivedAttributes, SurfaceDerivedAttributeReport{AttributeName: attr.Name})
	}
}

func surfaceClassRole(catalog *ClassCatalog, classInfo *ClassInfo) string {
	switch {
	case !classInfo.HasStates:
		return "liveness_only"
	case catalog.IsAssociationClass(classInfo.ClassKey):
		return "association_class"
	default:
		return "simulatable"
	}
}

func sortedStateNames(classInfo *ClassInfo) []string {
	names := make([]string, 0, len(classInfo.Class.States))
	for _, state := range classInfo.Class.States {
		names = append(names, state.Name)
	}
	sort.Strings(names)
	return names
}

func surfaceEventReport(
	catalog *ClassCatalog,
	classKey identity.Key,
	event model_state.Event,
	stateName string,
) SurfaceEventReport {
	report := SurfaceEventReport{EventName: event.Name}
	if action, ok := catalog.GetActionForEvent(classKey, event.Key, stateName); ok && action != nil {
		report.ActionName = action.Name
	}
	return report
}

func associationClassCreateNote(catalog *ClassCatalog, classInfo *ClassInfo) *SurfaceAssocCreateNote {
	acInfo := catalog.LookupAssociationClass(classInfo.ClassKey)
	if acInfo == nil || len(classInfo.CreationEvents) == 0 {
		return nil
	}

	event := classInfo.CreationEvents[0]
	note := &SurfaceAssocCreateNote{
		EventName:       event.Name,
		HostAssociation: acInfo.HostAssociation.Name,
		FromClassKey:    acInfo.FromClassKey.String(),
		ToClassKey:      acInfo.ToClassKey.String(),
	}
	if action, ok := catalog.GetActionForEvent(classInfo.ClassKey, event.Key, ""); ok && action != nil {
		note.ActionName = action.Name
	}
	return note
}

// FormatText renders the surface report for CLI output before a simulation run.
func (r *SurfaceReport) FormatText() string {
	if r == nil || len(r.Classes) == 0 {
		return "Simulation surface\n\n  (empty)\n"
	}

	var b strings.Builder
	b.WriteString("Simulation surface\n")

	for _, classEntry := range r.Classes {
		fmt.Fprintf(&b, "\n  %s (%s)\n", classEntry.ClassKey, classEntry.ClassName)
		fmt.Fprintf(&b, "    role: %s\n", surfaceRoleLabel(classEntry.Role))

		for _, ev := range classEntry.CreationEvents {
			b.WriteString(formatSurfaceEventLine("creation", ev))
		}

		if classEntry.AssociationCreate != nil {
			ac := classEntry.AssociationCreate
			line := fmt.Sprintf("    association creation: event %s", ac.EventName)
			if ac.ActionName != "" {
				line += fmt.Sprintf(" (action %s)", ac.ActionName)
			}
			line += fmt.Sprintf(" via %s (%s -> %s) on host association create only (not surface)\n",
				ac.HostAssociation, ac.FromClassKey, ac.ToClassKey)
			b.WriteString(line)
		}

		for _, stateEntry := range classEntry.States {
			fmt.Fprintf(&b, "    state %s:\n", stateEntry.StateName)
			for _, ev := range stateEntry.Events {
				b.WriteString(formatSurfaceEventLine("      transition", ev))
			}
			for _, action := range stateEntry.DoActions {
				fmt.Fprintf(&b, "      do-action: %s\n", action.ActionName)
			}
		}

		for _, query := range classEntry.Queries {
			fmt.Fprintf(&b, "    query: %s\n", query.QueryName)
		}

		for _, attr := range classEntry.DerivedAttributes {
			fmt.Fprintf(&b, "    derived: %s\n", attr.AttributeName)
		}
	}

	return b.String()
}

func surfaceRoleLabel(role string) string {
	switch role {
	case "liveness_only":
		return "liveness only (no state machine)"
	case "association_class":
		return "association class"
	default:
		return "simulatable"
	}
}

func formatSurfaceEventLine(prefix string, ev SurfaceEventReport) string {
	line := fmt.Sprintf("    %s: event %s", prefix, ev.EventName)
	if ev.ActionName != "" {
		line += fmt.Sprintf(" (action %s)", ev.ActionName)
	}
	return line + "\n"
}
