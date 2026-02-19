package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

func WriteModel(db *sql.DB, model req_model.Model) (err error) {

	// Validate the model tree before writing to database.
	if err = model.Validate(); err != nil {
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

		// Collect all logic rows to insert.
		allLogics := make([]model_logic.Logic, 0, len(model.Invariants)+len(model.GlobalFunctions))
		allLogics = append(allLogics, model.Invariants...)
		for _, gf := range model.GlobalFunctions {
			allLogics = append(allLogics, gf.Specification)
		}
		if err = AddLogics(tx, modelKey, allLogics); err != nil {
			return err
		}

		// Add invariant join rows.
		invariantKeys := make([]identity.Key, len(model.Invariants))
		for i, inv := range model.Invariants {
			invariantKeys[i] = inv.Key
		}
		if err = AddInvariants(tx, modelKey, invariantKeys); err != nil {
			return err
		}

		// Add global function rows.
		gfSlice := make([]model_logic.GlobalFunction, 0, len(model.GlobalFunctions))
		for _, gf := range model.GlobalFunctions {
			gfSlice = append(gfSlice, gf)
		}
		if err = AddGlobalFunctions(tx, modelKey, gfSlice); err != nil {
			return err
		}

		/*
		// Collect actors into a slice.
		actorsSlice := make([]model_actor.Actor, 0, len(model.Actors))
		for _, actor := range model.Actors {
			actorsSlice = append(actorsSlice, actor)
		}
		if err = AddActors(tx, modelKey, actorsSlice); err != nil {
			return err
		}

		// Collect domains into a slice.
		domainsSlice := make([]model_domain.Domain, 0, len(model.Domains))
		for _, domain := range model.Domains {
			domainsSlice = append(domainsSlice, domain)
		}
		if err = AddDomains(tx, modelKey, domainsSlice); err != nil {
			return err
		}

		// Collect domain associations (after all domains exist).
		// Domain associations are only at the model level.
		domainAssociationsSlice := make([]model_domain.Association, 0, len(model.DomainAssociations))
		for _, association := range model.DomainAssociations {
			domainAssociationsSlice = append(domainAssociationsSlice, association)
		}
		if err = AddDomainAssociations(tx, modelKey, domainAssociationsSlice); err != nil {
			return err
		}

		// Collect subdomains, classes, and other nested content into bulk structures.
		subdomainsMap := make(map[identity.Key][]model_domain.Subdomain)
		generalizationsSlice := make([]model_class.Generalization, 0)
		classesMap := make(map[identity.Key][]model_class.Class)
		attributesMap := make(map[identity.Key][]model_class.Attribute)
		eventsMap := make(map[identity.Key][]model_state.Event)
		guardsMap := make(map[identity.Key][]model_state.Guard)
		actionsMap := make(map[identity.Key][]model_state.Action)
		statesMap := make(map[identity.Key][]model_state.State)
		stateActionsMap := make(map[identity.Key][]model_state.StateAction)
		transitionsMap := make(map[identity.Key][]model_state.Transition)
		useCaseSubdomainKeys := make(map[identity.Key]identity.Key) // useCaseKey -> subdomainKey
		useCasesSlice := make([]model_use_case.UseCase, 0)
		useCaseActorsMap := make(map[identity.Key]map[identity.Key]model_use_case.Actor)
		scenariosMap := make(map[identity.Key][]model_scenario.Scenario)
		objectsMap := make(map[identity.Key][]model_scenario.Object)
		useCaseSharedsMap := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
		dataTypes := make(map[string]model_data_type.DataType)

		for _, domain := range model.Domains {
			domainKey := domain.Key

			// Collect subdomains.
			for _, subdomain := range domain.Subdomains {
				subdomainKey := subdomain.Key
				subdomainsMap[domainKey] = append(subdomainsMap[domainKey], subdomain)

				// Collect generalizations.
				for _, generalization := range subdomain.Generalizations {
					generalizationsSlice = append(generalizationsSlice, generalization)
				}

				// Collect classes.
				for _, class := range subdomain.Classes {
					classKey := class.Key
					classesMap[subdomainKey] = append(classesMap[subdomainKey], class)

					// Collect attributes.
					for _, attribute := range class.Attributes {
						attributesMap[classKey] = append(attributesMap[classKey], attribute)
						// Collect data types.
						if attribute.DataType != nil {
							dataTypes[attribute.DataType.Key] = *attribute.DataType
						}
					}

					// Collect events.
					for _, event := range class.Events {
						eventsMap[classKey] = append(eventsMap[classKey], event)
					}

					// Collect guards.
					for _, guard := range class.Guards {
						guardsMap[classKey] = append(guardsMap[classKey], guard)
					}

					// Collect actions.
					for _, action := range class.Actions {
						actionsMap[classKey] = append(actionsMap[classKey], action)
					}

					// Collect states and state actions.
					for _, state := range class.States {
						stateKey := state.Key
						statesMap[classKey] = append(statesMap[classKey], state)

						// Collect state actions.
						for _, stateAction := range state.Actions {
							stateActionsMap[stateKey] = append(stateActionsMap[stateKey], stateAction)
						}
					}

					// Collect transitions.
					for _, transition := range class.Transitions {
						transitionsMap[classKey] = append(transitionsMap[classKey], transition)
					}
				}

				// Collect use cases.
				for _, useCase := range subdomain.UseCases {
					useCaseKey := useCase.Key
					useCaseSubdomainKeys[useCaseKey] = subdomainKey
					useCasesSlice = append(useCasesSlice, useCase)

					// Collect use case actors.
					if len(useCase.Actors) > 0 {
						useCaseActorsMap[useCaseKey] = useCase.Actors
					}

					// Collect scenarios.
					for _, scenario := range useCase.Scenarios {
						scenarioKey := scenario.Key
						scenariosMap[useCaseKey] = append(scenariosMap[useCaseKey], scenario)

						// Collect objects.
						for _, object := range scenario.Objects {
							objectsMap[scenarioKey] = append(objectsMap[scenarioKey], object)
						}
					}
				}

				// Collect UseCaseShares.
				for seaLevelKey, mudLevelShares := range subdomain.UseCaseShares {
					if useCaseSharedsMap[seaLevelKey] == nil {
						useCaseSharedsMap[seaLevelKey] = make(map[identity.Key]model_use_case.UseCaseShared)
					}
					for mudLevelKey, shared := range mudLevelShares {
						useCaseSharedsMap[seaLevelKey][mudLevelKey] = shared
					}
				}
			}
		}

		// Bulk insert subdomains.
		if err = AddSubdomains(tx, modelKey, subdomainsMap); err != nil {
			return err
		}

		// Bulk insert generalizations.
		if err = AddGeneralizations(tx, modelKey, generalizationsSlice); err != nil {
			return err
		}

		// Bulk insert classes.
		if err = AddClasses(tx, modelKey, classesMap); err != nil {
			return err
		}

		// Bulk insert attributes.
		if err = AddAttributes(tx, modelKey, attributesMap); err != nil {
			return err
		}

		// Bulk insert class indexes (must be done individually since we need attribute.IndexNums).
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					classKey := class.Key
					for _, attribute := range class.Attributes {
						for _, indexNum := range attribute.IndexNums {
							if err = AddClassIndex(tx, modelKey, classKey, attribute.Key, indexNum); err != nil {
								return err
							}
						}
					}
				}
			}
		}

		// Bulk insert events.
		if err = AddEvents(tx, modelKey, eventsMap); err != nil {
			return err
		}

		// Bulk insert guards.
		if err = AddGuards(tx, modelKey, guardsMap); err != nil {
			return err
		}

		// Bulk insert actions (must be added before states with state actions).
		if err = AddActions(tx, modelKey, actionsMap); err != nil {
			return err
		}

		// Bulk insert states.
		if err = AddStates(tx, modelKey, statesMap); err != nil {
			return err
		}

		// Bulk insert state actions.
		if err = AddStateActions(tx, modelKey, stateActionsMap); err != nil {
			return err
		}

		// Bulk insert transitions.
		if err = AddTransitions(tx, modelKey, transitionsMap); err != nil {
			return err
		}

		// Bulk insert use cases.
		if err = AddUseCases(tx, modelKey, useCaseSubdomainKeys, useCasesSlice); err != nil {
			return err
		}

		// Bulk insert use case actors.
		if err = AddUseCaseActors(tx, modelKey, useCaseActorsMap); err != nil {
			return err
		}

		// Bulk insert scenarios.
		if err = AddScenarios(tx, modelKey, scenariosMap); err != nil {
			return err
		}

		// Bulk insert objects.
		if err = AddObjects(tx, modelKey, objectsMap); err != nil {
			return err
		}

		// Bulk insert use case shareds.
		if err = AddUseCaseShareds(tx, modelKey, useCaseSharedsMap); err != nil {
			return err
		}

		// Bulk insert class associations.
		classAssociations := model.GetClassAssociations()
		associationsSlice := make([]model_class.Association, 0, len(classAssociations))
		for _, association := range classAssociations {
			associationsSlice = append(associationsSlice, association)
		}
		if err = AddAssociations(tx, modelKey, associationsSlice); err != nil {
			return err
		}

		// Bulk insert data types.
		if err = AddTopLevelDataTypes(tx, modelKey, dataTypes); err != nil {
			return err
		}
		*/

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

		// Logics.
		logics, err := QueryLogics(tx, modelKey)
		if err != nil {
			return err
		}
		logicsByKey := make(map[identity.Key]model_logic.Logic, len(logics))
		for _, logic := range logics {
			logicsByKey[logic.Key] = logic
		}

		// Invariants — stitch logic data onto invariant keys.
		invariantKeys, err := QueryInvariants(tx, modelKey)
		if err != nil {
			return err
		}
		model.Invariants = make([]model_logic.Logic, len(invariantKeys))
		for i, key := range invariantKeys {
			model.Invariants[i] = logicsByKey[key]
		}

		// Global functions — stitch logic data onto global function rows.
		gfs, err := QueryGlobalFunctions(tx, modelKey)
		if err != nil {
			return err
		}
		if len(gfs) > 0 {
			model.GlobalFunctions = make(map[identity.Key]model_logic.GlobalFunction, len(gfs))
			for _, gf := range gfs {
				gf.Specification = logicsByKey[gf.Key]
				model.GlobalFunctions[gf.Key] = gf
			}
		}

		/*
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

		// Domain associations - returns slice (they are model-level, not domain-level).
		domainAssociationsSlice, err := QueryDomainAssociations(tx, modelKey)
		if err != nil {
			return err
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

		// Attach domain associations to the model (they are model-level, not domain-level).
		model.DomainAssociations = make(map[identity.Key]model_domain.Association)
		for _, assoc := range domainAssociationsSlice {
			model.DomainAssociations[assoc.Key] = assoc
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
		*/

		return nil
	})
	if err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
