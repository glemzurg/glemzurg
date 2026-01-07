package parser

import (
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseClass(classKey, filename, contents string) (class model_class.Class, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_class.Class{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_class.Class{}, errors.WithStack(err)
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

	class, err = model_class.NewClass(classKey, parsedFile.Title, parsedFile.Markdown, actorKey, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
	if err != nil {
		return model_class.Class{}, err
	}

	// Add any attributes we found.
	var attributesData map[string]any
	attributesAny, found := yamlData["attributes"]
	if found {
		attributesData = attributesAny.(map[string]any)
	}

	var attributes []model_class.Attribute
	for key, attributeAny := range attributesData {

		// Join the class to the key so it is unique in the model.
		key = classKey + "/" + key

		attribute, err := attributeFromYamlData(key, attributeAny)
		if err != nil {
			return model_class.Class{}, err
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

	var associations []model_class.Association
	for i, associationAny := range associationsData {
		association, err := associationFromYamlData(class.Key, i, associationAny)
		if err != nil {
			return model_class.Class{}, err
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

	var actions []model_state.Action
	actionKeyLookup := map[string]string{}
	for name, actionAny := range actionsData {
		action, err := actionFromYamlData(class.Key, name, actionAny)
		if err != nil {
			return model_class.Class{}, err
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

	var states []model_state.State
	stateKeyLookup := map[string]string{}
	for name, stateAny := range statesData {
		state, err := stateFromYamlData(actionKeyLookup, class.Key, name, stateAny)
		if err != nil {
			return model_class.Class{}, err
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

	var events []model_state.Event
	eventKeyLookup := map[string]string{}
	for name, eventAny := range eventsData {
		event, err := eventFromYamlData(class.Key, name, eventAny)
		if err != nil {
			return model_class.Class{}, err
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

	var guards []model_state.Guard
	guardKeyLookup := map[string]string{}
	for name, guardAny := range guardsData {
		guard, err := guardFromYamlData(class.Key, name, guardAny)
		if err != nil {
			return model_class.Class{}, err
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

	var transitions []model_state.Transition
	for i, transitionAny := range transitionsData {
		transition, err := transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup, class.Key, i, transitionAny)
		if err != nil {
			return model_class.Class{}, err
		}
		transitions = append(transitions, transition)
	}
	class.Transitions = transitions

	return class, nil
}

func attributeFromYamlData(key string, attributeAny any) (attribute model_class.Attribute, err error) {

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

		attribute, err = model_class.NewAttribute(
			key,
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

func associationFromYamlData(fromClassKey string, index int, associationAny any) (association model_class.Association, err error) {

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
		fromMultiplicity, err := model_class.NewMultiplicity(fromMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
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
		toMultiplicity, err := model_class.NewMultiplicity(toMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
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

		association, err = model_class.NewAssociation(
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
			return model_class.Association{}, err
		}
	}

	return association, nil
}

func stateFromYamlData(actionKeyLookup map[string]string, classKey, name string, stateAny any) (state model_state.State, err error) {

	key := classKey + "/state/" + strings.ToLower(name)

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

				actionKey := ""
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

				stateActionKey := key + "/action/" + strings.ToLower(when) + "/" + strings.ToLower(actionName)

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
		key,
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

func eventFromYamlData(classKey, name string, eventAny any) (event model_state.Event, err error) {

	key := classKey + "/event/" + strings.ToLower(name)

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
		key,
		name,
		details,
		params)
	if err != nil {
		return model_state.Event{}, err
	}

	return event, nil
}

func guardFromYamlData(classKey, name string, guardAny any) (guard model_state.Guard, err error) {

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

	guard, err = model_state.NewGuard(
		key,
		name,
		details)
	if err != nil {
		return model_state.Guard{}, err
	}

	return guard, nil
}

func actionFromYamlData(classKey, name string, actionAny any) (action model_state.Action, err error) {

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

	action, err = model_state.NewAction(
		key,
		name,
		details,
		requires,
		guarantees)
	if err != nil {
		return model_state.Action{}, err
	}

	return action, nil
}

func transitionFromYamlData(stateKeyLookup, eventKeyLookup, guardKeyLookup, actionKeyLookup map[string]string, fromClassKey string, index int, transitionAny any) (transition model_state.Transition, err error) {

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
					return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, stateName))
				}
			}
		}

		eventKey := ""
		eventNameAny, found := transitionData["event"]
		if found {
			eventName := eventNameAny.(string)
			eventKey, found = eventKeyLookup[eventName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown event: '%s'`, eventName))
			}
		}

		guardKey := ""
		guardNameAny, found := transitionData["guard"]
		if found {
			guardName := guardNameAny.(string)
			guardKey, found = guardKeyLookup[guardName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown guard: '%s'`, guardName))
			}
		}

		actionKey := ""
		actionNameAny, found := transitionData["action"]
		if found {
			actionName := actionNameAny.(string)
			actionKey, found = actionKeyLookup[actionName]
			if !found {
				return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
			}
		}

		toStateKey := ""
		toStateNameAny, found := transitionData["to"]
		if found {
			stateName := toStateNameAny.(string)
			if stateName != "" {
				toStateKey = stateKeyLookup[stateName]
				if !found {
					return model_state.Transition{}, errors.WithStack(errors.Errorf(`unknown state: '%s'`, stateName))
				}
			}
		}

		umlComment := ""
		umlCommentAny, found := transitionData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		transition, err = model_state.NewTransition(
			key,
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

func generateClassContent(class model_class.Class) string {
	yaml := ""
	if class.ActorKey != "" {
		yaml += "actor_key: " + class.ActorKey + "\n"
	}
	if class.SuperclassOfKey != "" {
		yaml += "superclass_of_key: " + class.SuperclassOfKey + "\n"
	}
	if class.SubclassOfKey != "" {
		yaml += "subclass_of_key: " + class.SubclassOfKey + "\n"
	}
	if len(class.Attributes) > 0 {
		yaml += "\n"
		yaml += "attributes:\n"
		for _, attr := range class.Attributes {
			yaml += "\n"
			name := strings.Split(attr.Key, "/")[len(strings.Split(attr.Key, "/"))-1]
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
	if len(class.Associations) > 0 {
		yaml += "\n"
		yaml += "associations:\n"

		for _, assoc := range class.Associations {
			yaml += "\n"

			// yaml cannot handle a 1 as string, so wrap it in quotes.
			fromMultiplicityValue := assoc.FromMultiplicity.ParsedString()
			if fromMultiplicityValue == "1" {
				fromMultiplicityValue = "\"1\""
			}
			toMultiplicityValue := assoc.ToMultiplicity.ParsedString()
			if toMultiplicityValue == "1" {
				toMultiplicityValue = "\"1\""
			}

			yaml += "    - name: " + assoc.Name + "\n"
			if assoc.Details != "" {
				yaml += "      details: " + assoc.Details + "\n"
			}
			yaml += "      from_multiplicity: " + fromMultiplicityValue + "\n"
			yaml += "      to_class_key: " + assoc.ToClassKey + "\n"
			yaml += "      to_multiplicity: " + toMultiplicityValue + "\n"
			if assoc.AssociationClassKey != "" {
				yaml += "      association_class_key: " + assoc.AssociationClassKey + "\n"
			}
			if assoc.UmlComment != "" {
				yaml += "      uml_comment: " + assoc.UmlComment + "\n"
			}
		}
	}

	// We need a lookup of actions to display names where they need to be.
	stateKeyLookups := map[string]model_state.State{}
	for _, state := range class.States {
		stateKeyLookups[state.Key] = state
	}
	actionKeyLookup := map[string]model_state.Action{}
	for _, action := range class.Actions {
		actionKeyLookup[action.Key] = action
	}
	eventKeyLookup := map[string]model_state.Event{}
	for _, event := range class.Events {
		eventKeyLookup[event.Key] = event
	}
	guardKeyLookup := map[string]model_state.Guard{}
	for _, guard := range class.Guards {
		guardKeyLookup[guard.Key] = guard
	}

	if len(class.States) > 0 {
		yaml += "\n"
		yaml += "states:\n"
		for _, state := range class.States {
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
					yaml += "        - action: " + actionKeyLookup[sa.ActionKey].Name + "\n"
					yaml += "          when: " + sa.When + "\n"
				}
			}
		}
	}
	if len(class.Events) > 0 {
		yaml += "\n"
		yaml += "events:\n"
		for _, event := range class.Events {
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
		for _, guard := range class.Guards {
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
		for _, action := range class.Actions {
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
		for _, trans := range class.Transitions {
			from := ""
			if trans.FromStateKey != "" {
				from = stateKeyLookups[trans.FromStateKey].Name
			}
			event := ""
			if trans.EventKey != "" {
				event = eventKeyLookup[trans.EventKey].Name
			}
			to := ""
			if trans.ToStateKey != "" {
				to = stateKeyLookups[trans.ToStateKey].Name
			}
			guard := ""
			if trans.GuardKey != "" {
				guard = guardKeyLookup[trans.GuardKey].Name
			}
			action := ""
			if trans.ActionKey != "" {
				action = actionKeyLookup[trans.ActionKey].Name
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
