package instance

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// LivenessHits records what the run exercised. Built by the engine from steps and
// final state; compared against obligations installed on State at construction.
// CheckLiveness does not consult schema.
type LivenessHits struct {
	Instantiated         map[identity.Key]bool
	WrittenAttrs         map[identity.Key]map[string]bool // classKey → attr SubKey
	LinkedAssocs         map[string]bool                  // association key string
	Events               map[identity.Key]bool
	Queries              map[identity.Key]bool
	DerivedReads         map[identity.Key]bool
	Actions              map[identity.Key]bool
	UsedSimulationParams map[identity.Key]bool
}

// LivenessStep is a schema-free summary of one simulation step for hit collection.
// Engine converts its step graph into this shape so instance need not import engine.
type LivenessStep struct {
	IsCreation           bool
	ClassKey             identity.Key
	EventKey             identity.Key
	EventName            string
	QueryKey             identity.Key
	QueryName            string
	DerivedAttributeKey  identity.Key
	DerivedAttributeName string
	ExecutedActionKeys   []identity.Key
	TransitionKey        identity.Key
	HasTransitionAction  bool // true when TransitionResult had an ActionResult
	PrimedAttrSubKeys    []string
	Cascaded             []LivenessStep
}

type livenessClassObl struct {
	key  identity.Key
	name string
}

type livenessAttrObl struct {
	classKey  identity.Key
	className string
	subKey    string
	name      string
}

type livenessAssocObl struct {
	key     identity.Key
	name    string
	fromKey identity.Key
	toKey   identity.Key
	keyStr  string
}

type livenessNamedObl struct {
	classKey  identity.Key
	className string
	memberKey identity.Key
	name      string
}

type livenessParamObl struct {
	classKey   identity.Key
	className  string
	actionName string
	paramKey   identity.Key
	paramName  string
}

// livenessObligations is the coverage contract for one run, installed once at NewState.
type livenessObligations struct {
	classes      []livenessClassObl
	attributes   []livenessAttrObl
	associations []livenessAssocObl
	events       []livenessNamedObl
	queries      []livenessNamedObl
	derived      []livenessNamedObl
	actions      []livenessNamedObl
	params       []livenessParamObl
	// transitionToAction maps transition keys to action keys for hit collection.
	transitionToAction map[identity.Key]identity.Key
}

func installLivenessObligations(sch *schema.Schema) *livenessObligations {
	obl := &livenessObligations{
		transitionToAction: make(map[identity.Key]identity.Key),
	}
	if sch == nil {
		return obl
	}
	sch.EachInScopeClassSim(func(sim *schema.ClassSimInfo) {
		if sim != nil {
			installClassLiveness(obl, sch, sim)
		}
	})
	for _, view := range sch.ScopedAssociations() {
		// Both ends are in scope for ScopedAssociations.
		obl.associations = append(obl.associations, livenessAssocObl{
			key:     view.Association.Key,
			name:    view.Association.Name,
			fromKey: view.FromClassKey,
			toKey:   view.ToClassKey,
			keyStr:  view.Association.Key.String(),
		})
	}
	return obl
}

func installClassLiveness(obl *livenessObligations, sch *schema.Schema, sim *schema.ClassSimInfo) {
	obl.classes = append(obl.classes, livenessClassObl{key: sim.ClassKey, name: sim.Class.Name})
	obl.attributes = append(obl.attributes, classAttributeObligations(sim)...)
	obl.events = append(obl.events, classEventObligations(sim)...)
	obl.queries = append(obl.queries, classQueryObligations(sim)...)
	obl.actions = append(obl.actions, classActionObligations(sim)...)
	for _, attr := range sch.ExternalDerivedAttributes(sim.ClassKey) {
		obl.derived = append(obl.derived, livenessNamedObl{
			classKey: sim.ClassKey, className: sim.Class.Name,
			memberKey: attr.Key, name: attr.Name,
		})
	}
	samplable := installTransitionMaps(obl, sch, sim)
	obl.params = append(obl.params, classParamObligations(sim, samplable)...)
}

func classAttributeObligations(sim *schema.ClassSimInfo) []livenessAttrObl {
	var attrs []livenessAttrObl
	for _, attr := range sim.Class.Attributes {
		if attr.DerivationPolicy != nil {
			continue
		}
		attrs = append(attrs, livenessAttrObl{
			classKey: sim.ClassKey, className: sim.Class.Name,
			subKey: attr.Key.SubKey, name: attr.Name,
		})
	}
	sort.Slice(attrs, func(i, j int) bool { return attrs[i].name < attrs[j].name })
	return attrs
}

func classEventObligations(sim *schema.ClassSimInfo) []livenessNamedObl {
	var out []livenessNamedObl
	for _, event := range sortedEvents(sim.Class) {
		out = append(out, livenessNamedObl{
			classKey: sim.ClassKey, className: sim.Class.Name,
			memberKey: event.Key, name: event.Name,
		})
	}
	return out
}

func classQueryObligations(sim *schema.ClassSimInfo) []livenessNamedObl {
	var out []livenessNamedObl
	for _, query := range sortedQueries(sim.Class) {
		out = append(out, livenessNamedObl{
			classKey: sim.ClassKey, className: sim.Class.Name,
			memberKey: query.Key, name: query.Name,
		})
	}
	return out
}

func classActionObligations(sim *schema.ClassSimInfo) []livenessNamedObl {
	var out []livenessNamedObl
	for _, action := range sortedActions(sim.Class) {
		out = append(out, livenessNamedObl{
			classKey: sim.ClassKey, className: sim.Class.Name,
			memberKey: action.Key, name: action.Name,
		})
	}
	return out
}

func installTransitionMaps(obl *livenessObligations, sch *schema.Schema, sim *schema.ClassSimInfo) map[identity.Key]bool {
	externalCreation := make(map[identity.Key]bool)
	for _, ev := range sch.ExternalCreationEvents(sim.ClassKey) {
		externalCreation[ev.Key] = true
	}
	samplable := make(map[identity.Key]bool)
	for _, t := range sim.Class.Transitions {
		if t.ActionKey == nil {
			continue
		}
		obl.transitionToAction[t.Key] = *t.ActionKey
		if t.FromStateKey == nil {
			if externalCreation[t.EventKey] {
				samplable[*t.ActionKey] = true
			}
			continue
		}
		samplable[*t.ActionKey] = true
	}
	return samplable
}

func classParamObligations(sim *schema.ClassSimInfo, samplable map[identity.Key]bool) []livenessParamObl {
	var out []livenessParamObl
	for _, action := range sim.Class.Actions {
		if !samplable[action.Key] {
			continue
		}
		for _, param := range action.Parameters {
			if !paramHasSimulationSpecification(param) {
				continue
			}
			out = append(out, livenessParamObl{
				classKey:   sim.ClassKey,
				className:  sim.Class.Name,
				actionName: action.Name,
				paramKey:   param.Key,
				paramName:  param.Name,
			})
		}
	}
	return out
}

func sortedEvents(class model_class.Class) []model_state.Event {
	events := make([]model_state.Event, 0, len(class.Events))
	for _, event := range class.Events {
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Name < events[j].Name })
	return events
}

func sortedQueries(class model_class.Class) []model_state.Query {
	queries := make([]model_state.Query, 0, len(class.Queries))
	for _, query := range class.Queries {
		queries = append(queries, query)
	}
	sort.Slice(queries, func(i, j int) bool { return queries[i].Name < queries[j].Name })
	return queries
}

func sortedActions(class model_class.Class) []model_state.Action {
	actions := make([]model_state.Action, 0, len(class.Actions))
	for _, action := range class.Actions {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(i, j int) bool { return actions[i].Name < actions[j].Name })
	return actions
}

func paramHasSimulationSpecification(param model_state.Parameter) bool {
	if param.Simulation == nil {
		return false
	}
	for _, rule := range param.Simulation.Rules {
		if rule.HasSpecification() {
			return true
		}
	}
	return false
}

// CollectLivenessHits folds step records into hits against this state's links and
// transition→action map (installed at NewState). Does not query schema.
func (s *State) CollectLivenessHits(steps []LivenessStep, usedParams map[identity.Key]bool) LivenessHits {
	hits := LivenessHits{
		Instantiated:         make(map[identity.Key]bool),
		WrittenAttrs:         make(map[identity.Key]map[string]bool),
		LinkedAssocs:         s.LinkedAssociationKeys(),
		Events:               make(map[identity.Key]bool),
		Queries:              make(map[identity.Key]bool),
		DerivedReads:         make(map[identity.Key]bool),
		Actions:              make(map[identity.Key]bool),
		UsedSimulationParams: usedParams,
	}
	if hits.UsedSimulationParams == nil {
		hits.UsedSimulationParams = make(map[identity.Key]bool)
	}
	var transitionMap map[identity.Key]identity.Key
	if s != nil && s.liveness != nil {
		transitionMap = s.liveness.transitionToAction
	}
	collectStepHits(steps, transitionMap, &hits)
	return hits
}

func collectStepHits(steps []LivenessStep, transitionToAction map[identity.Key]identity.Key, hits *LivenessHits) {
	for _, step := range steps {
		if step.IsCreation {
			hits.Instantiated[step.ClassKey] = true
		}
		if step.EventName != "" {
			hits.Events[step.EventKey] = true
		}
		if step.QueryName != "" {
			hits.Queries[step.QueryKey] = true
		}
		if step.DerivedAttributeName != "" {
			hits.DerivedReads[step.DerivedAttributeKey] = true
		}
		for _, ak := range step.ExecutedActionKeys {
			hits.Actions[ak] = true
		}
		if step.HasTransitionAction && transitionToAction != nil {
			if ak, ok := transitionToAction[step.TransitionKey]; ok {
				hits.Actions[ak] = true
			}
		}
		if len(step.PrimedAttrSubKeys) > 0 {
			if hits.WrittenAttrs[step.ClassKey] == nil {
				hits.WrittenAttrs[step.ClassKey] = make(map[string]bool)
			}
			for _, sk := range step.PrimedAttrSubKeys {
				hits.WrittenAttrs[step.ClassKey][identity.NormalizeSubKey(sk)] = true
			}
		}
		if len(step.Cascaded) > 0 {
			collectStepHits(step.Cascaded, transitionToAction, hits)
		}
	}
}

// LinkedAssociationKeys returns association key strings that have at least one link.
func (s *State) LinkedAssociationKeys() map[string]bool {
	out := make(map[string]bool)
	if s == nil {
		return out
	}
	for k := range s.links.AllAssociationKeys() {
		out[string(k)] = true
	}
	for hostKey := range s.associationLinks.AllHostAssociationKeys() {
		out[string(hostKey)] = true
	}
	return out
}

// CheckLiveness compares installed obligations against hits. Does not use schema.
func (s *State) CheckLiveness(hits LivenessHits) ViolationErrors {
	if s == nil || s.liveness == nil {
		return nil
	}
	obl := s.liveness
	var violations ViolationErrors
	violations = append(violations, checkClassLiveness(obl, hits)...)
	violations = append(violations, checkAttributeLiveness(obl, hits)...)
	violations = append(violations, checkAssociationLiveness(obl, hits)...)
	violations = append(violations, checkMemberLiveness(obl, hits)...)
	violations = append(violations, checkParamLiveness(obl, hits)...)
	return violations
}

func checkClassLiveness(obl *livenessObligations, hits LivenessHits) ViolationErrors {
	var violations ViolationErrors
	for _, c := range obl.classes {
		if !hits.Instantiated[c.key] {
			violations = append(violations, NewLivenessClassNotInstantiatedViolation(c.key, c.name))
		}
	}
	return violations
}

func checkAttributeLiveness(obl *livenessObligations, hits LivenessHits) ViolationErrors {
	var violations ViolationErrors
	for _, a := range obl.attributes {
		written := hits.WrittenAttrs[a.classKey]
		if written == nil || !written[a.subKey] {
			violations = append(violations, NewLivenessAttributeNotWrittenViolation(a.classKey, a.className, a.name))
		}
	}
	return violations
}

func checkAssociationLiveness(obl *livenessObligations, hits LivenessHits) ViolationErrors {
	var violations ViolationErrors
	for _, assoc := range obl.associations {
		// Link tables store evaluator.AssociationKey form of the model key string.
		assocKeyStr := string(evaluator.AssociationKey(assoc.keyStr))
		if !hits.LinkedAssocs[assocKeyStr] && !hits.LinkedAssocs[assoc.keyStr] {
			violations = append(violations, NewLivenessAssociationNotLinkedViolation(
				assoc.key, assoc.name, assoc.fromKey, assoc.toKey,
			))
		}
	}
	return violations
}

func checkMemberLiveness(obl *livenessObligations, hits LivenessHits) ViolationErrors {
	var violations ViolationErrors
	for _, e := range obl.events {
		if !hits.Events[e.memberKey] {
			violations = append(violations, NewLivenessEventNotSentViolation(e.classKey, e.className, e.name))
		}
	}
	for _, q := range obl.queries {
		if !hits.Queries[q.memberKey] {
			violations = append(violations, NewLivenessQueryNotRunViolation(q.classKey, q.className, q.name))
		}
	}
	for _, d := range obl.derived {
		if !hits.DerivedReads[d.memberKey] {
			violations = append(violations, NewLivenessAttributeNotReadViolation(d.classKey, d.className, d.name))
		}
	}
	for _, a := range obl.actions {
		if !hits.Actions[a.memberKey] {
			violations = append(violations, NewLivenessActionNotExecutedViolation(a.classKey, a.className, a.name))
		}
	}
	return violations
}

func checkParamLiveness(obl *livenessObligations, hits LivenessHits) ViolationErrors {
	var violations ViolationErrors
	for _, p := range obl.params {
		if !hits.UsedSimulationParams[p.paramKey] {
			violations = append(violations, NewLivenessParameterSimulationNotUsedViolation(
				p.classKey, p.className, p.actionName, p.paramName,
			))
		}
	}
	return violations
}
