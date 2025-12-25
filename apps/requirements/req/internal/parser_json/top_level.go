package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// UnpackRequirements converts a requirements.Requirements into a tree of parser_json objects.
func UnpackRequirements(reqs requirements.Requirements) modelInOut {
	return FromRequirementsModel(reqs.Model)
}

// PackRequirements converts a tree of parser_json objects back into requirements.Requirements.
func PackRequirements(tree modelInOut) requirements.Requirements {
	model := tree.ToRequirements()

	reqs := requirements.Requirements{
		Model: model,
	}

	// Flatten the nested structure into the top-level fields.

	// Actors
	reqs.Actors = model.Actors

	// Domains
	reqs.Domains = model.Domains

	// Domain Associations
	reqs.DomainAssociations = model.DomainAssociations

	// Associations
	reqs.Associations = model.Associations

	// Subdomains, Classes, etc.
	reqs.Subdomains = make(map[string][]requirements.Subdomain)
	reqs.Classes = make(map[string][]requirements.Class)
	reqs.UseCases = make(map[string][]requirements.UseCase)

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			reqs.Subdomains[domain.Key] = append(reqs.Subdomains[domain.Key], subdomain)

			// Generalizations
			reqs.Generalizations = append(reqs.Generalizations, subdomain.Generalizations...)

			// Classes
			for _, class := range subdomain.Classes {
				reqs.Classes[subdomain.Key] = append(reqs.Classes[subdomain.Key], class)

				// Attributes
				if reqs.Attributes == nil {
					reqs.Attributes = make(map[string][]requirements.Attribute)
				}
				reqs.Attributes[class.Key] = class.Attributes

				// States
				if reqs.States == nil {
					reqs.States = make(map[string][]requirements.State)
				}
				reqs.States[class.Key] = class.States

				// Events
				if reqs.Events == nil {
					reqs.Events = make(map[string][]requirements.Event)
				}
				reqs.Events[class.Key] = class.Events

				// Guards
				if reqs.Guards == nil {
					reqs.Guards = make(map[string][]requirements.Guard)
				}
				reqs.Guards[class.Key] = class.Guards

				// Actions
				if reqs.Actions == nil {
					reqs.Actions = make(map[string][]requirements.Action)
				}
				reqs.Actions[class.Key] = class.Actions

				// Transitions
				if reqs.Transitions == nil {
					reqs.Transitions = make(map[string][]requirements.Transition)
				}
				reqs.Transitions[class.Key] = class.Transitions

				// State Actions
				if reqs.StateActions == nil {
					reqs.StateActions = make(map[string][]requirements.StateAction)
				}
				for _, state := range class.States {
					reqs.StateActions[state.Key] = state.Actions
				}
			}

			// Use Cases
			for _, useCase := range subdomain.UseCases {
				reqs.UseCases[subdomain.Key] = append(reqs.UseCases[subdomain.Key], useCase)

				// Use Case Actors
				if reqs.UseCaseActors == nil {
					reqs.UseCaseActors = make(map[string]map[string]requirements.UseCaseActor)
				}
				reqs.UseCaseActors[useCase.Key] = useCase.Actors

				// Scenarios
				if reqs.Scenarios == nil {
					reqs.Scenarios = make(map[string][]requirements.Scenario)
				}
				reqs.Scenarios[useCase.Key] = useCase.Scenarios

				// Scenario Objects
				if reqs.ScenarioObjects == nil {
					reqs.ScenarioObjects = make(map[string][]requirements.ScenarioObject)
				}
				for _, scenario := range useCase.Scenarios {
					reqs.ScenarioObjects[scenario.Key] = scenario.Objects
				}
			}

			// Associations in subdomain
			reqs.Associations = append(reqs.Associations, subdomain.Associations...)
		}
	}

	return reqs
}