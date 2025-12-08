package parser

import (
	"github.com/glemzurg/futz/apps/req/requirements"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseClass(classKey, filename, contents string) (class requirements.Class, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.Class{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return requirements.Class{}, errors.WithStack(err)
	}

	actorKey := ""
	actorAny, found := yamlData["actor_key"]
	if found {
		actorKey = actorAny.(string)
	}

	superclassOfKey := ""
	superclassOfAny, found := yamlData["superclass_of_key"]
	if found {
		superclassOfKey = superclassOfAny.(string)
	}

	subclassOfKey := ""
	subclassOfAny, found := yamlData["subclass_of_key"]
	if found {
		subclassOfKey = subclassOfAny.(string)
	}

	class, err = requirements.NewClass(classKey, parsedFile.Title, parsedFile.Markdown, actorKey, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
	if err != nil {
		return requirements.Class{}, err
	}

	// Add any attributes we found.
	var attributesData map[string]any
	attributesAny, found := yamlData["attributes"]
	if found {
		attributesData = attributesAny.(map[string]any)
	}

	var attributes []requirements.Attribute
	for key, attributeAny := range attributesData {

		// Join the class to the key so it is unique in the model.
		key = classKey + "/" + key

		attribute, err := attributeFromYamlData(key, attributeAny)
		if err != nil {
			return requirements.Class{}, err
		}
		attributes = append(attributes, attribute)
	}
	class.SetAttributes(attributes)

	// Add any associations we found.
	var associationsData []any
	associationsAny, found := yamlData["associations"]
	if found {
		associationsData = associationsAny.([]any)
	}

	var associations []requirements.Association
	for i, associationAny := range associationsData {
		association, err := associationFromYamlData(class.Key, i, associationAny)
		if err != nil {
			return requirements.Class{}, err
		}
		associations = append(associations, association)
	}
	class.Associations = associations

	// Add any actions we found.
	var actionsData map[string]any
	actionsAny, found := yamlData["actions"]
	if found {
		actionsData = actionsAny.(map[string]any)
	}

	var actions []requirements.Action
	actionKeyLookup := map[string]string{}
	for name, actionAny := range actionsData {
		action, err := actionFromYamlData(class.Key, name, actionAny)
		if err != nil {
			return requirements.Class{}, err
		}
		actionKeyLookup[action.Name] = action.Key
		actions = append(actions, action)
	}
	class.SetActions(actions)

	// Add any states we found.
	var statesData map[string]any
	statesAny, found := yamlData["states"]
	if found {
		statesData = statesAny.(map[string]any)
	}

	var states []requirements.State
	stateKeyLookup := map[string]string{}
	for name, stateAny := range statesData {
		state, err := stateFromYamlData(actionKeyLookup, class.Key, name, stateAny)
		if err != nil {
			return requirements.Class{}, err
		}
		stateKeyLookup[state.Name] = state.Key
		states = append(states, state)
	}
	class.SetStates(states)

	// Add any events we found.
	var eventsData map[string]any
	eventsAny, found := yamlData["events"]
	if found {
		eventsData = eventsAny.(map[string]any)
	}

	var events []requirements.Event
	eventKeyLookup := map[string]string{}
	for name, eventAny := range eventsData {
		event, err := eventFromYamlData(class.Key, name, eventAny)
		if err != nil {
			return requirements.Class{}, err
		}
		eventKeyLookup[event.Name] = event.Key
		events = append(events, event)
	}
	class.SetEvents(events)

	// Add any guards we found.
	var guardsData map[string]any
	guardsAny, found := yamlData["guards"]
	if found {
		guardsData = guardsAny.(map[string]any)
	}

	var guards []requirements.Guard
	guardKeyLookup := map[string]string{}
	for name, guardAny := range guardsData {
		guard, err := guardFromYamlData(class.Key, name, guardAny)
		if err != nil {
			return requirements.Class{}, err
		}
		guardKeyLookup[guard.Name] = guard.Key
		guards = append(guards, guard)
	}
	class.SetGuards(guards)

	// Add any transitions we found.
	var transitionsData []any
	transitionsAny, found := yamlData["transitions"]
	if found {
		transitionsData = transitionsAny.([]any)
	}

	var transitions []requirements.Transition
	for i, transitionAny := range transitionsData {
		transition, err := transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup, class.Key, i, transitionAny)
		if err != nil {
			return requirements.Class{}, err
		}
		transitions = append(transitions, transition)
	}
	class.Transitions = transitions

	return class, nil
}

func attributeFromYamlData(key string, attributeAny any) (attribute requirements.Attribute, err error) {

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

		derivationPolicy := ""
		derivationPolicyAny, found := attributeData["derivation"]
		if found {
			derivationPolicy = derivationPolicyAny.(string)
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

		attribute, err = requirements.NewAttribute(
			key,
			name,
			normalizeWhitespace(details),
			dataTypeRules,
			normalizeWhitespace(derivationPolicy),
			nullable,
			umlComment,
			indexNums)
		if err != nil {
			return requirements.Attribute{}, err
		}
	}

	return attribute, nil
}

func associationFromYamlData(fromClassKey string, index int, associationAny any) (association requirements.Association, err error) {

	associationData, ok := associationAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		key := fromClassKey + "/association/" + strconv.Itoa(index+1) // Don't start at zero.

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
		fromMultiplicity, err := requirements.NewMultiplicity(fromMultiplicityValue)
		if err != nil {
			return requirements.Association{}, err
		}

		toClassKey := ""
		toClassKeyAny, found := associationData["to_class_key"]
		if found {
			toClassKey = toClassKeyAny.(string)
		}

		toMultiplicityValue := ""
		toMultiplicityAny, found := associationData["to_multiplicity"]
		if found {
			toMultiplicityValue = toMultiplicityAny.(string)
		}
		toMultiplicity, err := requirements.NewMultiplicity(toMultiplicityValue)
		if err != nil {
			return requirements.Association{}, err
		}

		associationClassKey := ""
		associationClassKeyAny, found := associationData["association_class_key"]
		if found {
			associationClassKey = associationClassKeyAny.(string)
		}

		umlComment := ""
		umlCommentAny, found := associationData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		association, err = requirements.NewAssociation(
			key,
			name,
			details,
			fromClassKey,
			fromMultiplicity,
			toClassKey,
			toMultiplicity,
			associationClassKey,
			umlComment)
		if err != nil {
			return requirements.Association{}, err
		}
	}

	return association, nil
}

func stateFromYamlData(actionKeyLookup map[string]string, classKey, name string, stateAny any) (state requirements.State, err error) {

	key := classKey + "/state/" + strings.ToLower(name)

	details := ""
	umlComment := ""
	var actions []requirements.StateAction

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

				actionKey := ""
				actionName := ""
				actionNameAny, found := actionData["action"]
				if found {
					actionName = actionNameAny.(string)
					actionKey, found = actionKeyLookup[actionName]
					if !found {
						return requirements.State{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
					}
				}

				when := ""
				whenAny, found := actionData["when"]
				if found {
					when = whenAny.(string)
				}

				stateActionKey := key + "/action/" + strings.ToLower(when) + "/" + strings.ToLower(actionName)

				action, err := requirements.NewStateAction(
					stateActionKey,
					actionKey,
					when,
				)
				if err != nil {
					return requirements.State{}, err
				}

				actions = append(actions, action)

			}
		}
	}

	state, err = requirements.NewState(
		key,
		name,
		details,
		umlComment)
	if err != nil {
		return requirements.State{}, err
	}

	// Attach the actions.
	state.SetActions(actions)

	return state, nil
}

func eventFromYamlData(classKey, name string, eventAny any) (event requirements.Event, err error) {

	key := classKey + "/event/" + strings.ToLower(name)

	var params []requirements.EventParameter
	details := ""

	eventData, ok := eventAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := eventData["details"]
		if found {
			details = detailsAny.(string)
		}

		paramsAny, found := eventData["parameters"]
		if found {
			paramsAny := paramsAny.([]any)
			for _, paramAny := range paramsAny {
				paramData, ok := paramAny.(map[string]any)
				if ok {

					name := ""
					nameAny, found := paramData["name"]
					if found {
						name = nameAny.(string)
					}

					source := ""
					sourceAny, found := paramData["source"]
					if found {
						source = sourceAny.(string)
					}

					param, err := requirements.NewEventParameter(name, source)
					if err != nil {
						return requirements.Event{}, err
					}

					params = append(params, param)
				}
			}
		}
	}

	event, err = requirements.NewEvent(
		key,
		name,
		details,
		params)
	if err != nil {
		return requirements.Event{}, err
	}

	return event, nil
}

func guardFromYamlData(classKey, name string, guardAny any) (guard requirements.Guard, err error) {

	key := classKey + "/guard/" + strings.ToLower(name)

	details := ""

	guardData, ok := guardAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := guardData["details"]
		if found {
			details = detailsAny.(string)
		}
	}

	guard, err = requirements.NewGuard(
		key,
		name,
		details)
	if err != nil {
		return requirements.Guard{}, err
	}

	return guard, nil
}

func actionFromYamlData(classKey, name string, actionAny any) (action requirements.Action, err error) {

	key := classKey + "/action/" + strings.ToLower(name)

	details := ""
	var requires []string
	var guarantees []string

	actionData, ok := actionAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := actionData["details"]
		if found {
			details = detailsAny.(string)
		}

		requiresAny, found := actionData["requires"]
		if found {
			requiresAny := requiresAny.([]any)
			for _, requireAny := range requiresAny {
				requires = append(requires, requireAny.(string))
			}
		}

		guaranteesAny, found := actionData["guarantees"]
		if found {
			guaranteesAny := guaranteesAny.([]any)
			for _, guaranteeAny := range guaranteesAny {
				guarantees = append(guarantees, guaranteeAny.(string))
			}
		}
	}

	action, err = requirements.NewAction(
		key,
		name,
		details,
		requires,
		guarantees)
	if err != nil {
		return requirements.Action{}, err
	}

	return action, nil
}

func transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup map[string]string, fromClassKey string, index int, transitionAny any) (transition requirements.Transition, err error) {

	transitionData, ok := transitionAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		key := fromClassKey + "/transition/" + strconv.Itoa(index+1) // Don't start at zero.

		fromStateKey := ""
		fromStateNameAny, found := transitionData["from"]
		if found {
			stateName := fromStateNameAny.(string)
			if stateName != "" {
				fromStateKey, found = stateKeyLookup[stateName]
				if !found {
					return requirements.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, stateName))
				}
			}
		}

		eventKey := ""
		eventNameAny, found := transitionData["event"]
		if found {
			eventName := eventNameAny.(string)
			eventKey, found = eventKeyLookup[eventName]
			if !found {
				return requirements.Transition{}, errors.WithStack(errors.Errorf(`unknown event: '%s'`, eventName))
			}
		}

		guardKey := ""
		guardNameAny, found := transitionData["guard"]
		if found {
			guardName := guardNameAny.(string)
			guardKey, found = guardKeyLookup[guardName]
			if !found {
				return requirements.Transition{}, errors.WithStack(errors.Errorf(`unknown guard: '%s'`, guardName))
			}
		}

		actionKey := ""
		actionNameAny, found := transitionData["action"]
		if found {
			actionName := actionNameAny.(string)
			actionKey, found = actionKeyLookup[actionName]
			if !found {
				return requirements.Transition{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
			}
		}

		toStateKey := ""
		toStateNameAny, found := transitionData["to"]
		if found {
			stateName := toStateNameAny.(string)
			if stateName != "" {
				toStateKey = stateKeyLookup[stateName]
				if !found {
					return requirements.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, stateName))
				}
			}
		}

		umlComment := ""
		umlCommentAny, found := transitionData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		transition, err = requirements.NewTransition(
			key,
			fromStateKey,
			eventKey,
			guardKey,
			actionKey,
			toStateKey,
			umlComment)
		if err != nil {
			return requirements.Transition{}, err
		}
	}

	return transition, nil
}
