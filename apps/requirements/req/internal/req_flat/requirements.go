package req_flat

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// Requirements provides flat lookups by key for all business logic objects in the model.
// This is the data structure for the view layer - templates use keys to look up objects.
type Requirements struct {
	Model req_model.Model

	// Flat lookups by key - populated from Model tree.
	Actors             map[identity.Key]model_actor.Actor
	Domains            map[identity.Key]model_domain.Domain
	Subdomains         map[identity.Key]model_domain.Subdomain
	DomainAssociations map[identity.Key]model_domain.Association
	Generalizations    map[identity.Key]model_class.Generalization
	Classes            map[identity.Key]model_class.Class
	Attributes         map[identity.Key]model_class.Attribute
	ClassAssociations  map[identity.Key]model_class.Association
	States             map[identity.Key]model_state.State
	Events             map[identity.Key]model_state.Event
	Guards             map[identity.Key]model_state.Guard
	Actions            map[identity.Key]model_state.Action
	Transitions        map[identity.Key]model_state.Transition
	StateActions       map[identity.Key]model_state.StateAction
	UseCases           map[identity.Key]model_use_case.UseCase
	Scenarios          map[identity.Key]model_scenario.Scenario
	Objects            map[identity.Key]model_scenario.Object

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
	r.Domains = make(map[identity.Key]model_domain.Domain)
	r.Subdomains = make(map[identity.Key]model_domain.Subdomain)
	r.DomainAssociations = make(map[identity.Key]model_domain.Association)
	r.Generalizations = make(map[identity.Key]model_class.Generalization)
	r.Classes = make(map[identity.Key]model_class.Class)
	r.Attributes = make(map[identity.Key]model_class.Attribute)
	r.ClassAssociations = make(map[identity.Key]model_class.Association)
	r.States = make(map[identity.Key]model_state.State)
	r.Events = make(map[identity.Key]model_state.Event)
	r.Guards = make(map[identity.Key]model_state.Guard)
	r.Actions = make(map[identity.Key]model_state.Action)
	r.Transitions = make(map[identity.Key]model_state.Transition)
	r.StateActions = make(map[identity.Key]model_state.StateAction)
	r.UseCases = make(map[identity.Key]model_use_case.UseCase)
	r.Scenarios = make(map[identity.Key]model_scenario.Scenario)
	r.Objects = make(map[identity.Key]model_scenario.Object)

	// Actors from model.
	for key, actor := range r.Model.Actors {
		r.Actors[key] = actor
	}

	// Domain associations from model.
	for key, assoc := range r.Model.DomainAssociations {
		r.DomainAssociations[key] = assoc
	}

	// Model-level class associations.
	for key, assoc := range r.Model.ClassAssociations {
		r.ClassAssociations[key] = assoc
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

	// Get the use cases that are fully loaded with data.
	for _, useCase := range inUseCases {
		populatedUseCase, found := useCaseLookup[useCase.Key.String()]
		if !found {
			return nil, nil, errors.New("use case not found in lookup: " + useCase.Key.String())
		}
		useCases = append(useCases, populatedUseCase)
	}

	// Collect unique actors.
	uniqueActors := map[string]model_actor.Actor{}
	for _, useCase := range useCases {
		for actorKey := range useCase.Actors {
			actor, found := actorLookup[actorKey.String()]
			if !found {
				return nil, nil, errors.New("actor not found in lookup: " + actorKey.String())
			}
			uniqueActors[actorKey.String()] = actor

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
