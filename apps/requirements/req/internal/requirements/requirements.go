package requirements

import (
	"sort"

	"github.com/pkg/errors"
)

type Requirements struct {
	Model Model
	// Generalizations.
	Generalizations []Generalization
	// Actors.
	Actors []Actor
	// Organization.
	Domains            []Domain
	Subdomains         map[string][]Subdomain // All the subdomains in a domain.
	DomainAssociations []DomainAssociation
	// Classes.
	Classes      map[string][]Class     // All the classes in a subdomain.
	Attributes   map[string][]Attribute // All the attributes in a class.
	Associations []Association
	// Class States.
	States       map[string][]State       // All the states in a class.
	Events       map[string][]Event       // All the state events in a class.
	Guards       map[string][]Guard       // All the state guards in a class.
	Actions      map[string][]Action      // All the state actions in a class.
	Transitions  map[string][]Transition  // All the state transitions in a class.
	StateActions map[string][]StateAction // All the state actions in a state.
	// Use Cases.
	UseCases      map[string][]UseCase               // All the use cases in a subdomain.
	UseCaseActors map[string]map[string]UseCaseActor // All the use cases actors.
	// Scenarios.
	Scenarios       map[string][]Scenario       // All scenarios in a use case.
	ScenarioObjects map[string][]ScenarioObject // All scenario objects in a scenario.
	// Convenience structures.
	generalizationLookup map[string]Generalization
	actorLookup          map[string]Actor
	domainLookup         map[string]Domain
	classLookup          map[string]Class
	attributeLookup      map[string]Attribute
	associationLookup    map[string]Association
	stateLookup          map[string]State
	eventLookup          map[string]Event
	guardLookup          map[string]Guard
	actionLookup         map[string]Action
	transitionLookup     map[string]Transition
	stateActionLookup    map[string]StateAction
	useCaseLookup        map[string]UseCase
	scenarioLookup       map[string]Scenario
	scenarioObjectLookup map[string]ScenarioObject
}

// Prepare data for templating.
func (r *Requirements) PrepLookups() {
	r.prepLookups()
}

func (r *Requirements) prepLookups() {
	if r.generalizationLookup == nil {

		// Put data into an easy to lookup format.
		r.generalizationLookup = createKeyGeneralizationLookup(r.Classes, r.Generalizations)
		r.actorLookup = createKeyActorLookup(r.Classes, r.Actors)
		r.domainLookup = createKeyDomainLookup(r.Classes, r.UseCases, r.Domains)
		r.classLookup = createKeyClassLookup(r.Attributes, r.States, r.Events, r.Guards, r.Actions, r.Transitions, r.Classes)
		r.attributeLookup = createKeyAttributeLookup(r.Attributes)
		r.associationLookup = createKeyAssociationLookup(r.Associations)
		r.stateLookup = createKeyStateLookup(r.StateActions, r.States)
		r.eventLookup = createKeyEventLookup(r.Events)
		r.guardLookup = createKeyGuardLookup(r.Guards)
		r.actionLookup = createKeyActionLookup(r.Transitions, r.StateActions, r.Actions)
		r.transitionLookup = createKeyTransitionLookup(r.Transitions)
		r.stateActionLookup = createKeyStateActionLookup(r.StateActions)
		r.useCaseLookup = createKeyUseCaseLookup(r.UseCases, r.UseCaseActors, r.Scenarios)
		r.scenarioLookup = createKeyScenarioLookup(r.Scenarios, r.ScenarioObjects)
		r.scenarioObjectLookup = createKeyScenarioObjectLookup(r.ScenarioObjects, r.classLookup)

		// Populate references in scenarios. Their steps are like and abstract symbol tree.
		// And any references to objects, events, attributes, or scenarios need to be populated.
		if err := populateScenarioStepReferences(r.scenarioLookup, r.scenarioObjectLookup, r.attributeLookup, r.eventLookup); err != nil {
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

func (r *Requirements) GeneralizationLookup() (generalizationLookup map[string]Generalization) {
	r.prepLookups()
	return r.generalizationLookup
}

func (r *Requirements) ActorLookup() (actorLookup map[string]Actor) {
	r.prepLookups()
	return r.actorLookup
}

func (r *Requirements) DomainLookup() (domainLookup map[string]Domain, associations []DomainAssociation) {
	r.prepLookups()
	return r.domainLookup, r.DomainAssociations
}

func (r *Requirements) ClassLookup() (classLookup map[string]Class, associations []Association) {
	r.prepLookups()
	return r.classLookup, r.Associations
}

func (r *Requirements) StateLookup() (eventLookup map[string]State) {
	r.prepLookups()
	return r.stateLookup
}

func (r *Requirements) EventLookup() (eventLookup map[string]Event) {
	r.prepLookups()
	return r.eventLookup
}

func (r *Requirements) GuardLookup() (guardLookup map[string]Guard) {
	r.prepLookups()
	return r.guardLookup
}

func (r *Requirements) ActionLookup() (actionLookup map[string]Action) {
	r.prepLookups()
	return r.actionLookup
}

func (r *Requirements) UseCaseLookup() (useCaseLookup map[string]UseCase) {
	r.prepLookups()
	return r.useCaseLookup
}

func (r *Requirements) ScenarioLookup() (scenarioLookup map[string]Scenario) {
	r.prepLookups()
	return r.scenarioLookup
}

func (r *Requirements) ScenarioObjectLookup() (scenarioObjectLookup map[string]ScenarioObject) {
	r.prepLookups()
	return r.scenarioObjectLookup
}

// Get all the objects connected to one or more classes for rending in a uml diagram.
func (r *Requirements) RegardingClasses(inClasses []Class) (generalizations []Generalization, classes []Class, associations []Association) {
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
func (r *Requirements) RegardingUseCases(inUseCases []UseCase) (useCases []UseCase, actors []Actor, err error) {
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
	uniqueActors := map[string]Actor{}
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

func mergeClassLookups(lookupA, lookupB map[string]Class) (lookup map[string]Class) {
	lookup = map[string]Class{}
	for _, class := range lookupA {
		lookup[class.Key] = class
	}
	for _, class := range lookupB {
		lookup[class.Key] = class
	}
	return lookup
}

func classesAsLookup(allClassLookup map[string]Class, classes []Class) (lookup map[string]Class) {
	lookup = map[string]Class{}
	for _, class := range classes {
		lookup[class.Key] = allClassLookup[class.Key]
	}
	return lookup
}

func classGeneralizationsAsLookup(classLookup map[string]Class, generalizationLookup map[string]Generalization) (lookup map[string]Generalization) {
	lookup = map[string]Generalization{}
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

func classAssociationsAsLookup(classLookup map[string]Class, associations []Association) (lookup map[string]Association) {
	lookup = map[string]Association{}
	for classKey := range classLookup {
		for _, association := range associations {
			if association.Includes(classKey) {
				lookup[association.Key] = association
			}
		}
	}
	return lookup
}

func classesFromGeneralizations(allClassLookup map[string]Class, generalizationLookup map[string]Generalization) (lookup map[string]Class) {
	lookup = map[string]Class{}
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

func classesFromAssociations(allClassLookup map[string]Class, associationLookup map[string]Association) (lookup map[string]Class) {
	lookup = map[string]Class{}
	for _, association := range associationLookup {
		lookup[association.FromClassKey] = allClassLookup[association.FromClassKey]
		lookup[association.ToClassKey] = allClassLookup[association.ToClassKey]
		if association.AssociationClassKey != "" {
			lookup[association.AssociationClassKey] = allClassLookup[association.AssociationClassKey]
		}
	}
	return lookup
}
