package requirements

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
)

type Requirements struct {
	Model Model
	// Generalizations.
	Generalizations []model_class.Generalization
	// Actors.
	Actors []model_actor.Actor
	// Organization.
	Domains            []model_domain.Domain
	Subdomains         map[string][]model_domain.Subdomain // All the subdomains in a domain.
	DomainAssociations []model_domain.Association
	// Classes.
	Classes      map[string][]model_class.Class     // All the classes in a subdomain.
	Attributes   map[string][]model_class.Attribute // All the attributes in a class.
	Associations []model_class.Association
	// Class States.
	States       map[string][]model_state.State       // All the states in a class.
	Events       map[string][]model_state.Event       // All the state events in a class.
	Guards       map[string][]model_state.Guard       // All the state guards in a class.
	Actions      map[string][]model_state.Action      // All the state actions in a class.
	Transitions  map[string][]model_state.Transition  // All the state transitions in a class.
	StateActions map[string][]model_state.StateAction // All the state actions in a state.
	// Use Cases.
	UseCases      map[string][]model_use_case.UseCase        // All the use cases in a subdomain.
	UseCaseActors map[string]map[string]model_use_case.Actor // All the use cases actors.
	// Scenarios.
	Scenarios map[string][]model_scenario.Scenario // All scenarios in a use case.
	Objects   map[string][]model_scenario.Object   // All scenario objects in a scenario.
	// Convenience structures.
	generalizationLookup map[string]model_class.Generalization
	actorLookup          map[string]model_actor.Actor
	domainLookup         map[string]model_domain.Domain
	classLookup          map[string]model_class.Class
	attributeLookup      map[string]model_class.Attribute
	associationLookup    map[string]model_class.Association
	stateLookup          map[string]model_state.State
	eventLookup          map[string]model_state.Event
	guardLookup          map[string]model_state.Guard
	actionLookup         map[string]model_state.Action
	transitionLookup     map[string]model_state.Transition
	stateActionLookup    map[string]model_state.StateAction
	useCaseLookup        map[string]model_use_case.UseCase
	scenarioLookup       map[string]model_scenario.Scenario
	objectLookup         map[string]model_scenario.Object
}

// Prepare data for templating.
func (r *Requirements) PrepLookups() {
	r.prepLookups()
}

func (r *Requirements) prepLookups() {
	if r.generalizationLookup == nil {

		// Put data into an easy to lookup format.
		r.generalizationLookup = model_class.CreateKeyGeneralizationLookup(r.Classes, r.Generalizations)
		r.actorLookup = model_actor.CreateKeyActorLookup(r.Classes, r.Actors)
		r.domainLookup = createKeyDomainLookup(r.Classes, r.UseCases, r.Domains)
		r.classLookup = model_class.CreateKeyClassLookup(r.Attributes, r.States, r.Events, r.Guards, r.Actions, r.Transitions, r.Classes)
		r.attributeLookup = model_class.CreateKeyAttributeLookup(r.Attributes)
		r.associationLookup = model_class.CreateKeyAssociationLookup(r.Associations)
		r.stateLookup = model_state.CreateKeyStateLookup(r.StateActions, r.States)
		r.eventLookup = model_state.CreateKeyEventLookup(r.Events)
		r.guardLookup = model_state.CreateKeyGuardLookup(r.Guards)
		r.actionLookup = model_state.CreateKeyActionLookup(r.Transitions, r.StateActions, r.Actions)
		r.transitionLookup = model_state.CreateKeyTransitionLookup(r.Transitions)
		r.stateActionLookup = model_state.CreateKeyStateActionLookup(r.StateActions)
		r.useCaseLookup = createKeyUseCaseLookup(r.UseCases, r.UseCaseActors, r.Scenarios)
		r.scenarioLookup = model_scenario.CreateKeyScenarioLookup(r.Scenarios, r.Objects)
		r.objectLookup = model_scenario.CreateKeyObjectLookup(r.Objects, r.classLookup)

		// Populate references in scenarios. Their steps are like and abstract symbol tree.
		// And any references to objects, events, attributes, or scenarios need to be populated.
		if err := model_scenario.PopulateScenarioStepReferences(r.scenarioLookup, r.objectLookup, r.attributeLookup, r.eventLookup); err != nil {
			panic(errors.Errorf("error populating scenario step references: %+v", err))
		}

		// Sort anything that should be sorted for templates.
		sort.Slice(r.Actors, func(i, j int) bool {
			return r.Actors[i].Key < r.Actors[j].Key
		})
		sort.Slice(r.DomainAssociations, func(i, j int) bool {
			return r.DomainAssociations[i].Key.String() < r.DomainAssociations[j].Key.String()
		})
		sort.Slice(r.Associations, func(i, j int) bool {
			return r.Associations[i].Key < r.Associations[j].Key
		})
	}
}

func (r *Requirements) GeneralizationLookup() (generalizationLookup map[string]model_class.Generalization) {
	r.prepLookups()
	return r.generalizationLookup
}

func (r *Requirements) ActorLookup() (actorLookup map[string]model_actor.Actor) {
	r.prepLookups()
	return r.actorLookup
}

func (r *Requirements) DomainLookup() (domainLookup map[string]model_domain.Domain, associations []model_domain.Association) {
	r.prepLookups()
	return r.domainLookup, r.DomainAssociations
}

func (r *Requirements) ClassLookup() (classLookup map[string]model_class.Class, associations []model_class.Association) {
	r.prepLookups()
	return r.classLookup, r.Associations
}

func (r *Requirements) StateLookup() (stateLookup map[string]model_state.State) {
	r.prepLookups()
	return r.stateLookup
}

func (r *Requirements) EventLookup() (eventLookup map[string]model_state.Event) {
	r.prepLookups()
	return r.eventLookup
}

func (r *Requirements) GuardLookup() (guardLookup map[string]model_state.Guard) {
	r.prepLookups()
	return r.guardLookup
}

func (r *Requirements) ActionLookup() (actionLookup map[string]model_state.Action) {
	r.prepLookups()
	return r.actionLookup
}

func (r *Requirements) UseCaseLookup() (useCaseLookup map[string]model_use_case.UseCase) {
	r.prepLookups()
	return r.useCaseLookup
}

func (r *Requirements) ScenarioLookup() (scenarioLookup map[string]model_scenario.Scenario) {
	r.prepLookups()
	return r.scenarioLookup
}

func (r *Requirements) ObjectLookup() (objectLookup map[string]model_scenario.Object) {
	r.prepLookups()
	return r.objectLookup
}

// Get all the objects connected to one or more classes for rending in a uml diagram.
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
		return useCases[i].Key.String() < useCases[j].Key.String()
	})
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Key < actors[j].Key
	})

	return useCases, actors, nil
}

func mergeClassLookups(lookupA, lookupB map[string]model_class.Class) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	for _, class := range lookupA {
		lookup[class.Key] = class
	}
	for _, class := range lookupB {
		lookup[class.Key] = class
	}
	return lookup
}

func classesAsLookup(allClassLookup map[string]model_class.Class, classes []model_class.Class) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
	for _, class := range classes {
		lookup[class.Key] = allClassLookup[class.Key]
	}
	return lookup
}

func classGeneralizationsAsLookup(classLookup map[string]model_class.Class, generalizationLookup map[string]model_class.Generalization) (lookup map[string]model_class.Generalization) {
	lookup = map[string]model_class.Generalization{}
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

func classAssociationsAsLookup(classLookup map[string]model_class.Class, associations []model_class.Association) (lookup map[string]model_class.Association) {
	lookup = map[string]model_class.Association{}
	for classKey := range classLookup {
		for _, association := range associations {
			if association.Includes(classKey) {
				lookup[association.Key] = association
			}
		}
	}
	return lookup
}

func classesFromGeneralizations(allClassLookup map[string]model_class.Class, generalizationLookup map[string]model_class.Generalization) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
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

func classesFromAssociations(allClassLookup map[string]model_class.Class, associationLookup map[string]model_class.Association) (lookup map[string]model_class.Class) {
	lookup = map[string]model_class.Class{}
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
func (r *Requirements) ToTree() Model {
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
		domain.Subdomains = r.Subdomains[domain.Key.String()]

		// Populate subdomains
		for j := range domain.Subdomains {
			subdomain := &domain.Subdomains[j]
			subdomain.Classes = r.Classes[subdomain.Key.String()]
			subdomain.UseCases = r.UseCases[subdomain.Key.String()]

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
				useCase.Actors = r.UseCaseActors[useCase.Key.String()]
				useCase.Scenarios = r.Scenarios[useCase.Key.String()]

				// Populate scenarios with objects
				for l := range useCase.Scenarios {
					scenario := &useCase.Scenarios[l]
					scenario.Objects = r.Objects[scenario.Key]
				}
			}
		}
	}

	// Group generalizations by subdomain
	for _, g := range r.Generalizations {
		subdomainKey := classToSubdomain[g.SuperclassKey]
		if subdomainKey != "" {
			for i := range tree.Domains {
				for j := range tree.Domains[i].Subdomains {
					if tree.Domains[i].Subdomains[j].Key.String() == subdomainKey {
						tree.Domains[i].Subdomains[j].Generalizations = append(tree.Domains[i].Subdomains[j].Generalizations, g)
					}
				}
			}
		}
	}

	// Group associations by subdomain or model
	for _, association := range r.Associations {
		fromSubdomain := classToSubdomain[association.FromClassKey]
		toSubdomain := classToSubdomain[association.ToClassKey]
		if fromSubdomain == toSubdomain && fromSubdomain != "" {
			for i := range tree.Domains {
				for j := range tree.Domains[i].Subdomains {
					if tree.Domains[i].Subdomains[j].Key.String() == fromSubdomain {
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
func (r *Requirements) FromTree(tree Model) {
	// Clear the maps
	r.Subdomains = make(map[string][]model_domain.Subdomain)
	r.Classes = make(map[string][]model_class.Class)
	r.Attributes = make(map[string][]model_class.Attribute)
	r.States = make(map[string][]model_state.State)
	r.Events = make(map[string][]model_state.Event)
	r.Guards = make(map[string][]model_state.Guard)
	r.Actions = make(map[string][]model_state.Action)
	r.Transitions = make(map[string][]model_state.Transition)
	r.StateActions = make(map[string][]model_state.StateAction)
	r.UseCases = make(map[string][]model_use_case.UseCase)
	r.UseCaseActors = make(map[string]map[string]model_use_case.Actor)
	r.Scenarios = make(map[string][]model_scenario.Scenario)
	r.Objects = make(map[string][]model_scenario.Object)

	// Populate from tree
	for _, domain := range tree.Domains {
		r.Subdomains[domain.Key.String()] = domain.Subdomains

		for _, subdomain := range domain.Subdomains {
			r.Classes[subdomain.Key.String()] = subdomain.Classes
			r.UseCases[subdomain.Key.String()] = subdomain.UseCases

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
				r.UseCaseActors[useCase.Key.String()] = useCase.Actors
				r.Scenarios[useCase.Key.String()] = useCase.Scenarios

				for _, scenario := range useCase.Scenarios {
					r.Objects[scenario.Key] = scenario.Objects
				}
			}
		}
	}

	// Collect generalizations from subdomains
	r.Generalizations = []model_class.Generalization{}
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
	r.Model = Model{
		Key:                tree.Key,
		Name:               tree.Name,
		Details:            tree.Details,
		Actors:             tree.Actors,
		Domains:            make([]model_domain.Domain, len(tree.Domains)),
		DomainAssociations: tree.DomainAssociations,
		Associations:       tree.Associations,
	}
	for i, d := range tree.Domains {
		r.Model.Domains[i] = model_domain.Domain{
			Key:        d.Key,
			Name:       d.Name,
			Details:    d.Details,
			Realized:   d.Realized,
			UmlComment: d.UmlComment,
			// Subdomains empty
		}
	}
}

func createKeyDomainLookup(domainClasses map[string][]model_class.Class, domainUseCases map[string][]model_use_case.UseCase, items []model_domain.Domain) (lookup map[string]model_domain.Domain) {

	lookup = map[string]model_domain.Domain{}
	for _, item := range items {

		item.Classes = domainClasses[item.Key.String()]
		item.UseCases = domainUseCases[item.Key.String()]

		lookup[item.Key.String()] = item
	}
	return lookup
}

func createKeyUseCaseLookup(
	byCategory map[string][]model_use_case.UseCase,
	actors map[string]map[string]model_use_case.Actor,
	scenarios map[string][]model_scenario.Scenario,
) (lookup map[string]model_use_case.UseCase) {

	lookup = map[string]model_use_case.UseCase{}
	for domainKey, items := range byCategory {
		for _, item := range items {

			item.SetDomainKey(domainKey)
			item.SetActors(actors[item.Key.String()])
			item.SetScenarios(scenarios[item.Key.String()])

			lookup[item.Key.String()] = item
		}
	}
	return lookup
}
