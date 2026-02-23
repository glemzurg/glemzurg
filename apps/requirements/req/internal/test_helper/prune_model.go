package test_helper

import (
	"fmt"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
)

// PruneToModelOnly returns a copy of the model with only its direct children:
// actors, actor generalizations, domains (with empty subdomains), domain associations,
// invariants, and global functions. No class associations.
func PruneToModelOnly(model req_model.Model) req_model.Model {
	model.ClassAssociations = nil

	// Strip subdomains from each domain, keeping only the domain-level fields.
	for domainKey, domain := range model.Domains {
		domain.ClassAssociations = nil
		for subdomainKey, subdomain := range domain.Subdomains {
			subdomain.Classes = nil
			subdomain.Generalizations = nil
			subdomain.UseCases = nil
			subdomain.UseCaseGeneralizations = nil
			subdomain.ClassAssociations = nil
			subdomain.UseCaseShares = nil
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

// PruneToClassAttributes returns a copy of the model with everything through class
// attributes. Includes subdomains, classes with attributes, and class generalizations.
// No class associations, no states/events/guards/actions/queries/transitions,
// no use cases, no use case generalizations.
func PruneToClassAttributes(model req_model.Model) req_model.Model {
	model.ClassAssociations = nil

	for domainKey, domain := range model.Domains {
		domain.ClassAssociations = nil
		for subdomainKey, subdomain := range domain.Subdomains {
			subdomain.UseCases = nil
			subdomain.UseCaseGeneralizations = nil
			subdomain.ClassAssociations = nil
			subdomain.UseCaseShares = nil

			// Strip state machine from each class.
			for classKey, class := range subdomain.Classes {
				class.States = nil
				class.Events = nil
				class.Guards = nil
				class.Actions = nil
				class.Queries = nil
				class.Transitions = nil
				subdomain.Classes[classKey] = class
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

// PruneToClassAssociations returns a copy of the model with everything through class
// associations. Includes subdomains, classes with attributes, class generalizations,
// and class associations at all levels (subdomain, domain, and model).
// No states/events/guards/actions/queries/transitions, no use cases.
func PruneToClassAssociations(model req_model.Model) req_model.Model {
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			subdomain.UseCases = nil
			subdomain.UseCaseGeneralizations = nil
			subdomain.UseCaseShares = nil

			// Strip state machine from each class.
			for classKey, class := range subdomain.Classes {
				class.States = nil
				class.Events = nil
				class.Guards = nil
				class.Actions = nil
				class.Queries = nil
				class.Transitions = nil
				subdomain.Classes[classKey] = class
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

// PruneToStateMachine returns a copy of the model with everything through the state
// machine. Includes subdomains, classes with attributes, class generalizations,
// class associations, and all state machine parts (states, events, guards, actions,
// queries, transitions). No use cases or use case generalizations.
func PruneToStateMachine(model req_model.Model) req_model.Model {
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			subdomain.UseCases = nil
			subdomain.UseCaseGeneralizations = nil
			subdomain.UseCaseShares = nil
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

// PruneToNoSteps returns a copy of the model with everything except scenario steps.
// All scenarios have their Steps set to nil.
func PruneToNoSteps(model req_model.Model) req_model.Model {
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for useCaseKey, useCase := range subdomain.UseCases {
				for scenarioKey, scenario := range useCase.Scenarios {
					scenario.Steps = nil
					useCase.Scenarios[scenarioKey] = scenario
				}
				subdomain.UseCases[useCaseKey] = useCase
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

// LocatedScenario pairs a scenario with a human-readable path describing its location.
type LocatedScenario struct {
	Path     string
	Scenario model_scenario.Scenario
}

// ExtractScenarios walks the model tree and returns all scenarios as a sorted slice.
// Each entry includes a path like "domain > subdomain > use_case > scenario" for diagnostics.
func ExtractScenarios(model req_model.Model) []LocatedScenario {
	var result []LocatedScenario

	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for useCaseKey, useCase := range subdomain.UseCases {
				for scenarioKey, scenario := range useCase.Scenarios {
					path := fmt.Sprintf("%s > %s > %s > %s",
						domainKey, subdomainKey, useCaseKey, scenarioKey)
					result = append(result, LocatedScenario{
						Path:     path,
						Scenario: scenario,
					})
				}
			}
		}
	}

	// Sort by path for deterministic comparison.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result
}
