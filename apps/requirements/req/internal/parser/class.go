package parser

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/view_helper"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseClass(subdomainKey identity.Key, classSubKey, filename, contents string) (class model_class.Class, associations []model_class.Association, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_class.Class{}, nil, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_class.Class{}, nil, errors.WithStack(err)
	}

	// Parse optional key references from YAML (stored as strings, converted to keys later as needed).
	var actorKey *identity.Key
	actorAny, found := yamlData["actor_key"]
	if found {
		actorKeyStr := actorAny.(string)
		if actorKeyStr != "" {
			key, err := identity.NewActorKey(actorKeyStr)
			if err != nil {
				return model_class.Class{}, nil, errors.WithStack(err)
			}
			actorKey = &key
		}
	}

	var superclassOfKey *identity.Key
	superclassOfAny, found := yamlData["superclass_of_key"]
	if found {
		superclassOfStr := superclassOfAny.(string)
		if superclassOfStr != "" {
			// If it's a simple key (no slashes), construct a generalization key in the same subdomain.
			var key identity.Key
			if !strings.Contains(superclassOfStr, "/") {
				key, err = identity.NewGeneralizationKey(subdomainKey, superclassOfStr)
			} else {
				key, err = identity.ParseKey(superclassOfStr)
			}
			if err != nil {
				return model_class.Class{}, nil, errors.WithStack(err)
			}
			superclassOfKey = &key
		}
	}

	var subclassOfKey *identity.Key
	subclassOfAny, found := yamlData["subclass_of_key"]
	if found {
		subclassOfStr := subclassOfAny.(string)
		if subclassOfStr != "" {
			// If it's a simple key (no slashes), construct a generalization key in the same subdomain.
			var key identity.Key
			if !strings.Contains(subclassOfStr, "/") {
				key, err = identity.NewGeneralizationKey(subdomainKey, subclassOfStr)
			} else {
				key, err = identity.ParseKey(subclassOfStr)
			}
			if err != nil {
				return model_class.Class{}, nil, errors.WithStack(err)
			}
			subclassOfKey = &key
		}
	}

	// Construct the identity key for this class.
	classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
	if err != nil {
		return model_class.Class{}, nil, errors.WithStack(err)
	}

	class, err = model_class.NewClass(classKey, parsedFile.Title, stripMarkdownTitle(parsedFile.Markdown), actorKey, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
	if err != nil {
		return model_class.Class{}, nil, err
	}

	// Add any invariants we found.
	invariants, err := logicListFromYamlData(yamlData, "invariants", classKey, identity.NewClassInvariantKey)
	if err != nil {
		return model_class.Class{}, nil, err
	}
	class.SetInvariants(invariants)

	// Add any attributes we found.
	var attributesData map[string]any
	attributesAny, found := yamlData["attributes"]
	if found {
		attributesData = attributesAny.(map[string]any)
	}

	attributes := make(map[identity.Key]model_class.Attribute)
	for attrSubKey, attributeAny := range attributesData {
		attribute, err := attributeFromYamlData(classKey, attrSubKey, attributeAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		attributes[attribute.Key] = attribute
	}
	class.SetAttributes(attributes)

	// Add any associations we found (returned separately, not stored in class).
	var associationsData []any
	associationsAny, found := yamlData["associations"]
	if found {
		associationsData = associationsAny.([]any)
	}

	for i, associationAny := range associationsData {
		association, err := associationFromYamlData(subdomainKey, classKey, i, associationAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		associations = append(associations, association)
	}

	// Add any actions we found.
	var actionsData map[string]any
	actionsAny, found := yamlData["actions"]
	if found {
		actionsData = actionsAny.(map[string]any)
	}

	actions := make(map[identity.Key]model_state.Action)
	actionKeyLookup := map[string]identity.Key{}
	for name, actionAny := range actionsData {
		action, err := actionFromYamlData(classKey, name, actionAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		actionKeyLookup[action.Name] = action.Key
		actions[action.Key] = action
	}
	class.SetActions(actions)

	// Add any states we found.
	var statesData map[string]any
	statesAny, found := yamlData["states"]
	if found {
		statesData = statesAny.(map[string]any)
	}

	states := make(map[identity.Key]model_state.State)
	stateKeyLookup := map[string]identity.Key{}
	for name, stateAny := range statesData {
		state, err := stateFromYamlData(actionKeyLookup, classKey, name, stateAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		stateKeyLookup[state.Name] = state.Key
		states[state.Key] = state
	}
	class.SetStates(states)

	// Add any events we found.
	var eventsData map[string]any
	eventsAny, found := yamlData["events"]
	if found {
		eventsData = eventsAny.(map[string]any)
	}

	events := make(map[identity.Key]model_state.Event)
	eventKeyLookup := map[string]identity.Key{}
	for name, eventAny := range eventsData {
		event, err := eventFromYamlData(classKey, name, eventAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		eventKeyLookup[event.Name] = event.Key
		events[event.Key] = event
	}
	class.SetEvents(events)

	// Add any guards we found.
	var guardsData map[string]any
	guardsAny, found := yamlData["guards"]
	if found {
		guardsData = guardsAny.(map[string]any)
	}

	guards := make(map[identity.Key]model_state.Guard)
	guardKeyLookup := map[string]identity.Key{}
	for name, guardAny := range guardsData {
		guard, err := guardFromYamlData(classKey, name, guardAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		guardKeyLookup[guard.Name] = guard.Key
		guards[guard.Key] = guard
	}
	class.SetGuards(guards)

	// Add any queries we found.
	var queriesData map[string]any
	queriesAny, found := yamlData["queries"]
	if found {
		queriesData = queriesAny.(map[string]any)
	}

	queries := make(map[identity.Key]model_state.Query)
	for name, queryAny := range queriesData {
		query, err := queryFromYamlData(classKey, name, queryAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		queries[query.Key] = query
	}
	class.SetQueries(queries)

	// Add any transitions we found.
	var transitionsData []any
	transitionsAny, found := yamlData["transitions"]
	if found {
		transitionsData = transitionsAny.([]any)
	}

	transitions := make(map[identity.Key]model_state.Transition)
	for i, transitionAny := range transitionsData {
		transition, err := transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup, classKey, i, transitionAny)
		if err != nil {
			return model_class.Class{}, nil, err
		}
		transitions[transition.Key] = transition
	}
	class.SetTransitions(transitions)

	return class, associations, nil
}

func attributeFromYamlData(classKey identity.Key, attrSubKey string, attributeAny any) (attribute model_class.Attribute, err error) {

	attributeData, ok := attributeAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		name := ""
		nameAny, found := attributeData["name"]
		if found {
			name = nameAny.(string)
		}

		details := ""
		detailsAny, found := attributeData["details"]
		if found {
			details = detailsAny.(string)
		}

		dataTypeRules := ""
		dataTypeRulesAny, found := attributeData["rules"]
		if found {
			dataTypeRules = dataTypeRulesAny.(string)
		}

		// Parse derivation policy as *model_logic.Logic.
		var derivationPolicy *model_logic.Logic
		derivationAny, found := attributeData["derivation"]
		if found {
			derivationMap, ok := derivationAny.(map[string]any)
			if ok {
				description := ""
				if descAny, ok := derivationMap["description"]; ok {
					description = descAny.(string)
				}
				specification := ""
				if specAny, ok := derivationMap["specification"]; ok {
					specification = specAny.(string)
				}
				// Construct the derivation key as a child of the attribute key.
				attrKey, err := identity.NewAttributeKey(classKey, attrSubKey)
				if err != nil {
					return model_class.Attribute{}, errors.WithStack(err)
				}
				derivKey, err := identity.NewAttributeDerivationKey(attrKey, "derivation")
				if err != nil {
					return model_class.Attribute{}, errors.WithStack(err)
				}
				logic, err := model_logic.NewLogic(derivKey, description, "tla_plus", specification)
				if err != nil {
					return model_class.Attribute{}, errors.Wrap(err, "failed to create derivation policy logic")
				}
				derivationPolicy = &logic
			}
		}

		nullable := false
		nullableAny, found := attributeData["nullable"]
		if found {
			nullable = nullableAny.(bool)
		}

		umlComment := ""
		umlCommentAny, found := attributeData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		var indexNums []uint
		indexNumsAny, found := attributeData["index_nums"]
		if found {
			indexNumsAnyList := indexNumsAny.([]any)
			for _, indexNumAny := range indexNumsAnyList {
				indexNumInt := indexNumAny.(int)
				indexNums = append(indexNums, uint(indexNumInt))
			}
		}

		// Construct the identity key for this attribute.
		attrKey, err := identity.NewAttributeKey(classKey, attrSubKey)
		if err != nil {
			return model_class.Attribute{}, errors.WithStack(err)
		}

		attribute, err = model_class.NewAttribute(
			attrKey,
			name,
			details,
			dataTypeRules,
			derivationPolicy,
			nullable,
			umlComment,
			indexNums)
		if err != nil {
			return model_class.Attribute{}, err
		}
	}

	return attribute, nil
}

func associationFromYamlData(subdomainKey, fromClassKey identity.Key, index int, associationAny any) (association model_class.Association, err error) {

	associationData, ok := associationAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		_ = strconv.Itoa(index + 1) // Don't start at zero (kept for reference but key constructed differently now).

		name := ""
		nameAny, found := associationData["name"]
		if found {
			name = nameAny.(string)
		}

		details := ""
		detailsAny, found := associationData["details"]
		if found {
			details = detailsAny.(string)
		}

		fromMultiplicityValue := ""
		fromMultiplicityAny, found := associationData["from_multiplicity"]
		if found {
			fromMultiplicityValue = fromMultiplicityAny.(string)
		}
		fromMultiplicity, err := model_class.NewMultiplicity(fromMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
		}

		toClassKeyStr := ""
		toClassKeyAny, found := associationData["to_class_key"]
		if found {
			toClassKeyStr = toClassKeyAny.(string)
		}

		toMultiplicityValue := ""
		toMultiplicityAny, found := associationData["to_multiplicity"]
		if found {
			toMultiplicityValue = toMultiplicityAny.(string)
		}
		toMultiplicity, err := model_class.NewMultiplicity(toMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
		}

		// Resolve the to-class key based on the path format.
		// Simple key (no /): same subdomain.
		// Starts with "subdomain/": same domain, different subdomain — prepend domain prefix.
		// Starts with "domain/": different domain — full key, parse directly.
		toClassKey, err := resolveClassKeyFromRelative(subdomainKey, toClassKeyStr)
		if err != nil {
			return model_class.Association{}, err
		}

		// Determine the association parent key based on which classes are connected.
		assocParentKey, err := determineAssociationParent(subdomainKey, fromClassKey, toClassKey)
		if err != nil {
			return model_class.Association{}, err
		}

		// Parse association class key if present (uses same relative format).
		var associationClassKey *identity.Key
		associationClassKeyAny, found := associationData["association_class_key"]
		if found {
			associationClassKeyStr := associationClassKeyAny.(string)
			if associationClassKeyStr != "" {
				key, err := resolveClassKeyFromRelative(subdomainKey, associationClassKeyStr)
				if err != nil {
					return model_class.Association{}, errors.WithStack(err)
				}
				associationClassKey = &key
			}
		}

		umlComment := ""
		umlCommentAny, found := associationData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		assocKey, err := identity.NewClassAssociationKey(assocParentKey, fromClassKey, toClassKey, name)
		if err != nil {
			return model_class.Association{}, errors.WithStack(err)
		}

		association, err = model_class.NewAssociation(
			assocKey,
			name,
			details,
			fromClassKey,
			fromMultiplicity,
			toClassKey,
			toMultiplicity,
			associationClassKey,
			umlComment)
		if err != nil {
			return model_class.Association{}, err
		}
	}

	return association, nil
}

// resolveClassKeyFromRelative resolves a class key string relative to a subdomain.
// Simple key (no /): same subdomain → NewClassKey(subdomainKey, str).
// Starts with "subdomain/": same domain, different subdomain → prepend domain prefix and parse.
// Starts with "domain/": different domain → parse as full key.
func resolveClassKeyFromRelative(subdomainKey identity.Key, keyStr string) (identity.Key, error) {
	if !strings.Contains(keyStr, "/") {
		// Simple key: same subdomain.
		return identity.NewClassKey(subdomainKey, keyStr)
	}
	if strings.HasPrefix(keyStr, identity.KEY_TYPE_SUBDOMAIN+"/") {
		// Relative to domain: prepend the domain prefix.
		fullKeyStr := subdomainKey.ParentKey + "/" + keyStr
		return identity.ParseKey(fullKeyStr)
	}
	// Full key (starts with "domain/" or similar): parse directly.
	return identity.ParseKey(keyStr)
}

// determineAssociationParent determines the correct parent key for a class association
// based on the relationship between the from-class and to-class.
// Same subdomain → subdomain parent. Same domain → domain parent. Different domains → empty (model) parent.
func determineAssociationParent(subdomainKey, fromClassKey, toClassKey identity.Key) (identity.Key, error) {
	// Same subdomain: both class keys have the same parent.
	if fromClassKey.ParentKey == toClassKey.ParentKey {
		return subdomainKey, nil
	}

	// Parse subdomain keys to check if same domain.
	fromSubParsed, err := identity.ParseKey(fromClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}
	toSubParsed, err := identity.ParseKey(toClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	if fromSubParsed.ParentKey == toSubParsed.ParentKey {
		// Same domain, different subdomain: use domain as parent.
		domainKey, err := identity.ParseKey(fromSubParsed.ParentKey)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return domainKey, nil
	}

	// Different domains: model-level (empty parent key).
	return identity.Key{}, nil
}

func stateFromYamlData(actionKeyLookup map[string]identity.Key, classKey identity.Key, name string, stateAny any) (state model_state.State, err error) {

	// Construct the state key.
	stateKey, err := identity.NewStateKey(classKey, strings.ToLower(name))
	if err != nil {
		return model_state.State{}, errors.WithStack(err)
	}

	details := ""
	umlComment := ""
	var actions []model_state.StateAction

	stateData, ok := stateAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := stateData["details"]
		if found {
			details = detailsAny.(string)
		}

		umlCommentAny, found := stateData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		actionsAny, found := stateData["actions"]
		if found {
			actionsData := actionsAny.([]any)
			for _, actionAny := range actionsData {
				actionData := actionAny.(map[string]any)

				var actionKey identity.Key
				actionName := ""
				actionNameAny, found := actionData["action"]
				if found {
					actionName = actionNameAny.(string)
					actionKey, found = actionKeyLookup[actionName]
					if !found {
						return model_state.State{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
					}
				}

				when := ""
				whenAny, found := actionData["when"]
				if found {
					when = whenAny.(string)
				}

				// Construct the state action key.
				stateActionKey, err := identity.NewStateActionKey(stateKey, strings.ToLower(when), strings.ToLower(actionName))
				if err != nil {
					return model_state.State{}, errors.WithStack(err)
				}

				action, err := model_state.NewStateAction(
					stateActionKey,
					actionKey,
					when,
				)
				if err != nil {
					return model_state.State{}, err
				}

				actions = append(actions, action)

			}
		}
	}

	state, err = model_state.NewState(
		stateKey,
		name,
		details,
		umlComment)
	if err != nil {
		return model_state.State{}, err
	}

	// Attach the actions.
	state.SetActions(actions)

	return state, nil
}

func eventFromYamlData(classKey identity.Key, name string, eventAny any) (event model_state.Event, err error) {

	// Construct the event key.
	eventKey, err := identity.NewEventKey(classKey, strings.ToLower(name))
	if err != nil {
		return model_state.Event{}, errors.WithStack(err)
	}

	details := ""
	var parameters []model_state.Parameter

	eventData, ok := eventAny.(map[string]any)
	if ok {
		detailsAny, found := eventData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse event parameters.
		parametersAny, found := eventData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Event{}, errors.Errorf("event '%s': parameters must be a sequence", name)
			}
			for _, paramAny := range paramsList {
				paramMap, ok := paramAny.(map[string]any)
				if !ok {
					return model_state.Event{}, errors.Errorf("event '%s': each parameter must be a mapping", name)
				}
				paramName, _ := paramMap["name"].(string)
				paramRules, _ := paramMap["rules"].(string)
				param, err := model_state.NewParameter(paramName, paramRules)
				if err != nil {
					return model_state.Event{}, errors.Wrapf(err, "event '%s' parameter '%s'", name, paramName)
				}
				parameters = append(parameters, param)
			}
		}
	}

	event, err = model_state.NewEvent(
		eventKey,
		name,
		details,
		parameters)
	if err != nil {
		return model_state.Event{}, err
	}

	return event, nil
}

func guardFromYamlData(classKey identity.Key, name string, guardAny any) (guard model_state.Guard, err error) {

	// Construct the guard key.
	guardKey, err := identity.NewGuardKey(classKey, strings.ToLower(name))
	if err != nil {
		return model_state.Guard{}, errors.WithStack(err)
	}

	details := ""

	specification := ""

	guardData, ok := guardAny.(map[string]any)
	if ok {
		detailsAny, found := guardData["details"]
		if found {
			details = detailsAny.(string)
		}
		specAny, found := guardData["specification"]
		if found {
			specification = specAny.(string)
		}
	}

	logic := model_logic.Logic{
		Key:           guardKey,
		Description:   details,
		Notation:      "tla_plus",
		Specification: specification,
	}

	guard, err = model_state.NewGuard(
		guardKey,
		name,
		logic)
	if err != nil {
		return model_state.Guard{}, err
	}

	return guard, nil
}

func actionFromYamlData(classKey identity.Key, name string, actionAny any) (action model_state.Action, err error) {

	// Construct the action key.
	actionKey, err := identity.NewActionKey(classKey, strings.ToLower(name))
	if err != nil {
		return model_state.Action{}, errors.WithStack(err)
	}

	details := ""

	var parameters []model_state.Parameter
	var requires []model_logic.Logic
	var guarantees []model_logic.Logic
	var safetyRules []model_logic.Logic

	actionData, ok := actionAny.(map[string]any)
	if ok {
		detailsAny, found := actionData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse parameters.
		parametersAny, found := actionData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Action{}, errors.Errorf("action '%s': parameters must be a sequence", name)
			}
			for _, paramAny := range paramsList {
				paramMap, ok := paramAny.(map[string]any)
				if !ok {
					return model_state.Action{}, errors.Errorf("action '%s': each parameter must be a mapping", name)
				}
				paramName, _ := paramMap["name"].(string)
				paramRules, _ := paramMap["rules"].(string)
				param, err := model_state.NewParameter(paramName, paramRules)
				if err != nil {
					return model_state.Action{}, errors.Wrapf(err, "action '%s' parameter '%s'", name, paramName)
				}
				parameters = append(parameters, param)
			}
		}

		// Parse requires.
		requires, err = logicListFromYamlData(actionData, "requires", actionKey, identity.NewActionRequireKey)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}

		// Parse guarantees.
		guarantees, err = logicListFromYamlData(actionData, "guarantees", actionKey, identity.NewActionGuaranteeKey)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}

		// Parse safety rules.
		safetyRules, err = logicListFromYamlData(actionData, "safety_rules", actionKey, identity.NewActionSafetyKey)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}
	}

	action, err = model_state.NewAction(
		actionKey,
		name,
		details,
		requires,
		guarantees,
		safetyRules,
		parameters)
	if err != nil {
		return model_state.Action{}, err
	}

	return action, nil
}

func queryFromYamlData(classKey identity.Key, name string, queryAny any) (query model_state.Query, err error) {

	// Construct the query key.
	queryKey, err := identity.NewQueryKey(classKey, strings.ToLower(name))
	if err != nil {
		return model_state.Query{}, errors.WithStack(err)
	}

	details := ""
	var parameters []model_state.Parameter
	var requires []model_logic.Logic
	var guarantees []model_logic.Logic

	queryData, ok := queryAny.(map[string]any)
	if ok {
		detailsAny, found := queryData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse parameters.
		parametersAny, found := queryData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Query{}, errors.Errorf("query '%s': parameters must be a sequence", name)
			}
			for _, paramAny := range paramsList {
				paramMap, ok := paramAny.(map[string]any)
				if !ok {
					return model_state.Query{}, errors.Errorf("query '%s': each parameter must be a mapping", name)
				}
				paramName, _ := paramMap["name"].(string)
				paramRules, _ := paramMap["rules"].(string)
				param, err := model_state.NewParameter(paramName, paramRules)
				if err != nil {
					return model_state.Query{}, errors.Wrapf(err, "query '%s' parameter '%s'", name, paramName)
				}
				parameters = append(parameters, param)
			}
		}

		// Parse requires.
		requires, err = logicListFromYamlData(queryData, "requires", queryKey, identity.NewQueryRequireKey)
		if err != nil {
			return model_state.Query{}, errors.Wrapf(err, "query '%s'", name)
		}

		// Parse guarantees.
		guarantees, err = logicListFromYamlData(queryData, "guarantees", queryKey, identity.NewQueryGuaranteeKey)
		if err != nil {
			return model_state.Query{}, errors.Wrapf(err, "query '%s'", name)
		}
	}

	query, err = model_state.NewQuery(
		queryKey,
		name,
		details,
		requires,
		guarantees,
		parameters)
	if err != nil {
		return model_state.Query{}, err
	}

	return query, nil
}

// logicListFromYamlData parses a YAML sequence of logic mappings (details + optional specification).
func logicListFromYamlData(data map[string]any, field string, parentKey identity.Key, newKey func(identity.Key, string) (identity.Key, error)) ([]model_logic.Logic, error) {
	listAny, found := data[field]
	if !found {
		return nil, nil
	}
	list, ok := listAny.([]any)
	if !ok {
		return nil, errors.Errorf("%s must be a sequence", field)
	}
	var logics []model_logic.Logic
	for i, itemAny := range list {
		itemMap, ok := itemAny.(map[string]any)
		if !ok {
			return nil, errors.Errorf("%s[%d] must be a mapping", field, i)
		}
		details, _ := itemMap["details"].(string)
		specification, _ := itemMap["specification"].(string)

		key, err := newKey(parentKey, strconv.Itoa(i))
		if err != nil {
			return nil, errors.Wrapf(err, "%s[%d]", field, i)
		}

		logics = append(logics, model_logic.Logic{
			Key:           key,
			Description:   details,
			Notation:      "tla_plus",
			Specification: specification,
		})
	}
	return logics, nil
}

func transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup map[string]identity.Key, classKey identity.Key, index int, transitionAny any) (transition model_state.Transition, err error) {

	transitionData, ok := transitionAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		// Parse values needed for the transition key.
		fromStateName := ""
		fromStateNameAny, found := transitionData["from"]
		if found {
			fromStateName = fromStateNameAny.(string)
		}

		eventName := ""
		eventNameAny, found := transitionData["event"]
		if found {
			eventName = eventNameAny.(string)
		}

		guardName := ""
		guardNameAny, found := transitionData["guard"]
		if found {
			guardName = guardNameAny.(string)
		}

		actionName := ""
		actionNameAny, found := transitionData["action"]
		if found {
			actionName = actionNameAny.(string)
		}

		toStateName := ""
		toStateNameAny, found := transitionData["to"]
		if found {
			toStateName = toStateNameAny.(string)
		}

		// Construct the transition key using the component names.
		transitionKey, err := identity.NewTransitionKey(classKey, fromStateName, eventName, guardName, actionName, toStateName)
		if err != nil {
			return model_state.Transition{}, errors.WithStack(err)
		}

		// Look up the state keys.
		var fromStateKey *identity.Key
		if fromStateName != "" {
			key, found := stateKeyLookup[fromStateName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, fromStateName))
			}
			fromStateKey = &key
		}

		var eventKey identity.Key
		if eventName != "" {
			eventKey, found = eventKeyLookup[eventName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown event: '%s'`, eventName))
			}
		}

		var guardKey *identity.Key
		if guardName != "" {
			key, found := guardKeyLookup[guardName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown guard: '%s'`, guardName))
			}
			guardKey = &key
		}

		var actionKey *identity.Key
		if actionName != "" {
			key, found := actionKeyLookup[actionName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
			}
			actionKey = &key
		}

		var toStateKey *identity.Key
		if toStateName != "" {
			key, found := stateKeyLookup[toStateName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, toStateName))
			}
			toStateKey = &key
		}

		umlComment := ""
		umlCommentAny, found := transitionData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		transition, err = model_state.NewTransition(
			transitionKey,
			fromStateKey,
			eventKey,
			guardKey,
			actionKey,
			toStateKey,
			umlComment)
		if err != nil {
			return model_state.Transition{}, err
		}
	}

	return transition, nil
}

func generateClassContent(class model_class.Class, associations []model_class.Association) string {
	builder := NewYamlBuilder()

	// Add top-level fields.
	if class.ActorKey != nil {
		builder.AddField("actor_key", class.ActorKey.SubKey)
	}
	if class.SuperclassOfKey != nil {
		builder.AddField("superclass_of_key", class.SuperclassOfKey.SubKey)
	}
	if class.SubclassOfKey != nil {
		builder.AddField("subclass_of_key", class.SubclassOfKey.SubKey)
	}

	// Add invariants section.
	generateLogicSequence(builder, "invariants", class.Invariants)

	// Add attributes section.
	if len(class.Attributes) > 0 {
		attrsBuilder := NewYamlBuilder()
		sortedAttrs := view_helper.GetAttributesSorted(class.Attributes)
		for _, attr := range sortedAttrs {
			attrBuilder := NewYamlBuilder()
			attrBuilder.AddField("name", attr.Name)
			attrBuilder.AddField("details", attr.Details)
			attrBuilder.AddField("rules", attr.DataTypeRules)
			attrBuilder.AddBoolField("nullable", attr.Nullable)
			if attr.DerivationPolicy != nil {
				derivBuilder := NewYamlBuilder()
				derivBuilder.AddField("description", attr.DerivationPolicy.Description)
				derivBuilder.AddQuotedField("specification", attr.DerivationPolicy.Specification)
				attrBuilder.AddMappingField("derivation", derivBuilder)
			}
			attrBuilder.AddField("uml_comment", attr.UmlComment)
			// Convert []uint to []int for index_nums.
			if len(attr.IndexNums) > 0 {
				intNums := make([]int, len(attr.IndexNums))
				for i, n := range attr.IndexNums {
					intNums[i] = int(n)
				}
				attrBuilder.AddIntSliceField("index_nums", intNums)
			}
			attrsBuilder.AddMappingField(attr.Key.SubKey, attrBuilder)
		}
		builder.AddMappingField("attributes", attrsBuilder)
	}

	// Add associations section.
	if len(associations) > 0 {
		var assocBuilders []*YamlBuilder
		for _, assoc := range associations {
			assocBuilder := NewYamlBuilder()
			assocBuilder.AddField("name", assoc.Name)
			assocBuilder.AddField("details", assoc.Details)
			addMultiplicityField(assocBuilder, "from_multiplicity", assoc.FromMultiplicity)
			assocBuilder.AddField("to_class_key", classAssociationRelativeKey(class, assoc.ToClassKey))
			addMultiplicityField(assocBuilder, "to_multiplicity", assoc.ToMultiplicity)
			if assoc.AssociationClassKey != nil {
				assocBuilder.AddField("association_class_key", classAssociationRelativeKey(class, *assoc.AssociationClassKey))
			}
			assocBuilder.AddField("uml_comment", assoc.UmlComment)
			assocBuilders = append(assocBuilders, assocBuilder)
		}
		builder.AddSequenceOfMappings("associations", assocBuilders)
	}

	// Create lookups for names.
	stateKeyLookups := map[string]model_state.State{}
	for _, state := range class.States {
		stateKeyLookups[state.Key.String()] = state
	}
	actionKeyLookup := map[string]model_state.Action{}
	for _, action := range class.Actions {
		actionKeyLookup[action.Key.String()] = action
	}
	eventKeyLookup := map[string]model_state.Event{}
	for _, event := range class.Events {
		eventKeyLookup[event.Key.String()] = event
	}
	guardKeyLookup := map[string]model_state.Guard{}
	for _, guard := range class.Guards {
		guardKeyLookup[guard.Key.String()] = guard
	}

	// Add states section.
	if len(class.States) > 0 {
		statesBuilder := NewYamlBuilder()
		stateKeys := make([]string, 0, len(class.States))
		for k := range class.States {
			stateKeys = append(stateKeys, k.String())
		}
		sort.Strings(stateKeys)
		for _, keyStr := range stateKeys {
			key, _ := identity.ParseKey(keyStr)
			state := class.States[key]
			stateBuilder := NewYamlBuilder()
			stateBuilder.AddField("details", state.Details)
			stateBuilder.AddField("uml_comment", state.UmlComment)
			if len(state.Actions) > 0 {
				var stateActionBuilders []*YamlBuilder
				for _, sa := range state.Actions {
					saBuilder := NewYamlBuilder()
					saBuilder.AddField("action", actionKeyLookup[sa.ActionKey.String()].Name)
					saBuilder.AddField("when", sa.When)
					stateActionBuilders = append(stateActionBuilders, saBuilder)
				}
				stateBuilder.AddSequenceOfMappings("actions", stateActionBuilders)
			}
			statesBuilder.AddMappingFieldAlways(state.Name, stateBuilder)
		}
		builder.AddMappingField("states", statesBuilder)
	}

	// Add events section.
	if len(class.Events) > 0 {
		eventsBuilder := NewYamlBuilder()
		eventKeys := make([]string, 0, len(class.Events))
		for k := range class.Events {
			eventKeys = append(eventKeys, k.String())
		}
		sort.Strings(eventKeys)
		for _, keyStr := range eventKeys {
			key, _ := identity.ParseKey(keyStr)
			event := class.Events[key]
			eventBuilder := NewYamlBuilder()
			eventBuilder.AddField("details", event.Details)
			generateParameterSequence(eventBuilder, event.Parameters)
			eventsBuilder.AddMappingFieldAlways(event.Name, eventBuilder)
		}
		builder.AddMappingField("events", eventsBuilder)
	}

	// Add guards section.
	if len(class.Guards) > 0 {
		guardsBuilder := NewYamlBuilder()
		guardKeys := make([]string, 0, len(class.Guards))
		for k := range class.Guards {
			guardKeys = append(guardKeys, k.String())
		}
		sort.Strings(guardKeys)
		for _, keyStr := range guardKeys {
			key, _ := identity.ParseKey(keyStr)
			guard := class.Guards[key]
			guardBuilder := NewYamlBuilder()
			guardBuilder.AddField("details", guard.Logic.Description)
			guardBuilder.AddField("specification", guard.Logic.Specification)
			guardsBuilder.AddMappingField(guard.Name, guardBuilder)
		}
		builder.AddMappingField("guards", guardsBuilder)
	}

	// Add actions section.
	if len(class.Actions) > 0 {
		actionsBuilder := NewYamlBuilder()
		actionKeys := make([]string, 0, len(class.Actions))
		for k := range class.Actions {
			actionKeys = append(actionKeys, k.String())
		}
		sort.Strings(actionKeys)
		for _, keyStr := range actionKeys {
			key, _ := identity.ParseKey(keyStr)
			action := class.Actions[key]
			actionBuilder := NewYamlBuilder()
			actionBuilder.AddField("details", action.Details)
			generateParameterSequence(actionBuilder, action.Parameters)
			generateLogicSequence(actionBuilder, "requires", action.Requires)
			generateLogicSequence(actionBuilder, "guarantees", action.Guarantees)
			generateLogicSequence(actionBuilder, "safety_rules", action.SafetyRules)
			actionsBuilder.AddMappingField(action.Name, actionBuilder)
		}
		builder.AddMappingField("actions", actionsBuilder)
	}

	// Add queries section.
	if len(class.Queries) > 0 {
		queriesBuilder := NewYamlBuilder()
		queryKeys := make([]string, 0, len(class.Queries))
		for k := range class.Queries {
			queryKeys = append(queryKeys, k.String())
		}
		sort.Strings(queryKeys)
		for _, keyStr := range queryKeys {
			key, _ := identity.ParseKey(keyStr)
			query := class.Queries[key]
			queryBuilder := NewYamlBuilder()
			queryBuilder.AddField("details", query.Details)
			generateParameterSequence(queryBuilder, query.Parameters)
			generateLogicSequence(queryBuilder, "requires", query.Requires)
			generateLogicSequence(queryBuilder, "guarantees", query.Guarantees)
			queriesBuilder.AddMappingField(query.Name, queryBuilder)
		}
		builder.AddMappingField("queries", queriesBuilder)
	}

	// Add transitions section.
	if len(class.Transitions) > 0 {
		var transitionBuilders []*YamlBuilder
		transitionKeys := make([]string, 0, len(class.Transitions))
		for k := range class.Transitions {
			transitionKeys = append(transitionKeys, k.String())
		}
		sort.Strings(transitionKeys)
		for _, keyStr := range transitionKeys {
			key, _ := identity.ParseKey(keyStr)
			trans := class.Transitions[key]
			transBuilder := NewYamlBuilder()
			from := ""
			if trans.FromStateKey != nil {
				from = stateKeyLookups[trans.FromStateKey.String()].Name
			}
			transBuilder.AddQuotedField("from", from)
			transBuilder.AddQuotedField("event", eventKeyLookup[trans.EventKey.String()].Name)
			to := ""
			if trans.ToStateKey != nil {
				to = stateKeyLookups[trans.ToStateKey.String()].Name
			}
			transBuilder.AddQuotedField("to", to)
			if trans.GuardKey != nil {
				transBuilder.AddQuotedField("guard", guardKeyLookup[trans.GuardKey.String()].Name)
			}
			if trans.ActionKey != nil {
				transBuilder.AddQuotedField("action", actionKeyLookup[trans.ActionKey.String()].Name)
			}
			if trans.UmlComment != "" {
				transBuilder.AddQuotedField("uml_comment", trans.UmlComment)
			}
			transitionBuilders = append(transitionBuilders, transBuilder)
		}
		builder.AddFlowSequence("transitions", transitionBuilders)
	}

	yamlStr, _ := builder.Build()
	return generateFileContent(prependMarkdownTitle(class.Name, class.Details), class.UmlComment, yamlStr)
}

// generateParameterSequence adds a parameters sequence of mappings to the builder.
func generateParameterSequence(builder *YamlBuilder, params []model_state.Parameter) {
	if len(params) == 0 {
		return
	}
	items := make([]*YamlBuilder, 0, len(params))
	for _, param := range params {
		paramBuilder := NewYamlBuilder()
		paramBuilder.AddField("name", param.Name)
		paramBuilder.AddField("rules", param.DataTypeRules)
		items = append(items, paramBuilder)
	}
	builder.AddSequenceOfMappings("parameters", items)
}

// generateLogicSequence adds a logic sequence of mappings to the builder.
func generateLogicSequence(builder *YamlBuilder, field string, logics []model_logic.Logic) {
	if len(logics) == 0 {
		return
	}
	items := make([]*YamlBuilder, 0, len(logics))
	for _, logic := range logics {
		logicBuilder := NewYamlBuilder()
		logicBuilder.AddField("details", logic.Description)
		logicBuilder.AddField("specification", logic.Specification)
		items = append(items, logicBuilder)
	}
	builder.AddSequenceOfMappings(field, items)
}

// classAssociationRelativeKey returns the shortest relative key string for a target class key
// relative to the from-class. If both classes share the same subdomain, returns just the SubKey.
// If they share the same domain, returns the path relative to the domain (subdomain/X/class/Y).
// Otherwise returns the full key string (domain/X/subdomain/Y/class/Z).
func classAssociationRelativeKey(fromClass model_class.Class, targetClassKey identity.Key) string {
	fromSubdomain := fromClass.Key.ParentKey
	targetSubdomain := targetClassKey.ParentKey

	// Same subdomain: just the class subkey.
	if fromSubdomain == targetSubdomain {
		return targetClassKey.SubKey
	}

	// Check if same domain by parsing subdomain parent keys.
	fromSubdomainParsed, err1 := identity.ParseKey(fromSubdomain)
	targetSubdomainParsed, err2 := identity.ParseKey(targetSubdomain)
	if err1 == nil && err2 == nil && fromSubdomainParsed.ParentKey == targetSubdomainParsed.ParentKey {
		// Same domain, different subdomain: path relative to domain.
		// Strip the domain prefix to get "subdomain/X/class/Y".
		domainPrefix := fromSubdomainParsed.ParentKey + "/"
		return strings.TrimPrefix(targetClassKey.String(), domainPrefix)
	}

	// Different domains: full key string.
	return targetClassKey.String()
}

// addMultiplicityField adds a multiplicity field to the builder.
// Numeric multiplicities are quoted, "any" is not quoted.
func addMultiplicityField(builder *YamlBuilder, key string, m model_class.Multiplicity) {
	s := m.ParsedString()
	// Convert UML "N..*" format to parseable "N..many" format.
	if strings.HasSuffix(s, "..*") {
		s = strings.TrimSuffix(s, "..*") + "..many"
	}
	if s == "any" {
		builder.AddField(key, s)
	} else {
		builder.AddQuotedField(key, s)
	}
}
