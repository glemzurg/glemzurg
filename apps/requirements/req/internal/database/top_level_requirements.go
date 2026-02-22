package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
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
			allLogics = append(allLogics, gf.Logic)
		}
		// Collect derivation policy logics from attributes.
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, attr := range class.Attributes {
						if attr.DerivationPolicy != nil {
							allLogics = append(allLogics, *attr.DerivationPolicy)
						}
					}
				}
			}
		}
		// Collect guard logics (guard key is also the logic key).
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, guard := range class.Guards {
						allLogics = append(allLogics, guard.Logic)
					}
				}
			}
		}
		// Collect action require, guarantee, and safety logics.
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, action := range class.Actions {
						allLogics = append(allLogics, action.Requires...)
						allLogics = append(allLogics, action.Guarantees...)
						allLogics = append(allLogics, action.SafetyRules...)
					}
				}
			}
		}
		// Collect query require and guarantee logics.
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, query := range class.Queries {
						allLogics = append(allLogics, query.Requires...)
						allLogics = append(allLogics, query.Guarantees...)
					}
				}
			}
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

		// Collect actor generalizations into a slice (must be inserted before actors due to FK).
		actorGeneralizationsSlice := make([]model_actor.Generalization, 0, len(model.ActorGeneralizations))
		for _, ag := range model.ActorGeneralizations {
			actorGeneralizationsSlice = append(actorGeneralizationsSlice, ag)
		}
		if err = AddActorGeneralizations(tx, modelKey, actorGeneralizationsSlice); err != nil {
			return err
		}

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

		// Collect subdomains, generalizations, use case generalizations, classes, and attributes into bulk structures.
		subdomainsMap := make(map[identity.Key][]model_domain.Subdomain)
		generalizationsMap := make(map[identity.Key][]model_class.Generalization)
		useCaseGeneralizationsMap := make(map[identity.Key][]model_use_case.Generalization)
		classesMap := make(map[identity.Key][]model_class.Class)
		attributesMap := make(map[identity.Key][]model_class.Attribute)

		for _, domain := range model.Domains {
			domainKey := domain.Key

			// Collect subdomains.
			for _, subdomain := range domain.Subdomains {
				subdomainKey := subdomain.Key
				subdomainsMap[domainKey] = append(subdomainsMap[domainKey], subdomain)

				// Collect generalizations.
				for _, generalization := range subdomain.Generalizations {
					generalizationsMap[subdomainKey] = append(generalizationsMap[subdomainKey], generalization)
				}

				// Collect use case generalizations.
				for _, ucGen := range subdomain.UseCaseGeneralizations {
					useCaseGeneralizationsMap[subdomainKey] = append(useCaseGeneralizationsMap[subdomainKey], ucGen)
				}

				// Collect classes.
				for _, class := range subdomain.Classes {
					classKey := class.Key
					classesMap[subdomainKey] = append(classesMap[subdomainKey], class)

					// Collect attributes.
					for _, attribute := range class.Attributes {
						attributesMap[classKey] = append(attributesMap[classKey], attribute)
					}
				}
			}
		}

		// Bulk insert subdomains.
		if err = AddSubdomains(tx, modelKey, subdomainsMap); err != nil {
			return err
		}

		// Bulk insert generalizations.
		if err = AddGeneralizations(tx, modelKey, generalizationsMap); err != nil {
			return err
		}

		// Bulk insert use case generalizations.
		if err = AddUseCaseGeneralizations(tx, modelKey, useCaseGeneralizationsMap); err != nil {
			return err
		}

		// Collect use cases from subdomains (must be inserted after use_case_generalization due to FK).
		useCaseSubdomainKeys := make(map[identity.Key]identity.Key) // useCaseKey -> subdomainKey
		useCasesSlice := make([]model_use_case.UseCase, 0)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, uc := range subdomain.UseCases {
					useCaseSubdomainKeys[uc.Key] = subdomain.Key
					useCasesSlice = append(useCasesSlice, uc)
				}
			}
		}
		if err = AddUseCases(tx, modelKey, useCaseSubdomainKeys, useCasesSlice); err != nil {
			return err
		}

		// Bulk insert classes.
		if err = AddClasses(tx, modelKey, classesMap); err != nil {
			return err
		}

		// Collect data types from attributes and query parameters (must be inserted before attributes/query_parameter due to FK).
		dataTypes := make(map[string]model_data_type.DataType)
		for _, attrs := range attributesMap {
			for _, attr := range attrs {
				if attr.DataType != nil {
					dataTypes[attr.DataType.Key] = *attr.DataType
				}
			}
		}
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, query := range class.Queries {
						for _, param := range query.Parameters {
							if param.DataType != nil {
								dataTypes[param.DataType.Key] = *param.DataType
							}
						}
					}
					for _, event := range class.Events {
						for _, param := range event.Parameters {
							if param.DataType != nil {
								dataTypes[param.DataType.Key] = *param.DataType
							}
						}
					}
					for _, action := range class.Actions {
						for _, param := range action.Parameters {
							if param.DataType != nil {
								dataTypes[param.DataType.Key] = *param.DataType
							}
						}
					}
				}
			}
		}
		if err = AddTopLevelDataTypes(tx, modelKey, dataTypes); err != nil {
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

		// Collect class associations from subdomains and model level.
		var allClassAssociations []model_class.Association
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, assoc := range subdomain.ClassAssociations {
					allClassAssociations = append(allClassAssociations, assoc)
				}
			}
		}
		for _, assoc := range model.ClassAssociations {
			allClassAssociations = append(allClassAssociations, assoc)
		}
		if err = AddAssociations(tx, modelKey, allClassAssociations); err != nil {
			return err
		}

		// Collect states from classes.
		statesMap := make(map[identity.Key][]model_state.State)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, state := range class.States {
						statesMap[class.Key] = append(statesMap[class.Key], state)
					}
				}
			}
		}
		if err = AddStates(tx, modelKey, statesMap); err != nil {
			return err
		}

		// Collect guards from classes.
		guardsMap := make(map[identity.Key][]model_state.Guard)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, guard := range class.Guards {
						guardsMap[class.Key] = append(guardsMap[class.Key], guard)
					}
				}
			}
		}
		if err = AddGuards(tx, modelKey, guardsMap); err != nil {
			return err
		}

		// Collect actions from classes.
		actionsMap := make(map[identity.Key][]model_state.Action)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, action := range class.Actions {
						actionsMap[class.Key] = append(actionsMap[class.Key], action)
					}
				}
			}
		}
		if err = AddActions(tx, modelKey, actionsMap); err != nil {
			return err
		}

		// Collect action parameters from actions (must be inserted after actions due to FK).
		actionParamsMap := make(map[identity.Key][]model_state.Parameter)
		for _, actionList := range actionsMap {
			for _, action := range actionList {
				for _, param := range action.Parameters {
					actionParamsMap[action.Key] = append(actionParamsMap[action.Key], param)
				}
			}
		}
		if err = AddActionParameters(tx, modelKey, actionParamsMap); err != nil {
			return err
		}

		// Collect action require join rows from actions.
		actionRequiresMap := make(map[identity.Key][]identity.Key)
		for _, actionList := range actionsMap {
			for _, action := range actionList {
				for _, req := range action.Requires {
					actionRequiresMap[action.Key] = append(actionRequiresMap[action.Key], req.Key)
				}
			}
		}
		if err = AddActionRequires(tx, modelKey, actionRequiresMap); err != nil {
			return err
		}

		// Collect action guarantee join rows from actions.
		actionGuaranteesMap := make(map[identity.Key][]identity.Key)
		for _, actionList := range actionsMap {
			for _, action := range actionList {
				for _, guar := range action.Guarantees {
					actionGuaranteesMap[action.Key] = append(actionGuaranteesMap[action.Key], guar.Key)
				}
			}
		}
		if err = AddActionGuarantees(tx, modelKey, actionGuaranteesMap); err != nil {
			return err
		}

		// Collect action safety join rows from actions.
		actionSafetiesMap := make(map[identity.Key][]identity.Key)
		for _, actionList := range actionsMap {
			for _, action := range actionList {
				for _, rule := range action.SafetyRules {
					actionSafetiesMap[action.Key] = append(actionSafetiesMap[action.Key], rule.Key)
				}
			}
		}
		if err = AddActionSafeties(tx, modelKey, actionSafetiesMap); err != nil {
			return err
		}

		// Collect events from classes.
		eventsMap := make(map[identity.Key][]model_state.Event)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, event := range class.Events {
						eventsMap[class.Key] = append(eventsMap[class.Key], event)
					}
				}
			}
		}
		if err = AddEvents(tx, modelKey, eventsMap); err != nil {
			return err
		}

		// Collect queries from classes.
		queriesMap := make(map[identity.Key][]model_state.Query)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, query := range class.Queries {
						queriesMap[class.Key] = append(queriesMap[class.Key], query)
					}
				}
			}
		}
		if err = AddQueries(tx, modelKey, queriesMap); err != nil {
			return err
		}

		// Collect query parameters from queries (must be inserted after queries due to FK).
		queryParamsMap := make(map[identity.Key][]model_state.Parameter)
		for _, queryList := range queriesMap {
			for _, query := range queryList {
				for _, param := range query.Parameters {
					queryParamsMap[query.Key] = append(queryParamsMap[query.Key], param)
				}
			}
		}
		if err = AddQueryParameters(tx, modelKey, queryParamsMap); err != nil {
			return err
		}

		// Collect query require join rows from queries.
		queryRequiresMap := make(map[identity.Key][]identity.Key)
		for _, queryList := range queriesMap {
			for _, query := range queryList {
				for _, req := range query.Requires {
					queryRequiresMap[query.Key] = append(queryRequiresMap[query.Key], req.Key)
				}
			}
		}
		if err = AddQueryRequires(tx, modelKey, queryRequiresMap); err != nil {
			return err
		}

		// Collect query guarantee join rows from queries.
		queryGuaranteesMap := make(map[identity.Key][]identity.Key)
		for _, queryList := range queriesMap {
			for _, query := range queryList {
				for _, guar := range query.Guarantees {
					queryGuaranteesMap[query.Key] = append(queryGuaranteesMap[query.Key], guar.Key)
				}
			}
		}
		if err = AddQueryGuarantees(tx, modelKey, queryGuaranteesMap); err != nil {
			return err
		}

		// Collect event parameters from events (must be inserted after events due to FK).
		eventParamsMap := make(map[identity.Key][]model_state.Parameter)
		for _, eventList := range eventsMap {
			for _, event := range eventList {
				for _, param := range event.Parameters {
					eventParamsMap[event.Key] = append(eventParamsMap[event.Key], param)
				}
			}
		}
		if err = AddEventParameters(tx, modelKey, eventParamsMap); err != nil {
			return err
		}

		// Collect state actions from states (must be inserted after states and actions due to FK).
		stateActionsMap := make(map[identity.Key][]model_state.StateAction)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, state := range class.States {
						if len(state.Actions) > 0 {
							stateActionsMap[state.Key] = append(stateActionsMap[state.Key], state.Actions...)
						}
					}
				}
			}
		}
		if err = AddStateActions(tx, modelKey, stateActionsMap); err != nil {
			return err
		}

		// Collect transitions from classes (must be inserted after states, events, guards, and actions due to FK).
		transitionsMap := make(map[identity.Key][]model_state.Transition)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					for _, transition := range class.Transitions {
						transitionsMap[class.Key] = append(transitionsMap[class.Key], transition)
					}
				}
			}
		}
		if err = AddTransitions(tx, modelKey, transitionsMap); err != nil {
			return err
		}

		// Collect use case actors from use cases (must be inserted after use cases and classes due to FK).
		useCaseActorsMap := make(map[identity.Key]map[identity.Key]model_use_case.Actor)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, uc := range subdomain.UseCases {
					if len(uc.Actors) > 0 {
						useCaseActorsMap[uc.Key] = uc.Actors
					}
				}
			}
		}
		if err = AddUseCaseActors(tx, modelKey, useCaseActorsMap); err != nil {
			return err
		}

		// Collect use case shared entries from subdomains (must be inserted after use cases due to FK).
		useCaseSharedsMap := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for seaKey, mudMap := range subdomain.UseCaseShares {
					useCaseSharedsMap[seaKey] = mudMap
				}
			}
		}
		if err = AddUseCaseShareds(tx, modelKey, useCaseSharedsMap); err != nil {
			return err
		}

		// Collect scenarios from use cases (must be inserted after use cases due to FK).
		scenariosMap := make(map[identity.Key][]model_scenario.Scenario)
		objectsMap := make(map[identity.Key][]model_scenario.Object)
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, uc := range subdomain.UseCases {
					for _, scenario := range uc.Scenarios {
						scenariosMap[uc.Key] = append(scenariosMap[uc.Key], scenario)
						// Collect objects from this scenario.
						for _, obj := range scenario.Objects {
							objectsMap[scenario.Key] = append(objectsMap[scenario.Key], obj)
						}
					}
				}
			}
		}
		if err = AddScenarios(tx, modelKey, scenariosMap); err != nil {
			return err
		}

		// Bulk insert scenario objects (must be inserted after scenarios due to FK).
		if err = AddObjects(tx, modelKey, objectsMap); err != nil {
			return err
		}

		// Collect steps from scenarios and flatten for bulk insert (must be after scenarios and objects due to FK).
		var allStepRows []stepRow
		for _, domain := range model.Domains {
			for _, subdomain := range domain.Subdomains {
				for _, uc := range subdomain.UseCases {
					for _, scenario := range uc.Scenarios {
						if scenario.Steps != nil {
							allStepRows = append(allStepRows, flattenSteps(scenario.Key, scenario.Steps)...)
						}
					}
				}
			}
		}
		if err = AddSteps(tx, modelKey, allStepRows); err != nil {
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
				gf.Logic = logicsByKey[gf.Key]
				model.GlobalFunctions[gf.Key] = gf
			}
		}

		// Actor generalizations - returns slice, convert to map.
		actorGeneralizationsSlice, err := QueryActorGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}
		if len(actorGeneralizationsSlice) > 0 {
			model.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization, len(actorGeneralizationsSlice))
			for _, ag := range actorGeneralizationsSlice {
				model.ActorGeneralizations[ag.Key] = ag
			}
		}

		// Actors - returns slice, convert to map.
		actorsSlice, err := QueryActors(tx, modelKey)
		if err != nil {
			return err
		}
		if len(actorsSlice) > 0 {
			model.Actors = make(map[identity.Key]model_actor.Actor)
			for _, actor := range actorsSlice {
				model.Actors[actor.Key] = actor
			}
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

		// Generalizations grouped by subdomain key.
		generalizationsMap, err := QueryGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}

		// Use case generalizations grouped by subdomain key.
		useCaseGeneralizationsMap, err := QueryUseCaseGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}

		// Use cases - returns subdomainKeys map and slice.
		useCaseSubdomainKeys, useCasesSlice, err := QueryUseCases(tx, modelKey)
		if err != nil {
			return err
		}

		// Use case actors grouped by use case key -> actor class key -> Actor.
		useCaseActorsMap, err := QueryUseCaseActors(tx, modelKey)
		if err != nil {
			return err
		}

		// Use case shared entries grouped by sea-level key -> mud-level key -> UseCaseShared.
		useCaseSharedsMap, err := QueryUseCaseShareds(tx, modelKey)
		if err != nil {
			return err
		}

		// Scenarios grouped by use case key.
		scenariosMap, err := QueryScenarios(tx, modelKey)
		if err != nil {
			return err
		}

		// Scenario objects grouped by scenario key.
		scenarioObjectsMap, err := QueryObjects(tx, modelKey)
		if err != nil {
			return err
		}

		// Steps grouped by scenario key (reconstructed trees).
		stepsMap, err := QuerySteps(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch objects and steps onto scenarios.
		for useCaseKey, scenList := range scenariosMap {
			for i, scenario := range scenList {
				if objs, ok := scenarioObjectsMap[scenario.Key]; ok {
					scenList[i].Objects = make(map[identity.Key]model_scenario.Object, len(objs))
					for _, obj := range objs {
						scenList[i].Objects[obj.Key] = obj
					}
				}
				if rootStep, ok := stepsMap[scenario.Key]; ok {
					scenList[i].Steps = rootStep
				}
			}
			scenariosMap[useCaseKey] = scenList
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

		// Guards grouped by class key.
		guardsMap, err := QueryGuards(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch logic data onto guards.
		for classKey, guards := range guardsMap {
			for i, guard := range guards {
				if logic, ok := logicsByKey[guard.Key]; ok {
					guards[i].Logic = logic
				}
			}
			guardsMap[classKey] = guards
		}

		// Actions grouped by class key.
		actionsMap, err := QueryActions(tx, modelKey)
		if err != nil {
			return err
		}

		// Action parameters grouped by action key.
		actionParamsMap, err := QueryActionParameters(tx, modelKey)
		if err != nil {
			return err
		}

		// Action require join rows grouped by action key.
		actionRequiresMap, err := QueryActionRequires(tx, modelKey)
		if err != nil {
			return err
		}

		// Action guarantee join rows grouped by action key.
		actionGuaranteesMap, err := QueryActionGuarantees(tx, modelKey)
		if err != nil {
			return err
		}

		// Action safety join rows grouped by action key.
		actionSafetiesMap, err := QueryActionSafeties(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch parameters, requires, guarantees, and safety rules onto actions.
		for classKey, actions := range actionsMap {
			for i, action := range actions {
				if params, ok := actionParamsMap[action.Key]; ok {
					actions[i].Parameters = params
				}
				// Stitch requires from logic data.
				if reqKeys, ok := actionRequiresMap[action.Key]; ok {
					actions[i].Requires = make([]model_logic.Logic, len(reqKeys))
					for j, key := range reqKeys {
						actions[i].Requires[j] = logicsByKey[key]
					}
				}
				// Stitch guarantees from logic data.
				if guarKeys, ok := actionGuaranteesMap[action.Key]; ok {
					actions[i].Guarantees = make([]model_logic.Logic, len(guarKeys))
					for j, key := range guarKeys {
						actions[i].Guarantees[j] = logicsByKey[key]
					}
				}
				// Stitch safety rules from logic data.
				if safetyKeys, ok := actionSafetiesMap[action.Key]; ok {
					actions[i].SafetyRules = make([]model_logic.Logic, len(safetyKeys))
					for j, key := range safetyKeys {
						actions[i].SafetyRules[j] = logicsByKey[key]
					}
				}
			}
			actionsMap[classKey] = actions
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

		// Stitch state actions onto states.
		for classKey, states := range statesMap {
			for i, state := range states {
				if stateActions, ok := stateActionsMap[state.Key]; ok {
					states[i].Actions = stateActions
				}
			}
			statesMap[classKey] = states
		}

		// Transitions grouped by class key.
		transitionsMap, err := QueryTransitions(tx, modelKey)
		if err != nil {
			return err
		}

		// Events grouped by class key.
		eventsMap, err := QueryEvents(tx, modelKey)
		if err != nil {
			return err
		}

		// Event parameters grouped by event key.
		eventParamsMap, err := QueryEventParameters(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch parameters onto events.
		for classKey, events := range eventsMap {
			for i, event := range events {
				if params, ok := eventParamsMap[event.Key]; ok {
					events[i].Parameters = params
				}
			}
			eventsMap[classKey] = events
		}

		// Queries grouped by class key.
		queriesMap, err := QueryQueries(tx, modelKey)
		if err != nil {
			return err
		}

		// Query parameters grouped by query key.
		queryParamsMap, err := QueryQueryParameters(tx, modelKey)
		if err != nil {
			return err
		}

		// Query require join rows grouped by query key.
		queryRequiresMap, err := QueryQueryRequires(tx, modelKey)
		if err != nil {
			return err
		}

		// Query guarantee join rows grouped by query key.
		queryGuaranteesMap, err := QueryQueryGuarantees(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch parameters, requires, and guarantees onto queries (data types are stitched onto parameters below after dataTypes are loaded).
		for classKey, queries := range queriesMap {
			for i, query := range queries {
				if params, ok := queryParamsMap[query.Key]; ok {
					queries[i].Parameters = params
				}
				// Stitch requires from logic data.
				if reqKeys, ok := queryRequiresMap[query.Key]; ok {
					queries[i].Requires = make([]model_logic.Logic, len(reqKeys))
					for j, key := range reqKeys {
						queries[i].Requires[j] = logicsByKey[key]
					}
				}
				// Stitch guarantees from logic data.
				if guarKeys, ok := queryGuaranteesMap[query.Key]; ok {
					queries[i].Guarantees = make([]model_logic.Logic, len(guarKeys))
					for j, key := range guarKeys {
						queries[i].Guarantees[j] = logicsByKey[key]
					}
				}
			}
			queriesMap[classKey] = queries
		}

		// Load data types for stitching onto attributes.
		dataTypes, err := LoadTopLevelDataTypes(tx, modelKey)
		if err != nil {
			return err
		}

		// Stitch derivation policy logics, data types, and class indexes onto attributes.
		for classKey, attrs := range attributesMap {
			for i, attr := range attrs {
				// Stitch derivation policy from logics table.
				if attr.DerivationPolicy != nil {
					logic := logicsByKey[attr.DerivationPolicy.Key]
					attrs[i].DerivationPolicy = &logic
				}
				// Stitch data type from data types table.
				if dt, ok := dataTypes[attr.Key.String()]; ok {
					attrs[i].DataType = &dt
				}
				// Load class indexes for this attribute.
				indexNums, err := LoadClassAttributeIndexes(tx, modelKey, classKey, attr.Key)
				if err != nil {
					return err
				}
				attrs[i].IndexNums = indexNums
			}
			attributesMap[classKey] = attrs
		}

		// Stitch data types onto query parameters.
		for classKey, queries := range queriesMap {
			for i, query := range queries {
				for j, param := range query.Parameters {
					if param.DataType != nil {
						if dt, ok := dataTypes[param.DataType.Key]; ok {
							queries[i].Parameters[j].DataType = &dt
						}
					}
				}
			}
			queriesMap[classKey] = queries
		}

		// Stitch data types onto event parameters.
		for classKey, events := range eventsMap {
			for i, event := range events {
				for j, param := range event.Parameters {
					if param.DataType != nil {
						if dt, ok := dataTypes[param.DataType.Key]; ok {
							events[i].Parameters[j].DataType = &dt
						}
					}
				}
			}
			eventsMap[classKey] = events
		}

		// Stitch data types onto action parameters.
		for classKey, actions := range actionsMap {
			for i, action := range actions {
				for j, param := range action.Parameters {
					if param.DataType != nil {
						if dt, ok := dataTypes[param.DataType.Key]; ok {
							actions[i].Parameters[j].DataType = &dt
						}
					}
				}
			}
			actionsMap[classKey] = actions
		}

		// Now assemble the tree structure.
		if len(domainsSlice) > 0 {
			model.Domains = make(map[identity.Key]model_domain.Domain)
			for _, domain := range domainsSlice {
				domainKey := domain.Key

				// Attach subdomains to domain.
				if subdomains, ok := subdomainsMap[domainKey]; ok {
					domain.Subdomains = make(map[identity.Key]model_domain.Subdomain)
					for _, subdomain := range subdomains {
						subdomainKey := subdomain.Key

						// Attach generalizations to subdomain.
						if generalizations, ok := generalizationsMap[subdomainKey]; ok {
							subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
							for _, gen := range generalizations {
								subdomain.Generalizations[gen.Key] = gen
							}
						}

						// Attach use case generalizations to subdomain.
						if ucGens, ok := useCaseGeneralizationsMap[subdomainKey]; ok {
							subdomain.UseCaseGeneralizations = make(map[identity.Key]model_use_case.Generalization)
							for _, ucGen := range ucGens {
								subdomain.UseCaseGeneralizations[ucGen.Key] = ucGen
							}
						}

						// Attach use cases to subdomain, stitching actors and scenarios onto each use case.
						{
							useCasesForSubdomain := make(map[identity.Key]model_use_case.UseCase)
							for _, uc := range useCasesSlice {
								if useCaseSubdomainKeys[uc.Key] == subdomainKey {
									// Stitch actors onto use case.
									if actors, ok := useCaseActorsMap[uc.Key]; ok {
										uc.Actors = actors
									}
									// Stitch scenarios onto use case.
									if scenList, ok := scenariosMap[uc.Key]; ok {
										uc.Scenarios = make(map[identity.Key]model_scenario.Scenario, len(scenList))
										for _, scenario := range scenList {
											uc.Scenarios[scenario.Key] = scenario
										}
									}
									useCasesForSubdomain[uc.Key] = uc
								}
							}
							if len(useCasesForSubdomain) > 0 {
								subdomain.UseCases = useCasesForSubdomain
							}
						}

						// Attach use case shares to subdomain.
						{
							sharesForSubdomain := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
							for seaKey, mudMap := range useCaseSharedsMap {
								if useCaseSubdomainKeys[seaKey] == subdomainKey {
									sharesForSubdomain[seaKey] = mudMap
								}
							}
							if len(sharesForSubdomain) > 0 {
								subdomain.UseCaseShares = sharesForSubdomain
							}
						}

						// Attach classes to subdomain.
						if classes, ok := classesMap[subdomainKey]; ok {
							subdomain.Classes = make(map[identity.Key]model_class.Class)
							for _, class := range classes {
								classKey := class.Key

								// Attach attributes to class.
								if attributes, ok := attributesMap[classKey]; ok {
									class.Attributes = make(map[identity.Key]model_class.Attribute)
									for _, attr := range attributes {
										class.Attributes[attr.Key] = attr
									}
								}

								// Attach guards to class.
								if guards, ok := guardsMap[classKey]; ok {
									class.Guards = make(map[identity.Key]model_state.Guard)
									for _, guard := range guards {
										class.Guards[guard.Key] = guard
									}
								}

								// Attach actions to class.
								if actions, ok := actionsMap[classKey]; ok {
									class.Actions = make(map[identity.Key]model_state.Action)
									for _, action := range actions {
										class.Actions[action.Key] = action
									}
								}

								// Attach states to class.
								if states, ok := statesMap[classKey]; ok {
									class.States = make(map[identity.Key]model_state.State)
									for _, state := range states {
										class.States[state.Key] = state
									}
								}

								// Attach events to class.
								if events, ok := eventsMap[classKey]; ok {
									class.Events = make(map[identity.Key]model_state.Event)
									for _, event := range events {
										class.Events[event.Key] = event
									}
								}

								// Attach queries to class.
								if queries, ok := queriesMap[classKey]; ok {
									class.Queries = make(map[identity.Key]model_state.Query)
									for _, query := range queries {
										class.Queries[query.Key] = query
									}
								}

								// Attach transitions to class.
								if transitions, ok := transitionsMap[classKey]; ok {
									class.Transitions = make(map[identity.Key]model_state.Transition)
									for _, transition := range transitions {
										class.Transitions[transition.Key] = transition
									}
								}

								subdomain.Classes[class.Key] = class
							}
						}

						domain.Subdomains[subdomain.Key] = subdomain
					}
				}

				model.Domains[domain.Key] = domain
			}
		}

		// Attach domain associations to the model (they are model-level, not domain-level).
		if len(domainAssociationsSlice) > 0 {
			model.DomainAssociations = make(map[identity.Key]model_domain.Association)
			for _, assoc := range domainAssociationsSlice {
				model.DomainAssociations[assoc.Key] = assoc
			}
		}

		// Class associations — query all and route to subdomains or model level.
		classAssociationsSlice, err := QueryAssociations(tx, modelKey)
		if err != nil {
			return err
		}
		if len(classAssociationsSlice) > 0 {
			// Route each association: if its key is a child of a subdomain, attach there; otherwise model-level.
			for _, assoc := range classAssociationsSlice {
				routed := false
				for domainKey, domain := range model.Domains {
					for subdomainKey, subdomain := range domain.Subdomains {
						if assoc.Key.IsParent(subdomainKey) {
							if subdomain.ClassAssociations == nil {
								subdomain.ClassAssociations = make(map[identity.Key]model_class.Association)
							}
							subdomain.ClassAssociations[assoc.Key] = assoc
							domain.Subdomains[subdomainKey] = subdomain
							routed = true
							break
						}
					}
					if routed {
						model.Domains[domainKey] = domain
						break
					}
				}
				if !routed {
					// Model-level class association.
					if model.ClassAssociations == nil {
						model.ClassAssociations = make(map[identity.Key]model_class.Association)
					}
					model.ClassAssociations[assoc.Key] = assoc
				}
			}
		}

		return nil
	})
	if err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
