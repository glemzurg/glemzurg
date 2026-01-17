package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

func WriteModel(db *sql.DB, model req_model.Model) (err error) {

	// Validate the model tree before writing to database.
	if err = model.ValidateWithParent(); err != nil {
		return err
	}

	// Everything should be written in order, as a transaction.
	err = dbTransaction(db, func(tx *sql.Tx) (err error) {

		modelKey := model.Key

		// Clear out the prior model first.
		if err = RemoveModel(tx, modelKey); err != nil {
			return err
		}

		// Add the model.
		if err = AddModel(tx, model); err != nil {
			return err
		}

		// Actors.
		for _, actor := range model.Actors {
			if err = AddActor(tx, modelKey, actor); err != nil {
				return err
			}
		}

		// First pass: Add all domains.
		for _, domain := range model.Domains {
			if err = AddDomain(tx, modelKey, domain); err != nil {
				return err
			}
		}

		// Second pass: Add domain associations (after all domains exist).
		for _, domain := range model.Domains {
			for _, association := range domain.DomainAssociations {
				if err = AddDomainAssociation(tx, modelKey, association); err != nil {
					return err
				}
			}
		}

		// Model-level domain associations.
		for _, association := range model.DomainAssociations {
			if err = AddDomainAssociation(tx, modelKey, association); err != nil {
				return err
			}
		}

		// Third pass: Add subdomains, classes, and other nested content.
		for _, domain := range model.Domains {
			domainKey := domain.Key

			// Subdomains.
			for _, subdomain := range domain.Subdomains {
				subdomainKey := subdomain.Key

				if err = AddSubdomain(tx, modelKey, domainKey, subdomain); err != nil {
					return err
				}

				// Generalizations.
				for _, generalization := range subdomain.Generalizations {
					if err = AddGeneralization(tx, modelKey, generalization); err != nil {
						return err
					}
				}

				// Classes.
				for _, class := range subdomain.Classes {
					classKey := class.Key

					if err = AddClass(tx, modelKey, subdomainKey, class); err != nil {
						return err
					}

					// Attributes.
					for _, attribute := range class.Attributes {
						if err = AddAttribute(tx, modelKey, classKey, attribute); err != nil {
							return err
						}
						// Add any indexes.
						for _, indexNum := range attribute.IndexNums {
							if err = AddClassIndex(tx, modelKey, classKey, attribute.Key, indexNum); err != nil {
								return err
							}
						}
					}

					// Events.
					for _, event := range class.Events {
						if err = AddEvent(tx, modelKey, classKey, event); err != nil {
							return err
						}
					}

					// Guards.
					for _, guard := range class.Guards {
						if err = AddGuard(tx, modelKey, classKey, guard); err != nil {
							return err
						}
					}

					// Actions (must be added before States with StateActions).
					for _, action := range class.Actions {
						if err = AddAction(tx, modelKey, classKey, action); err != nil {
							return err
						}
					}

					// States (after Actions since StateActions reference Actions).
					for _, state := range class.States {
						stateKey := state.Key

						if err = AddState(tx, modelKey, classKey, state); err != nil {
							return err
						}

						// State actions.
						for _, stateAction := range state.Actions {
							if err = AddStateAction(tx, modelKey, stateKey, stateAction); err != nil {
								return err
							}
						}
					}

					// Transitions.
					for _, transition := range class.Transitions {
						if err = AddTransition(tx, modelKey, classKey, transition); err != nil {
							return err
						}
					}
				}

				// Use cases.
				for _, useCase := range subdomain.UseCases {
					useCaseKey := useCase.Key

					if err = AddUseCase(tx, modelKey, subdomainKey, useCase); err != nil {
						return err
					}

					// Use case actors.
					for actorKey, actor := range useCase.Actors {
						if err = AddUseCaseActor(tx, modelKey, useCaseKey, actorKey, actor); err != nil {
							return err
						}
					}

					// Scenarios.
					for _, scenario := range useCase.Scenarios {
						scenarioKey := scenario.Key

						if err = AddScenario(tx, modelKey, useCaseKey, scenario); err != nil {
							return err
						}

						// Objects.
						for _, object := range scenario.Objects {
							if err = AddObject(tx, modelKey, scenarioKey, object); err != nil {
								return err
							}
						}
					}
				}

				// UseCaseShares.
				for seaLevelKey, mudLevelShares := range subdomain.UseCaseShares {
					for mudLevelKey, shared := range mudLevelShares {
						if err = AddUseCaseShared(tx, modelKey, seaLevelKey, mudLevelKey, shared); err != nil {
							return err
						}
					}
				}
			}
		}

		// Class associations - use GetClassAssociations to get all associations from all levels.
		classAssociations := model.GetClassAssociations()
		for _, association := range classAssociations {
			if err = AddAssociation(tx, modelKey, association); err != nil {
				return err
			}
		}

		// Collect all data types from attributes and add them.
		dataTypes := make(map[string]model_data_type.DataType)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, attribute := range class.Attributes {
						if attribute.DataType != nil {
							dataTypes[attribute.DataType.Key] = *attribute.DataType
						}
					}
				}
			}
		}

		// Add data types.
		if err = AddTopLevelDataTypes(tx, modelKey, dataTypes); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func ReadModel(db *sql.DB, modelKey string) (model req_model.Model, err error) {

	// Read from within a transaction.
	err = dbTransaction(db, func(tx *sql.Tx) (err error) {

		// Model.
		model, err = LoadModel(tx, modelKey)
		if err != nil {
			return err
		}

		// Actors - returns slice, convert to map.
		actorsSlice, err := QueryActors(tx, modelKey)
		if err != nil {
			return err
		}
		model.Actors = make(map[identity.Key]model_actor.Actor)
		for _, actor := range actorsSlice {
			model.Actors[actor.Key] = actor
		}

		// Domains - returns slice.
		domainsSlice, err := QueryDomains(tx, modelKey)
		if err != nil {
			return err
		}

		// Subdomains grouped by domain key.
		subdomainsMap, err := QuerySubdomains(tx, modelKey)
		if err != nil {
			return err
		}

		// Domain associations - returns slice, group by parent domain key.
		domainAssociationsSlice, err := QueryDomainAssociations(tx, modelKey)
		if err != nil {
			return err
		}
		// Group domain associations by their parent (problem domain) key.
		domainAssociationsMap := make(map[identity.Key][]model_domain.Association)
		for _, assoc := range domainAssociationsSlice {
			// The parent key is the problem domain key.
			parentKeyStr := assoc.Key.ParentKey()
			parentKey, parseErr := identity.ParseKey(parentKeyStr)
			if parseErr == nil {
				domainAssociationsMap[parentKey] = append(domainAssociationsMap[parentKey], assoc)
			}
		}

		// Generalizations - returned as slice, need to group by subdomain (parent) key.
		generalizationsSlice, err := QueryGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}
		generalizationsMap := make(map[identity.Key][]model_class.Generalization)
		for _, gen := range generalizationsSlice {
			// Extract parent (subdomain) key from the generalization key.
			parentKeyStr := gen.Key.ParentKey()
			parentKey, parseErr := identity.ParseKey(parentKeyStr)
			if parseErr == nil {
				generalizationsMap[parentKey] = append(generalizationsMap[parentKey], gen)
			}
		}

		// Classes grouped by subdomain key.
		classesMap, err := QueryClasses(tx, modelKey)
		if err != nil {
			return err
		}

		// Attributes grouped by class key.
		attributesMap, err := QueryAttributes(tx, modelKey)
		if err != nil {
			return err
		}

		// Class associations - all associations from DB as a flat slice.
		associationsSlice, err := QueryAssociations(tx, modelKey)
		if err != nil {
			return err
		}

		// Load all data types.
		dataTypes, err := LoadTopLevelDataTypes(tx, modelKey)
		if err != nil {
			return err
		}

		// States grouped by class key.
		statesMap, err := QueryStates(tx, modelKey)
		if err != nil {
			return err
		}

		// State actions grouped by state key.
		stateActionsMap, err := QueryStateActions(tx, modelKey)
		if err != nil {
			return err
		}

		// Events grouped by class key.
		eventsMap, err := QueryEvents(tx, modelKey)
		if err != nil {
			return err
		}

		// Guards grouped by class key.
		guardsMap, err := QueryGuards(tx, modelKey)
		if err != nil {
			return err
		}

		// Actions grouped by class key.
		actionsMap, err := QueryActions(tx, modelKey)
		if err != nil {
			return err
		}

		// Transitions grouped by class key.
		transitionsMap, err := QueryTransitions(tx, modelKey)
		if err != nil {
			return err
		}

		// Use cases grouped by subdomain key.
		subdomainKeysForUseCases, useCasesSlice, err := QueryUseCases(tx, modelKey)
		if err != nil {
			return err
		}
		useCasesMap := make(map[identity.Key][]model_use_case.UseCase)
		for _, uc := range useCasesSlice {
			subdomainKey := subdomainKeysForUseCases[uc.Key]
			useCasesMap[subdomainKey] = append(useCasesMap[subdomainKey], uc)
		}

		// Use case actors grouped by use case key.
		useCaseActorsMap, err := QueryUseCaseActors(tx, modelKey)
		if err != nil {
			return err
		}

		// Scenarios grouped by use case key.
		scenariosMap, err := QueryScenarios(tx, modelKey)
		if err != nil {
			return err
		}

		// Objects grouped by scenario key.
		objectsMap, err := QueryObjects(tx, modelKey)
		if err != nil {
			return err
		}

		// UseCaseShareds grouped by subdomain key (outer key is sea-level use case, inner is mud-level).
		useCaseSharedsMap, err := QueryUseCaseShareds(tx, modelKey)
		if err != nil {
			return err
		}

		// Now assemble the tree structure.
		model.Domains = make(map[identity.Key]model_domain.Domain)
		for _, domain := range domainsSlice {
			domainKey := domain.Key

			// Attach domain associations to domain.
			domain.DomainAssociations = make(map[identity.Key]model_domain.Association)
			if domainAssocs, ok := domainAssociationsMap[domainKey]; ok {
				for _, assoc := range domainAssocs {
					domain.DomainAssociations[assoc.Key] = assoc
				}
			}

			// Attach subdomains to domain.
			domain.Subdomains = make(map[identity.Key]model_domain.Subdomain)
			if subdomains, ok := subdomainsMap[domainKey]; ok {
				for _, subdomain := range subdomains {
					subdomainKey := subdomain.Key

					// Attach generalizations to subdomain.
					subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
					if generalizations, ok := generalizationsMap[subdomainKey]; ok {
						for _, gen := range generalizations {
							subdomain.Generalizations[gen.Key] = gen
						}
					}

					// Attach classes to subdomain.
					subdomain.Classes = make(map[identity.Key]model_class.Class)
					if classes, ok := classesMap[subdomainKey]; ok {
						for _, class := range classes {
							classKey := class.Key

							// Attach attributes to class.
							class.Attributes = make(map[identity.Key]model_class.Attribute)
							if attributes, ok := attributesMap[classKey]; ok {
								// Attach data types to attributes.
								for _, attr := range attributes {
									if dt, ok := dataTypes[attr.Key.String()]; ok {
										attr.DataType = &dt
									}
									class.Attributes[attr.Key] = attr
								}
							}

							// Attach states to class.
							class.States = make(map[identity.Key]model_state.State)
							if states, ok := statesMap[classKey]; ok {
								for _, state := range states {
									stateKey := state.Key
									// Attach state actions to state (remains as slice).
									if stateActions, ok := stateActionsMap[stateKey]; ok {
										state.Actions = stateActions
									}
									class.States[state.Key] = state
								}
							}

							// Attach events to class.
							class.Events = make(map[identity.Key]model_state.Event)
							if events, ok := eventsMap[classKey]; ok {
								for _, event := range events {
									class.Events[event.Key] = event
								}
							}

							// Attach guards to class.
							class.Guards = make(map[identity.Key]model_state.Guard)
							if guards, ok := guardsMap[classKey]; ok {
								for _, guard := range guards {
									class.Guards[guard.Key] = guard
								}
							}

							// Attach actions to class.
							class.Actions = make(map[identity.Key]model_state.Action)
							if actions, ok := actionsMap[classKey]; ok {
								for _, action := range actions {
									class.Actions[action.Key] = action
								}
							}

							// Attach transitions to class.
							class.Transitions = make(map[identity.Key]model_state.Transition)
							if transitions, ok := transitionsMap[classKey]; ok {
								for _, transition := range transitions {
									class.Transitions[transition.Key] = transition
								}
							}

							subdomain.Classes[class.Key] = class
						}
					}

					// Attach use cases to subdomain.
					subdomain.UseCases = make(map[identity.Key]model_use_case.UseCase)
					if useCases, ok := useCasesMap[subdomainKey]; ok {
						for _, useCase := range useCases {
							useCaseKey := useCase.Key

							// Attach use case actors.
							useCase.Actors = make(map[identity.Key]model_use_case.Actor)
							if actorsMap, ok := useCaseActorsMap[useCaseKey]; ok {
								for actorKey, actor := range actorsMap {
									useCase.Actors[actorKey] = actor
								}
							}

							// Attach scenarios to use case.
							useCase.Scenarios = make(map[identity.Key]model_scenario.Scenario)
							if scenarios, ok := scenariosMap[useCaseKey]; ok {
								for _, scenario := range scenarios {
									scenarioKey := scenario.Key
									// Attach objects to scenario.
									scenario.Objects = make(map[identity.Key]model_scenario.Object)
									if objects, ok := objectsMap[scenarioKey]; ok {
										for _, obj := range objects {
											scenario.Objects[obj.Key] = obj
										}
									}
									useCase.Scenarios[scenario.Key] = scenario
								}
							}

							subdomain.UseCases[useCase.Key] = useCase
						}
					}

					// Attach UseCaseShares to subdomain.
					subdomain.UseCaseShares = make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
					// UseCaseShareds are keyed by sea-level use case key.
					// We need to filter to those belonging to this subdomain.
					for seaLevelKey, mudLevelShares := range useCaseSharedsMap {
						// Check if the sea-level use case belongs to this subdomain.
						if _, exists := subdomain.UseCases[seaLevelKey]; exists {
							subdomain.UseCaseShares[seaLevelKey] = mudLevelShares
						}
					}

					domain.Subdomains[subdomain.Key] = subdomain
				}
			}

			model.Domains[domain.Key] = domain
		}

		// Class associations - use SetClassAssociations to route them to the correct level.
		allAssociations := make(map[identity.Key]model_class.Association)
		for _, assoc := range associationsSlice {
			allAssociations[assoc.Key] = assoc
		}
		if len(allAssociations) > 0 {
			if err = model.SetClassAssociations(allAssociations); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
