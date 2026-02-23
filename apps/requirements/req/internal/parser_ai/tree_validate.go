package parser_ai

import (
	"fmt"
	"regexp"
	"strings"
)

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
	for assocKey, assoc := range model.Associations {
		if err := validateModelAssociation(model, assocKey, assoc); err != nil {
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
		).WithField("actors")
	}

	// Check model has at least one domain
	if len(model.Domains) == 0 {
		return NewParseError(
			ErrTreeModelNoDomains,
			"model must have at least one domain defined - domains are high-level subject areas that group related functionality; create a domain directory under 'domains/' with a 'domain.json' file",
			"model.json",
		).WithField("domains")
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
		).WithField("subdomains")
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
// - If there are multiple subdomains, none can be named "default"
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
			).WithField("subdomain_key")
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
			).WithField("subdomain_key")
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
		).WithField("classes")
	}

	// Check subdomain has at least one association
	if len(subdomain.Associations) == 0 {
		return NewParseError(
			ErrTreeSubdomainNoAssociations,
			fmt.Sprintf("subdomain '%s' must have at least one association defined - associations describe how classes relate to each other; create association files under 'domains/%s/subdomains/%s/associations/' with '.assoc.json' extension",
				subdomainKey, domainKey, subdomainKey),
			subdomainPath,
		).WithField("associations")
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
		).WithField("attributes")
	}

	// Check class has a state machine
	if class.StateMachine == nil {
		return NewParseError(
			ErrTreeClassNoStateMachine,
			fmt.Sprintf("class '%s' must have a state machine defined - state machines describe the lifecycle and behavior of a class; create a 'state_machine.json' file in the class directory with states, events, and transitions",
				classKey),
			classPath,
		).WithField("state_machine")
	}

	// Check state machine has at least one transition
	if len(class.StateMachine.Transitions) == 0 {
		smPath := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKey, subdomainKey, classKey)
		return NewParseError(
			ErrTreeStateMachineNoTransitions,
			fmt.Sprintf("state machine for class '%s' must have at least one transition defined - transitions describe how the class moves between states in response to events; add transitions to the 'transitions' array with event_key and state references",
				classKey),
			smPath,
		).WithField("transitions")
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
	for assocKey, assoc := range domain.Associations {
		if err := validateDomainAssociation(model, domainKey, domain, assocKey, assoc); err != nil {
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
	for genKey, gen := range subdomain.Generalizations {
		if err := validateClassGeneralizationTree(subdomain, domainKey, subdomainKey, genKey, gen); err != nil {
			return err
		}
	}

	// Validate subdomain-level associations
	for assocKey, assoc := range subdomain.Associations {
		if err := validateSubdomainAssociation(subdomain, domainKey, subdomainKey, assocKey, assoc); err != nil {
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
			).WithField("actor_key")
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
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j))
			}
			seen[attrKey] = true

			// Check that the attribute exists
			if _, ok := class.Attributes[attrKey]; !ok {
				return NewParseError(
					ErrTreeClassIndexAttrNotFound,
					fmt.Sprintf("class '%s' index[%d] references attribute '%s' which does not exist", classKey, i, attrKey),
					classPath,
				).WithField(fmt.Sprintf("indexes[%d][%d]", i, j))
			}
		}
	}

	// Validate state machine if present
	if class.StateMachine != nil {
		if err := validateStateMachineTree(class, domainKey, subdomainKey, classKey); err != nil {
			return err
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
				).WithField(fmt.Sprintf("states.%s.actions[%d].action_key", stateKey, i))
			}
		}
	}

	// Validate transitions
	for i, transition := range sm.Transitions {
		// Check that at least one state is specified
		if transition.FromStateKey == nil && transition.ToStateKey == nil {
			return NewParseError(
				ErrTreeTransitionNoStates,
				fmt.Sprintf("transition[%d] must have at least one of from_state_key or to_state_key", i),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d]", i))
		}

		// Check that it's not both initial and final (from=nil and to=nil is already caught above)
		// Initial: from_state_key is nil, to_state_key is set
		// Final: from_state_key is set, to_state_key is nil
		// This check is redundant given the above, but kept for clarity

		// Validate from_state_key if present
		if transition.FromStateKey != nil {
			if _, ok := sm.States[*transition.FromStateKey]; !ok {
				return NewParseError(
					ErrTreeStateMachineStateNotFound,
					fmt.Sprintf("transition[%d] from_state_key '%s' does not exist", i, *transition.FromStateKey),
					smPath,
				).WithField(fmt.Sprintf("transitions[%d].from_state_key", i))
			}
		}

		// Validate to_state_key if present
		if transition.ToStateKey != nil {
			if _, ok := sm.States[*transition.ToStateKey]; !ok {
				return NewParseError(
					ErrTreeStateMachineStateNotFound,
					fmt.Sprintf("transition[%d] to_state_key '%s' does not exist", i, *transition.ToStateKey),
					smPath,
				).WithField(fmt.Sprintf("transitions[%d].to_state_key", i))
			}
		}

		// Validate event_key
		if _, ok := sm.Events[transition.EventKey]; !ok {
			return NewParseError(
				ErrTreeStateMachineEventNotFound,
				fmt.Sprintf("transition[%d] event_key '%s' does not exist", i, transition.EventKey),
				smPath,
			).WithField(fmt.Sprintf("transitions[%d].event_key", i))
		}

		// Validate guard_key if present
		if transition.GuardKey != nil {
			if _, ok := sm.Guards[*transition.GuardKey]; !ok {
				return NewParseError(
					ErrTreeStateMachineGuardNotFound,
					fmt.Sprintf("transition[%d] guard_key '%s' does not exist", i, *transition.GuardKey),
					smPath,
				).WithField(fmt.Sprintf("transitions[%d].guard_key", i))
			}
		}

		// Validate action_key if present
		if transition.ActionKey != nil {
			if _, ok := class.Actions[*transition.ActionKey]; !ok {
				return NewParseError(
					ErrTreeStateMachineActionNotFound,
					fmt.Sprintf("transition[%d] action_key '%s' does not exist in class '%s'", i, *transition.ActionKey, classKey),
					smPath,
				).WithField(fmt.Sprintf("transitions[%d].action_key", i))
			}
		}
	}

	// Validate that every action is referenced by at least one state action or transition
	if err := validateActionsReferenced(class, domainKey, subdomainKey, classKey); err != nil {
		return err
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
			).WithField("action_key")
		}
	}

	return nil
}

// validateClassGeneralizationTree validates a class generalization's cross-references.
func validateClassGeneralizationTree(subdomain *inputSubdomain, domainKey, subdomainKey, genKey string, gen *inputClassGeneralization) error {
	genPath := fmt.Sprintf("domains/%s/subdomains/%s/generalizations/%s.gen.json", domainKey, subdomainKey, genKey)

	// Validate superclass_key exists
	if _, ok := subdomain.Classes[gen.SuperclassKey]; !ok {
		return NewParseError(
			ErrTreeClassGenSuperclassNotFound,
			fmt.Sprintf("class generalization '%s' superclass_key '%s' does not exist in subdomain '%s'",
				genKey, gen.SuperclassKey, subdomainKey),
			genPath,
		).WithField("superclass_key")
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
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
		seen[subclassKey] = true

		// Check that the subclass exists
		if _, ok := subdomain.Classes[subclassKey]; !ok {
			return NewParseError(
				ErrTreeClassGenSubclassNotFound,
				fmt.Sprintf("class generalization '%s' subclass_key '%s' does not exist in subdomain '%s'",
					genKey, subclassKey, subdomainKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}

		// Check that superclass is not also a subclass
		if subclassKey == gen.SuperclassKey {
			return NewParseError(
				ErrTreeClassGenSuperclassIsSubclass,
				fmt.Sprintf("class generalization '%s' superclass '%s' cannot also be a subclass",
					genKey, gen.SuperclassKey),
				genPath,
			).WithField(fmt.Sprintf("subclass_keys[%d]", i))
		}
	}

	return nil
}

// validateSubdomainAssociation validates an association at the subdomain level.
// Keys are scoped to the subdomain (just class names).
func validateSubdomainAssociation(subdomain *inputSubdomain, domainKey, subdomainKey, assocKey string, assoc *inputAssociation) error {
	assocPath := fmt.Sprintf("domains/%s/subdomains/%s/associations/%s.assoc.json", domainKey, subdomainKey, assocKey)

	// Validate from_class_key
	if _, ok := subdomain.Classes[assoc.FromClassKey]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' does not exist in subdomain '%s'",
				assocKey, assoc.FromClassKey, subdomainKey),
			assocPath,
		).WithField("from_class_key")
	}

	// Validate to_class_key
	if _, ok := subdomain.Classes[assoc.ToClassKey]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' does not exist in subdomain '%s'",
				assocKey, assoc.ToClassKey, subdomainKey),
			assocPath,
		).WithField("to_class_key")
	}

	// Validate association_class_key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		if _, ok := subdomain.Classes[*assoc.AssociationClassKey]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' does not exist in subdomain '%s'",
					assocKey, *assoc.AssociationClassKey, subdomainKey),
				assocPath,
			).WithField("association_class_key")
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity")
	}

	return nil
}

// validateDomainAssociation validates an association at the domain level.
// Keys include subdomain to disambiguate (subdomain/class).
func validateDomainAssociation(model *inputModel, domainKey string, domain *inputDomain, assocKey string, assoc *inputAssociation) error {
	assocPath := fmt.Sprintf("domains/%s/associations/%s.assoc.json", domainKey, assocKey)

	// Parse from_class_key (subdomain/class format)
	fromSubdomain, fromClass, err := parseDomainScopedKey(assoc.FromClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' is invalid: %s",
				assocKey, assoc.FromClassKey, err.Error()),
			assocPath,
		).WithField("from_class_key")
	}

	// Check from subdomain exists
	subdomain, ok := domain.Subdomains[fromSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references subdomain '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromSubdomain),
			assocPath,
		).WithField("from_class_key")
	}

	// Check from class exists
	if _, ok := subdomain.Classes[fromClass]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references class '%s' which does not exist in subdomain '%s'",
				assocKey, assoc.FromClassKey, fromClass, fromSubdomain),
			assocPath,
		).WithField("from_class_key")
	}

	// Parse to_class_key (subdomain/class format)
	toSubdomain, toClass, err := parseDomainScopedKey(assoc.ToClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' is invalid: %s",
				assocKey, assoc.ToClassKey, err.Error()),
			assocPath,
		).WithField("to_class_key")
	}

	// Check to subdomain exists
	subdomain, ok = domain.Subdomains[toSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references subdomain '%s' which does not exist",
				assocKey, assoc.ToClassKey, toSubdomain),
			assocPath,
		).WithField("to_class_key")
	}

	// Check to class exists
	if _, ok := subdomain.Classes[toClass]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references class '%s' which does not exist in subdomain '%s'",
				assocKey, assoc.ToClassKey, toClass, toSubdomain),
			assocPath,
		).WithField("to_class_key")
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
			).WithField("association_class_key")
		}

		subdomain, ok := domain.Subdomains[assocSubdomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references subdomain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocSubdomain),
				assocPath,
			).WithField("association_class_key")
		}

		if _, ok := subdomain.Classes[assocClass]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references class '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocClass),
				assocPath,
			).WithField("association_class_key")
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity")
	}

	return nil
}

// validateModelAssociation validates an association at the model level.
// Keys include domain and subdomain (domain/subdomain/class).
func validateModelAssociation(model *inputModel, assocKey string, assoc *inputAssociation) error {
	assocPath := fmt.Sprintf("associations/%s.assoc.json", assocKey)

	// Parse from_class_key (domain/subdomain/class format)
	fromDomain, fromSubdomain, fromClass, err := parseModelScopedKey(assoc.FromClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' is invalid: %s",
				assocKey, assoc.FromClassKey, err.Error()),
			assocPath,
		).WithField("from_class_key")
	}

	// Check from domain exists
	domain, ok := model.Domains[fromDomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references domain '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromDomain),
			assocPath,
		).WithField("from_class_key")
	}

	// Check from subdomain exists
	subdomain, ok := domain.Subdomains[fromSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references subdomain '%s' which does not exist in domain '%s'",
				assocKey, assoc.FromClassKey, fromSubdomain, fromDomain),
			assocPath,
		).WithField("from_class_key")
	}

	// Check from class exists
	if _, ok := subdomain.Classes[fromClass]; !ok {
		return NewParseError(
			ErrTreeAssocFromClassNotFound,
			fmt.Sprintf("association '%s' from_class_key '%s' references class '%s' which does not exist",
				assocKey, assoc.FromClassKey, fromClass),
			assocPath,
		).WithField("from_class_key")
	}

	// Parse to_class_key (domain/subdomain/class format)
	toDomain, toSubdomain, toClass, err := parseModelScopedKey(assoc.ToClassKey)
	if err != nil {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' is invalid: %s",
				assocKey, assoc.ToClassKey, err.Error()),
			assocPath,
		).WithField("to_class_key")
	}

	// Check to domain exists
	domain, ok = model.Domains[toDomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references domain '%s' which does not exist",
				assocKey, assoc.ToClassKey, toDomain),
			assocPath,
		).WithField("to_class_key")
	}

	// Check to subdomain exists
	subdomain, ok = domain.Subdomains[toSubdomain]
	if !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references subdomain '%s' which does not exist in domain '%s'",
				assocKey, assoc.ToClassKey, toSubdomain, toDomain),
			assocPath,
		).WithField("to_class_key")
	}

	// Check to class exists
	if _, ok := subdomain.Classes[toClass]; !ok {
		return NewParseError(
			ErrTreeAssocToClassNotFound,
			fmt.Sprintf("association '%s' to_class_key '%s' references class '%s' which does not exist",
				assocKey, assoc.ToClassKey, toClass),
			assocPath,
		).WithField("to_class_key")
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
			).WithField("association_class_key")
		}

		domain, ok := model.Domains[assocDomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references domain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocDomain),
				assocPath,
			).WithField("association_class_key")
		}

		subdomain, ok := domain.Subdomains[assocSubdomain]
		if !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references subdomain '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocSubdomain),
				assocPath,
			).WithField("association_class_key")
		}

		if _, ok := subdomain.Classes[assocClass]; !ok {
			return NewParseError(
				ErrTreeAssocClassNotFound,
				fmt.Sprintf("association '%s' association_class_key '%s' references class '%s' which does not exist",
					assocKey, *assoc.AssociationClassKey, assocClass),
				assocPath,
			).WithField("association_class_key")
		}
		// Association class cannot be the same as from or to class
		if *assoc.AssociationClassKey == assoc.FromClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as from_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
		if *assoc.AssociationClassKey == assoc.ToClassKey {
			return NewParseError(
				ErrTreeAssocClassSameAsEndpoint,
				fmt.Sprintf("association '%s' association_class_key '%s' cannot be the same as to_class_key",
					assocKey, *assoc.AssociationClassKey),
				assocPath,
			).WithField("association_class_key")
		}
	}

	// Validate multiplicity formats
	if err := validateMultiplicity(assoc.FromMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' from_multiplicity '%s' is invalid: %s",
				assocKey, assoc.FromMultiplicity, err.Error()),
			assocPath,
		).WithField("from_multiplicity")
	}

	if err := validateMultiplicity(assoc.ToMultiplicity); err != nil {
		return NewParseError(
			ErrTreeAssocMultiplicityInvalid,
			fmt.Sprintf("association '%s' to_multiplicity '%s' is invalid: %s",
				assocKey, assoc.ToMultiplicity, err.Error()),
			assocPath,
		).WithField("to_multiplicity")
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
			fmt.Sscanf(lower, "%d", &lowerNum)
			fmt.Sscanf(upper, "%d", &upperNum)
			if upperNum < lowerNum {
				return fmt.Errorf("upper bound %d cannot be less than lower bound %d", upperNum, lowerNum)
			}
		}
	}

	return nil
}
