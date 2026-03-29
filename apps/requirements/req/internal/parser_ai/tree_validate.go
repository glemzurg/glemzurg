package parser_ai

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
)

// sortedKeys returns sorted keys from a map.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// domainScopedClassKeys returns sorted "subdomain/class" keys across all subdomains in a domain.
func domainScopedClassKeys(domain *inputDomain) []string {
	var keys []string
	for sdKey, sd := range domain.Subdomains {
		for cKey := range sd.Classes {
			keys = append(keys, sdKey+"/"+cKey)
		}
	}
	sort.Strings(keys)
	return keys
}

// modelScopedClassKeys returns sorted "domain/subdomain/class" keys across the entire model.
func modelScopedClassKeys(model *inputModel) []string {
	var keys []string
	for dKey, d := range model.Domains {
		for sdKey, sd := range d.Subdomains {
			for cKey := range sd.Classes {
				keys = append(keys, dKey+"/"+sdKey+"/"+cKey)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

// validateModelTree validates a complete model tree for cross-reference integrity.
// This is called automatically after the tree has been successfully loaded from the filesystem.
// It checks that all keys referenced in the tree point to valid entities.
func validateModelTree(model *inputModel) error {
	// Validate each domain
	for domainKey, domain := range model.Domains {
		if err := validateDomainTree(model, domainKey, domain); err != nil {
			return err
		}
	}

	// Validate model-level associations
	for assocKey, assoc := range model.ClassAssociations {
		if err := validateModelAssociation(model, assocKey, assoc); err != nil {
			return err
		}
	}

	// Validate domain associations (cross-domain references)
	for daKey, da := range model.DomainAssociations {
		if err := validateModelDomainAssociation(model, daKey, da); err != nil {
			return err
		}
	}

	// Validate actor generalizations reference real actors
	for agKey, ag := range model.ActorGeneralizations {
		if err := validateActorGeneralizationTree(model, agKey, ag); err != nil {
			return err
		}
	}

	return nil
}

// validateModelCompleteness validates that a model is complete enough to be useful.
// This provides guidance to AI about what elements are still needed.
// It checks that the model has all required structural components.
func validateModelCompleteness(model *inputModel) error {
	// Check model has at least one actor
	if len(model.Actors) == 0 {
		return NewParseError(
			ErrTreeModelNoActors,
			"model must have at least one actor defined - actors represent the users, systems, or external entities that interact with your system; define actors in the 'actors/' directory with files like 'actors/user.actor.json'",
			"model.json",
		).WithField("actors").WithHint("create actors/{key}.actor.json with {\"name\": ..., \"type\": \"person|external_system|time\"}")
	}

	// Check model has at least one domain
	if len(model.Domains) == 0 {
		return NewParseError(
			ErrTreeModelNoDomains,
			"model must have at least one domain defined - domains are high-level subject areas that group related functionality; create a domain directory under 'domains/' with a 'domain.json' file",
			"model.json",
		).WithField("domains").WithHint("create domains/{key}/domain.json with {\"name\": ...}")
	}

	// Validate each domain's completeness
	for domainKey, domain := range model.Domains {
		if err := validateDomainCompleteness(domainKey, domain); err != nil {
			return err
		}
	}

	return nil
}

// validateDomainCompleteness validates that a domain is complete.
func validateDomainCompleteness(domainKey string, domain *inputDomain) error {
	domainPath := fmt.Sprintf("domains/%s/domain.json", domainKey)

	// Check domain has at least one subdomain
	if len(domain.Subdomains) == 0 {
		return NewParseError(
			ErrTreeDomainNoSubdomains,
			fmt.Sprintf("domain '%s' must have at least one subdomain defined - subdomains organize classes within a domain; create a subdomain directory under 'domains/%s/' with a 'subdomain.json' file",
				domainKey, domainKey),
			domainPath,
		).WithField("subdomains").WithHint(fmt.Sprintf("create domains/%s/subdomains/default/subdomain.json", domainKey))
	}

	// Validate subdomain naming rules
	if err := validateSubdomainNaming(domainKey, domain); err != nil {
		return err
	}

	// Validate each subdomain's completeness
	for subdomainKey, subdomain := range domain.Subdomains {
		if err := validateSubdomainCompleteness(domainKey, subdomainKey, subdomain); err != nil {
			return err
		}
	}

	return nil
}

// validateSubdomainNaming validates subdomain naming rules:
// - If there's only one subdomain, it must be named "default"
// - If there are multiple subdomains, none can be named "default".
func validateSubdomainNaming(domainKey string, domain *inputDomain) error {
	subdomainCount := len(domain.Subdomains)

	// Check if "default" subdomain exists
	_, hasDefault := domain.Subdomains["default"]

	if subdomainCount == 1 {
		// Single subdomain must be named "default"
		if !hasDefault {
			// Get the actual subdomain name
			var actualName string
			for name := range domain.Subdomains {
				actualName = name
				break
			}
			subdomainPath := fmt.Sprintf("domains/%s/subdomains/%s", domainKey, actualName)
			return NewParseError(
				ErrTreeSingleSubdomainNotDefault,
				fmt.Sprintf("domain '%s' has a single subdomain '%s' which must be renamed to 'default' - "+
					"when a domain has only one subdomain, it must be named 'default'; "+
					"rename 'domains/%s/subdomains/%s/' to 'domains/%s/subdomains/default/'",
					domainKey, actualName, domainKey, actualName, domainKey),
				subdomainPath,
			).WithField("subdomain_key").WithHint("rename subdomain directory to 'default'")
		}
	} else if subdomainCount > 1 {
		// Multiple subdomains cannot include one named "default"
		if hasDefault {
			subdomainPath := fmt.Sprintf("domains/%s/subdomains/default", domainKey)
			return NewParseError(
				ErrTreeMultipleSubdomainsHasDefault,
				fmt.Sprintf("domain '%s' has multiple subdomains but one is named 'default' - "+
					"when a domain has multiple subdomains, none should be named 'default'; "+
					"rename 'domains/%s/subdomains/default/' to a more descriptive name that reflects its purpose",
					domainKey, domainKey),
				subdomainPath,
			).WithField("subdomain_key").WithHint("rename 'default' subdomain to a descriptive name")
		}
	}

	return nil
}

// validateSubdomainCompleteness validates that a subdomain is complete.
func validateSubdomainCompleteness(domainKey, subdomainKey string, subdomain *inputSubdomain) error {
	subdomainPath := fmt.Sprintf("domains/%s/subdomains/%s/subdomain.json", domainKey, subdomainKey)

	// Check subdomain has at least 2 classes
	if len(subdomain.Classes) < 2 {
		return NewParseError(
			ErrTreeSubdomainTooFewClasses,
			fmt.Sprintf("subdomain '%s' must have at least 2 classes defined (has %d) - a subdomain needs multiple classes to represent meaningful relationships; create class directories under 'domains/%s/subdomains/%s/classes/' with 'class.json' files",
				subdomainKey, len(subdomain.Classes), domainKey, subdomainKey),
			subdomainPath,
		).WithField("classes").WithHint("create class directories under classes/ with class.json files")
	}

	// Check subdomain has at least one association
	if len(subdomain.ClassAssociations) == 0 {
		return NewParseError(
			ErrTreeSubdomainNoAssociations,
			fmt.Sprintf("subdomain '%s' must have at least one association defined - associations describe how classes relate to each other; create association files under 'domains/%s/subdomains/%s/class_associations/' with '.assoc.json' extension",
				subdomainKey, domainKey, subdomainKey),
			subdomainPath,
		).WithField("class_associations").WithHint("create {from}--{to}--{name}.assoc.json in class_associations/")
	}

	// Validate each class's completeness
	for classKey, class := range subdomain.Classes {
		if err := validateClassCompleteness(domainKey, subdomainKey, classKey, class); err != nil {
			return err
		}
	}

	return nil
}

// validateClassCompleteness validates that a class is complete.
func validateClassCompleteness(domainKey, subdomainKey, classKey string, class *inputClass) error {
	classPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKey, subdomainKey, classKey)

	// Check class has at least one attribute
	if len(class.Attributes) == 0 {
		return NewParseError(
			ErrTreeClassNoAttributes,
			fmt.Sprintf("class '%s' must have at least one attribute defined - attributes describe the data properties of a class; add attributes to the 'attributes' map in the class.json file with name, data type rules, and details",
				classKey),
			classPath,
		).WithField("attributes").WithHint("add attributes map to class.json: {\"attributes\": {\"attr_key\": {\"name\": ...}}}")
	}

	// Check class has a state machine
	if class.StateMachine == nil {
		return NewParseError(
			ErrTreeClassNoStateMachine,
			fmt.Sprintf("class '%s' must have a state machine defined - state machines describe the lifecycle and behavior of a class; create a 'state_machine.json' file in the class directory with states, events, and transitions",
				classKey),
			classPath,
		).WithField("state_machine").WithHint("create state_machine.json with states, events, and transitions")
	}

	// Check state machine has at least one transition
	if len(class.StateMachine.Transitions) == 0 {
		smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)
		return NewParseError(
			ErrTreeStateMachineNoTransitions,
			fmt.Sprintf("state machine for class '%s' must have at least one transition defined - transitions describe how the class moves between states in response to events; add transitions to the 'transitions' array with event_key and state references",
				classKey),
			smPath,
		).WithField("transitions").WithHint("add transitions: [{\"event_key\": ..., \"from_state_key\": ..., \"to_state_key\": ...}]")
	}

	return nil
}

// validateDomainTree validates a domain and its children.
func validateDomainTree(model *inputModel, domainKey string, domain *inputDomain) error {
	// Validate each subdomain
	for subdomainKey, subdomain := range domain.Subdomains {
		if err := validateSubdomainTree(model, domainKey, subdomainKey, subdomain); err != nil {
			return err
		}
	}

	// Validate domain-level associations
	for assocKey, assoc := range domain.ClassAssociations {
		if err := validateDomainAssociation(domainKey, domain, assocKey, assoc); err != nil {
			return err
		}
	}

	return nil
}

// validateSubdomainTree validates a subdomain and its children.
func validateSubdomainTree(model *inputModel, domainKey, subdomainKey string, subdomain *inputSubdomain) error {
	// Validate each class
	for classKey, class := range subdomain.Classes {
		if err := validateClassTree(model, domainKey, subdomainKey, classKey, class); err != nil {
			return err
		}
	}

	// Validate class generalizations
	for genKey, gen := range subdomain.ClassGeneralizations {
		if err := validateClassGeneralizationTree(subdomain, domainKey, subdomainKey, genKey, gen); err != nil {
			return err
		}
	}

	// Validate subdomain-level associations
	for assocKey, assoc := range subdomain.ClassAssociations {
		if err := validateSubdomainAssociation(subdomain, domainKey, subdomainKey, assocKey, assoc); err != nil {
			return err
		}
	}

	// Validate each use case
	for useCaseKey, useCase := range subdomain.UseCases {
		if err := validateUseCaseTree(model, domainKey, subdomainKey, useCaseKey, useCase); err != nil {
			return err
		}
	}

	// Validate use case generalizations
	for genKey, gen := range subdomain.UseCaseGeneralizations {
		if err := validateUseCaseGeneralizationTree(subdomain, domainKey, subdomainKey, genKey, gen); err != nil {
			return err
		}
	}

	return nil
}

// validateClassTree validates a class and its children.
func validateClassTree(model *inputModel, domainKey, subdomainKey, classKey string, class *inputClass) error {
	classPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKey, subdomainKey, classKey)

	// Validate actor_key if present
	if class.ActorKey != "" {
		if _, ok := model.Actors[class.ActorKey]; !ok {
			return NewParseError(
				ErrTreeClassActorNotFound,
				fmt.Sprintf("class '%s' references actor '%s' which does not exist", classKey, class.ActorKey),
				classPath,
			).WithField("actor_key").WithHint(fmt.Sprintf("available actors: %s", strings.Join(sortedKeys(model.Actors), ", ")))
		}
	}

	// Validate indexes reference valid attributes
	for i, index := range class.Indexes {
		seen := make(map[string]bool)
		for j, attrKey := range index {
			// Check for duplicates within this index
			if seen[attrKey] {
				return NewParseError(
					ErrTreeClassIndexAttrNotFound,
					fmt.Sprintf("class '%s' index[%d] contains duplicate attribute key '%s'", classKey, i, attrKey),
					classPath,
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j)).WithHint(fmt.Sprintf("available attributes: %s", strings.Join(sortedKeys(class.Attributes), ", ")))
			}
			seen[attrKey] = true

			// Check that the attribute exists
			if _, ok := class.Attributes[attrKey]; !ok {
				return NewParseError(
					ErrTreeClassIndexAttrNotFound,
					fmt.Sprintf("class '%s' index[%d] references attribute '%s' which does not exist", classKey, i, attrKey),
					classPath,
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j)).WithHint(fmt.Sprintf("available attributes: %s", strings.Join(sortedKeys(class.Attributes), ", ")))
			}
		}
	}

	// Validate attribute key-name consistency
	if err := validateAttributeKeyNameConsistency(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate attribute data_type_rules are parseable
	if err := validateClassDataTypes(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate name uniqueness across actions, queries, states, events, guards
	if err := validateClassNameUniqueness(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate state machine key-name consistency
	if err := validateStateMachineKeyNameConsistency(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate state machine if present
	if class.StateMachine != nil {
		if err := validateStateMachineTree(class, domainKey, subdomainKey, classKey); err != nil {
			return err
		}
	}

	return nil
}

// dataTypeHint is the concise hint shown for data type parse errors.
const dataTypeHint = "valid types: unconstrained, enum of v1, v2, v3, [1..100] at 1 unit, ordered/unordered/stack/queue of <type>, { field: <type> }. integers and floats are spans e.g. [0..unconstrained] at 1 count or [0..unconstrained] at 0.01 dollars. booleans are enum of true, false. strings are unconstrained for free text, enum of x, y for a fixed set, or ref from Source Name for externally documented values (e.g. ISO codes)"

// validateClassDataTypes validates that all attribute data_type_rules in a class are parseable.
func validateClassDataTypes(class *inputClass, domainKey, subdomainKey, classKey string) error {
	classPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKey, subdomainKey, classKey)

	// Validate attribute data_type_rules
	for attrKey, attr := range class.Attributes {
		if attr.DataTypeRules == "" {
			continue
		}
		_, err := model_data_type.New(attrKey, attr.DataTypeRules, nil)
		if err != nil {
			return NewParseError(
				ErrClassDataTypeUnparseable,
				fmt.Sprintf("class '%s' attribute '%s' data_type_rules could not be parsed: %s", classKey, attrKey, err.Error()),
				classPath,
			).WithField(fmt.Sprintf("attributes.%s.data_type_rules", attrKey)).WithHint(dataTypeHint)
		}
	}

	// Validate action parameter data_type_rules
	if err := validateActionParamDataTypes(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate query parameter data_type_rules
	if err := validateQueryParamDataTypes(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	// Validate event parameter data_type_rules
	if err := validateEventParamDataTypes(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	return nil
}

// validateActionParamDataTypes validates data_type_rules on action parameters.
func validateActionParamDataTypes(class *inputClass, domainKey, subdomainKey, classKey string) error {
	for actionKey, action := range class.Actions {
		for i, param := range action.Parameters {
			if param.DataTypeRules == "" {
				continue
			}
			_, err := model_data_type.New(param.Name, param.DataTypeRules, nil)
			if err != nil {
				actionPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/actions/%s.json", domainKey, subdomainKey, classKey, actionKey)
				return NewParseError(
					ErrParamDataTypeUnparseable,
					fmt.Sprintf("action '%s' parameter[%d] '%s' data_type_rules could not be parsed: %s", actionKey, i, param.Name, err.Error()),
					actionPath,
				).WithField(fmt.Sprintf("parameters[%d].data_type_rules", i)).WithHint(dataTypeHint)
			}
		}
	}
	return nil
}

// validateQueryParamDataTypes validates data_type_rules on query parameters.
func validateQueryParamDataTypes(class *inputClass, domainKey, subdomainKey, classKey string) error {
	for queryKey, query := range class.Queries {
		for i, param := range query.Parameters {
			if param.DataTypeRules == "" {
				continue
			}
			_, err := model_data_type.New(param.Name, param.DataTypeRules, nil)
			if err != nil {
				queryPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/queries/%s.json", domainKey, subdomainKey, classKey, queryKey)
				return NewParseError(
					ErrParamDataTypeUnparseable,
					fmt.Sprintf("query '%s' parameter[%d] '%s' data_type_rules could not be parsed: %s", queryKey, i, param.Name, err.Error()),
					queryPath,
				).WithField(fmt.Sprintf("parameters[%d].data_type_rules", i)).WithHint(dataTypeHint)
			}
		}
	}
	return nil
}

// validateEventParamDataTypes validates data_type_rules on event parameters.
func validateEventParamDataTypes(class *inputClass, domainKey, subdomainKey, classKey string) error {
	if class.StateMachine == nil {
		return nil
	}
	for eventKey, event := range class.StateMachine.Events {
		for i, param := range event.Parameters {
			if param.DataTypeRules == "" {
				continue
			}
			_, err := model_data_type.New(param.Name, param.DataTypeRules, nil)
			if err != nil {
				smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)
				return NewParseError(
					ErrEventParamDataTypeUnparseable,
					fmt.Sprintf("event '%s' parameter[%d] '%s' data_type_rules could not be parsed: %s", eventKey, i, param.Name, err.Error()),
					smPath,
				).WithField(fmt.Sprintf("events.%s.parameters[%d].data_type_rules", eventKey, i)).WithHint(dataTypeHint)
			}
		}
	}
	return nil
}

// validateStateMachineTree validates a state machine's cross-references.
func validateStateMachineTree(class *inputClass, domainKey, subdomainKey, classKey string) error {
	sm := class.StateMachine
	smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)

	// Validate state actions reference existing actions
	for stateKey, state := range sm.States {
		for i, stateAction := range state.Actions {
			if _, ok := class.Actions[stateAction.ActionKey]; !ok {
				return NewParseError(
					ErrTreeStateMachineActionNotFound,
					fmt.Sprintf("state '%s' action[%d] references action '%s' which does not exist in class '%s'",
						stateKey, i, stateAction.ActionKey, classKey),
					smPath,
				).WithField(fmt.Sprintf("states.%s.actions[%d].action_key", stateKey, i)).WithHint(fmt.Sprintf("available actions: %s", strings.Join(sortedKeys(class.Actions), ", ")))
			}
		}
	}

	// Validate transitions
	for i, transition := range sm.Transitions {
		if err := validateSingleTransitionTree(class, sm, i, transition, classKey, smPath); err != nil {
			return err
		}
	}

	// Validate that every action is referenced by at least one state action or transition
	if err := validateActionsReferenced(class, domainKey, subdomainKey, classKey); err != nil {
		return err
	}

	return nil
}

// validateSingleTransitionTree validates a single transition's cross-references.
func validateSingleTransitionTree(class *inputClass, sm *inputStateMachine, i int, transition inputTransition, classKey, smPath string) error {
	if transition.FromStateKey == nil && transition.ToStateKey == nil {
		return NewParseError(
			ErrTreeTransitionNoStates,
			fmt.Sprintf("transition[%d] must have at least one of from_state_key or to_state_key", i),
			smPath,
		).WithField(fmt.Sprintf("transitions[%d]", i)).WithHint("add from_state_key, to_state_key, or both")
	}

	if transition.FromStateKey != nil {
		if _, ok := sm.States[*transition.FromStateKey]; !ok {
			return NewParseError(
				ErrTreeStateMachineStateNotFound,
				fmt.Sprintf("transition[%d] from_state_key '%s' does not exist", i, *transition.FromStateKey),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d].from_state_key", i)).WithHint(fmt.Sprintf("available states: %s", strings.Join(sortedKeys(sm.States), ", ")))
		}
	}

	if transition.ToStateKey != nil {
		if _, ok := sm.States[*transition.ToStateKey]; !ok {
			return NewParseError(
				ErrTreeStateMachineStateNotFound,
				fmt.Sprintf("transition[%d] to_state_key '%s' does not exist", i, *transition.ToStateKey),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d].to_state_key", i)).WithHint(fmt.Sprintf("available states: %s", strings.Join(sortedKeys(sm.States), ", ")))
		}
	}

	if _, ok := sm.Events[transition.EventKey]; !ok {
		return NewParseError(
			ErrTreeStateMachineEventNotFound,
			fmt.Sprintf("transition[%d] event_key '%s' does not exist", i, transition.EventKey),
			smPath,
		).WithField(fmt.Sprintf("transitions[%d].event_key", i)).WithHint(fmt.Sprintf("available events: %s", strings.Join(sortedKeys(sm.Events), ", ")))
	}

	if transition.GuardKey != nil {
		if _, ok := sm.Guards[*transition.GuardKey]; !ok {
			return NewParseError(
				ErrTreeStateMachineGuardNotFound,
				fmt.Sprintf("transition[%d] guard_key '%s' does not exist", i, *transition.GuardKey),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d].guard_key", i)).WithHint(fmt.Sprintf("available guards: %s", strings.Join(sortedKeys(sm.Guards), ", ")))
		}
	}

	if transition.ActionKey != nil {
		if _, ok := class.Actions[*transition.ActionKey]; !ok {
			return NewParseError(
				ErrTreeStateMachineActionNotFound,
				fmt.Sprintf("transition[%d] action_key '%s' does not exist in class '%s'", i, *transition.ActionKey, classKey),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d].action_key", i)).WithHint(fmt.Sprintf("available actions: %s", strings.Join(sortedKeys(class.Actions), ", ")))
		}
	}

	return nil
}

// validateActionsReferenced ensures every action in a class is referenced by at least one
// state action (entry/exit/do) or transition action.
func validateActionsReferenced(class *inputClass, domainKey, subdomainKey, classKey string) error {
	if len(class.Actions) == 0 {
		return nil
	}

	sm := class.StateMachine
	if sm == nil {
		return nil
	}

	// Build a set of all referenced action keys
	referencedActions := make(map[string]bool)

	// Check state actions (entry, exit, do)
	for _, state := range sm.States {
		for _, stateAction := range state.Actions {
			referencedActions[stateAction.ActionKey] = true
		}
	}

	// Check transition actions
	for _, transition := range sm.Transitions {
		if transition.ActionKey != nil {
			referencedActions[*transition.ActionKey] = true
		}
	}

	// Check each action is referenced
	for actionKey := range class.Actions {
		if !referencedActions[actionKey] {
			actionPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/actions/%s.json", domainKey, subdomainKey, classKey, actionKey)
			return NewParseError(
				ErrTreeActionUnreferenced,
				fmt.Sprintf("action '%s' in class '%s' is defined but not referenced by any state action or transition - "+
					"every action must be used in the state machine either as a state entry/exit/do action or as a transition action",
					actionKey, classKey),
				actionPath,
			).WithField("action_key").WithHint("reference this action in a state entry/exit/do or transition action_key")
		}
	}

	return nil
}

// validateClassGeneralizationTree validates a class generalization's cross-references.
func validateClassGeneralizationTree(subdomain *inputSubdomain, domainKey, subdomainKey, genKey string, gen *inputClassGeneralization) error {
	genPath := fmt.Sprintf("domains/%s/subdomains/%s/class_generalizations/%s.cgen.json", domainKey, subdomainKey, genKey)

	// Validate superclass_key exists
	if _, ok := subdomain.Classes[gen.SuperclassKey]; !ok {
		return NewParseError(
			ErrTreeClassGenSuperclassNotFound,
			fmt.Sprintf("class generalization '%s' superclass_key '%s' does not exist in subdomain '%s'",
				genKey, gen.SuperclassKey, subdomainKey),
			genPath,
		).WithField("superclass_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Validate subclass_keys exist and are unique
	seen := make(map[string]bool)
	for i, subclassKey := range gen.SubclassKeys {
		// Check for duplicates
		if seen[subclassKey] {
			return NewParseError(
				ErrTreeClassGenSubclassDuplicate,
				fmt.Sprintf("class generalization '%s' has duplicate subclass_key '%s'", genKey, subclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("remove duplicate entries from subclass_keys")
		}
		seen[subclassKey] = true

		// Check that the subclass exists
		if _, ok := subdomain.Classes[subclassKey]; !ok {
			return NewParseError(
				ErrTreeClassGenSubclassNotFound,
				fmt.Sprintf("class generalization '%s' subclass_key '%s' does not exist in subdomain '%s'",
					genKey, subclassKey, subdomainKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
		}

		// Check that superclass is not also a subclass
		if subclassKey == gen.SuperclassKey {
			return NewParseError(
				ErrTreeClassGenSuperclassIsSubclass,
				fmt.Sprintf("class generalization '%s' superclass '%s' cannot also be a subclass",
					genKey, gen.SuperclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("superclass cannot also appear in subclass_keys")
		}
	}

	return nil
}

// validateModelDomainAssociation validates that a parsed domain association refers to existing domains.
func validateModelDomainAssociation(model *inputModel, key string, da *inputDomainAssociation) error {
	assocPath := fmt.Sprintf("domain_associations/%s.domain_assoc.json", key)

	if _, ok := model.Domains[da.ProblemDomainKey]; !ok {
		return NewParseError(
			ErrTreeDomainAssocDomainNotFound,
			fmt.Sprintf("domain association '%s' problem_domain_key '%s' references domain which does not exist",
				key, da.ProblemDomainKey),
			assocPath,
		).WithField("problem_domain_key").WithHint(fmt.Sprintf("available domains: %s", strings.Join(sortedKeys(model.Domains), ", ")))
	}
	if _, ok := model.Domains[da.SolutionDomainKey]; !ok {
		return NewParseError(
			ErrTreeDomainAssocDomainNotFound,
			fmt.Sprintf("domain association '%s' solution_domain_key '%s' references domain which does not exist",
				key, da.SolutionDomainKey),
			assocPath,
		).WithField("solution_domain_key").WithHint(fmt.Sprintf("available domains: %s", strings.Join(sortedKeys(model.Domains), ", ")))
	}
	return nil
}

// validateActorGeneralizationTree validates that actor generalizations reference real actors.
func validateActorGeneralizationTree(model *inputModel, genKey string, gen *inputActorGeneralization) error {
	genPath := fmt.Sprintf("actor_generalizations/%s.agen.json", genKey)

	if _, ok := model.Actors[gen.SuperclassKey]; !ok {
		return NewParseError(
			ErrTreeActorGenActorNotFound,
			fmt.Sprintf("actor generalization '%s' superclass_key '%s' does not exist",
				genKey, gen.SuperclassKey),
			genPath,
		).WithField("superclass_key").WithHint(fmt.Sprintf("available actors: %s", strings.Join(sortedKeys(model.Actors), ", ")))
	}

	seen := make(map[string]bool)
	for i, subclassKey := range gen.SubclassKeys {
		if seen[subclassKey] {
			// reuse existing actor gen duplicate error code if present
			return NewParseError(
				ErrActorGenSubclassesEmpty,
				fmt.Sprintf("actor generalization '%s' has duplicate subclass_key '%s'", genKey, subclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("remove duplicate entries from subclass_keys")
		}
		seen[subclassKey] = true

		if _, ok := model.Actors[subclassKey]; !ok {
			return NewParseError(
				ErrTreeActorGenActorNotFound,
				fmt.Sprintf("actor generalization '%s' subclass_key '%s' does not exist",
					genKey, subclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint(fmt.Sprintf("available actors: %s", strings.Join(sortedKeys(model.Actors), ", ")))
		}

		if subclassKey == gen.SuperclassKey {
			return NewParseError(
				ErrTreeActorGenActorNotFound,
				fmt.Sprintf("actor generalization '%s' superclass '%s' cannot also be a subclass",
					genKey, gen.SuperclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("superclass cannot also appear in subclass_keys")
		}
	}
	return nil
}

// validateSubdomainAssociation validates an association at the subdomain level.
// Keys are scoped to the subdomain (just class names).
func validateSubdomainAssociation(subdomain *inputSubdomain, domainKey, subdomainKey, assocKey string, assoc *inputClassAssociation) error {
	assocPath := fmt.Sprintf("domains/%s/subdomains/%s/class_associations/%s.assoc.json", domainKey, subdomainKey, assocKey)

	// Validate from_class_key
	if _, ok := subdomain.Classes[assoc.FromClassKey]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' does not exist in subdomain '%s'",
				assocKey, assoc.FromClassKey, subdomainKey),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Validate to_class_key
	if _, ok := subdomain.Classes[assoc.ToClassKey]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' does not exist in subdomain '%s'",
				assocKey, assoc.ToClassKey, subdomainKey),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes in this subdomain: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Validate association_class_key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		if _, ok := subdomain.Classes[*assoc.AssociationClassKey]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' does not exist in subdomain '%s'",
					assocKey, *assoc.AssociationClassKey, subdomainKey),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	return nil
}

// validateDomainAssociation validates an association at the domain level.
// Keys include subdomain to disambiguate (subdomain/class).
func validateDomainAssociation(domainKey string, domain *inputDomain, assocKey string, assoc *inputClassAssociation) error {
	assocPath := fmt.Sprintf("domains/%s/class_associations/%s.assoc.json", domainKey, assocKey)

	domainClassKeys := domainScopedClassKeys(domain)

	// Parse from_class_key (subdomain/class format)
	fromSubdomain, fromClass, err := parseDomainScopedKey(assoc.FromClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' is invalid: %s",
				assocKey, assoc.FromClassKey, err.Error()),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(domainClassKeys, ", ")))
	}

	// Check from subdomain exists
	subdomain, ok := domain.Subdomains[fromSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references subdomain '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromSubdomain),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(domainClassKeys, ", ")))
	}

	// Check from class exists
	if _, ok := subdomain.Classes[fromClass]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references class '%s' which does not exist in subdomain '%s'",
				assocKey, assoc.FromClassKey, fromClass, fromSubdomain),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes in subdomain '%s': %s", fromSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Parse to_class_key (subdomain/class format)
	toSubdomain, toClass, err := parseDomainScopedKey(assoc.ToClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' is invalid: %s",
				assocKey, assoc.ToClassKey, err.Error()),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Domain-level key format is 'subdomain/class'. Available classes: %s", strings.Join(domainClassKeys, ", ")))
	}

	// Check to subdomain exists
	subdomain, ok = domain.Subdomains[toSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references subdomain '%s' which does not exist",
				assocKey, assoc.ToClassKey, toSubdomain),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes: %s", strings.Join(domainClassKeys, ", ")))
	}

	// Check to class exists
	if _, ok := subdomain.Classes[toClass]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references class '%s' which does not exist in subdomain '%s'",
				assocKey, assoc.ToClassKey, toClass, toSubdomain),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes in subdomain '%s': %s", toSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Validate association_class_key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		assocSubdomain, assocClass, err := parseDomainScopedKey(*assoc.AssociationClassKey)
		if err != nil {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' is invalid: %s",
					assocKey, *assoc.AssociationClassKey, err.Error()),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(domainClassKeys, ", ")))
		}

		subdomain, ok := domain.Subdomains[assocSubdomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references subdomain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocSubdomain),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(domainClassKeys, ", ")))
		}

		if _, ok := subdomain.Classes[assocClass]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references class '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocClass),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes in subdomain '%s': %s", assocSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	return nil
}

// validateModelAssociation validates an association at the model level.
// Keys include domain and subdomain (domain/subdomain/class).
func validateModelAssociation(model *inputModel, assocKey string, assoc *inputClassAssociation) error {
	assocPath := fmt.Sprintf("class_associations/%s.assoc.json", assocKey)

	allClassKeys := modelScopedClassKeys(model)

	// Parse from_class_key (domain/subdomain/class format)
	fromDomain, fromSubdomain, fromClass, err := parseModelScopedKey(assoc.FromClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' is invalid: %s",
				assocKey, assoc.FromClassKey, err.Error()),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check from domain exists
	domain, ok := model.Domains[fromDomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references domain '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromDomain),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check from subdomain exists
	subdomain, ok := domain.Subdomains[fromSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references subdomain '%s' which does not exist in domain '%s'",
				assocKey, assoc.FromClassKey, fromSubdomain, fromDomain),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check from class exists
	if _, ok := subdomain.Classes[fromClass]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references class '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromClass),
			assocPath,
		).WithField("from_class_key").WithHint(fmt.Sprintf("available classes in subdomain '%s/%s': %s", fromDomain, fromSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Parse to_class_key (domain/subdomain/class format)
	toDomain, toSubdomain, toClass, err := parseModelScopedKey(assoc.ToClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' is invalid: %s",
				assocKey, assoc.ToClassKey, err.Error()),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Model-level key format is 'domain/subdomain/class'. Available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check to domain exists
	domain, ok = model.Domains[toDomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references domain '%s' which does not exist",
				assocKey, assoc.ToClassKey, toDomain),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check to subdomain exists
	subdomain, ok = domain.Subdomains[toSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references subdomain '%s' which does not exist in domain '%s'",
				assocKey, assoc.ToClassKey, toSubdomain, toDomain),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes: %s", strings.Join(allClassKeys, ", ")))
	}

	// Check to class exists
	if _, ok := subdomain.Classes[toClass]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references class '%s' which does not exist",
				assocKey, assoc.ToClassKey, toClass),
			assocPath,
		).WithField("to_class_key").WithHint(fmt.Sprintf("first verify the association is at the correct level (subdomain/domain/model), then check the key format matches that level, then create the class if needed. Available classes in subdomain '%s/%s': %s", toDomain, toSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
	}

	// Validate association_class_key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		assocDomain, assocSubdomain, assocClass, err := parseModelScopedKey(*assoc.AssociationClassKey)
		if err != nil {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' is invalid: %s",
					assocKey, *assoc.AssociationClassKey, err.Error()),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
		}

		domain, ok := model.Domains[assocDomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references domain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocDomain),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
		}

		subdomain, ok := domain.Subdomains[assocSubdomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references subdomain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocSubdomain),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes: %s", strings.Join(allClassKeys, ", ")))
		}

		if _, ok := subdomain.Classes[assocClass]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references class '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocClass),
				assocPath,
			).WithField("association_class_key").WithHint(fmt.Sprintf("available classes in subdomain '%s/%s': %s", assocDomain, assocSubdomain, strings.Join(sortedKeys(subdomain.Classes), ", ")))
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key").WithHint("association_class_key must be a different class than from_class_key and to_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity").WithHint("valid multiplicities: 1, 0..1, *, 0..*, 1..*")
	}

	return nil
}

// parseDomainScopedKey parses a key in subdomain/class format.
func parseDomainScopedKey(key string) (subdomain, class string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'subdomain/class', got '%s'", key)
	}
	return parts[0], parts[1], nil
}

// parseModelScopedKey parses a key in domain/subdomain/class format.
func parseModelScopedKey(key string) (domain, subdomain, class string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected format 'domain/subdomain/class', got '%s'", key)
	}
	return parts[0], parts[1], parts[2], nil
}

// multiplicityPattern matches valid multiplicity formats.
var multiplicityPattern = regexp.MustCompile(`^(\d+|\*)$|^(\d+)\.\.(\d+|\*)$`)

// validateMultiplicity checks if a multiplicity string is valid.
func validateMultiplicity(mult string) error {
	if mult == "" {
		return fmt.Errorf("multiplicity cannot be empty")
	}

	if !multiplicityPattern.MatchString(mult) {
		return fmt.Errorf("invalid format")
	}

	// Additional validation for ranges
	parts := strings.Split(mult, "..")
	if len(parts) == 2 {
		lower := parts[0]
		upper := parts[1]

		// If upper is not *, compare numerically
		if upper != "*" {
			var lowerNum, upperNum int
			_, _ = fmt.Sscanf(lower, "%d", &lowerNum)
			_, _ = fmt.Sscanf(upper, "%d", &upperNum)
			if upperNum < lowerNum {
				return fmt.Errorf("upper bound %d cannot be less than lower bound %d", upperNum, lowerNum)
			}
		}
	}

	return nil
}

// validateUseCaseTree validates a use case and its children.
func validateUseCaseTree(model *inputModel, domainKey, subdomainKey, useCaseKey string, useCase *inputUseCase) error {
	useCasePath := fmt.Sprintf("domains/%s/subdomains/%s/use_cases/%s/use_case.json", domainKey, subdomainKey, useCaseKey)

	// Validate actor keys if present (actors reference classes in the same subdomain)
	subdomain := model.Domains[domainKey].Subdomains[subdomainKey]
	for actorKey := range useCase.Actors {
		if _, ok := subdomain.Classes[actorKey]; !ok {
			return NewParseError(
				ErrTreeClassActorNotFound,
				fmt.Sprintf("use case '%s' references actor class '%s' which does not exist in subdomain '%s/%s'", useCaseKey, actorKey, domainKey, subdomainKey),
				useCasePath,
			).WithField("actors").WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(subdomain.Classes), ", ")))
		}
	}

	// Validate scenarios (basic validation for now)
	for scenarioKey, scenario := range useCase.Scenarios {
		if err := validateScenarioTree(model, domainKey, subdomainKey, useCaseKey, scenarioKey, scenario); err != nil {
			return err
		}
	}

	return nil
}

// scenarioValidationContext holds context for scenario step validation.
type scenarioValidationContext struct {
	model        *inputModel
	domainKey    string
	subdomainKey string
	scenarioKey  string
	scenario     *inputScenario
	scenarioPath string
}

// validateScenarioTree validates a scenario's cross-references.
func validateScenarioTree(model *inputModel, domainKey, subdomainKey, useCaseKey, scenarioKey string, scenario *inputScenario) error {
	scenarioPath := fmt.Sprintf("domains/%s/subdomains/%s/use_cases/%s/scenarios/%s.scenario.json", domainKey, subdomainKey, useCaseKey, scenarioKey)

	// Validate object class_keys exist
	for objectKey, object := range scenario.Objects {
		if _, ok := model.Domains[domainKey].Subdomains[subdomainKey].Classes[object.ClassKey]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("scenario '%s' object '%s' references class '%s' which does not exist", scenarioKey, objectKey, object.ClassKey),
				scenarioPath,
			).WithField("objects").WithHint(fmt.Sprintf("available classes: %s", strings.Join(sortedKeys(model.Domains[domainKey].Subdomains[subdomainKey].Classes), ", ")))
		}
	}

	// Validate steps recursively
	if scenario.Steps != nil {
		ctx := scenarioValidationContext{
			model:        model,
			domainKey:    domainKey,
			subdomainKey: subdomainKey,
			scenarioKey:  scenarioKey,
			scenario:     scenario,
			scenarioPath: scenarioPath,
		}
		if err := walkValidateSteps(*scenario.Steps, "steps", ctx); err != nil {
			return err
		}
	}

	return nil
}

// walkValidateSteps recursively validates step cross-references.
func walkValidateSteps(step inputStep, path string, ctx scenarioValidationContext) error {
	if err := validateStepObjectRefs(step, path, ctx); err != nil {
		return err
	}
	if err := validateStepEventRef(step, path, ctx); err != nil {
		return err
	}
	if err := validateStepQueryRef(step, path, ctx); err != nil {
		return err
	}
	for i, st := range step.Statements {
		if err := walkValidateSteps(st, fmt.Sprintf("%s.statements[%d]", path, i), ctx); err != nil {
			return err
		}
	}
	return nil
}

// validateStepObjectRefs validates from_object_key and to_object_key references in a step.
func validateStepObjectRefs(step inputStep, path string, ctx scenarioValidationContext) error {
	if step.FromObjectKey != nil {
		if _, ok := ctx.scenario.Objects[*step.FromObjectKey]; !ok {
			return NewParseError(
				ErrTreeScenarioStepObjectNotFound,
				fmt.Sprintf("scenario '%s' step at '%s' references from_object_key '%s' which does not exist",
					ctx.scenarioKey, path, *step.FromObjectKey),
				ctx.scenarioPath,
			).WithField(path + ".from_object_key").WithHint(fmt.Sprintf("available objects: %s", strings.Join(sortedKeys(ctx.scenario.Objects), ", ")))
		}
	}
	if step.ToObjectKey != nil {
		if _, ok := ctx.scenario.Objects[*step.ToObjectKey]; !ok {
			return NewParseError(
				ErrTreeScenarioStepObjectNotFound,
				fmt.Sprintf("scenario '%s' step at '%s' references to_object_key '%s' which does not exist",
					ctx.scenarioKey, path, *step.ToObjectKey),
				ctx.scenarioPath,
			).WithField(path + ".to_object_key").WithHint(fmt.Sprintf("available objects: %s", strings.Join(sortedKeys(ctx.scenario.Objects), ", ")))
		}
	}
	return nil
}

// validateStepEventRef validates event_key references in a step.
func validateStepEventRef(step inputStep, path string, ctx scenarioValidationContext) error {
	if step.EventKey == nil {
		return nil
	}
	if step.FromObjectKey == nil && step.ToObjectKey == nil {
		return NewParseError(
			ErrTreeScenarioStepEventNotFound,
			fmt.Sprintf("scenario '%s' step at '%s' references event '%s' but no object is specified to resolve the class",
				ctx.scenarioKey, path, *step.EventKey),
			ctx.scenarioPath,
		).WithField(path + ".event_key").WithHint(fmt.Sprintf("available objects: %s", strings.Join(sortedKeys(ctx.scenario.Objects), ", ")))
	}
	found := false
	var availableEvents []string
	for _, objKey := range []*string{step.ToObjectKey, step.FromObjectKey} {
		if objKey == nil {
			continue
		}
		obj := ctx.scenario.Objects[*objKey]
		class := ctx.model.Domains[ctx.domainKey].Subdomains[ctx.subdomainKey].Classes[obj.ClassKey]
		if class != nil && class.StateMachine != nil {
			if _, ok := class.StateMachine.Events[*step.EventKey]; ok {
				found = true
				break
			}
			availableEvents = append(availableEvents, sortedKeys(class.StateMachine.Events)...)
		}
	}
	if !found {
		classNames := collectStepClassNames(step, ctx.scenario)
		sort.Strings(availableEvents)
		return NewParseError(
			ErrTreeScenarioStepEventNotFound,
			fmt.Sprintf("scenario '%s' step at '%s' references event '%s' which does not exist on classes %v",
				ctx.scenarioKey, path, *step.EventKey, classNames),
			ctx.scenarioPath,
		).WithField(path + ".event_key").WithHint(fmt.Sprintf("available events: %s", strings.Join(availableEvents, ", ")))
	}
	return nil
}

// validateStepQueryRef validates query_key references in a step.
func validateStepQueryRef(step inputStep, path string, ctx scenarioValidationContext) error {
	if step.QueryKey == nil {
		return nil
	}
	if step.FromObjectKey == nil && step.ToObjectKey == nil {
		return NewParseError(
			ErrTreeScenarioStepQueryNotFound,
			fmt.Sprintf("scenario '%s' step at '%s' references query '%s' but no object is specified to resolve the class",
				ctx.scenarioKey, path, *step.QueryKey),
			ctx.scenarioPath,
		).WithField(path + ".query_key").WithHint(fmt.Sprintf("available objects: %s", strings.Join(sortedKeys(ctx.scenario.Objects), ", ")))
	}
	found := false
	var availableQueries []string
	for _, objKey := range []*string{step.ToObjectKey, step.FromObjectKey} {
		if objKey == nil {
			continue
		}
		obj := ctx.scenario.Objects[*objKey]
		class := ctx.model.Domains[ctx.domainKey].Subdomains[ctx.subdomainKey].Classes[obj.ClassKey]
		if class != nil {
			if _, ok := class.Queries[*step.QueryKey]; ok {
				found = true
				break
			}
			availableQueries = append(availableQueries, sortedKeys(class.Queries)...)
		}
	}
	if !found {
		classNames := collectStepClassNames(step, ctx.scenario)
		sort.Strings(availableQueries)
		return NewParseError(
			ErrTreeScenarioStepQueryNotFound,
			fmt.Sprintf("scenario '%s' step at '%s' references query '%s' which does not exist on classes %v",
				ctx.scenarioKey, path, *step.QueryKey, classNames),
			ctx.scenarioPath,
		).WithField(path + ".query_key").WithHint(fmt.Sprintf("available queries: %s", strings.Join(availableQueries, ", ")))
	}
	return nil
}

// collectStepClassNames collects class names from a step's object references.
func collectStepClassNames(step inputStep, scenario *inputScenario) []string {
	var classNames []string
	for _, objKey := range []*string{step.ToObjectKey, step.FromObjectKey} {
		if objKey != nil {
			classNames = append(classNames, scenario.Objects[*objKey].ClassKey)
		}
	}
	return classNames
}

// validateUseCaseGeneralizationTree validates a use case generalization's cross-references.
func validateUseCaseGeneralizationTree(subdomain *inputSubdomain, domainKey, subdomainKey, genKey string, gen *inputUseCaseGeneralization) error {
	genPath := fmt.Sprintf("domains/%s/subdomains/%s/use_case_generalizations/%s.ucgen.json", domainKey, subdomainKey, genKey)

	// Validate superclass_key exists
	if _, ok := subdomain.UseCases[gen.SuperclassKey]; !ok {
		return NewParseError(
			ErrTreeClassGenSuperclassNotFound, // reusing error code
			fmt.Sprintf("use case generalization '%s' superclass_key '%s' does not exist in subdomain '%s'",
				genKey, gen.SuperclassKey, subdomainKey),
			genPath,
		).WithField("superclass_key").WithHint(fmt.Sprintf("available use cases: %s", strings.Join(sortedKeys(subdomain.UseCases), ", ")))
	}

	// Validate subclass_keys exist and are unique
	seen := make(map[string]bool)
	for i, subclassKey := range gen.SubclassKeys {
		// Check for duplicates
		if seen[subclassKey] {
			return NewParseError(
				ErrTreeClassGenSubclassDuplicate, // reusing error code
				fmt.Sprintf("use case generalization '%s' has duplicate subclass_key '%s'", genKey, subclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("remove duplicate entries from subclass_keys")
		}
		seen[subclassKey] = true

		// Check that the subclass exists
		if _, ok := subdomain.UseCases[subclassKey]; !ok {
			return NewParseError(
				ErrTreeClassGenSubclassNotFound, // reusing error code
				fmt.Sprintf("use case generalization '%s' subclass_key '%s' does not exist in subdomain '%s'",
					genKey, subclassKey, subdomainKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint(fmt.Sprintf("available use cases: %s", strings.Join(sortedKeys(subdomain.UseCases), ", ")))
		}

		// Check that superclass is not also a subclass
		if subclassKey == gen.SuperclassKey {
			return NewParseError(
				ErrTreeClassGenSuperclassIsSubclass, // reusing error code
				fmt.Sprintf("use case generalization '%s' superclass '%s' cannot also be a subclass",
					genKey, gen.SuperclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i)).WithHint("superclass cannot also appear in subclass_keys")
		}
	}

	return nil
}

// validateClassNameUniqueness checks that action names, query names, state names,
// event names, and guard names are unique within a class.
func validateClassNameUniqueness(class *inputClass, domainKey, subdomainKey, classKey string) error {
	// Validate action name uniqueness.
	actionNames := make(map[string]string) // name -> first key
	for key, action := range class.Actions {
		if first, ok := actionNames[action.Name]; ok {
			actionPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/actions/%s.json", domainKey, subdomainKey, classKey, key)
			return NewParseError(
				ErrActionDuplicateName,
				fmt.Sprintf("class '%s' has duplicate action name '%s' — keys '%s' and '%s' both use this name", classKey, action.Name, first, key),
				actionPath,
			).WithField("name").WithHint("each action within a class must have a unique \"name\" value")
		}
		actionNames[action.Name] = key
	}

	// Validate query name uniqueness.
	queryNames := make(map[string]string) // name -> first key
	for key, query := range class.Queries {
		if first, ok := queryNames[query.Name]; ok {
			queryPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/queries/%s.json", domainKey, subdomainKey, classKey, key)
			return NewParseError(
				ErrQueryDuplicateName,
				fmt.Sprintf("class '%s' has duplicate query name '%s' — keys '%s' and '%s' both use this name", classKey, query.Name, first, key),
				queryPath,
			).WithField("name").WithHint("each query within a class must have a unique \"name\" value")
		}
		queryNames[query.Name] = key
	}

	// Validate state machine name uniqueness (states, events, guards).
	if class.StateMachine != nil {
		smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)

		// State name uniqueness.
		stateNames := make(map[string]string)
		for key, state := range class.StateMachine.States {
			if first, ok := stateNames[state.Name]; ok {
				return NewParseError(
					ErrStateDuplicateName,
					fmt.Sprintf("class '%s' has duplicate state name '%s' — keys '%s' and '%s' both use this name", classKey, state.Name, first, key),
					smPath,
				).WithField("states." + key + ".name").WithHint("each state within a state machine must have a unique \"name\" value")
			}
			stateNames[state.Name] = key
		}

		// Event name uniqueness.
		eventNames := make(map[string]string)
		for key, event := range class.StateMachine.Events {
			if first, ok := eventNames[event.Name]; ok {
				return NewParseError(
					ErrEventDuplicateName,
					fmt.Sprintf("class '%s' has duplicate event name '%s' — keys '%s' and '%s' both use this name", classKey, event.Name, first, key),
					smPath,
				).WithField("events." + key + ".name").WithHint("each event within a state machine must have a unique \"name\" value")
			}
			eventNames[event.Name] = key
		}

		// Guard name uniqueness.
		guardNames := make(map[string]string)
		for key, guard := range class.StateMachine.Guards {
			if first, ok := guardNames[guard.Name]; ok {
				return NewParseError(
					ErrGuardDuplicateName,
					fmt.Sprintf("class '%s' has duplicate guard name '%s' — keys '%s' and '%s' both use this name", classKey, guard.Name, first, key),
					smPath,
				).WithField("guards." + key + ".name").WithHint("each guard within a state machine must have a unique \"name\" value")
			}
			guardNames[guard.Name] = key
		}
	}

	return nil
}

// validateAttributeKeyNameConsistency checks that each attribute map key
// matches keyFromName(name).
func validateAttributeKeyNameConsistency(class *inputClass, domainKey, subdomainKey, classKey string) error {
	classPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKey, subdomainKey, classKey)
	for key, attr := range class.Attributes {
		if err := validateMapKeyMatchesName(key, attr.Name, "attribute", ErrClassAttrKeyNameMismatch, classPath); err != nil {
			return err
		}
	}
	return nil
}

// validateStateMachineKeyNameConsistency checks that each state machine map key
// matches keyFromName(name) for states, events, and guards.
func validateStateMachineKeyNameConsistency(class *inputClass, domainKey, subdomainKey, classKey string) error {
	if class.StateMachine == nil {
		return nil
	}
	smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)

	for key, state := range class.StateMachine.States {
		if expected := keyFromName(state.Name); key != expected {
			return NewParseError(
				ErrStateKeyNameMismatch,
				fmt.Sprintf("state key '%s' does not match name '%s' — expected key '%s'", key, state.Name, expected),
				smPath,
			).WithField("states." + key).WithHint(fmt.Sprintf("rename the key to '%s' or change the name to match the key", expected))
		}
	}
	for key, event := range class.StateMachine.Events {
		if expected := keyFromName(event.Name); key != expected {
			return NewParseError(
				ErrEventKeyNameMismatch,
				fmt.Sprintf("event key '%s' does not match name '%s' — expected key '%s'", key, event.Name, expected),
				smPath,
			).WithField("events." + key).WithHint(fmt.Sprintf("rename the key to '%s' or change the name to match the key", expected))
		}
	}
	for key, guard := range class.StateMachine.Guards {
		if expected := keyFromName(guard.Name); key != expected {
			return NewParseError(
				ErrGuardKeyNameMismatch,
				fmt.Sprintf("guard key '%s' does not match name '%s' — expected key '%s'", key, guard.Name, expected),
				smPath,
			).WithField("guards." + key).WithHint(fmt.Sprintf("rename the key to '%s' or change the name to match the key", expected))
		}
	}
	return nil
}
