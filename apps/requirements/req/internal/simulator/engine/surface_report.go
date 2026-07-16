package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// SurfaceReport catalogs simulation scope and external drivers for a run.
// Scope is which classes/subdomains participate; Classes lists only top-level
// drivers (events, do-actions, queries, derived attributes) for human testers.
type SurfaceReport struct {
	// Scope summarizes include-list participation: whole subdomains or individual classes.
	Scope []surface.ScopeEntry `json:"scope,omitempty"`
	// Classes lists only classes with external drivers (not peer-only scoped classes).
	Classes []SurfaceClassReport `json:"classes"`
	// UnavailableMembers are derived attributes and queries kept off the surface
	// because they depend on out-of-scope association data (pass-through pattern).
	UnavailableMembers []SurfaceUnavailableMemberReport `json:"unavailable_members,omitempty"`
}

// SurfaceUnavailableMemberReport documents a derived attribute or query not on the surface.
type SurfaceUnavailableMemberReport struct {
	ClassKey       string   `json:"class_key"`
	ClassName      string   `json:"class_name"`
	Kind           string   `json:"kind"` // "derived" or "query"
	MemberName     string   `json:"member_name"`
	MissingClasses []string `json:"missing_classes,omitempty"`
	Reason         string   `json:"reason"`
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

// BuildSurfaceReport lists only classes that have at least one external driver
// (creation/state event, do-action, query, or derived attribute). Peer-only classes,
// association classes, and empty scoped classes are omitted so the report matches what
// a human tester can treat as under test at the top level. Out-of-scope pass-through
// derived/queries appear under UnavailableMembers, not as drivers.
func BuildSurfaceReport(catalog *ClassCatalog) *SurfaceReport {
	report := &SurfaceReport{
		Classes: make([]SurfaceClassReport, 0, len(catalog.classes)),
	}

	for _, classInfo := range catalog.AllScopedClasses() {
		if catalog.IsAssociationClass(classInfo.ClassKey) {
			continue
		}
		entry := buildSurfaceClassReport(catalog, classInfo)
		if !surfaceClassHasDrivers(entry) {
			continue
		}
		report.Classes = append(report.Classes, entry)
	}

	for _, m := range catalog.SurfaceUnavailableMembers() {
		report.UnavailableMembers = append(report.UnavailableMembers, SurfaceUnavailableMemberReport{
			ClassKey:       m.ClassKey.String(),
			ClassName:      m.ClassName,
			Kind:           string(m.Kind),
			MemberName:     m.MemberName,
			MissingClasses: append([]string(nil), m.MissingClasses...),
			Reason:         m.Reason(),
		})
	}

	return report
}

// surfaceClassHasDrivers reports whether the class row lists any top-level selector hooks.
func surfaceClassHasDrivers(entry SurfaceClassReport) bool {
	if len(entry.CreationEvents) > 0 || len(entry.Queries) > 0 || len(entry.DerivedAttributes) > 0 {
		return true
	}
	for _, st := range entry.States {
		if len(st.Events) > 0 || len(st.DoActions) > 0 {
			return true
		}
	}
	return false
}

func buildSurfaceClassReport(catalog *ClassCatalog, classInfo *ClassInfo) SurfaceClassReport {
	entry := SurfaceClassReport{
		ClassKey:  classInfo.ClassKey.String(),
		ClassName: classInfo.Class.Name,
		Role:      surfaceClassRole(catalog, classInfo),
	}

	if !classInfo.HasStates {
		appendSurfaceReadEntries(catalog, classInfo.ClassKey, &entry)
		return entry
	}

	for _, ev := range catalog.ExternalCreationEvents(classInfo.ClassKey) {
		entry.CreationEvents = append(entry.CreationEvents, surfaceEventReport(catalog, classInfo.ClassKey, ev, ""))
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

// FormatText renders scope (what is loaded) then surface drivers (what is tested at top level).
func (r *SurfaceReport) FormatText() string {
	if r == nil {
		return "Simulation scope\n\n  (empty)\n\nSimulation surface\n\n  (empty)\n"
	}

	var b strings.Builder
	b.WriteString("Simulation scope\n")
	if len(r.Scope) == 0 {
		b.WriteString("\n  (empty)\n")
	} else {
		for _, entry := range r.Scope {
			switch entry.Kind {
			case surface.ScopeSubdomain:
				fmt.Fprintf(&b, "  subdomain %s\n", entry.Path)
			case surface.ScopeClass:
				fmt.Fprintf(&b, "  class %s\n", entry.Path)
			default:
				fmt.Fprintf(&b, "  %s\n", entry.Path)
			}
		}
	}

	b.WriteString("\nSimulation surface\n")
	if len(r.Classes) == 0 {
		b.WriteString("\n  (empty)\n")
	} else {
		for _, classEntry := range r.Classes {
			fmt.Fprintf(&b, "\n  %s (%s)\n", classEntry.ClassKey, classEntry.ClassName)

			for _, ev := range classEntry.CreationEvents {
				b.WriteString(formatSurfaceEventLine("creation", ev))
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
	}

	if len(r.UnavailableMembers) > 0 {
		b.WriteString("\n  off-surface (out-of-scope association data):\n")
		for _, m := range r.UnavailableMembers {
			fmt.Fprintf(&b, "    %s %s.%s — %s\n", m.Kind, m.ClassName, m.MemberName, m.Reason)
		}
	}

	return b.String()
}

func formatSurfaceEventLine(prefix string, ev SurfaceEventReport) string {
	line := fmt.Sprintf("    %s: event %s", prefix, ev.EventName)
	if ev.ActionName != "" {
		line += fmt.Sprintf(" (action %s)", ev.ActionName)
	}
	return line + "\n"
}
