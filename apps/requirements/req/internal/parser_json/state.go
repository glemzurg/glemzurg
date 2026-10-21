package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeState struct {
	Key        string            `json:"key"`
	Name       string            `json:"name"`
	Details    string            `json:"details,omitempty"`
	UmlComment string            `json:"uml_comment,omitempty"`
	Actions    []pipeStateAction `json:"actions,omitempty"`
}

type pipeStateAction struct {
	Key       string `json:"key"`
	ActionKey string `json:"action_key"`
	When      string `json:"when"`
}

type pipeEvent struct {
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	Details        string   `json:"details,omitempty"`
	ParameterNames []string `json:"parameter_names,omitempty"`
}

type pipeGuard struct {
	Key   string    `json:"key"`
	Name  string    `json:"name"`
	Logic pipeLogic `json:"logic"`
}

type pipeAction struct {
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Details     string          `json:"details,omitempty"`
	Parameters  []pipeParameter `json:"parameters,omitempty"`
	Requires    []pipeLogic     `json:"requires,omitempty"`
	Guarantees  []pipeLogic     `json:"guarantees,omitempty"`
	SafetyRules []pipeLogic     `json:"safety_rules,omitempty"`
}

type pipeQuery struct {
	Key        string          `json:"key"`
	Name       string          `json:"name"`
	Details    string          `json:"details,omitempty"`
	Parameters []pipeParameter `json:"parameters,omitempty"`
	Requires   []pipeLogic     `json:"requires,omitempty"`
	Guarantees []pipeLogic     `json:"guarantees,omitempty"`
}

// pipeParameter is an action or query parameter. The parsed data type is omitted;
// NewParameter rebuilds it from data_type_rules.
type pipeParameter struct {
	Key           string                   `json:"key"`
	Name          string                   `json:"name"`
	DataTypeRules string                   `json:"data_type_rules,omitempty"`
	Nullable      bool                     `json:"nullable,omitempty"`
	Invariants    []pipeLogic              `json:"invariants,omitempty"`
	Simulation    *pipeParameterSimulation `json:"simulation,omitempty"`
}

type pipeParameterSimulation struct {
	Details string                        `json:"details,omitempty"`
	Rules   []pipeParameterSimulationRule `json:"rules,omitempty"`
}

type pipeParameterSimulationRule struct {
	Details       string      `json:"details,omitempty"`
	Requires      []pipeLogic `json:"requires,omitempty"`
	Specification *pipeLogic  `json:"specification,omitempty"`
}

type pipeTransition struct {
	Key          string `json:"key"`
	FromStateKey string `json:"from_state_key,omitempty"`
	EventKey     string `json:"event_key"`
	GuardKey     string `json:"guard_key,omitempty"`
	ActionKey    string `json:"action_key,omitempty"`
	ToStateKey   string `json:"to_state_key,omitempty"`
	UmlComment   string `json:"uml_comment,omitempty"`
}

func setBehaviorFromCore(item *pipeClass, class model_class.Class) {
	item.States = statesFromCore(class.States)
	item.Events = eventsFromCore(class.Events)
	item.Guards = guardsFromCore(class.Guards)
	item.Actions = actionsFromCore(class.Actions)
	item.Queries = queriesFromCore(class.Queries)
	item.Transitions = transitionsFromCore(class.Transitions)
}

func setBehaviorToCore(class *model_class.Class, item pipeClass) error {
	if err := setStateGraphToCore(class, item); err != nil {
		return err
	}
	return setStateOperationsToCore(class, item)
}

func setStateGraphToCore(class *model_class.Class, item pipeClass) error {
	states, err := statesToCore(item.States)
	if err != nil {
		return err
	}
	events, err := eventsToCore(item.Events)
	if err != nil {
		return err
	}
	transitions, err := transitionsToCore(item.Transitions)
	if err != nil {
		return err
	}
	class.SetStates(states)
	class.SetEvents(events)
	class.SetTransitions(transitions)
	return nil
}

func setStateOperationsToCore(class *model_class.Class, item pipeClass) error {
	guards, err := guardsToCore(item.Guards)
	if err != nil {
		return err
	}
	actions, err := actionsToCore(item.Actions)
	if err != nil {
		return err
	}
	queries, err := queriesToCore(item.Queries)
	if err != nil {
		return err
	}
	class.SetGuards(guards)
	class.SetActions(actions)
	class.SetQueries(queries)
	return nil
}

func statesFromCore(states map[identity.Key]model_state.State) []pipeState {
	if len(states) == 0 {
		return nil
	}
	out := make([]pipeState, 0, len(states))
	for _, state := range states {
		out = append(out, stateFromCore(state))
	}
	sortByKey(out, func(item pipeState) string { return item.Key })
	return out
}

func stateFromCore(state model_state.State) pipeState {
	item := pipeState{
		Key:        keyString(state.Key),
		Name:       state.Name,
		Details:    state.Details,
		UmlComment: state.UmlComment,
	}
	for _, action := range state.Actions {
		item.Actions = append(item.Actions, pipeStateAction{
			Key:       keyString(action.Key),
			ActionKey: keyString(action.ActionKey),
			When:      action.When,
		})
	}
	return item
}

func statesToCore(items []pipeState) (map[identity.Key]model_state.State, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.State, len(items))
	for i, item := range items {
		state, err := stateToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "states[%d]", i)
		}
		if err := putUnique(out, state.Key, state, i); err != nil {
			return nil, errors.Wrap(err, "states")
		}
	}
	return out, nil
}

func stateToCore(item pipeState) (model_state.State, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_state.State{}, err
	}
	actions, err := stateActionsToCore(item.Actions)
	if err != nil {
		return model_state.State{}, err
	}
	state := model_state.NewState(key, item.Name, item.Details, item.UmlComment)
	state.SetActions(actions)
	return state, nil
}

func stateActionsToCore(items []pipeStateAction) ([]model_state.StateAction, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_state.StateAction, 0, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "actions[%d]", i)
		}
		actionKey, err := parseKey("action_key", item.ActionKey)
		if err != nil {
			return nil, errors.Wrapf(err, "actions[%d]", i)
		}
		out = append(out, model_state.NewStateAction(key, actionKey, item.When))
	}
	return out, nil
}

func eventsFromCore(events map[identity.Key]model_state.Event) []pipeEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]pipeEvent, 0, len(events))
	for _, event := range events {
		out = append(out, pipeEvent{
			Key:            keyString(event.Key),
			Name:           event.Name,
			Details:        event.Details,
			ParameterNames: event.ParameterNames,
		})
	}
	sortByKey(out, func(item pipeEvent) string { return item.Key })
	return out
}

func eventsToCore(items []pipeEvent) (map[identity.Key]model_state.Event, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.Event, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "events[%d]", i)
		}
		event := model_state.NewEvent(key, item.Name, item.Details, item.ParameterNames)
		if err := putUnique(out, event.Key, event, i); err != nil {
			return nil, errors.Wrap(err, "events")
		}
	}
	return out, nil
}

func guardsFromCore(guards map[identity.Key]model_state.Guard) []pipeGuard {
	if len(guards) == 0 {
		return nil
	}
	out := make([]pipeGuard, 0, len(guards))
	for _, guard := range guards {
		out = append(out, pipeGuard{
			Key:   keyString(guard.Key),
			Name:  guard.Name,
			Logic: logicFromCore(guard.Logic),
		})
	}
	sortByKey(out, func(item pipeGuard) string { return item.Key })
	return out
}

func guardsToCore(items []pipeGuard) (map[identity.Key]model_state.Guard, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.Guard, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "guards[%d]", i)
		}
		logic, err := logicToCore(item.Logic)
		if err != nil {
			return nil, errors.Wrapf(err, "guards[%d].logic", i)
		}
		guard := model_state.NewGuard(key, item.Name, logic)
		if err := putUnique(out, guard.Key, guard, i); err != nil {
			return nil, errors.Wrap(err, "guards")
		}
	}
	return out, nil
}

func actionsFromCore(actions map[identity.Key]model_state.Action) []pipeAction {
	if len(actions) == 0 {
		return nil
	}
	out := make([]pipeAction, 0, len(actions))
	for _, action := range actions {
		out = append(out, pipeAction{
			Key:         keyString(action.Key),
			Name:        action.Name,
			Details:     action.Details,
			Parameters:  parametersFromCore(action.Parameters),
			Requires:    logicsFromCore(action.Requires),
			Guarantees:  logicsFromCore(action.Guarantees),
			SafetyRules: logicsFromCore(action.SafetyRules),
		})
	}
	sortByKey(out, func(item pipeAction) string { return item.Key })
	return out
}

func actionsToCore(items []pipeAction) (map[identity.Key]model_state.Action, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.Action, len(items))
	for i, item := range items {
		action, err := actionToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "actions[%d]", i)
		}
		if err := putUnique(out, action.Key, action, i); err != nil {
			return nil, errors.Wrap(err, "actions")
		}
	}
	return out, nil
}

func actionToCore(item pipeAction) (model_state.Action, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_state.Action{}, err
	}
	parameters, err := parametersToCore(key, item.Parameters)
	if err != nil {
		return model_state.Action{}, err
	}
	requires, err := logicsToCore(item.Requires)
	if err != nil {
		return model_state.Action{}, errors.Wrap(err, "requires")
	}
	guarantees, err := logicsToCore(item.Guarantees)
	if err != nil {
		return model_state.Action{}, errors.Wrap(err, "guarantees")
	}
	safetyRules, err := logicsToCore(item.SafetyRules)
	if err != nil {
		return model_state.Action{}, errors.Wrap(err, "safety_rules")
	}
	return model_state.NewAction(key, model_state.ActionDetails{
		Name:    item.Name,
		Details: item.Details,
	}, requires, guarantees, safetyRules, parameters), nil
}

func queriesFromCore(queries map[identity.Key]model_state.Query) []pipeQuery {
	if len(queries) == 0 {
		return nil
	}
	out := make([]pipeQuery, 0, len(queries))
	for _, query := range queries {
		out = append(out, pipeQuery{
			Key:        keyString(query.Key),
			Name:       query.Name,
			Details:    query.Details,
			Parameters: parametersFromCore(query.Parameters),
			Requires:   logicsFromCore(query.Requires),
			Guarantees: logicsFromCore(query.Guarantees),
		})
	}
	sortByKey(out, func(item pipeQuery) string { return item.Key })
	return out
}

func queriesToCore(items []pipeQuery) (map[identity.Key]model_state.Query, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.Query, len(items))
	for i, item := range items {
		query, err := queryToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "queries[%d]", i)
		}
		if err := putUnique(out, query.Key, query, i); err != nil {
			return nil, errors.Wrap(err, "queries")
		}
	}
	return out, nil
}

func queryToCore(item pipeQuery) (model_state.Query, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_state.Query{}, err
	}
	parameters, err := parametersToCore(key, item.Parameters)
	if err != nil {
		return model_state.Query{}, err
	}
	requires, err := logicsToCore(item.Requires)
	if err != nil {
		return model_state.Query{}, errors.Wrap(err, "requires")
	}
	guarantees, err := logicsToCore(item.Guarantees)
	if err != nil {
		return model_state.Query{}, errors.Wrap(err, "guarantees")
	}
	return model_state.NewQuery(key, item.Name, item.Details, requires, guarantees, parameters), nil
}

func parametersFromCore(parameters []model_state.Parameter) []pipeParameter {
	if len(parameters) == 0 {
		return nil
	}
	out := make([]pipeParameter, 0, len(parameters))
	for _, param := range parameters {
		out = append(out, pipeParameter{
			Key:           keyString(param.Key),
			Name:          param.Name,
			DataTypeRules: param.DataTypeRules,
			Nullable:      param.Nullable,
			Invariants:    logicsFromCore(param.Invariants),
			Simulation:    simulationFromCore(param.Simulation),
		})
	}
	return out
}

func parametersToCore(parent identity.Key, items []pipeParameter) ([]model_state.Parameter, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_state.Parameter, 0, len(items))
	for i, item := range items {
		param, err := parameterToCore(parent, item)
		if err != nil {
			return nil, errors.Wrapf(err, "parameters[%d]", i)
		}
		out = append(out, param)
	}
	return out, nil
}

func parameterToCore(parent identity.Key, item pipeParameter) (model_state.Parameter, error) {
	// NewParameter derives the key from the parent and the normalized name.
	// The pipe still carries the key so a mismatch is a bad document rather than a renamed parameter.
	param, err := model_state.NewParameter(parent, item.Name, item.DataTypeRules, item.Nullable)
	if err != nil {
		return model_state.Parameter{}, err
	}
	if keyString(param.Key) != item.Key {
		return model_state.Parameter{}, errors.Errorf("parameter key %q does not match name %q", item.Key, item.Name)
	}
	invariants, err := logicsToCore(item.Invariants)
	if err != nil {
		return model_state.Parameter{}, errors.Wrap(err, "invariants")
	}
	param.SetInvariants(invariants)
	simulation, err := simulationToCore(item.Simulation)
	if err != nil {
		return model_state.Parameter{}, err
	}
	param.SetSimulation(simulation)
	return param, nil
}

func simulationFromCore(simulation *model_state.ParameterSimulation) *pipeParameterSimulation {
	if simulation == nil || !simulation.HasSimulation() {
		return nil
	}
	item := &pipeParameterSimulation{Details: simulation.Details}
	for _, rule := range simulation.Rules {
		encoded := pipeParameterSimulationRule{
			Details:  rule.Details,
			Requires: logicsFromCore(rule.Requires),
		}
		if rule.Specification != nil {
			logic := logicFromCore(*rule.Specification)
			encoded.Specification = &logic
		}
		item.Rules = append(item.Rules, encoded)
	}
	return item
}

func simulationToCore(item *pipeParameterSimulation) (*model_state.ParameterSimulation, error) {
	if item == nil {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	rules := make([]model_state.ParameterSimulationRule, 0, len(item.Rules))
	for i, rule := range item.Rules {
		requires, err := logicsToCore(rule.Requires)
		if err != nil {
			return nil, errors.Wrapf(err, "simulation.rules[%d].requires", i)
		}
		var specification *model_logic.Logic
		if rule.Specification != nil {
			logic, err := logicToCore(*rule.Specification)
			if err != nil {
				return nil, errors.Wrapf(err, "simulation.rules[%d].specification", i)
			}
			specification = &logic
		}
		rules = append(rules, model_state.ParameterSimulationRule{
			Details:       rule.Details,
			Requires:      requires,
			Specification: specification,
		})
	}
	if len(rules) == 0 {
		rules = nil
	}
	return &model_state.ParameterSimulation{
		Details: item.Details,
		Rules:   rules,
	}, nil
}

func transitionsFromCore(transitions map[identity.Key]model_state.Transition) []pipeTransition {
	if len(transitions) == 0 {
		return nil
	}
	out := make([]pipeTransition, 0, len(transitions))
	for _, transition := range transitions {
		out = append(out, pipeTransition{
			Key:          keyString(transition.Key),
			FromStateKey: optionalKeyString(transition.FromStateKey),
			EventKey:     keyString(transition.EventKey),
			GuardKey:     optionalKeyString(transition.GuardKey),
			ActionKey:    optionalKeyString(transition.ActionKey),
			ToStateKey:   optionalKeyString(transition.ToStateKey),
			UmlComment:   transition.UmlComment,
		})
	}
	sortByKey(out, func(item pipeTransition) string { return item.Key })
	return out
}

func transitionsToCore(items []pipeTransition) (map[identity.Key]model_state.Transition, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_state.Transition, len(items))
	for i, item := range items {
		transition, err := transitionToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "transitions[%d]", i)
		}
		if err := putUnique(out, transition.Key, transition, i); err != nil {
			return nil, errors.Wrap(err, "transitions")
		}
	}
	return out, nil
}

func transitionToCore(item pipeTransition) (model_state.Transition, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_state.Transition{}, err
	}
	eventKey, err := parseKey("event_key", item.EventKey)
	if err != nil {
		return model_state.Transition{}, err
	}
	from, err := optionalKeyToCore("from_state_key", item.FromStateKey)
	if err != nil {
		return model_state.Transition{}, err
	}
	to, err := optionalKeyToCore("to_state_key", item.ToStateKey)
	if err != nil {
		return model_state.Transition{}, err
	}
	guard, err := optionalKeyToCore("guard_key", item.GuardKey)
	if err != nil {
		return model_state.Transition{}, err
	}
	action, err := optionalKeyToCore("action_key", item.ActionKey)
	if err != nil {
		return model_state.Transition{}, err
	}
	return model_state.NewTransition(key, eventKey, model_state.TransitionStateKeys{
		FromStateKey: from,
		ToStateKey:   to,
	}, model_state.TransitionLogicKeys{
		GuardKey:  guard,
		ActionKey: action,
	}, item.UmlComment), nil
}
