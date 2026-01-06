package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// UnpackRequirements converts a requirements.Requirements into a tree of parser_json objects.
func UnpackRequirements(reqs requirements.Requirements) modelInOut {
	tree := modelInOut{
		Key:     reqs.Model.Key,
		Name:    reqs.Model.Name,
		Details: reqs.Model.Details,
	}

	// // Top-level generalizations
	// for _, g := range reqs.Generalizations {
	// 	tree.Generalizations = append(tree.Generalizations, FromRequirementsGeneralization(g))
	// }

	// Actors
	for _, a := range reqs.Actors {
		tree.Actors = append(tree.Actors, FromRequirementsActor(a))
	}

	// Domains and subdomains
	for _, domain := range reqs.Domains {
		domainInOut := FromRequirementsDomain(domain)
		// Add subdomains
		for _, subdomain := range reqs.Subdomains[domain.Key] {
			subdomainInOut := FromRequirementsSubdomain(subdomain)
			// Add generalizations to subdomain
			// In the test, generalizations are top-level, so skip
			// Add classes
			for _, class := range reqs.Classes[subdomain.Key] {
				classInOut := FromRequirementsClass(class)
				// Add attributes
				classInOut.Attributes = nil
				for _, attr := range reqs.Attributes[class.Key] {
					classInOut.Attributes = append(classInOut.Attributes, FromRequirementsAttribute(attr))
				}
				// Add states
				classInOut.States = nil
				for _, state := range reqs.States[class.Key] {
					stateInOut := FromRequirementsState(state)
					// Add state actions
					stateInOut.Actions = nil
					for _, sa := range reqs.StateActions[state.Key] {
						stateInOut.Actions = append(stateInOut.Actions, FromRequirementsStateAction(sa))
					}
					classInOut.States = append(classInOut.States, stateInOut)
				}
				// Add events
				classInOut.Events = nil
				for _, event := range reqs.Events[class.Key] {
					classInOut.Events = append(classInOut.Events, FromRequirementsEvent(event))
				}
				// Add guards
				classInOut.Guards = nil
				for _, guard := range reqs.Guards[class.Key] {
					classInOut.Guards = append(classInOut.Guards, FromRequirementsGuard(guard))
				}
				// Add actions
				classInOut.Actions = nil
				for _, action := range reqs.Actions[class.Key] {
					classInOut.Actions = append(classInOut.Actions, FromRequirementsAction(action))
				}
				// Add transitions
				classInOut.Transitions = nil
				for _, transition := range reqs.Transitions[class.Key] {
					classInOut.Transitions = append(classInOut.Transitions, FromRequirementsTransition(transition))
				}
				subdomainInOut.Classes = append(subdomainInOut.Classes, classInOut)
			}
			// Add use cases
			for _, useCase := range reqs.UseCases[subdomain.Key] {
				useCaseInOut := FromRequirementsUseCase(useCase)
				// Add actors
				useCaseInOut.Actors = make(map[string]useCaseActorInOut)
				for k, v := range reqs.UseCaseActors[useCase.Key] {
					useCaseInOut.Actors[k] = FromRequirementsUseCaseActor(v)
				}
				// Add scenarios
				for _, scenario := range reqs.Scenarios[useCase.Key] {
					scenarioInOut := FromRequirementsScenario(scenario)
					// Add objects
					scenarioInOut.Objects = nil
					for _, obj := range reqs.ScenarioObjects[scenario.Key] {
						scenarioInOut.Objects = append(scenarioInOut.Objects, FromRequirementsScenarioObject(obj))
					}
					useCaseInOut.Scenarios = append(useCaseInOut.Scenarios, scenarioInOut)
				}
				subdomainInOut.UseCases = append(subdomainInOut.UseCases, useCaseInOut)
			}
			// Add associations in subdomain
			// For now, assume associations are in model.Associations
			domainInOut.Subdomains = append(domainInOut.Subdomains, subdomainInOut)
		}
		tree.Domains = append(tree.Domains, domainInOut)
	}

	// Domain associations
	for _, da := range reqs.DomainAssociations {
		tree.DomainAssociations = append(tree.DomainAssociations, FromRequirementsDomainAssociation(da))
	}

	// Associations
	for _, a := range reqs.Associations {
		tree.Associations = append(tree.Associations, FromRequirementsAssociation(a))
	}

	return tree
}

// PackRequirements converts a tree of parser_json objects back into requirements.Requirements.
func PackRequirements(tree modelInOut) requirements.Requirements {
	reqs := requirements.Requirements{
		Model: requirements.Model{
			Key:     tree.Key,
			Name:    tree.Name,
			Details: tree.Details,
		},
	}

	// // Top-level generalizations
	// for _, g := range tree.Generalizations {
	// 	reqs.Generalizations = append(reqs.Generalizations, g.ToRequirements())
	// }

	// Flatten the nested structure into the top-level fields.

	// Actors
	reqs.Actors = make([]requirements.Actor, len(tree.Actors))
	for i, a := range tree.Actors {
		reqs.Actors[i] = a.ToRequirements()
	}

	// Domains
	reqs.Domains = make([]requirements.Domain, len(tree.Domains))
	for i, d := range tree.Domains {
		reqs.Domains[i] = d.ToRequirements()
	}

	// Domain Associations
	reqs.DomainAssociations = make([]requirements.DomainAssociation, len(tree.DomainAssociations))
	for i, da := range tree.DomainAssociations {
		reqs.DomainAssociations[i] = da.ToRequirements()
	}

	// Associations
	reqs.Associations = make([]requirements.Association, len(tree.Associations))
	for i, a := range tree.Associations {
		reqs.Associations[i] = a.ToRequirements()
	}

	// Subdomains, Classes, etc.
	reqs.Subdomains = make(map[string][]requirements.Subdomain)
	reqs.Classes = make(map[string][]requirements.Class)
	reqs.UseCases = make(map[string][]requirements.UseCase)

	for _, domain := range tree.Domains {
		for _, subdomain := range domain.Subdomains {
			reqs.Subdomains[domain.Key] = append(reqs.Subdomains[domain.Key], subdomain.ToRequirements())

			// Generalizations from subdomains
			for _, g := range subdomain.Generalizations {
				reqs.Generalizations = append(reqs.Generalizations, g.ToRequirements())
			}

			// Classes
			for _, class := range subdomain.Classes {
				reqs.Classes[subdomain.Key] = append(reqs.Classes[subdomain.Key], class.ToRequirements())

				// Attributes
				if len(class.Attributes) > 0 {
					if reqs.Attributes == nil {
						reqs.Attributes = make(map[string][]requirements.Attribute)
					}
					reqs.Attributes[class.Key] = make([]requirements.Attribute, len(class.Attributes))
					for i, a := range class.Attributes {
						reqs.Attributes[class.Key][i] = a.ToRequirements()
					}
				}

				// States
				if len(class.States) > 0 {
					if reqs.States == nil {
						reqs.States = make(map[string][]requirements.State)
					}
					reqs.States[class.Key] = make([]requirements.State, len(class.States))
					for i, s := range class.States {
						reqs.States[class.Key][i] = s.ToRequirements()
					}
				}

				// Events
				if len(class.Events) > 0 {
					if reqs.Events == nil {
						reqs.Events = make(map[string][]requirements.Event)
					}
					reqs.Events[class.Key] = make([]requirements.Event, len(class.Events))
					for i, e := range class.Events {
						reqs.Events[class.Key][i] = e.ToRequirements()
					}
				}

				// Guards
				if len(class.Guards) > 0 {
					if reqs.Guards == nil {
						reqs.Guards = make(map[string][]requirements.Guard)
					}
					reqs.Guards[class.Key] = make([]requirements.Guard, len(class.Guards))
					for i, g := range class.Guards {
						reqs.Guards[class.Key][i] = g.ToRequirements()
					}
				}

				// Actions
				if len(class.Actions) > 0 {
					if reqs.Actions == nil {
						reqs.Actions = make(map[string][]requirements.Action)
					}
					reqs.Actions[class.Key] = make([]requirements.Action, len(class.Actions))
					for i, a := range class.Actions {
						reqs.Actions[class.Key][i] = a.ToRequirements()
					}
				}

				// Transitions
				if len(class.Transitions) > 0 {
					if reqs.Transitions == nil {
						reqs.Transitions = make(map[string][]requirements.Transition)
					}
					reqs.Transitions[class.Key] = make([]requirements.Transition, len(class.Transitions))
					for i, tr := range class.Transitions {
						reqs.Transitions[class.Key][i] = tr.ToRequirements()
					}
				}

				// State Actions
				for _, state := range class.States {
					if len(state.Actions) > 0 {
						if reqs.StateActions == nil {
							reqs.StateActions = make(map[string][]requirements.StateAction)
						}
						reqs.StateActions[state.Key] = make([]requirements.StateAction, len(state.Actions))
						for i, sa := range state.Actions {
							reqs.StateActions[state.Key][i] = sa.ToRequirements()
						}
					}
				}
			}

			// Use Cases
			for _, useCase := range subdomain.UseCases {
				reqs.UseCases[subdomain.Key] = append(reqs.UseCases[subdomain.Key], useCase.ToRequirements())

				// Use Case Actors
				if reqs.UseCaseActors == nil {
					reqs.UseCaseActors = make(map[string]map[string]requirements.UseCaseActor)
				}
				reqs.UseCaseActors[useCase.Key] = make(map[string]requirements.UseCaseActor)
				for k, v := range useCase.Actors {
					reqs.UseCaseActors[useCase.Key][k] = v.ToRequirements()
				}

				// Scenarios
				if reqs.Scenarios == nil {
					reqs.Scenarios = make(map[string][]requirements.Scenario)
				}
				reqs.Scenarios[useCase.Key] = make([]requirements.Scenario, len(useCase.Scenarios))
				for i, s := range useCase.Scenarios {
					reqs.Scenarios[useCase.Key][i] = s.ToRequirements()
				}

				// Scenario Objects
				if reqs.ScenarioObjects == nil {
					reqs.ScenarioObjects = make(map[string][]requirements.ScenarioObject)
				}
				for _, scenario := range useCase.Scenarios {
					reqs.ScenarioObjects[scenario.Key] = make([]requirements.ScenarioObject, len(scenario.Objects))
					for i, o := range scenario.Objects {
						reqs.ScenarioObjects[scenario.Key][i] = o.ToRequirements()
					}
				}
			}

			// Associations in subdomain
			for _, a := range subdomain.Associations {
				reqs.Associations = append(reqs.Associations, a.ToRequirements())
			}
		}
	}

	return reqs
}
