package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
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

		// Domains and their nested content.
		for _, domain := range model.Domains {
			domainKey := domain.Key.String()

			if err = AddDomain(tx, modelKey, domain); err != nil {
				return err
			}

			// Domain associations (at domain level).
			for _, association := range domain.Associations {
				if err = AddDomainAssociation(tx, modelKey, association); err != nil {
					return err
				}
			}

			// Subdomains.
			for _, subdomain := range domain.Subdomains {
				subdomainKey := subdomain.Key.String()

				if err = AddSubdomain(tx, modelKey, domainKey, subdomain); err != nil {
					return err
				}

				// Generalizations.
				for _, generalization := range subdomain.Generalizations {
					if err = AddGeneralization(tx, modelKey, generalization); err != nil {
						return err
					}
				}

				// Subdomain-level associations.
				for _, association := range subdomain.Associations {
					if err = AddAssociation(tx, modelKey, association); err != nil {
						return err
					}
				}

				// Classes.
				for _, class := range subdomain.Classes {
					classKey := class.Key.String()

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
							if err = AddClassIndex(tx, modelKey, classKey, attribute.Key.String(), indexNum); err != nil {
								return err
							}
						}
					}

					// Class associations.
					for _, association := range class.Associations {
						if err = AddAssociation(tx, modelKey, association); err != nil {
							return err
						}
					}

					// States.
					for _, state := range class.States {
						stateKey := state.Key.String()

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

					// Actions.
					for _, action := range class.Actions {
						if err = AddAction(tx, modelKey, classKey, action); err != nil {
							return err
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
					useCaseKey := useCase.Key.String()

					if err = AddUseCase(tx, modelKey, subdomainKey, useCase); err != nil {
						return err
					}

					// Use case actors.
					for actorKey, actor := range useCase.Actors {
						if err = AddUseCaseActor(tx, modelKey, useCaseKey, actorKey.String(), actor); err != nil {
							return err
						}
					}

					// Scenarios.
					for _, scenario := range useCase.Scenarios {
						scenarioKey := scenario.Key.String()

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
			}
		}

		// Model-level domain associations.
		for _, association := range model.DomainAssociations {
			if err = AddDomainAssociation(tx, modelKey, association); err != nil {
				return err
			}
		}

		// Model-level class associations (spanning domains).
		for _, association := range model.Associations {
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

		// Actors.
		model.Actors, err = QueryActors(tx, modelKey)
		if err != nil {
			return err
		}

		// Domains.
		domains, err := QueryDomains(tx, modelKey)
		if err != nil {
			return err
		}

		// Subdomains grouped by domain key.
		subdomainsMap, err := QuerySubdomains(tx, modelKey)
		if err != nil {
			return err
		}

		// Domain associations.
		model.DomainAssociations, err = QueryDomainAssociations(tx, modelKey)
		if err != nil {
			return err
		}

		// Generalizations - returned as slice, need to group by subdomain (parent) key.
		generalizationsSlice, err := QueryGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}
		generalizationsMap := make(map[string][]model_class.Generalization)
		for _, gen := range generalizationsSlice {
			// Extract parent (subdomain) key from the generalization key.
			parentKey := gen.Key.ParentKey()
			generalizationsMap[parentKey] = append(generalizationsMap[parentKey], gen)
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

		// Class associations.
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
		useCasesMap := make(map[string][]model_use_case.UseCase)
		for _, uc := range useCasesSlice {
			subdomainKey := subdomainKeysForUseCases[uc.Key.String()]
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

		// Now assemble the tree structure.
		for i := range domains {
			domainKey := domains[i].Key.String()

			// Attach subdomains to domain.
			if subdomains, ok := subdomainsMap[domainKey]; ok {
				for j := range subdomains {
					subdomainKey := subdomains[j].Key.String()

					// Attach generalizations to subdomain.
					if generalizations, ok := generalizationsMap[subdomainKey]; ok {
						subdomains[j].Generalizations = generalizations
					}

					// Attach classes to subdomain.
					if classes, ok := classesMap[subdomainKey]; ok {
						for k := range classes {
							classKey := classes[k].Key.String()

							// Attach attributes to class.
							if attributes, ok := attributesMap[classKey]; ok {
								// Attach data types to attributes.
								for l := range attributes {
									if dt, ok := dataTypes[attributes[l].Key.String()]; ok {
										attributes[l].DataType = &dt
									}
								}
								classes[k].Attributes = attributes
							}

							// Attach states to class.
							if states, ok := statesMap[classKey]; ok {
								for m := range states {
									stateKey := states[m].Key.String()
									// Attach state actions to state.
									if stateActions, ok := stateActionsMap[stateKey]; ok {
										states[m].Actions = stateActions
									}
								}
								classes[k].States = states
							}

							// Attach events to class.
							if events, ok := eventsMap[classKey]; ok {
								classes[k].Events = events
							}

							// Attach guards to class.
							if guards, ok := guardsMap[classKey]; ok {
								classes[k].Guards = guards
							}

							// Attach actions to class.
							if actions, ok := actionsMap[classKey]; ok {
								classes[k].Actions = actions
							}

							// Attach transitions to class.
							if transitions, ok := transitionsMap[classKey]; ok {
								classes[k].Transitions = transitions
							}
						}
						subdomains[j].Classes = classes
					}

					// Attach use cases to subdomain.
					if useCases, ok := useCasesMap[subdomainKey]; ok {
						for k := range useCases {
							useCaseKey := useCases[k].Key.String()

							// Attach use case actors - convert string keys to identity.Key.
							if actorsStringMap, ok := useCaseActorsMap[useCaseKey]; ok {
								actorsKeyMap := make(map[identity.Key]model_use_case.Actor)
								for actorKeyStr, actor := range actorsStringMap {
									actorKey, parseErr := identity.ParseKey(actorKeyStr)
									if parseErr == nil {
										actorsKeyMap[actorKey] = actor
									}
								}
								useCases[k].Actors = actorsKeyMap
							}

							// Attach scenarios to use case.
							if scenarios, ok := scenariosMap[useCaseKey]; ok {
								for m := range scenarios {
									scenarioKey := scenarios[m].Key.String()
									// Attach objects to scenario.
									if objects, ok := objectsMap[scenarioKey]; ok {
										scenarios[m].Objects = objects
									}
								}
								useCases[k].Scenarios = scenarios
							}
						}
						subdomains[j].UseCases = useCases
					}
				}
				domains[i].Subdomains = subdomains
			}
		}

		model.Domains = domains

		// Model-level associations - these are associations that span domains.
		model.Associations = associationsSlice

		return nil
	})
	if err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
