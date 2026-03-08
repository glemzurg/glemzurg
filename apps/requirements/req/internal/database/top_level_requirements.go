package database

import (
	"database/sql"
	"maps"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

func WriteModel(db *sql.DB, model core.Model) (err error) {
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

		// Write model-level data (logics, invariants, global functions, named sets, actors).
		if err = writeModelLevelData(tx, modelKey, model); err != nil {
			return err
		}

		// Write structural data (domains, subdomains, use cases, classes, attributes, associations).
		if err = writeStructuralData(tx, modelKey, model); err != nil {
			return err
		}

		// Write behavioral data (state items, use case details).
		if err = writeBehavioralData(tx, modelKey, model); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// writeModelLevelData writes logics, invariants, global functions, named sets, and actors.
func writeModelLevelData(tx *sql.Tx, modelKey string, model core.Model) error {
	if err := writeLogics(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeInvariants(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeGlobalFunctions(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeNamedSets(tx, modelKey, model); err != nil {
		return err
	}
	return writeActors(tx, modelKey, model)
}

// writeStructuralData writes domains, subdomains, use cases, classes, attributes, and associations.
func writeStructuralData(tx *sql.Tx, modelKey string, model core.Model) error {
	if err := writeDomains(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeSubdomainHierarchy(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeUseCases(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeClassesAndInvariants(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeAttributeData(tx, modelKey, model); err != nil {
		return err
	}
	return writeClassAssociations(tx, modelKey, model)
}

// writeBehavioralData writes state items and use case details.
func writeBehavioralData(tx *sql.Tx, modelKey string, model core.Model) error {
	if err := writeStateItems(tx, modelKey, model); err != nil {
		return err
	}
	return writeUseCaseDetails(tx, modelKey, model)
}

// writeLogics collects all logic rows from the model and inserts them.
func writeLogics(tx *sql.Tx, modelKey string, model core.Model) error {
	allLogics := make([]model_logic.Logic, 0, len(model.Invariants)+len(model.GlobalFunctions))
	sortOrders := make(map[string]int)

	// Invariants: sort_order = index within the invariants slice.
	for i, inv := range model.Invariants {
		sortOrders[inv.Key.String()] = i
	}
	allLogics = append(allLogics, model.Invariants...)

	// Global functions: single logic each, sort_order = 0.
	for _, gf := range model.GlobalFunctions {
		sortOrders[gf.Logic.Key.String()] = 0
		allLogics = append(allLogics, gf.Logic)
	}

	// Collect logics from domain hierarchy.
	collectDomainLogics(model, &allLogics, sortOrders)

	return AddLogics(tx, modelKey, allLogics, sortOrders)
}

// collectDomainLogics collects derivation policy, guard, action, query, class invariant, and attribute invariant logics.
func collectDomainLogics(model core.Model, allLogics *[]model_logic.Logic, sortOrders map[string]int) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				collectClassLogics(class, allLogics, sortOrders)
			}
		}
	}
}

// collectClassLogics collects all logic objects from a single class.
func collectClassLogics(class model_class.Class, allLogics *[]model_logic.Logic, sortOrders map[string]int) {
	// Derivation policy logics from attributes.
	for _, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			sortOrders[attr.DerivationPolicy.Key.String()] = 0
			*allLogics = append(*allLogics, *attr.DerivationPolicy)
		}
	}

	// Guard logics.
	for _, guard := range class.Guards {
		sortOrders[guard.Logic.Key.String()] = 0
		*allLogics = append(*allLogics, guard.Logic)
	}

	// Action require, guarantee, and safety logics.
	for _, action := range class.Actions {
		for i, req := range action.Requires {
			sortOrders[req.Key.String()] = i
		}
		for i, guar := range action.Guarantees {
			sortOrders[guar.Key.String()] = i
		}
		for i, rule := range action.SafetyRules {
			sortOrders[rule.Key.String()] = i
		}
		*allLogics = append(*allLogics, action.Requires...)
		*allLogics = append(*allLogics, action.Guarantees...)
		*allLogics = append(*allLogics, action.SafetyRules...)
	}

	// Query require and guarantee logics.
	for _, query := range class.Queries {
		for i, req := range query.Requires {
			sortOrders[req.Key.String()] = i
		}
		for i, guar := range query.Guarantees {
			sortOrders[guar.Key.String()] = i
		}
		*allLogics = append(*allLogics, query.Requires...)
		*allLogics = append(*allLogics, query.Guarantees...)
	}

	// Class invariant logics.
	for i, inv := range class.Invariants {
		sortOrders[inv.Key.String()] = i
	}
	*allLogics = append(*allLogics, class.Invariants...)

	// Attribute invariant logics.
	for _, attr := range class.Attributes {
		for i, inv := range attr.Invariants {
			sortOrders[inv.Key.String()] = i
		}
		*allLogics = append(*allLogics, attr.Invariants...)
	}
}

// writeInvariants writes invariant join rows.
func writeInvariants(tx *sql.Tx, modelKey string, model core.Model) error {
	invariantKeys := make([]identity.Key, len(model.Invariants))
	for i, inv := range model.Invariants {
		invariantKeys[i] = inv.Key
	}
	return AddInvariants(tx, modelKey, invariantKeys)
}

// writeGlobalFunctions writes global function rows.
func writeGlobalFunctions(tx *sql.Tx, modelKey string, model core.Model) error {
	gfSlice := make([]model_logic.GlobalFunction, 0, len(model.GlobalFunctions))
	for _, gf := range model.GlobalFunctions {
		gfSlice = append(gfSlice, gf)
	}
	return AddGlobalFunctions(tx, modelKey, gfSlice)
}

// writeNamedSets writes named set rows.
func writeNamedSets(tx *sql.Tx, modelKey string, model core.Model) error {
	if len(model.NamedSets) == 0 {
		return nil
	}
	nsSlice := make([]model_named_set.NamedSet, 0, len(model.NamedSets))
	for _, ns := range model.NamedSets {
		nsSlice = append(nsSlice, ns)
	}
	return AddNamedSets(tx, modelKey, nsSlice)
}

// writeActors writes actor generalizations and actors.
func writeActors(tx *sql.Tx, modelKey string, model core.Model) error {
	// Collect actor generalizations into a slice (must be inserted before actors due to FK).
	actorGeneralizationsSlice := make([]model_actor.Generalization, 0, len(model.ActorGeneralizations))
	for _, ag := range model.ActorGeneralizations {
		actorGeneralizationsSlice = append(actorGeneralizationsSlice, ag)
	}
	if err := AddActorGeneralizations(tx, modelKey, actorGeneralizationsSlice); err != nil {
		return err
	}

	// Collect actors into a slice.
	actorsSlice := make([]model_actor.Actor, 0, len(model.Actors))
	for _, actor := range model.Actors {
		actorsSlice = append(actorsSlice, actor)
	}
	return AddActors(tx, modelKey, actorsSlice)
}

// writeDomains writes domains and domain associations.
func writeDomains(tx *sql.Tx, modelKey string, model core.Model) error {
	// Collect domains into a slice.
	domainsSlice := make([]model_domain.Domain, 0, len(model.Domains))
	for _, domain := range model.Domains {
		domainsSlice = append(domainsSlice, domain)
	}
	if err := AddDomains(tx, modelKey, domainsSlice); err != nil {
		return err
	}

	// Collect domain associations (after all domains exist).
	domainAssociationsSlice := make([]model_domain.Association, 0, len(model.DomainAssociations))
	for _, association := range model.DomainAssociations {
		domainAssociationsSlice = append(domainAssociationsSlice, association)
	}
	return AddDomainAssociations(tx, modelKey, domainAssociationsSlice)
}

// writeSubdomainHierarchy collects and inserts subdomains, generalizations, use case generalizations,
// classes, and attributes from the domain hierarchy.
func writeSubdomainHierarchy(tx *sql.Tx, modelKey string, model core.Model) error {
	subdomainsMap := make(map[identity.Key][]model_domain.Subdomain)
	generalizationsMap := make(map[identity.Key][]model_class.Generalization)
	useCaseGeneralizationsMap := make(map[identity.Key][]model_use_case.Generalization)
	classesMap := make(map[identity.Key][]model_class.Class)

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			subdomainsMap[domain.Key] = append(subdomainsMap[domain.Key], subdomain)

			for _, generalization := range subdomain.Generalizations {
				generalizationsMap[subdomain.Key] = append(generalizationsMap[subdomain.Key], generalization)
			}

			for _, ucGen := range subdomain.UseCaseGeneralizations {
				useCaseGeneralizationsMap[subdomain.Key] = append(useCaseGeneralizationsMap[subdomain.Key], ucGen)
			}

			for _, class := range subdomain.Classes {
				classesMap[subdomain.Key] = append(classesMap[subdomain.Key], class)
			}
		}
	}

	// Bulk insert subdomains.
	if err := AddSubdomains(tx, modelKey, subdomainsMap); err != nil {
		return err
	}

	// Bulk insert generalizations.
	if err := AddGeneralizations(tx, modelKey, generalizationsMap); err != nil {
		return err
	}

	// Bulk insert use case generalizations.
	return AddUseCaseGeneralizations(tx, modelKey, useCaseGeneralizationsMap)
}

// writeUseCases collects and inserts use cases from subdomains.
func writeUseCases(tx *sql.Tx, modelKey string, model core.Model) error {
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
	return AddUseCases(tx, modelKey, useCaseSubdomainKeys, useCasesSlice)
}

// writeClassesAndInvariants inserts classes and class invariant join rows.
func writeClassesAndInvariants(tx *sql.Tx, modelKey string, model core.Model) error {
	classesMap := make(map[identity.Key][]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				classesMap[subdomain.Key] = append(classesMap[subdomain.Key], class)
			}
		}
	}

	// Bulk insert classes.
	if err := AddClasses(tx, modelKey, classesMap); err != nil {
		return err
	}

	// Collect class invariant join rows from classes (must be inserted after classes due to FK).
	classInvariantsMap := make(map[identity.Key][]identity.Key)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, inv := range class.Invariants {
					classInvariantsMap[class.Key] = append(classInvariantsMap[class.Key], inv.Key)
				}
			}
		}
	}
	return AddClassInvariants(tx, modelKey, classInvariantsMap)
}

// writeAttributeData collects and inserts attribute invariants, data types, attributes, and class indexes.
func writeAttributeData(tx *sql.Tx, modelKey string, model core.Model) error {
	// Collect attribute invariant join rows.
	attrInvariantsMap := make(map[identity.Key][]identity.Key)
	attributesMap := make(map[identity.Key][]model_class.Attribute)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, attr := range class.Attributes {
					attributesMap[class.Key] = append(attributesMap[class.Key], attr)
					for _, inv := range attr.Invariants {
						attrInvariantsMap[attr.Key] = append(attrInvariantsMap[attr.Key], inv.Key)
					}
				}
			}
		}
	}

	// Collect data types from attributes and parameters.
	dataTypes := collectDataTypes(model, attributesMap)
	if err := AddTopLevelDataTypes(tx, modelKey, dataTypes); err != nil {
		return err
	}

	// Bulk insert attributes.
	if err := AddAttributes(tx, modelKey, attributesMap); err != nil {
		return err
	}

	// Insert attribute invariant join rows (must be after attributes due to FK).
	if err := AddAttributeInvariants(tx, modelKey, attrInvariantsMap); err != nil {
		return err
	}

	// Bulk insert class indexes.
	return writeClassIndexes(tx, modelKey, model)
}

// collectDataTypes collects data types from attributes and parameters across the model.
func collectDataTypes(model core.Model, attributesMap map[identity.Key][]model_class.Attribute) map[string]model_data_type.DataType {
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
				collectParameterDataTypes(class, dataTypes)
			}
		}
	}
	return dataTypes
}

// collectParameterDataTypes collects data types from query, event, and action parameters.
func collectParameterDataTypes(class model_class.Class, dataTypes map[string]model_data_type.DataType) {
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

// writeClassIndexes inserts class index rows for all attributes.
func writeClassIndexes(tx *sql.Tx, modelKey string, model core.Model) error {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, attribute := range class.Attributes {
					for _, indexNum := range attribute.IndexNums {
						if err := AddClassIndex(tx, modelKey, class.Key, attribute.Key, indexNum); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

// writeClassAssociations inserts class associations.
func writeClassAssociations(tx *sql.Tx, modelKey string, model core.Model) error {
	allClassAssociations := model.GetClassAssociations()
	var classAssociationsList []model_class.Association
	for _, assoc := range allClassAssociations {
		classAssociationsList = append(classAssociationsList, assoc)
	}
	return AddAssociations(tx, modelKey, classAssociationsList)
}

// writeStateItems inserts states, guards, actions, events, queries and their related sub-items.
func writeStateItems(tx *sql.Tx, modelKey string, model core.Model) error {
	statesMap, guardsMap, actionsMap, eventsMap, queriesMap := collectStateItemMaps(model)

	// Write core state item rows.
	if err := writeStateItemCoreRows(tx, modelKey, statesMap, guardsMap, actionsMap, eventsMap, queriesMap); err != nil {
		return err
	}

	// Write sub-items (parameters, requires, guarantees, etc.).
	if err := writeStateItemSubRows(tx, modelKey, actionsMap, eventsMap, queriesMap); err != nil {
		return err
	}

	// Write state-action links and transitions.
	if err := writeStateActions(tx, modelKey, model); err != nil {
		return err
	}
	return writeTransitions(tx, modelKey, model)
}

// writeStateItemCoreRows writes the primary rows for states, guards, actions, events, and queries.
func writeStateItemCoreRows(tx *sql.Tx, modelKey string,
	statesMap map[identity.Key][]model_state.State,
	guardsMap map[identity.Key][]model_state.Guard,
	actionsMap map[identity.Key][]model_state.Action,
	eventsMap map[identity.Key][]model_state.Event,
	queriesMap map[identity.Key][]model_state.Query,
) error {
	if err := AddStates(tx, modelKey, statesMap); err != nil {
		return err
	}
	if err := AddGuards(tx, modelKey, guardsMap); err != nil {
		return err
	}
	if err := AddActions(tx, modelKey, actionsMap); err != nil {
		return err
	}
	if err := AddEvents(tx, modelKey, eventsMap); err != nil {
		return err
	}
	return AddQueries(tx, modelKey, queriesMap)
}

// writeStateItemSubRows writes sub-items for actions, events, and queries.
func writeStateItemSubRows(tx *sql.Tx, modelKey string,
	actionsMap map[identity.Key][]model_state.Action,
	eventsMap map[identity.Key][]model_state.Event,
	queriesMap map[identity.Key][]model_state.Query,
) error {
	if err := writeActionSubItems(tx, modelKey, actionsMap); err != nil {
		return err
	}
	if err := writeQuerySubItems(tx, modelKey, queriesMap); err != nil {
		return err
	}
	return writeEventParameters(tx, modelKey, eventsMap)
}

// collectStateItemMaps collects states, guards, actions, events, and queries from all classes.
func collectStateItemMaps(model core.Model) (
	statesMap map[identity.Key][]model_state.State,
	guardsMap map[identity.Key][]model_state.Guard,
	actionsMap map[identity.Key][]model_state.Action,
	eventsMap map[identity.Key][]model_state.Event,
	queriesMap map[identity.Key][]model_state.Query,
) {
	statesMap = make(map[identity.Key][]model_state.State)
	guardsMap = make(map[identity.Key][]model_state.Guard)
	actionsMap = make(map[identity.Key][]model_state.Action)
	eventsMap = make(map[identity.Key][]model_state.Event)
	queriesMap = make(map[identity.Key][]model_state.Query)

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, state := range class.States {
					statesMap[class.Key] = append(statesMap[class.Key], state)
				}
				for _, guard := range class.Guards {
					guardsMap[class.Key] = append(guardsMap[class.Key], guard)
				}
				for _, action := range class.Actions {
					actionsMap[class.Key] = append(actionsMap[class.Key], action)
				}
				for _, event := range class.Events {
					eventsMap[class.Key] = append(eventsMap[class.Key], event)
				}
				for _, query := range class.Queries {
					queriesMap[class.Key] = append(queriesMap[class.Key], query)
				}
			}
		}
	}
	return
}

// writeActionSubItems writes action parameters, requires, guarantees, and safety join rows.
func writeActionSubItems(tx *sql.Tx, modelKey string, actionsMap map[identity.Key][]model_state.Action) error {
	actionParamsMap := make(map[identity.Key][]model_state.Parameter)
	actionRequiresMap := make(map[identity.Key][]identity.Key)
	actionGuaranteesMap := make(map[identity.Key][]identity.Key)
	actionSafetiesMap := make(map[identity.Key][]identity.Key)

	for _, actionList := range actionsMap {
		for _, action := range actionList {
			actionParamsMap[action.Key] = append(actionParamsMap[action.Key], action.Parameters...)
			for _, req := range action.Requires {
				actionRequiresMap[action.Key] = append(actionRequiresMap[action.Key], req.Key)
			}
			for _, guar := range action.Guarantees {
				actionGuaranteesMap[action.Key] = append(actionGuaranteesMap[action.Key], guar.Key)
			}
			for _, rule := range action.SafetyRules {
				actionSafetiesMap[action.Key] = append(actionSafetiesMap[action.Key], rule.Key)
			}
		}
	}

	if err := AddActionParameters(tx, modelKey, actionParamsMap); err != nil {
		return err
	}
	if err := AddActionRequires(tx, modelKey, actionRequiresMap); err != nil {
		return err
	}
	if err := AddActionGuarantees(tx, modelKey, actionGuaranteesMap); err != nil {
		return err
	}
	return AddActionSafeties(tx, modelKey, actionSafetiesMap)
}

// writeQuerySubItems writes query parameters, requires, and guarantee join rows.
func writeQuerySubItems(tx *sql.Tx, modelKey string, queriesMap map[identity.Key][]model_state.Query) error {
	queryParamsMap := make(map[identity.Key][]model_state.Parameter)
	queryRequiresMap := make(map[identity.Key][]identity.Key)
	queryGuaranteesMap := make(map[identity.Key][]identity.Key)

	for _, queryList := range queriesMap {
		for _, query := range queryList {
			queryParamsMap[query.Key] = append(queryParamsMap[query.Key], query.Parameters...)
			for _, req := range query.Requires {
				queryRequiresMap[query.Key] = append(queryRequiresMap[query.Key], req.Key)
			}
			for _, guar := range query.Guarantees {
				queryGuaranteesMap[query.Key] = append(queryGuaranteesMap[query.Key], guar.Key)
			}
		}
	}

	if err := AddQueryParameters(tx, modelKey, queryParamsMap); err != nil {
		return err
	}
	if err := AddQueryRequires(tx, modelKey, queryRequiresMap); err != nil {
		return err
	}
	return AddQueryGuarantees(tx, modelKey, queryGuaranteesMap)
}

// writeEventParameters writes event parameter rows.
func writeEventParameters(tx *sql.Tx, modelKey string, eventsMap map[identity.Key][]model_state.Event) error {
	eventParamsMap := make(map[identity.Key][]model_state.Parameter)
	for _, eventList := range eventsMap {
		for _, event := range eventList {
			eventParamsMap[event.Key] = append(eventParamsMap[event.Key], event.Parameters...)
		}
	}
	return AddEventParameters(tx, modelKey, eventParamsMap)
}

// writeStateActions collects and inserts state action rows.
func writeStateActions(tx *sql.Tx, modelKey string, model core.Model) error {
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
	return AddStateActions(tx, modelKey, stateActionsMap)
}

// writeTransitions collects and inserts transition rows.
func writeTransitions(tx *sql.Tx, modelKey string, model core.Model) error {
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
	return AddTransitions(tx, modelKey, transitionsMap)
}

// writeUseCaseDetails writes use case actors, shared entries, scenarios, objects, and steps.
func writeUseCaseDetails(tx *sql.Tx, modelKey string, model core.Model) error {
	if err := writeUseCaseActors(tx, modelKey, model); err != nil {
		return err
	}
	if err := writeUseCaseShareds(tx, modelKey, model); err != nil {
		return err
	}
	return writeScenarioData(tx, modelKey, model)
}

// writeUseCaseActors collects and inserts use case actor rows.
func writeUseCaseActors(tx *sql.Tx, modelKey string, model core.Model) error {
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
	return AddUseCaseActors(tx, modelKey, useCaseActorsMap)
}

// writeUseCaseShareds collects and inserts use case shared entries.
func writeUseCaseShareds(tx *sql.Tx, modelKey string, model core.Model) error {
	useCaseSharedsMap := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			maps.Copy(useCaseSharedsMap, subdomain.UseCaseShares)
		}
	}
	return AddUseCaseShareds(tx, modelKey, useCaseSharedsMap)
}

// writeScenarioData collects and inserts scenarios, objects, and steps.
func writeScenarioData(tx *sql.Tx, modelKey string, model core.Model) error {
	scenariosMap := make(map[identity.Key][]model_scenario.Scenario)
	objectsMap := make(map[identity.Key][]model_scenario.Object)
	var allStepRows []stepRow

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, uc := range subdomain.UseCases {
				for _, scenario := range uc.Scenarios {
					scenariosMap[uc.Key] = append(scenariosMap[uc.Key], scenario)
					for _, obj := range scenario.Objects {
						objectsMap[scenario.Key] = append(objectsMap[scenario.Key], obj)
					}
					if scenario.Steps != nil {
						allStepRows = append(allStepRows, flattenSteps(scenario.Key, scenario.Steps)...)
					}
				}
			}
		}
	}

	if err := AddScenarios(tx, modelKey, scenariosMap); err != nil {
		return err
	}
	if err := AddObjects(tx, modelKey, objectsMap); err != nil {
		return err
	}
	return AddSteps(tx, modelKey, allStepRows)
}

func ReadModel(db *sql.DB, modelKey string) (model core.Model, err error) {
	// Read from within a transaction.
	err = dbTransaction(db, func(tx *sql.Tx) (err error) {
		// Model.
		model, err = LoadModel(tx, modelKey)
		if err != nil {
			return err
		}

		// Logics.
		logicsByKey, err := readLogicsByKey(tx, modelKey)
		if err != nil {
			return err
		}

		// Read model-level data (invariants, global functions, named sets, actors).
		if err = readModelLevelData(tx, modelKey, &model, logicsByKey); err != nil {
			return err
		}

		// Read domain-level data and assemble the tree.
		if err = readAndAssembleDomains(tx, modelKey, &model, logicsByKey); err != nil {
			return err
		}

		// Parse all TLA+ expressions with full model context.
		if err = convert.LowerAllExpressions(&model); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return core.Model{}, err
	}

	return model, nil
}

// readLogicsByKey loads all logics and returns them indexed by key.
func readLogicsByKey(tx *sql.Tx, modelKey string) (map[identity.Key]model_logic.Logic, error) {
	logics, err := QueryLogics(tx, modelKey)
	if err != nil {
		return nil, err
	}
	logicsByKey := make(map[identity.Key]model_logic.Logic, len(logics))
	for _, logic := range logics {
		logicsByKey[logic.Key] = logic
	}
	return logicsByKey, nil
}

// readModelLevelData reads invariants, global functions, named sets, actor generalizations, and actors.
func readModelLevelData(tx *sql.Tx, modelKey string, model *core.Model, logicsByKey map[identity.Key]model_logic.Logic) error {
	// Invariants.
	invariantKeys, err := QueryInvariants(tx, modelKey)
	if err != nil {
		return err
	}
	model.Invariants = make([]model_logic.Logic, len(invariantKeys))
	for i, key := range invariantKeys {
		model.Invariants[i] = logicsByKey[key]
	}

	// Global functions.
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

	// Named sets.
	namedSetSlice, err := QueryNamedSets(tx, modelKey)
	if err != nil {
		return err
	}
	if len(namedSetSlice) > 0 {
		model.NamedSets = make(map[identity.Key]model_named_set.NamedSet, len(namedSetSlice))
		for _, ns := range namedSetSlice {
			model.NamedSets[ns.Key] = ns
		}
	}

	// Actor generalizations.
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

	// Actors.
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

	return nil
}

// readDomainStructure holds all the data queried at the domain level for tree assembly.
type readDomainStructure struct {
	domainsSlice              []model_domain.Domain
	domainAssociationsSlice   []model_domain.Association
	subdomainsMap             map[identity.Key][]model_domain.Subdomain
	generalizationsMap        map[identity.Key][]model_class.Generalization
	useCaseGeneralizationsMap map[identity.Key][]model_use_case.Generalization
	useCaseSubdomainKeys      map[identity.Key]identity.Key
	useCasesSlice             []model_use_case.UseCase
	useCaseActorsMap          map[identity.Key]map[identity.Key]model_use_case.Actor
	useCaseSharedsMap         map[identity.Key]map[identity.Key]model_use_case.UseCaseShared
	scenariosMap              map[identity.Key][]model_scenario.Scenario
	classesMap                map[identity.Key][]model_class.Class
	classInvariantsMap        map[identity.Key][]identity.Key
	attrInvariantsMap         map[identity.Key][]identity.Key
	attributesMap             map[identity.Key][]model_class.Attribute
	guardsMap                 map[identity.Key][]model_state.Guard
	actionsMap                map[identity.Key][]model_state.Action
	statesMap                 map[identity.Key][]model_state.State
	transitionsMap            map[identity.Key][]model_state.Transition
	eventsMap                 map[identity.Key][]model_state.Event
	queriesMap                map[identity.Key][]model_state.Query
	dataTypes                 map[string]model_data_type.DataType
}

// readAndAssembleDomains reads all domain-level data and assembles the model tree.
func readAndAssembleDomains(tx *sql.Tx, modelKey string, model *core.Model, logicsByKey map[identity.Key]model_logic.Logic) error {
	ds, err := queryAllDomainData(tx, modelKey)
	if err != nil {
		return err
	}

	// Stitch logic data onto guards, actions, queries, attributes.
	stitchGuardLogics(ds.guardsMap, logicsByKey)
	stitchActionLogics(ds.actionsMap, ds, logicsByKey)
	stitchQueryLogics(ds.queriesMap, ds, logicsByKey)
	if err = stitchAttributeData(ds, logicsByKey, tx, modelKey); err != nil {
		return err
	}
	stitchParamDataTypes(ds)

	// Assemble the tree structure.
	assembleDomainTree(model, ds, logicsByKey)

	// Attach domain associations to the model.
	if len(ds.domainAssociationsSlice) > 0 {
		model.DomainAssociations = make(map[identity.Key]model_domain.Association)
		for _, assoc := range ds.domainAssociationsSlice {
			model.DomainAssociations[assoc.Key] = assoc
		}
	}

	// Class associations.
	classAssociationsSlice, err := QueryAssociations(tx, modelKey)
	if err != nil {
		return err
	}
	if len(classAssociationsSlice) > 0 {
		allClassAssocs := make(map[identity.Key]model_class.Association)
		for _, assoc := range classAssociationsSlice {
			allClassAssocs[assoc.Key] = assoc
		}
		if err = model.SetClassAssociations(allClassAssocs); err != nil {
			return err
		}
	}

	return nil
}

// queryAllDomainData runs all the domain-level queries and returns the results.
func queryAllDomainData(tx *sql.Tx, modelKey string) (*readDomainStructure, error) {
	ds := &readDomainStructure{}

	if err := queryDomainStructure(tx, modelKey, ds); err != nil {
		return nil, err
	}

	if err := queryUseCaseData(tx, modelKey, ds); err != nil {
		return nil, err
	}

	if err := queryClassStructure(tx, modelKey, ds); err != nil {
		return nil, err
	}

	if err := queryStateBehavior(tx, modelKey, ds); err != nil {
		return nil, err
	}

	var err error
	ds.dataTypes, err = LoadTopLevelDataTypes(tx, modelKey)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// queryDomainStructure queries domains, subdomains, associations, and generalizations.
func queryDomainStructure(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error

	ds.domainsSlice, err = QueryDomains(tx, modelKey)
	if err != nil {
		return err
	}

	ds.subdomainsMap, err = QuerySubdomains(tx, modelKey)
	if err != nil {
		return err
	}

	ds.domainAssociationsSlice, err = QueryDomainAssociations(tx, modelKey)
	if err != nil {
		return err
	}

	ds.generalizationsMap, err = QueryGeneralizations(tx, modelKey)
	if err != nil {
		return err
	}

	ds.useCaseGeneralizationsMap, err = QueryUseCaseGeneralizations(tx, modelKey)
	if err != nil {
		return err
	}

	return nil
}

// queryUseCaseData queries use cases, actors, shares, and scenario data.
func queryUseCaseData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error

	ds.useCaseSubdomainKeys, ds.useCasesSlice, err = QueryUseCases(tx, modelKey)
	if err != nil {
		return err
	}

	ds.useCaseActorsMap, err = QueryUseCaseActors(tx, modelKey)
	if err != nil {
		return err
	}

	ds.useCaseSharedsMap, err = QueryUseCaseShareds(tx, modelKey)
	if err != nil {
		return err
	}

	return queryScenarioData(tx, modelKey, ds)
}

// queryClassStructure queries classes, invariants, and attributes.
func queryClassStructure(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error

	ds.classesMap, err = QueryClasses(tx, modelKey)
	if err != nil {
		return err
	}

	ds.classInvariantsMap, err = QueryClassInvariants(tx, modelKey)
	if err != nil {
		return err
	}

	ds.attrInvariantsMap, err = QueryAttributeInvariants(tx, modelKey)
	if err != nil {
		return err
	}

	ds.attributesMap, err = QueryAttributes(tx, modelKey)
	if err != nil {
		return err
	}

	return nil
}

// queryStateBehavior queries guards, actions, states, transitions, events, and queries.
func queryStateBehavior(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error

	ds.guardsMap, err = QueryGuards(tx, modelKey)
	if err != nil {
		return err
	}

	if err = queryActionData(tx, modelKey, ds); err != nil {
		return err
	}

	ds.statesMap, err = QueryStates(tx, modelKey)
	if err != nil {
		return err
	}

	if err = queryStateActionData(tx, modelKey, ds); err != nil {
		return err
	}

	ds.transitionsMap, err = QueryTransitions(tx, modelKey)
	if err != nil {
		return err
	}

	if err = queryEventData(tx, modelKey, ds); err != nil {
		return err
	}

	return queryQueryData(tx, modelKey, ds)
}

// queryScenarioData queries scenarios, objects, and steps and stitches them together.
func queryScenarioData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error
	ds.scenariosMap, err = QueryScenarios(tx, modelKey)
	if err != nil {
		return err
	}

	scenarioObjectsMap, err := QueryObjects(tx, modelKey)
	if err != nil {
		return err
	}

	stepsMap, err := QuerySteps(tx, modelKey)
	if err != nil {
		return err
	}

	// Stitch objects and steps onto scenarios.
	for useCaseKey, scenList := range ds.scenariosMap {
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
		ds.scenariosMap[useCaseKey] = scenList
	}
	return nil
}

// queryActionData queries actions and their sub-items (parameters, requires, guarantees, safeties).
func queryActionData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error
	ds.actionsMap, err = QueryActions(tx, modelKey)
	if err != nil {
		return err
	}

	actionParamsMap, err := QueryActionParameters(tx, modelKey)
	if err != nil {
		return err
	}

	actionRequiresMap, err := QueryActionRequires(tx, modelKey)
	if err != nil {
		return err
	}

	actionGuaranteesMap, err := QueryActionGuarantees(tx, modelKey)
	if err != nil {
		return err
	}

	actionSafetiesMap, err := QueryActionSafeties(tx, modelKey)
	if err != nil {
		return err
	}

	// Stitch parameters onto actions.
	for classKey, actions := range ds.actionsMap {
		for i, action := range actions {
			if params, ok := actionParamsMap[action.Key]; ok {
				actions[i].Parameters = params
			}
		}
		ds.actionsMap[classKey] = actions
	}

	// Store requires/guarantees/safeties maps for logic stitching later.
	// We use a trick: store them temporarily so stitchActionLogics can use them.
	// Actually, we need to pass them. Let's stitch requires/guarantees/safeties keys here too.
	for classKey, actions := range ds.actionsMap {
		for i, action := range actions {
			if reqKeys, ok := actionRequiresMap[action.Key]; ok {
				actions[i].Requires = make([]model_logic.Logic, len(reqKeys))
				for j := range reqKeys {
					actions[i].Requires[j] = model_logic.Logic{Key: reqKeys[j]}
				}
			}
			if guarKeys, ok := actionGuaranteesMap[action.Key]; ok {
				actions[i].Guarantees = make([]model_logic.Logic, len(guarKeys))
				for j := range guarKeys {
					actions[i].Guarantees[j] = model_logic.Logic{Key: guarKeys[j]}
				}
			}
			if safetyKeys, ok := actionSafetiesMap[action.Key]; ok {
				actions[i].SafetyRules = make([]model_logic.Logic, len(safetyKeys))
				for j := range safetyKeys {
					actions[i].SafetyRules[j] = model_logic.Logic{Key: safetyKeys[j]}
				}
			}
		}
		ds.actionsMap[classKey] = actions
	}

	return nil
}

// queryStateActionData queries state actions and stitches them onto states.
func queryStateActionData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	stateActionsMap, err := QueryStateActions(tx, modelKey)
	if err != nil {
		return err
	}

	for classKey, states := range ds.statesMap {
		for i, state := range states {
			if stateActions, ok := stateActionsMap[state.Key]; ok {
				states[i].SetActions(stateActions)
			}
		}
		ds.statesMap[classKey] = states
	}
	return nil
}

// queryEventData queries events and their parameters.
func queryEventData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error
	ds.eventsMap, err = QueryEvents(tx, modelKey)
	if err != nil {
		return err
	}

	eventParamsMap, err := QueryEventParameters(tx, modelKey)
	if err != nil {
		return err
	}

	for classKey, events := range ds.eventsMap {
		for i, event := range events {
			if params, ok := eventParamsMap[event.Key]; ok {
				events[i].Parameters = params
			}
		}
		ds.eventsMap[classKey] = events
	}
	return nil
}

// queryQueryData loads queries and their parameters, requires, and guarantees.
func queryQueryData(tx *sql.Tx, modelKey string, ds *readDomainStructure) error {
	var err error
	ds.queriesMap, err = QueryQueries(tx, modelKey)
	if err != nil {
		return err
	}

	queryParamsMap, err := QueryQueryParameters(tx, modelKey)
	if err != nil {
		return err
	}

	queryRequiresMap, err := QueryQueryRequires(tx, modelKey)
	if err != nil {
		return err
	}

	queryGuaranteesMap, err := QueryQueryGuarantees(tx, modelKey)
	if err != nil {
		return err
	}

	for classKey, queries := range ds.queriesMap {
		for i, query := range queries {
			if params, ok := queryParamsMap[query.Key]; ok {
				queries[i].Parameters = params
			}
			if reqKeys, ok := queryRequiresMap[query.Key]; ok {
				queries[i].Requires = make([]model_logic.Logic, len(reqKeys))
				for j := range reqKeys {
					queries[i].Requires[j] = model_logic.Logic{Key: reqKeys[j]}
				}
			}
			if guarKeys, ok := queryGuaranteesMap[query.Key]; ok {
				queries[i].Guarantees = make([]model_logic.Logic, len(guarKeys))
				for j := range guarKeys {
					queries[i].Guarantees[j] = model_logic.Logic{Key: guarKeys[j]}
				}
			}
		}
		ds.queriesMap[classKey] = queries
	}
	return nil
}

// stitchGuardLogics stitches logic data onto guards.
func stitchGuardLogics(guardsMap map[identity.Key][]model_state.Guard, logicsByKey map[identity.Key]model_logic.Logic) {
	for classKey, guards := range guardsMap {
		for i, guard := range guards {
			if logic, ok := logicsByKey[guard.Key]; ok {
				guards[i].Logic = logic
			}
		}
		guardsMap[classKey] = guards
	}
}

// stitchActionLogics stitches full logic data onto action requires, guarantees, and safety rules.
func stitchActionLogics(actionsMap map[identity.Key][]model_state.Action, _ *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic) {
	for classKey, actions := range actionsMap {
		for i, action := range actions {
			for j, req := range action.Requires {
				actions[i].Requires[j] = logicsByKey[req.Key]
			}
			for j, guar := range action.Guarantees {
				actions[i].Guarantees[j] = logicsByKey[guar.Key]
			}
			for j, rule := range action.SafetyRules {
				actions[i].SafetyRules[j] = logicsByKey[rule.Key]
			}
		}
		actionsMap[classKey] = actions
	}
}

// stitchQueryLogics stitches full logic data onto query requires and guarantees.
func stitchQueryLogics(queriesMap map[identity.Key][]model_state.Query, _ *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic) {
	for classKey, queries := range queriesMap {
		for i, query := range queries {
			for j, req := range query.Requires {
				queries[i].Requires[j] = logicsByKey[req.Key]
			}
			for j, guar := range query.Guarantees {
				queries[i].Guarantees[j] = logicsByKey[guar.Key]
			}
		}
		queriesMap[classKey] = queries
	}
}

// stitchAttributeData stitches derivation policy, data types, invariants, and class indexes onto attributes.
func stitchAttributeData(ds *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic, tx *sql.Tx, modelKey string) error {
	for classKey, attrs := range ds.attributesMap {
		for i, attr := range attrs {
			// Stitch derivation policy from logics table.
			if attr.DerivationPolicy != nil {
				logic := logicsByKey[attr.DerivationPolicy.Key]
				attrs[i].DerivationPolicy = &logic
			}
			// Stitch data type from data types table.
			if dt, ok := ds.dataTypes[attr.Key.String()]; ok {
				attrs[i].DataType = &dt
			}
			// Stitch attribute invariants from logic data.
			if invKeys, ok := ds.attrInvariantsMap[attr.Key]; ok {
				attrs[i].Invariants = make([]model_logic.Logic, len(invKeys))
				for j, key := range invKeys {
					attrs[i].Invariants[j] = logicsByKey[key]
				}
			}
			// Load class indexes for this attribute.
			indexNums, err := LoadClassAttributeIndexes(tx, modelKey, classKey, attr.Key)
			if err != nil {
				return err
			}
			attrs[i].IndexNums = indexNums
		}
		ds.attributesMap[classKey] = attrs
	}
	return nil
}

// stitchParamDataTypes stitches data types onto query, event, and action parameters.
func stitchParamDataTypes(ds *readDomainStructure) {
	// Stitch data types onto query parameters.
	for classKey, queries := range ds.queriesMap {
		for i := range queries {
			stitchParameterDataTypes(queries[i].Parameters, ds.dataTypes)
		}
		ds.queriesMap[classKey] = queries
	}

	// Stitch data types onto event parameters.
	for classKey, events := range ds.eventsMap {
		for i := range events {
			stitchParameterDataTypes(events[i].Parameters, ds.dataTypes)
		}
		ds.eventsMap[classKey] = events
	}

	// Stitch data types onto action parameters.
	for classKey, actions := range ds.actionsMap {
		for i := range actions {
			stitchParameterDataTypes(actions[i].Parameters, ds.dataTypes)
		}
		ds.actionsMap[classKey] = actions
	}
}

// stitchParameterDataTypes resolves data type references in a parameter slice.
func stitchParameterDataTypes(params []model_state.Parameter, dataTypes map[string]model_data_type.DataType) {
	for j, param := range params {
		if param.DataType != nil {
			if dt, ok := dataTypes[param.DataType.Key]; ok {
				params[j].DataType = &dt
			}
		}
	}
}

// assembleDomainTree assembles the domain/subdomain/class hierarchy onto the model.
func assembleDomainTree(model *core.Model, ds *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic) {
	if len(ds.domainsSlice) == 0 {
		return
	}

	model.Domains = make(map[identity.Key]model_domain.Domain)
	for _, domain := range ds.domainsSlice {
		domainKey := domain.Key

		// Attach subdomains to domain.
		if subdomains, ok := ds.subdomainsMap[domainKey]; ok {
			domain.Subdomains = make(map[identity.Key]model_domain.Subdomain)
			for _, subdomain := range subdomains {
				assembleSubdomain(&subdomain, ds, logicsByKey)
				domain.Subdomains[subdomain.Key] = subdomain
			}
		}

		model.Domains[domain.Key] = domain
	}
}

// assembleSubdomain stitches generalizations, use cases, classes, and shares onto a subdomain.
func assembleSubdomain(subdomain *model_domain.Subdomain, ds *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic) {
	subdomainKey := subdomain.Key

	// Attach generalizations to subdomain.
	if generalizations, ok := ds.generalizationsMap[subdomainKey]; ok {
		subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
		for _, gen := range generalizations {
			subdomain.Generalizations[gen.Key] = gen
		}
	}

	// Attach use case generalizations to subdomain.
	if ucGens, ok := ds.useCaseGeneralizationsMap[subdomainKey]; ok {
		subdomain.UseCaseGeneralizations = make(map[identity.Key]model_use_case.Generalization)
		for _, ucGen := range ucGens {
			subdomain.UseCaseGeneralizations[ucGen.Key] = ucGen
		}
	}

	// Attach use cases to subdomain.
	assembleSubdomainUseCases(subdomain, ds)

	// Attach use case shares to subdomain.
	assembleSubdomainUseCaseShares(subdomain, ds)

	// Attach classes to subdomain.
	if classes, ok := ds.classesMap[subdomainKey]; ok {
		subdomain.Classes = make(map[identity.Key]model_class.Class)
		for _, class := range classes {
			assembleClass(&class, ds, logicsByKey)
			subdomain.Classes[class.Key] = class
		}
	}
}

// assembleSubdomainUseCases stitches use cases (with actors and scenarios) onto a subdomain.
func assembleSubdomainUseCases(subdomain *model_domain.Subdomain, ds *readDomainStructure) {
	useCasesForSubdomain := make(map[identity.Key]model_use_case.UseCase)
	for _, uc := range ds.useCasesSlice {
		if ds.useCaseSubdomainKeys[uc.Key] != subdomain.Key {
			continue
		}
		// Stitch actors onto use case.
		if actors, ok := ds.useCaseActorsMap[uc.Key]; ok {
			uc.Actors = actors
		}
		// Stitch scenarios onto use case.
		if scenList, ok := ds.scenariosMap[uc.Key]; ok {
			uc.Scenarios = make(map[identity.Key]model_scenario.Scenario, len(scenList))
			for _, scenario := range scenList {
				uc.Scenarios[scenario.Key] = scenario
			}
		}
		useCasesForSubdomain[uc.Key] = uc
	}
	if len(useCasesForSubdomain) > 0 {
		subdomain.UseCases = useCasesForSubdomain
	}
}

// assembleSubdomainUseCaseShares stitches use case shares onto a subdomain.
func assembleSubdomainUseCaseShares(subdomain *model_domain.Subdomain, ds *readDomainStructure) {
	sharesForSubdomain := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
	for seaKey, mudMap := range ds.useCaseSharedsMap {
		if ds.useCaseSubdomainKeys[seaKey] == subdomain.Key {
			sharesForSubdomain[seaKey] = mudMap
		}
	}
	if len(sharesForSubdomain) > 0 {
		subdomain.UseCaseShares = sharesForSubdomain
	}
}

// assembleClass stitches invariants, attributes, guards, actions, states, events, queries, and transitions onto a class.
func assembleClass(class *model_class.Class, ds *readDomainStructure, logicsByKey map[identity.Key]model_logic.Logic) {
	classKey := class.Key

	assembleClassInvariants(class, ds.classInvariantsMap[classKey], logicsByKey)
	assembleClassStructure(class, classKey, ds)
	assembleClassBehavior(class, classKey, ds)
}

// assembleClassInvariants attaches invariants to a class by resolving logic keys.
func assembleClassInvariants(class *model_class.Class, invKeys []identity.Key, logicsByKey map[identity.Key]model_logic.Logic) {
	if len(invKeys) == 0 {
		return
	}
	class.Invariants = make([]model_logic.Logic, len(invKeys))
	for j, key := range invKeys {
		class.Invariants[j] = logicsByKey[key]
	}
}

// assembleClassStructure attaches attributes to a class.
func assembleClassStructure(class *model_class.Class, classKey identity.Key, ds *readDomainStructure) {
	if attributes, ok := ds.attributesMap[classKey]; ok {
		class.Attributes = make(map[identity.Key]model_class.Attribute)
		for _, attr := range attributes {
			class.Attributes[attr.Key] = attr
		}
	}
}

// assembleClassBehavior attaches guards, actions, states, events, queries, and transitions to a class.
func assembleClassBehavior(class *model_class.Class, classKey identity.Key, ds *readDomainStructure) {
	if guards, ok := ds.guardsMap[classKey]; ok {
		class.Guards = make(map[identity.Key]model_state.Guard)
		for _, guard := range guards {
			class.Guards[guard.Key] = guard
		}
	}

	if actions, ok := ds.actionsMap[classKey]; ok {
		class.Actions = make(map[identity.Key]model_state.Action)
		for _, action := range actions {
			class.Actions[action.Key] = action
		}
	}

	if states, ok := ds.statesMap[classKey]; ok {
		class.States = make(map[identity.Key]model_state.State)
		for _, state := range states {
			class.States[state.Key] = state
		}
	}

	if events, ok := ds.eventsMap[classKey]; ok {
		class.Events = make(map[identity.Key]model_state.Event)
		for _, event := range events {
			class.Events[event.Key] = event
		}
	}

	if queries, ok := ds.queriesMap[classKey]; ok {
		class.Queries = make(map[identity.Key]model_state.Query)
		for _, query := range queries {
			class.Queries[query.Key] = query
		}
	}

	if transitions, ok := ds.transitionsMap[classKey]; ok {
		class.Transitions = make(map[identity.Key]model_state.Transition)
		for _, transition := range transitions {
			class.Transitions[transition.Key] = transition
		}
	}
}
