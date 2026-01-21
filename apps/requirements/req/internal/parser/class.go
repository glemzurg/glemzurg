package parser

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

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

	class, err = model_class.NewClass(classKey, parsedFile.Title, parsedFile.Markdown, actorKey, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
	if err != nil {
		return model_class.Class{}, nil, err
	}

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

		// Construct the identity key for this attribute.
		attrKey, err := identity.NewAttributeKey(classKey, attrSubKey)
		if err != nil {
			return model_class.Attribute{}, errors.WithStack(err)
		}

		attribute, err = model_class.NewAttribute(
			attrKey,
			name,
			normalizeWhitespace(details),
			dataTypeRules,
			normalizeWhitespace(derivationPolicy),
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

		// Parse association class key if present.
		var associationClassKey *identity.Key
		associationClassKeyAny, found := associationData["association_class_key"]
		if found {
			associationClassKeyStr := associationClassKeyAny.(string)
			if associationClassKeyStr != "" {
				// Parse the key - it should be a class key relative to subdomain.
				key, err := identity.NewClassKey(subdomainKey, associationClassKeyStr)
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

		// Construct the to class key - assuming it's in the same subdomain.
		toClassKey, err := identity.NewClassKey(subdomainKey, toClassKeyStr)
		if err != nil {
			return model_class.Association{}, errors.WithStack(err)
		}

		// Construct the class association key using the subdomain as parent
		// since both classes are in the same subdomain.
		assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey)
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

	var params []model_state.EventParameter
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

					param, err := model_state.NewEventParameter(name, source)
					if err != nil {
						return model_state.Event{}, err
					}

					params = append(params, param)
				}
			}
		}
	}

	event, err = model_state.NewEvent(
		eventKey,
		name,
		details,
		params)
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

	guardData, ok := guardAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := guardData["details"]
		if found {
			details = detailsAny.(string)
		}
	}

	guard, err = model_state.NewGuard(
		guardKey,
		name,
		details)
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

	action, err = model_state.NewAction(
		actionKey,
		name,
		details,
		requires,
		guarantees)
	if err != nil {
		return model_state.Action{}, err
	}

	return action, nil
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
	yaml := ""
	if class.ActorKey != nil {
		yaml += "actor_key: " + class.ActorKey.SubKey() + "\n"
	}
	if class.SuperclassOfKey != nil {
		yaml += "superclass_of_key: " + class.SuperclassOfKey.SubKey() + "\n"
	}
	if class.SubclassOfKey != nil {
		yaml += "subclass_of_key: " + class.SubclassOfKey.SubKey() + "\n"
	}
	if len(class.Attributes) > 0 {
		yaml += "\n"
		yaml += "attributes:\n"
		// Sort attributes for deterministic output.
		sortedAttrs := req_flat.GetAttributesSorted(class.Attributes)
		for _, attr := range sortedAttrs {
			yaml += "\n"
			name := attr.Key.SubKey()
			yaml += "    " + name + ":\n"
			yaml += "        name: " + attr.Name + "\n"
			if attr.Details != "" {
				yaml += "        details: " + attr.Details + "\n"
			}
			if attr.DataTypeRules != "" {
				yaml += "        rules: " + attr.DataTypeRules + "\n"
			}
			if attr.DerivationPolicy != "" {
				yaml += "        derivation: " + attr.DerivationPolicy + "\n"
			}
			if attr.Nullable {
				yaml += "        nullable: true\n"
			}
			if attr.UmlComment != "" {
				yaml += "        uml_comment: " + attr.UmlComment + "\n"
			}
			if len(attr.IndexNums) > 0 {
				yaml += "        index_nums: ["
				for i, num := range attr.IndexNums {
					if i > 0 {
						yaml += ", "
					}
					yaml += strconv.Itoa(int(num))
				}
				yaml += "]\n"
			}
		}
	}
	// Generate associations if present.
	if len(associations) > 0 {
		yaml += "\nassociations:\n"
		for _, assoc := range associations {
			yaml += "\n    - name: " + assoc.Name + "\n"
			if assoc.Details != "" {
				yaml += "      details: " + assoc.Details + "\n"
			}
			yaml += "      from_multiplicity: " + formatMultiplicity(assoc.FromMultiplicity) + "\n"
			yaml += "      to_class_key: " + assoc.ToClassKey.SubKey() + "\n"
			yaml += "      to_multiplicity: " + formatMultiplicity(assoc.ToMultiplicity) + "\n"
			if assoc.AssociationClassKey != nil {
				yaml += "      association_class_key: " + assoc.AssociationClassKey.SubKey() + "\n"
			}
			if assoc.UmlComment != "" {
				yaml += "      uml_comment: " + assoc.UmlComment + "\n"
			}
		}
	}

	// We need a lookup of actions to display names where they need to be.
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

	if len(class.States) > 0 {
		yaml += "\n"
		yaml += "states:\n"
		// Sort state keys for deterministic output.
		stateKeys := make([]string, 0, len(class.States))
		for k := range class.States {
			stateKeys = append(stateKeys, k.String())
		}
		sort.Strings(stateKeys)
		for _, keyStr := range stateKeys {
			key, _ := identity.ParseKey(keyStr)
			state := class.States[key]
			yaml += "\n"
			yaml += "  " + state.Name + ":\n"
			if state.Details != "" {
				yaml += "    details: " + state.Details + "\n"
			}
			if state.UmlComment != "" {
				yaml += "    uml_comment: " + state.UmlComment + "\n"
			}
			if len(state.Actions) > 0 {
				yaml += "    actions:\n"
				for _, sa := range state.Actions {
					yaml += "        - action: " + actionKeyLookup[sa.ActionKey.String()].Name + "\n"
					yaml += "          when: " + sa.When + "\n"
				}
			}
		}
	}
	if len(class.Events) > 0 {
		yaml += "\n"
		yaml += "events:\n"
		// Sort event keys for deterministic output.
		eventKeys := make([]string, 0, len(class.Events))
		for k := range class.Events {
			eventKeys = append(eventKeys, k.String())
		}
		sort.Strings(eventKeys)
		for _, keyStr := range eventKeys {
			key, _ := identity.ParseKey(keyStr)
			event := class.Events[key]
			yaml += "\n"
			yaml += "    " + event.Name + ":\n"
			if event.Details != "" {
				yaml += "        details: " + event.Details + "\n"
			}
			if len(event.Parameters) > 0 {
				yaml += "        parameters:\n"
				for _, param := range event.Parameters {
					yaml += "            - name: " + param.Name + "\n"
					if param.Source != "" {
						yaml += "              source: " + param.Source + "\n"
					}
				}
			}
		}
	}
	if len(class.Guards) > 0 {
		yaml += "\n"
		yaml += "guards:\n"
		// Sort guard keys for deterministic output.
		guardKeys := make([]string, 0, len(class.Guards))
		for k := range class.Guards {
			guardKeys = append(guardKeys, k.String())
		}
		sort.Strings(guardKeys)
		for _, keyStr := range guardKeys {
			key, _ := identity.ParseKey(keyStr)
			guard := class.Guards[key]
			yaml += "\n"
			yaml += "    " + guard.Name + ":\n"
			if guard.Details != "" {
				yaml += "        details: " + guard.Details + "\n"
			}
		}
	}
	if len(class.Actions) > 0 {
		yaml += "\n"
		yaml += "actions:\n"
		// Sort action keys for deterministic output.
		actionKeys := make([]string, 0, len(class.Actions))
		for k := range class.Actions {
			actionKeys = append(actionKeys, k.String())
		}
		sort.Strings(actionKeys)
		for _, keyStr := range actionKeys {
			key, _ := identity.ParseKey(keyStr)
			action := class.Actions[key]
			yaml += "\n"
			yaml += "    " + action.Name + ":\n"
			if action.Details != "" {
				yaml += "        details: " + action.Details + "\n"
			}
			if len(action.Requires) > 0 {
				yaml += "        requires:\n"
				for _, req := range action.Requires {
					yaml += "            - " + req + "\n"
				}
			}
			if len(action.Guarantees) > 0 {
				yaml += "        guarantees:\n"
				for _, gua := range action.Guarantees {
					yaml += "            - " + gua + "\n"
				}
			}
		}
	}
	if len(class.Transitions) > 0 {
		yaml += "\n"
		yaml += "transitions:\n"
		yaml += "\n"
		// Sort transition keys for deterministic output.
		transitionKeys := make([]string, 0, len(class.Transitions))
		for k := range class.Transitions {
			transitionKeys = append(transitionKeys, k.String())
		}
		sort.Strings(transitionKeys)
		for _, keyStr := range transitionKeys {
			key, _ := identity.ParseKey(keyStr)
			trans := class.Transitions[key]
			from := ""
			if trans.FromStateKey != nil {
				from = stateKeyLookups[trans.FromStateKey.String()].Name
			}
			event := ""
			event = eventKeyLookup[trans.EventKey.String()].Name
			to := ""
			if trans.ToStateKey != nil {
				to = stateKeyLookups[trans.ToStateKey.String()].Name
			}
			guard := ""
			if trans.GuardKey != nil {
				guard = guardKeyLookup[trans.GuardKey.String()].Name
			}
			action := ""
			if trans.ActionKey != nil {
				action = actionKeyLookup[trans.ActionKey.String()].Name
			}
			yaml += "    - {from: \"" + from + "\", event: \"" + event + "\", to: \"" + to + "\""
			if guard != "" {
				yaml += ", guard: \"" + guard + "\""
			}
			if action != "" {
				yaml += ", action: \"" + action + "\""
			}
			if trans.UmlComment != "" {
				yaml += ", uml_comment: \"" + trans.UmlComment + "\""
			}
			yaml += "}\n"
		}
	}
	yamlStr := strings.TrimSpace(yaml)
	return generateFileContent(class.Details, class.UmlComment, yamlStr)
}

// formatMultiplicity formats a multiplicity for YAML output.
// Numeric multiplicities are quoted, "any" is not quoted.
func formatMultiplicity(m model_class.Multiplicity) string {
	s := m.ParsedString()
	if s == "any" {
		return s
	}
	return "\"" + s + "\""
}
