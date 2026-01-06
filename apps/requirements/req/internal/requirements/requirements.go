package requirements

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/use_case"
	"github.com/pkg/errors"
)

type Requirements struct {
	Model model.Model
	// Generalizations.
	Generalizations []class.Generalization
	// Actors.
	Actors []actor.Actor
	// Organization.
	Domains            []domain.Domain
	Subdomains         map[string][]domain.Subdomain // All the subdomains in a domain.
	DomainAssociations []domain.DomainAssociation
	// Classes.
	Classes      map[string][]class.Class     // All the classes in a subdomain.
	Attributes   map[string][]class.Attribute // All the attributes in a class.
	Associations []class.Association
	// Class States.
	States       map[string][]state.State       // All the states in a class.
	Events       map[string][]state.Event       // All the state events in a class.
	Guards       map[string][]state.Guard       // All the state guards in a class.
	Actions      map[string][]state.Action      // All the state actions in a class.
	Transitions  map[string][]state.Transition  // All the state transitions in a class.
	StateActions map[string][]state.StateAction // All the state actions in a state.
	// Use Cases.
	UseCases      map[string][]use_case.UseCase               // All the use cases in a subdomain.
	UseCaseActors map[string]map[string]use_case.UseCaseActor // All the use cases actors.
	// Scenarios.
	Scenarios       map[string][]scenario.Scenario       // All scenarios in a use case.
	ScenarioObjects map[string][]scenario.ScenarioObject // All scenario objects in a scenario.
	// Convenience structures.
	generalizationLookup map[string]class.Generalization
	actorLookup          map[string]actor.Actor
	domainLookup         map[string]domain.Domain
	classLookup          map[string]class.Class
	attributeLookup      map[string]class.Attribute
	associationLookup    map[string]class.Association
	stateLookup          map[string]state.State
	eventLookup          map[string]state.Event
	guardLookup          map[string]state.Guard
	actionLookup         map[string]state.Action
	transitionLookup     map[string]state.Transition
	stateActionLookup    map[string]state.StateAction
	useCaseLookup        map[string]use_case.UseCase
	scenarioLookup       map[string]scenario.Scenario
	scenarioObjectLookup map[string]scenario.ScenarioObject
}

// Prepare data for templating.
func (r *Requirements) PrepLookups() {
	r.prepLookups()
}

func (r *Requirements) prepLookups() {
	if r.generalizationLookup == nil {

		// Put data into an easy to lookup format.
		r.generalizationLookup = class.CreateKeyGeneralizationLookup(r.Classes, r.Generalizations)
		r.actorLookup = actor.CreateKeyActorLookup(r.Classes, r.Actors)
		r.domainLookup = domain.CreateKeyDomainLookup(r.Classes, r.UseCases, r.Domains)
		r.classLookup = class.CreateKeyClassLookup(r.Attributes, r.States, r.Events, r.Guards, r.Actions, r.Transitions, r.Classes)
		r.attributeLookup = class.CreateKeyAttributeLookup(r.Attributes)
		r.associationLookup = class.CreateKeyAssociationLookup(r.Associations)
		r.stateLookup = state.CreateKeyStateLookup(r.StateActions, r.States)
		r.eventLookup = state.CreateKeyEventLookup(r.Events)
		r.guardLookup = state.CreateKeyGuardLookup(r.Guards)
		r.actionLookup = state.CreateKeyActionLookup(r.Transitions, r.StateActions, r.Actions)
		r.transitionLookup = state.CreateKeyTransitionLookup(r.Transitions)
		r.stateActionLookup = state.CreateKeyStateActionLookup(r.StateActions)
		r.useCaseLookup = use_case.CreateKeyUseCaseLookup(r.UseCases, r.UseCaseActors, r.Scenarios)
		r.scenarioLookup = scenario.CreateKeyScenarioLookup(r.Scenarios, r.ScenarioObjects)
		r.scenarioObjectLookup = scenario.CreateKeyScenarioObjectLookup(r.ScenarioObjects, r.classLookup)

		// Populate references in scenarios. Their steps are like and abstract symbol tree.
		// And any references to objects, events, attributes, or scenarios need to be populated.
		if err := scenario.PopulateScenarioStepReferences(r.scenarioLookup, r.scenarioObjectLookup, r.attributeLookup, r.eventLookup); err != nil {
			panic(errors.Errorf("error populating scenario step references: %+v", err))
		}

		// Sort anything that should be sorted for templates.
		sort.Slice(r.Actors, func(i, j int) bool {
			return r.Actors[i].Key < r.Actors[j].Key
		})
		sort.Slice(r.DomainAssociations, func(i, j int) bool {
			return r.DomainAssociations[i].Key < r.DomainAssociations[j].Key
		})
		sort.Slice(r.Associations, func(i, j int) bool {
			return r.Associations[i].Key < r.Associations[j].Key
		})
	}
}

func (r *Requirements) GeneralizationLookup() (generalizationLookup map[string]class.Generalization) {
	r.prepLookups()
	return r.generalizationLookup
}

func (r *Requirements) ActorLookup() (actorLookup map[string]actor.Actor) {
	r.prepLookups()
	return r.actorLookup
}

func (r *Requirements) DomainLookup() (domainLookup map[string]domain.Domain, associations []domain.DomainAssociation) {
	r.prepLookups()
	return r.domainLookup, r.DomainAssociations
}

func (r *Requirements) ClassLookup() (classLookup map[string]class.Class, associations []class.Association) {
	r.prepLookups()
	return r.classLookup, r.Associations
}

func (r *Requirements) StateLookup() (eventLookup map[string]state.State) {
	r.prepLookups()
	return r.stateLookup
}

func (r *Requirements) EventLookup() (eventLookup map[string]state.Event) {
	r.prepLookups()
	return r.eventLookup
}

func (r *Requirements) GuardLookup() (guardLookup map[string]state.Guard) {
	r.prepLookups()
	return r.guardLookup
}

func (r *Requirements) ActionLookup() (actionLookup map[string]state.Action) {
	r.prepLookups()
	return r.actionLookup
}

func (r *Requirements) UseCaseLookup() (useCaseLookup map[string]use_case.UseCase) {
	r.prepLookups()
	return r.useCaseLookup
}

func (r *Requirements) ScenarioLookup() (scenarioLookup map[string]scenario.Scenario) {
	r.prepLookups()
	return r.scenarioLookup
}

func (r *Requirements) ScenarioObjectLookup() (scenarioObjectLookup map[string]scenario.ScenarioObject) {
	r.prepLookups()
	return r.scenarioObjectLookup
}

// Get all the objects connected to one or more classes for rending in a uml diagram.
func (r *Requirements) RegardingClasses(inClasses []class.Class) (generalizations []class.Generalization, classes []class.Class, associations []class.Association) {
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
		if class.SuperclassOfKey == "" && class.SubclassOfKey == "" {
			classes = append(classes, class)
		}
	}
	for _, association := range relevantAssociationsLookup {
		associations = append(associations, association)
	}

	// Sort everything.
	sort.Slice(generalizations, func(i, j int) bool {
		return generalizations[i].Key < generalizations[j].Key
	})
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Key < classes[j].Key
	})
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key < associations[j].Key
	})

	return generalizations, classes, associations
}

// Get all the actors connected to one or more use cases for rendering in a uml diagram.
func (r *Requirements) RegardingUseCases(inUseCases []use_case.UseCase) (useCases []use_case.UseCase, actors []actor.Actor, err error) {
	actorLookup := r.ActorLookup()
	useCaseLookup := r.UseCaseLookup()

	// Get the use cases that are fully loaded with data.
	for _, useCase := range inUseCases {
		populatedUseCase, found := useCaseLookup[useCase.Key]
		if !found {
			return nil, nil, errors.New("use case not found in lookup: " + useCase.Key)
		}
		useCases = append(useCases, populatedUseCase)
	}

	// Collect unique actors.
	uniqueActors := map[string]actor.Actor{}
	for _, useCase := range useCases {
		for actorKey := range useCase.Actors {
			actor, found := actorLookup[actorKey]
			if !found {
				return nil, nil, errors.New("actor not found in lookup: " + actorKey)
			}
			uniqueActors[actorKey] = actor

		}
	}

	// Convert to slice.
	for _, actor := range uniqueActors {
		actors = append(actors, actor)
	}

	// Sort.
	sort.Slice(useCases, func(i, j int) bool {
		return useCases[i].Key < useCases[j].Key
	})
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Key < actors[j].Key
	})

	return useCases, actors, nil
}

func mergeClassLookups(lookupA, lookupB map[string]class.Class) (lookup map[string]class.Class) {
	lookup = map[string]class.Class{}
	for _, class := range lookupA {
		lookup[class.Key] = class
	}
	for _, class := range lookupB {
		lookup[class.Key] = class
	}
	return lookup
}

func classesAsLookup(allClassLookup map[string]class.Class, classes []class.Class) (lookup map[string]class.Class) {
	lookup = map[string]class.Class{}
	for _, class := range classes {
		lookup[class.Key] = allClassLookup[class.Key]
	}
	return lookup
}

func classGeneralizationsAsLookup(classLookup map[string]class.Class, generalizationLookup map[string]class.Generalization) (lookup map[string]class.Generalization) {
	lookup = map[string]class.Generalization{}
	for _, class := range classLookup {
		for _, generalization := range generalizationLookup {
			if class.SuperclassOfKey == generalization.Key {
				lookup[generalization.Key] = generalization
			}
			if class.SubclassOfKey == generalization.Key {
				lookup[generalization.Key] = generalization
			}
		}
	}
	return lookup
}

func classAssociationsAsLookup(classLookup map[string]class.Class, associations []class.Association) (lookup map[string]class.Association) {
	lookup = map[string]class.Association{}
	for classKey := range classLookup {
		for _, association := range associations {
			if association.Includes(classKey) {
				lookup[association.Key] = association
			}
		}
	}
	return lookup
}

func classesFromGeneralizations(allClassLookup map[string]class.Class, generalizationLookup map[string]class.Generalization) (lookup map[string]class.Class) {
	lookup = map[string]class.Class{}
	for _, generalization := range generalizationLookup {
		if generalization.SuperclassKey != "" {
			lookup[generalization.SuperclassKey] = allClassLookup[generalization.SuperclassKey]
		}
		for _, subclassKey := range generalization.SubclassKeys {
			lookup[subclassKey] = allClassLookup[subclassKey]
		}
	}
	return lookup
}

func classesFromAssociations(allClassLookup map[string]class.Class, associationLookup map[string]class.Association) (lookup map[string]class.Class) {
	lookup = map[string]class.Class{}
	for _, association := range associationLookup {
		lookup[association.FromClassKey] = allClassLookup[association.FromClassKey]
		lookup[association.ToClassKey] = allClassLookup[association.ToClassKey]
		if association.AssociationClassKey != "" {
			lookup[association.AssociationClassKey] = allClassLookup[association.AssociationClassKey]
		}
	}
	return lookup
}

// ToTree builds a nested tree structure from the flattened Requirements.
func (r *Requirements) ToTree() model.Model {
	tree := r.Model

	// Build class to subdomain map
	classToSubdomain := make(map[string]string)
	for subdomainKey, classes := range r.Classes {
		for _, class := range classes {
			classToSubdomain[class.Key] = subdomainKey
		}
	}

	// Populate domains
	for i := range tree.Domains {
		domain := &tree.Domains[i]
		domain.Subdomains = r.Subdomains[domain.Key]

		// Populate subdomains
		for j := range domain.Subdomains {
			subdomain := &domain.Subdomains[j]
			subdomain.Classes = r.Classes[subdomain.Key]
			subdomain.UseCases = r.UseCases[subdomain.Key]

			// Populate classes
			for k := range subdomain.Classes {
				class := &subdomain.Classes[k]
				class.Attributes = r.Attributes[class.Key]
				class.States = r.States[class.Key]
				class.Events = r.Events[class.Key]
				class.Guards = r.Guards[class.Key]
				class.Actions = r.Actions[class.Key]
				class.Transitions = r.Transitions[class.Key]

				// Populate states with actions
				for l := range class.States {
					state := &class.States[l]
					state.Actions = r.StateActions[state.Key]
				}
			}

			// Populate use cases
			for k := range subdomain.UseCases {
				useCase := &subdomain.UseCases[k]
				useCase.Actors = r.UseCaseActors[useCase.Key]
				useCase.Scenarios = r.Scenarios[useCase.Key]

				// Populate scenarios with objects
				for l := range useCase.Scenarios {
					scenario := &useCase.Scenarios[l]
					scenario.Objects = r.ScenarioObjects[scenario.Key]
				}
			}
		}
	}

	// Populate domain associations
	tree.DomainAssociations = r.DomainAssociations

	// Group generalizations by subdomain
	for _, g := range r.Generalizations {
		subdomainKey := classToSubdomain[g.SuperclassKey]
		if subdomainKey != "" {
			for i := range tree.Domains {
				for j := range tree.Domains[i].Subdomains {
					if tree.Domains[i].Subdomains[j].Key == subdomainKey {
						tree.Domains[i].Subdomains[j].Generalizations = append(tree.Domains[i].Subdomains[j].Generalizations, g)
					}
				}
			}
		}
	}

	// Group associations by subdomain or model
	tree.Associations = nil
	for _, association := range r.Associations {
		fromSubdomain := classToSubdomain[association.FromClassKey]
		toSubdomain := classToSubdomain[association.ToClassKey]
		if fromSubdomain == toSubdomain && fromSubdomain != "" {
			for i := range tree.Domains {
				for j := range tree.Domains[i].Subdomains {
					if tree.Domains[i].Subdomains[j].Key == fromSubdomain {
						tree.Domains[i].Subdomains[j].Associations = append(tree.Domains[i].Subdomains[j].Associations, association)
					}
				}
			}
		} else {
			tree.Associations = append(tree.Associations, association)
		}
	}

	return tree
}

// FromTree flattens the nested tree structure into the Requirements maps and slices.
func (r *Requirements) FromTree(tree model.Model) {
	// Clear the maps
	r.Subdomains = make(map[string][]domain.Subdomain)
	r.Classes = make(map[string][]class.Class)
	r.Attributes = make(map[string][]class.Attribute)
	r.States = make(map[string][]state.State)
	r.Events = make(map[string][]state.Event)
	r.Guards = make(map[string][]state.Guard)
	r.Actions = make(map[string][]state.Action)
	r.Transitions = make(map[string][]state.Transition)
	r.StateActions = make(map[string][]state.StateAction)
	r.UseCases = make(map[string][]use_case.UseCase)
	r.UseCaseActors = make(map[string]map[string]use_case.UseCaseActor)
	r.Scenarios = make(map[string][]scenario.Scenario)
	r.ScenarioObjects = make(map[string][]scenario.ScenarioObject)

	// Populate from tree
	for _, domain := range tree.Domains {
		r.Subdomains[domain.Key] = domain.Subdomains

		for _, subdomain := range domain.Subdomains {
			r.Classes[subdomain.Key] = subdomain.Classes
			r.UseCases[subdomain.Key] = subdomain.UseCases

			for _, class := range subdomain.Classes {
				r.Attributes[class.Key] = class.Attributes
				r.States[class.Key] = class.States
				r.Events[class.Key] = class.Events
				r.Guards[class.Key] = class.Guards
				r.Actions[class.Key] = class.Actions
				r.Transitions[class.Key] = class.Transitions

				for _, state := range class.States {
					r.StateActions[state.Key] = state.Actions
				}
			}

			for _, useCase := range subdomain.UseCases {
				r.UseCaseActors[useCase.Key] = useCase.Actors
				r.Scenarios[useCase.Key] = useCase.Scenarios

				for _, scenario := range useCase.Scenarios {
					r.ScenarioObjects[scenario.Key] = scenario.Objects
				}
			}
		}
	}

	// Collect generalizations from subdomains
	r.Generalizations = []class.Generalization{}
	for _, domain := range tree.Domains {
		for _, subdomain := range domain.Subdomains {
			r.Generalizations = append(r.Generalizations, subdomain.Generalizations...)
		}
	}

	// Collect associations from model and subdomains
	r.Associations = tree.Associations
	for _, domain := range tree.Domains {
		for _, subdomain := range domain.Subdomains {
			r.Associations = append(r.Associations, subdomain.Associations...)
		}
	}

	// Set r.Model with empty nested
	r.Model = model.Model{
		Key:                tree.Key,
		Name:               tree.Name,
		Details:            tree.Details,
		Actors:             tree.Actors,
		Domains:            make([]domain.Domain, len(tree.Domains)),
		DomainAssociations: nil, // Associations are in r.DomainAssociations
		Associations:       nil, // Associations are in r.Associations
	}
	for i, d := range tree.Domains {
		r.Model.Domains[i] = domain.Domain{
			Key:        d.Key,
			Name:       d.Name,
			Details:    d.Details,
			Realized:   d.Realized,
			UmlComment: d.UmlComment,
			// Subdomains empty
			Associations: nil, // Associations are in r.Associations
			Classes:      nil, // Classes are in r.Classes
			UseCases:     nil, // UseCases are in r.UseCases
		}
	}
}
