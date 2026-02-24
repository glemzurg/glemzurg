package req_flat

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// Requirements provides flat lookups by key for all business logic objects in the model.
// This is the data structure for the view layer - templates use keys to look up objects.
type Requirements struct {
	Model req_model.Model

	// Flat lookups by key - populated from Model tree.
	Actors                 map[identity.Key]model_actor.Actor
	ActorGeneralizations   map[identity.Key]model_actor.Generalization
	Domains                map[identity.Key]model_domain.Domain
	Subdomains             map[identity.Key]model_domain.Subdomain
	DomainAssociations     map[identity.Key]model_domain.Association
	Generalizations        map[identity.Key]model_class.Generalization
	Classes                map[identity.Key]model_class.Class
	Attributes             map[identity.Key]model_class.Attribute
	ClassAssociations      map[identity.Key]model_class.Association
	Invariants             map[identity.Key]model_logic.Logic // Model-level invariants.
	ClassInvariants        map[identity.Key]model_logic.Logic // Class-level invariants (from all classes).
	States                 map[identity.Key]model_state.State
	Events                 map[identity.Key]model_state.Event
	Guards                 map[identity.Key]model_state.Guard
	Actions                map[identity.Key]model_state.Action
	Queries                map[identity.Key]model_state.Query
	Transitions            map[identity.Key]model_state.Transition
	StateActions           map[identity.Key]model_state.StateAction
	GlobalFunctions        map[identity.Key]model_logic.GlobalFunction
	UseCases               map[identity.Key]model_use_case.UseCase
	UseCaseGeneralizations map[identity.Key]model_use_case.Generalization
	Scenarios              map[identity.Key]model_scenario.Scenario
	Objects                map[identity.Key]model_scenario.Object

	// Prepared flag to avoid re-processing.
	prepared bool
}

// NewRequirements creates a Requirements from a Model, flattening the tree into lookups.
func NewRequirements(model req_model.Model) *Requirements {
	r := &Requirements{
		Model: model,
	}
	r.flattenModel()
	return r
}

// flattenModel walks the model tree and populates all flat lookup maps.
func (r *Requirements) flattenModel() {
	// Initialize all maps.
	r.Actors = make(map[identity.Key]model_actor.Actor)
	r.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization)
	r.Domains = make(map[identity.Key]model_domain.Domain)
	r.Subdomains = make(map[identity.Key]model_domain.Subdomain)
	r.DomainAssociations = make(map[identity.Key]model_domain.Association)
	r.Generalizations = make(map[identity.Key]model_class.Generalization)
	r.Classes = make(map[identity.Key]model_class.Class)
	r.Attributes = make(map[identity.Key]model_class.Attribute)
	r.ClassAssociations = make(map[identity.Key]model_class.Association)
	r.Invariants = make(map[identity.Key]model_logic.Logic)
	r.ClassInvariants = make(map[identity.Key]model_logic.Logic)
	r.States = make(map[identity.Key]model_state.State)
	r.Events = make(map[identity.Key]model_state.Event)
	r.Guards = make(map[identity.Key]model_state.Guard)
	r.Actions = make(map[identity.Key]model_state.Action)
	r.Queries = make(map[identity.Key]model_state.Query)
	r.Transitions = make(map[identity.Key]model_state.Transition)
	r.StateActions = make(map[identity.Key]model_state.StateAction)
	r.GlobalFunctions = make(map[identity.Key]model_logic.GlobalFunction)
	r.UseCases = make(map[identity.Key]model_use_case.UseCase)
	r.UseCaseGeneralizations = make(map[identity.Key]model_use_case.Generalization)
	r.Scenarios = make(map[identity.Key]model_scenario.Scenario)
	r.Objects = make(map[identity.Key]model_scenario.Object)

	// Actors from model.
	for key, actor := range r.Model.Actors {
		r.Actors[key] = actor
	}

	// Actor generalizations from model.
	for key, ag := range r.Model.ActorGeneralizations {
		r.ActorGeneralizations[key] = ag
	}

	// Global functions from model.
	for key, gf := range r.Model.GlobalFunctions {
		r.GlobalFunctions[key] = gf
	}

	// Domain associations from model.
	for key, assoc := range r.Model.DomainAssociations {
		r.DomainAssociations[key] = assoc
	}

	// Model-level class associations.
	for key, assoc := range r.Model.ClassAssociations {
		r.ClassAssociations[key] = assoc
	}

	// Model-level invariants (slice, not map).
	for _, inv := range r.Model.Invariants {
		r.Invariants[inv.Key] = inv
	}

	// Walk domains.
	for domainKey, domain := range r.Model.Domains {
		r.Domains[domainKey] = domain

		// Domain-level class associations.
		for key, assoc := range domain.ClassAssociations {
			r.ClassAssociations[key] = assoc
		}

		// Walk subdomains.
		for subdomainKey, subdomain := range domain.Subdomains {
			r.Subdomains[subdomainKey] = subdomain

			// Generalizations.
			for key, gen := range subdomain.Generalizations {
				r.Generalizations[key] = gen
			}

			// Use case generalizations.
			for key, ucGen := range subdomain.UseCaseGeneralizations {
				r.UseCaseGeneralizations[key] = ucGen
			}

			// Subdomain-level class associations.
			for key, assoc := range subdomain.ClassAssociations {
				r.ClassAssociations[key] = assoc
			}

			// Walk classes.
			for classKey, class := range subdomain.Classes {
				r.Classes[classKey] = class

				// Attributes.
				for key, attr := range class.Attributes {
					r.Attributes[key] = attr
				}

				// Class invariants (slice, not map).
				for _, inv := range class.Invariants {
					r.ClassInvariants[inv.Key] = inv
				}

				// States.
				for key, state := range class.States {
					r.States[key] = state

					// State actions (slice, not map).
					for _, stateAction := range state.Actions {
						r.StateActions[stateAction.Key] = stateAction
					}
				}

				// Events.
				for key, event := range class.Events {
					r.Events[key] = event
				}

				// Guards.
				for key, guard := range class.Guards {
					r.Guards[key] = guard
				}

				// Actions.
				for key, action := range class.Actions {
					r.Actions[key] = action
				}

				// Queries.
				for key, query := range class.Queries {
					r.Queries[key] = query
				}

				// Transitions.
				for key, transition := range class.Transitions {
					r.Transitions[key] = transition
				}
			}

			// Walk use cases.
			for useCaseKey, useCase := range subdomain.UseCases {
				r.UseCases[useCaseKey] = useCase

				// Walk scenarios.
				for scenarioKey, scenario := range useCase.Scenarios {
					r.Scenarios[scenarioKey] = scenario

					// Walk objects.
					for objectKey, object := range scenario.Objects {
						r.Objects[objectKey] = object
					}
				}
			}
		}
	}
}

// PrepLookups prepares any additional lookups needed for templates.
// This populates cross-references for scenario step references.
func (r *Requirements) PrepLookups() {
	if r.prepared {
		return
	}
	r.prepared = true
}

// ActorLookup returns actors by key (as string for template use).
func (r *Requirements) ActorLookup() map[string]model_actor.Actor {
	r.PrepLookups()
	lookup := make(map[string]model_actor.Actor)
	for key, actor := range r.Actors {
		lookup[key.String()] = actor
	}
	return lookup
}

// ActorClassesLookup returns a map of actor key to the classes that implement that actor.
// Classes reference actors via their ActorKey field.
func (r *Requirements) ActorClassesLookup() map[string][]model_class.Class {
	r.PrepLookups()
	lookup := make(map[string][]model_class.Class)

	// Initialize with empty slices for all actors.
	for actorKey := range r.Actors {
		lookup[actorKey.String()] = []model_class.Class{}
	}

	// Find all classes that reference an actor.
	for _, class := range r.Classes {
		if class.ActorKey != nil {
			actorKeyStr := class.ActorKey.String()
			lookup[actorKeyStr] = append(lookup[actorKeyStr], class)
		}
	}

	// Sort classes by key for consistent output.
	for actorKey := range lookup {
		classes := lookup[actorKey]
		sort.Slice(classes, func(i, j int) bool {
			return classes[i].Key.String() < classes[j].Key.String()
		})
		lookup[actorKey] = classes
	}

	return lookup
}

// ClassDomainLookup returns a map of class key to the domain that contains it.
// Classes are children of subdomains, which are children of domains.
func (r *Requirements) ClassDomainLookup() map[string]model_domain.Domain {
	r.PrepLookups()
	lookup := make(map[string]model_domain.Domain)

	// Walk the domain → subdomain → class hierarchy to build the mapping.
	for _, domain := range r.Model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey := range subdomain.Classes {
				lookup[classKey.String()] = domain
			}
		}
	}

	return lookup
}

// DomainUseCasesLookup returns a map of domain key to all use cases in that domain's subdomains.
func (r *Requirements) DomainUseCasesLookup() map[string][]model_use_case.UseCase {
	r.PrepLookups()
	lookup := make(map[string][]model_use_case.UseCase)

	// Walk the domain → subdomain → use case hierarchy.
	for domainKey, domain := range r.Model.Domains {
		var useCases []model_use_case.UseCase
		for _, subdomain := range domain.Subdomains {
			for _, useCase := range subdomain.UseCases {
				useCases = append(useCases, useCase)
			}
		}
		// Sort for consistent output.
		sort.Slice(useCases, func(i, j int) bool {
			return useCases[i].Key.String() < useCases[j].Key.String()
		})
		lookup[domainKey.String()] = useCases
	}

	return lookup
}

// DomainClassesLookup returns a map of domain key to all classes in that domain's subdomains.
func (r *Requirements) DomainClassesLookup() map[string][]model_class.Class {
	r.PrepLookups()
	lookup := make(map[string][]model_class.Class)

	// Walk the domain → subdomain → class hierarchy.
	for domainKey, domain := range r.Model.Domains {
		var classes []model_class.Class
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				classes = append(classes, class)
			}
		}
		// Sort for consistent output.
		sort.Slice(classes, func(i, j int) bool {
			return classes[i].Key.String() < classes[j].Key.String()
		})
		lookup[domainKey.String()] = classes
	}

	return lookup
}

// UseCaseDomainLookup returns a map of use case key to the domain that contains it.
// Use cases are children of subdomains, which are children of domains.
func (r *Requirements) UseCaseDomainLookup() map[string]model_domain.Domain {
	r.PrepLookups()
	lookup := make(map[string]model_domain.Domain)

	// Walk the domain → subdomain → use case hierarchy.
	for _, domain := range r.Model.Domains {
		for _, subdomain := range domain.Subdomains {
			for useCaseKey := range subdomain.UseCases {
				lookup[useCaseKey.String()] = domain
			}
		}
	}

	return lookup
}

// ClassSubdomainLookup returns a map of class key to the subdomain that contains it.
func (r *Requirements) ClassSubdomainLookup() map[string]model_domain.Subdomain {
	r.PrepLookups()
	lookup := make(map[string]model_domain.Subdomain)

	// Walk the domain → subdomain → class hierarchy to build the mapping.
	for _, domain := range r.Model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey := range subdomain.Classes {
				lookup[classKey.String()] = subdomain
			}
		}
	}

	return lookup
}

// UseCaseSubdomainLookup returns a map of use case key to the subdomain that contains it.
func (r *Requirements) UseCaseSubdomainLookup() map[string]model_domain.Subdomain {
	r.PrepLookups()
	lookup := make(map[string]model_domain.Subdomain)

	// Walk the domain → subdomain → use case hierarchy.
	for _, domain := range r.Model.Domains {
		for _, subdomain := range domain.Subdomains {
			for useCaseKey := range subdomain.UseCases {
				lookup[useCaseKey.String()] = subdomain
			}
		}
	}

	return lookup
}

// DomainHasMultipleSubdomains returns true if the domain has more than just the default subdomain.
func (r *Requirements) DomainHasMultipleSubdomains(domainKey identity.Key) bool {
	r.PrepLookups()
	domain, ok := r.Domains[domainKey]
	if !ok {
		return false
	}
	// Check if there's more than one subdomain, or if the only subdomain is not "default"
	if len(domain.Subdomains) > 1 {
		return true
	}
	// If there's exactly one subdomain, check if it's not the default
	for _, subdomain := range domain.Subdomains {
		if subdomain.Key.SubKey != "default" {
			return true
		}
	}
	return false
}

// DomainLookup returns domains by key and domain associations.
func (r *Requirements) DomainLookup() (map[string]model_domain.Domain, []model_domain.Association) {
	r.PrepLookups()
	lookup := make(map[string]model_domain.Domain)
	for key, domain := range r.Domains {
		lookup[key.String()] = domain
	}
	associations := make([]model_domain.Association, 0, len(r.DomainAssociations))
	for _, assoc := range r.DomainAssociations {
		associations = append(associations, assoc)
	}
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key.String() < associations[j].Key.String()
	})
	return lookup, associations
}

// GeneralizationLookup returns generalizations by key.
func (r *Requirements) GeneralizationLookup() map[string]model_class.Generalization {
	r.PrepLookups()
	lookup := make(map[string]model_class.Generalization)
	for key, gen := range r.Generalizations {
		lookup[key.String()] = gen
	}
	return lookup
}

// GeneralizationSuperclassLookup returns a map of generalization key to the superclass of that generalization.
// The superclass is the class that has SuperclassOfKey pointing to the generalization.
func (r *Requirements) GeneralizationSuperclassLookup() map[string]model_class.Class {
	r.PrepLookups()
	lookup := make(map[string]model_class.Class)

	for _, class := range r.Classes {
		if class.SuperclassOfKey != nil {
			lookup[class.SuperclassOfKey.String()] = class
		}
	}

	return lookup
}

// GeneralizationSubclassesLookup returns a map of generalization key to its subclasses.
// Subclasses are classes that have SubclassOfKey pointing to the generalization.
func (r *Requirements) GeneralizationSubclassesLookup() map[string][]model_class.Class {
	r.PrepLookups()
	lookup := make(map[string][]model_class.Class)

	// Initialize with empty slices for all generalizations.
	for genKey := range r.Generalizations {
		lookup[genKey.String()] = []model_class.Class{}
	}

	// Find all classes that are subclasses of a generalization.
	for _, class := range r.Classes {
		if class.SubclassOfKey != nil {
			genKeyStr := class.SubclassOfKey.String()
			lookup[genKeyStr] = append(lookup[genKeyStr], class)
		}
	}

	// Sort subclasses by key for consistent output.
	for genKey := range lookup {
		classes := lookup[genKey]
		sort.Slice(classes, func(i, j int) bool {
			return classes[i].Key.String() < classes[j].Key.String()
		})
		lookup[genKey] = classes
	}

	return lookup
}

// ClassLookup returns classes by key and class associations.
func (r *Requirements) ClassLookup() (map[string]model_class.Class, []model_class.Association) {
	r.PrepLookups()
	lookup := make(map[string]model_class.Class)
	for key, class := range r.Classes {
		lookup[key.String()] = class
	}
	associations := make([]model_class.Association, 0, len(r.ClassAssociations))
	for _, assoc := range r.ClassAssociations {
		associations = append(associations, assoc)
	}
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key.String() < associations[j].Key.String()
	})
	return lookup, associations
}

// StateLookup returns states by key.
func (r *Requirements) StateLookup() map[string]model_state.State {
	r.PrepLookups()
	lookup := make(map[string]model_state.State)
	for key, state := range r.States {
		lookup[key.String()] = state
	}
	return lookup
}

// EventLookup returns events by key.
func (r *Requirements) EventLookup() map[string]model_state.Event {
	r.PrepLookups()
	lookup := make(map[string]model_state.Event)
	for key, event := range r.Events {
		lookup[key.String()] = event
	}
	return lookup
}

// GuardLookup returns guards by key.
func (r *Requirements) GuardLookup() map[string]model_state.Guard {
	r.PrepLookups()
	lookup := make(map[string]model_state.Guard)
	for key, guard := range r.Guards {
		lookup[key.String()] = guard
	}
	return lookup
}

// ActionLookup returns actions by key.
func (r *Requirements) ActionLookup() map[string]model_state.Action {
	r.PrepLookups()
	lookup := make(map[string]model_state.Action)
	for key, action := range r.Actions {
		lookup[key.String()] = action
	}
	return lookup
}

// ActionTransitionsLookup returns a map of action key to the transitions that call that action.
func (r *Requirements) ActionTransitionsLookup() map[string][]model_state.Transition {
	r.PrepLookups()
	lookup := make(map[string][]model_state.Transition)

	// Initialize with empty slices for all actions.
	for actionKey := range r.Actions {
		lookup[actionKey.String()] = []model_state.Transition{}
	}

	// Find all transitions that call each action.
	for _, transition := range r.Transitions {
		if transition.ActionKey != nil {
			actionKeyStr := transition.ActionKey.String()
			lookup[actionKeyStr] = append(lookup[actionKeyStr], transition)
		}
	}

	// Sort transitions by key for consistent output.
	for actionKey := range lookup {
		transitions := lookup[actionKey]
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].Key.String() < transitions[j].Key.String()
		})
		lookup[actionKey] = transitions
	}

	return lookup
}

// ActionStateActionsLookup returns a map of action key to the state actions that call that action.
func (r *Requirements) ActionStateActionsLookup() map[string][]model_state.StateAction {
	r.PrepLookups()
	lookup := make(map[string][]model_state.StateAction)

	// Initialize with empty slices for all actions.
	for actionKey := range r.Actions {
		lookup[actionKey.String()] = []model_state.StateAction{}
	}

	// Find all state actions that call each action.
	for _, stateAction := range r.StateActions {
		actionKeyStr := stateAction.ActionKey.String()
		lookup[actionKeyStr] = append(lookup[actionKeyStr], stateAction)
	}

	// Sort state actions by key for consistent output.
	for actionKey := range lookup {
		stateActions := lookup[actionKey]
		sort.Slice(stateActions, func(i, j int) bool {
			return stateActions[i].Key.String() < stateActions[j].Key.String()
		})
		lookup[actionKey] = stateActions
	}

	return lookup
}

// QueryLookup returns queries by key.
func (r *Requirements) QueryLookup() map[string]model_state.Query {
	r.PrepLookups()
	lookup := make(map[string]model_state.Query)
	for key, query := range r.Queries {
		lookup[key.String()] = query
	}
	return lookup
}

// GlobalFunctionLookup returns global functions by key.
func (r *Requirements) GlobalFunctionLookup() map[string]model_logic.GlobalFunction {
	r.PrepLookups()
	lookup := make(map[string]model_logic.GlobalFunction)
	for key, gf := range r.GlobalFunctions {
		lookup[key.String()] = gf
	}
	return lookup
}

// InvariantLookup returns model-level invariants by key (as string for template use).
func (r *Requirements) InvariantLookup() map[string]model_logic.Logic {
	r.PrepLookups()
	lookup := make(map[string]model_logic.Logic)
	for key, inv := range r.Invariants {
		lookup[key.String()] = inv
	}
	return lookup
}

// ClassInvariantLookup returns class-level invariants by key (as string for template use).
func (r *Requirements) ClassInvariantLookup() map[string]model_logic.Logic {
	r.PrepLookups()
	lookup := make(map[string]model_logic.Logic)
	for key, inv := range r.ClassInvariants {
		lookup[key.String()] = inv
	}
	return lookup
}

// ActorGeneralizationLookup returns actor generalizations by key.
func (r *Requirements) ActorGeneralizationLookup() map[string]model_actor.Generalization {
	r.PrepLookups()
	lookup := make(map[string]model_actor.Generalization)
	for key, ag := range r.ActorGeneralizations {
		lookup[key.String()] = ag
	}
	return lookup
}

// ActorGeneralizationSuperclassLookup returns a map of actor generalization key to the superclass actor.
// The superclass is the actor that has SuperclassOfKey pointing to the generalization.
func (r *Requirements) ActorGeneralizationSuperclassLookup() map[string]model_actor.Actor {
	r.PrepLookups()
	lookup := make(map[string]model_actor.Actor)

	for _, actor := range r.Actors {
		if actor.SuperclassOfKey != nil {
			lookup[actor.SuperclassOfKey.String()] = actor
		}
	}

	return lookup
}

// ActorGeneralizationSubclassesLookup returns a map of actor generalization key to its subclass actors.
// Subclasses are actors that have SubclassOfKey pointing to the generalization.
func (r *Requirements) ActorGeneralizationSubclassesLookup() map[string][]model_actor.Actor {
	r.PrepLookups()
	lookup := make(map[string][]model_actor.Actor)

	// Initialize with empty slices for all actor generalizations.
	for agKey := range r.ActorGeneralizations {
		lookup[agKey.String()] = []model_actor.Actor{}
	}

	// Find all actors that are subclasses of a generalization.
	for _, actor := range r.Actors {
		if actor.SubclassOfKey != nil {
			agKeyStr := actor.SubclassOfKey.String()
			lookup[agKeyStr] = append(lookup[agKeyStr], actor)
		}
	}

	// Sort subclasses by key for consistent output.
	for agKey := range lookup {
		actors := lookup[agKey]
		sort.Slice(actors, func(i, j int) bool {
			return actors[i].Key.String() < actors[j].Key.String()
		})
		lookup[agKey] = actors
	}

	return lookup
}

// UseCaseGeneralizationLookup returns use case generalizations by key.
func (r *Requirements) UseCaseGeneralizationLookup() map[string]model_use_case.Generalization {
	r.PrepLookups()
	lookup := make(map[string]model_use_case.Generalization)
	for key, ucGen := range r.UseCaseGeneralizations {
		lookup[key.String()] = ucGen
	}
	return lookup
}

// UseCaseGeneralizationSuperclassLookup returns a map of use case generalization key to the superclass use case.
// The superclass is the use case that has SuperclassOfKey pointing to the generalization.
func (r *Requirements) UseCaseGeneralizationSuperclassLookup() map[string]model_use_case.UseCase {
	r.PrepLookups()
	lookup := make(map[string]model_use_case.UseCase)

	for _, useCase := range r.UseCases {
		if useCase.SuperclassOfKey != nil {
			lookup[useCase.SuperclassOfKey.String()] = useCase
		}
	}

	return lookup
}

// UseCaseGeneralizationSubclassesLookup returns a map of use case generalization key to its subclass use cases.
// Subclasses are use cases that have SubclassOfKey pointing to the generalization.
func (r *Requirements) UseCaseGeneralizationSubclassesLookup() map[string][]model_use_case.UseCase {
	r.PrepLookups()
	lookup := make(map[string][]model_use_case.UseCase)

	// Initialize with empty slices for all use case generalizations.
	for ucGenKey := range r.UseCaseGeneralizations {
		lookup[ucGenKey.String()] = []model_use_case.UseCase{}
	}

	// Find all use cases that are subclasses of a generalization.
	for _, useCase := range r.UseCases {
		if useCase.SubclassOfKey != nil {
			ucGenKeyStr := useCase.SubclassOfKey.String()
			lookup[ucGenKeyStr] = append(lookup[ucGenKeyStr], useCase)
		}
	}

	// Sort subclasses by key for consistent output.
	for ucGenKey := range lookup {
		useCases := lookup[ucGenKey]
		sort.Slice(useCases, func(i, j int) bool {
			return useCases[i].Key.String() < useCases[j].Key.String()
		})
		lookup[ucGenKey] = useCases
	}

	return lookup
}

// UseCaseLookup returns use cases by key.
func (r *Requirements) UseCaseLookup() map[string]model_use_case.UseCase {
	r.PrepLookups()
	lookup := make(map[string]model_use_case.UseCase)
	for key, useCase := range r.UseCases {
		lookup[key.String()] = useCase
	}
	return lookup
}

// ScenarioLookup returns scenarios by key.
func (r *Requirements) ScenarioLookup() map[string]model_scenario.Scenario {
	r.PrepLookups()
	lookup := make(map[string]model_scenario.Scenario)
	for key, scenario := range r.Scenarios {
		lookup[key.String()] = scenario
	}
	return lookup
}

// ObjectLookup returns scenario objects by key.
func (r *Requirements) ObjectLookup() map[string]model_scenario.Object {
	r.PrepLookups()
	lookup := make(map[string]model_scenario.Object)
	for key, object := range r.Objects {
		lookup[key.String()] = object
	}
	return lookup
}

// RegardingClasses returns all objects connected to the given classes for UML diagrams.
func (r *Requirements) RegardingClasses(inClasses []model_class.Class) (generalizations []model_class.Generalization, classes []model_class.Class, associations []model_class.Association) {
	allClassLookup, allAssociations := r.ClassLookup()
	allGeneralizationLookup := r.GeneralizationLookup()

	// Create a class lookup.
	relevantClassLookup := classesAsLookup(allClassLookup, inClasses)

	// Create an association lookup.
	relevantAssociationsLookup := classAssociationsAsLookup(relevantClassLookup, allAssociations)

	// Find relevant associations to the the classes.
	classesFromAssocaitionsLookup := classesFromAssociations(allClassLookup, relevantAssociationsLookup)
	relevantClassLookup = mergeClassLookups(relevantClassLookup, classesFromAssocaitionsLookup)

	// Create generalization lookup.
	relevantGeneralizationLookup := classGeneralizationsAsLookup(relevantClassLookup, allGeneralizationLookup)

	// Include any class related to those generalizations.
	classesFromGeneralizationsLookup := classesFromGeneralizations(allClassLookup, relevantGeneralizationLookup)
	relevantClassLookup = mergeClassLookups(relevantClassLookup, classesFromGeneralizationsLookup)

	// Get all that we want to return.
	for _, generalization := range relevantGeneralizationLookup {
		generalizations = append(generalizations, generalization)
	}
	for _, class := range relevantClassLookup {
		// Only include classes *not* in a generalization.
		// The classes in a generalization will be drawn by the generalization code.
		if class.SuperclassOfKey == nil && class.SubclassOfKey == nil {
			classes = append(classes, class)
		}
	}
	for _, association := range relevantAssociationsLookup {
		associations = append(associations, association)
	}

	// Sort everything.
	sort.Slice(generalizations, func(i, j int) bool {
		return generalizations[i].Key.String() < generalizations[j].Key.String()
	})
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Key.String() < classes[j].Key.String()
	})
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key.String() < associations[j].Key.String()
	})

	return generalizations, classes, associations
}

// RegardingUseCases returns all actors connected to the given use cases for UML diagrams.
func (r *Requirements) RegardingUseCases(inUseCases []model_use_case.UseCase) (useCases []model_use_case.UseCase, actors []model_actor.Actor, err error) {
	actorLookup := r.ActorLookup()
	useCaseLookup := r.UseCaseLookup()
	classLookup, _ := r.ClassLookup()

	// Get the use cases that are fully loaded with data.
	for _, useCase := range inUseCases {
		populatedUseCase, found := useCaseLookup[useCase.Key.String()]
		if !found {
			return nil, nil, errors.New("use case not found in lookup: " + useCase.Key.String())
		}
		useCases = append(useCases, populatedUseCase)
	}

	// Collect unique actors.
	// UseCase.Actors is keyed by class key (the class that implements the actor).
	// We need to look up the class to get its ActorKey, then look up the actual actor.
	uniqueActors := map[string]model_actor.Actor{}
	for _, useCase := range useCases {
		for actorClassKey := range useCase.Actors {
			// Look up the class that implements this actor.
			class, found := classLookup[actorClassKey.String()]
			if !found {
				return nil, nil, errors.New("actor class not found in lookup: " + actorClassKey.String())
			}
			// The class should have an ActorKey pointing to the actual actor.
			if class.ActorKey == nil {
				return nil, nil, errors.New("class does not have an ActorKey: " + actorClassKey.String())
			}
			// Look up the actual actor.
			actor, found := actorLookup[class.ActorKey.String()]
			if !found {
				return nil, nil, errors.New("actor not found in lookup: " + class.ActorKey.String())
			}
			uniqueActors[class.ActorKey.String()] = actor
		}
	}

	// Convert to slice.
	for _, actor := range uniqueActors {
		actors = append(actors, actor)
	}

	// Sort.
	sort.Slice(useCases, func(i, j int) bool {
		return useCases[i].Key.String() < useCases[j].Key.String()
	})
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Key.String() < actors[j].Key.String()
	})

	return useCases, actors, nil
}

// Helper functions for RegardingClasses.

func mergeClassLookups(lookupA, lookupB map[string]model_class.Class) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	for _, class := range lookupA {
		lookup[class.Key.String()] = class
	}
	for _, class := range lookupB {
		lookup[class.Key.String()] = class
	}
	return lookup
}

func classesAsLookup(allClassLookup map[string]model_class.Class, classes []model_class.Class) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	for _, class := range classes {
		lookup[class.Key.String()] = allClassLookup[class.Key.String()]
	}
	return lookup
}

func classGeneralizationsAsLookup(classLookup map[string]model_class.Class, generalizationLookup map[string]model_class.Generalization) (lookup map[string]model_class.Generalization) {
	lookup = map[string]model_class.Generalization{}
	for _, class := range classLookup {
		for _, generalization := range generalizationLookup {
			if class.SuperclassOfKey != nil && *class.SuperclassOfKey == generalization.Key {
				lookup[generalization.Key.String()] = generalization
			}
			if class.SubclassOfKey != nil && *class.SubclassOfKey == generalization.Key {
				lookup[generalization.Key.String()] = generalization
			}
		}
	}
	return lookup
}

func classAssociationsAsLookup(classLookup map[string]model_class.Class, associations []model_class.Association) (lookup map[string]model_class.Association) {
	lookup = map[string]model_class.Association{}
	for _, class := range classLookup {
		for _, association := range associations {
			if association.Includes(class.Key) {
				lookup[association.Key.String()] = association
			}
		}
	}
	return lookup
}

func classesFromGeneralizations(allClassLookup map[string]model_class.Class, generalizationLookup map[string]model_class.Generalization) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	// Find classes that reference these generalizations via SuperclassOfKey or SubclassOfKey.
	for _, class := range allClassLookup {
		if class.SuperclassOfKey != nil {
			if _, inGen := generalizationLookup[class.SuperclassOfKey.String()]; inGen {
				lookup[class.Key.String()] = class
			}
		}
		if class.SubclassOfKey != nil {
			if _, inGen := generalizationLookup[class.SubclassOfKey.String()]; inGen {
				lookup[class.Key.String()] = class
			}
		}
	}
	return lookup
}

func classesFromAssociations(allClassLookup map[string]model_class.Class, associationLookup map[string]model_class.Association) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	for _, association := range associationLookup {
		lookup[association.FromClassKey.String()] = allClassLookup[association.FromClassKey.String()]
		lookup[association.ToClassKey.String()] = allClassLookup[association.ToClassKey.String()]
		if association.AssociationClassKey != nil {
			lookup[association.AssociationClassKey.String()] = allClassLookup[association.AssociationClassKey.String()]
		}
	}
	return lookup
}
