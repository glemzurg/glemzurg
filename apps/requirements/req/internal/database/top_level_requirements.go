package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
)

func WriteRequirements(db *sql.DB, reqs requirements.Requirements) (err error) {

	// Everything should be written in order, as a transaction.
	err = dbTransaction(db, func(tx *sql.Tx) (err error) {

		// Model.
		modelKey := reqs.Model.Key

		// Clear out the prior model first.
		if err = RemoveModel(tx, modelKey); err != nil {
			return err
		}

		// Add the model.
		if err = AddModel(tx, reqs.Model); err != nil {
			return err
		}

		// Generalizations.
		for _, generalization := range reqs.Generalizations {
			if err = AddGeneralization(tx, modelKey, generalization); err != nil {
				return err
			}
		}

		// Actors.
		for _, actor := range reqs.Actors {
			if err = AddActor(tx, modelKey, actor); err != nil {
				return err
			}
		}

		// Organizations.
		for _, domain := range reqs.Domains {
			if err = AddDomain(tx, modelKey, domain); err != nil {
				return err
			}
		}
		for domainKey, subdomains := range reqs.Subdomains {
			for _, subdomain := range subdomains {
				if err = AddSubdomain(tx, modelKey, domainKey, subdomain); err != nil {
					return err
				}
			}
		}
		for _, association := range reqs.DomainAssociations {
			if err = AddDomainAssociation(tx, modelKey, association); err != nil {
				return err
			}
		}

		// Classes.
		for subdomainKey, classes := range reqs.Classes {
			for _, class := range classes {
				if err = AddClass(tx, modelKey, subdomainKey, class); err != nil {
					return err
				}
			}
		}
		for classKey, attributes := range reqs.Attributes {
			for _, attribute := range attributes {
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
		}
		for _, association := range reqs.Associations {
			if err = AddAssociation(tx, modelKey, association); err != nil {
				return err
			}
		}

		// State machines.
		for classKey, states := range reqs.States {
			for _, state := range states {
				if err = AddState(tx, modelKey, classKey, state); err != nil {
					return err
				}
			}
		}
		for classKey, events := range reqs.Events {
			for _, event := range events {
				if err = AddEvent(tx, modelKey, classKey, event); err != nil {
					return err
				}
			}
		}
		for classKey, guards := range reqs.Guards {
			for _, guard := range guards {
				if err = AddGuard(tx, modelKey, classKey, guard); err != nil {
					return err
				}
			}
		}
		for classKey, actions := range reqs.Actions {
			for _, action := range actions {
				if err = AddAction(tx, modelKey, classKey, action); err != nil {
					return err
				}
			}
		}
		for classKey, transitions := range reqs.Transitions {
			for _, transition := range transitions {
				if err = AddTransition(tx, modelKey, classKey, transition); err != nil {
					return err
				}
			}
		}
		for stateKey, stateActions := range reqs.StateActions {
			for _, stateAction := range stateActions {
				if err = AddStateAction(tx, modelKey, stateKey, stateAction); err != nil {
					return err
				}
			}
		}

		// The use cases.
		for subdomainKey, useCases := range reqs.UseCases {
			for _, useCase := range useCases {
				if err = AddUseCase(tx, modelKey, subdomainKey, useCase); err != nil {
					return err
				}
			}
		}

		for useCaseKey, actors := range reqs.UseCaseActors {
			for actorKey, actor := range actors {
				if err = AddUseCaseActor(tx, modelKey, useCaseKey, actorKey, actor); err != nil {
					return err
				}
			}
		}

		// Collect all data types from attributes.
		dataTypes := make(map[string]model_data_type.DataType)
		for _, attributes := range reqs.Attributes {
			for _, attribute := range attributes {
				if attribute.DataType != nil {
					dataTypes[attribute.DataType.Key] = *attribute.DataType
				}
			}
		}

		// Add data types.
		if err = AddTopLevelDataTypes(tx, modelKey, dataTypes); err != nil {
			return err
		}

		// Scenarios.
		for useCaseKey, useCaseScenarios := range reqs.Scenarios {
			for _, scenario := range useCaseScenarios {
				if err = AddScenario(tx, modelKey, useCaseKey, scenario); err != nil {
					return err
				}
			}
		}
		for scenarioKey, objects := range reqs.Objects {
			for _, object := range objects {
				if err = AddObject(tx, modelKey, scenarioKey, object); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func ReadRequirements(db *sql.DB, modelKey string) (reqs requirements.Requirements, err error) {

	// Read from within a transaction.
	err = dbTransaction(db, func(tx *sql.Tx) (err error) {

		// Model.
		reqs.Model, err = LoadModel(tx, modelKey)
		if err != nil {
			return err
		}

		// Generalizations.
		reqs.Generalizations, err = QueryGeneralizations(tx, modelKey)
		if err != nil {
			return err
		}

		// Actors.
		reqs.Actors, err = QueryActors(tx, modelKey)
		if err != nil {
			return err
		}

		// Organizations.
		reqs.Domains, err = QueryDomains(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Subdomains, err = QuerySubdomains(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.DomainAssociations, err = QueryDomainAssociations(tx, modelKey)
		if err != nil {
			return err
		}

		// Classes.
		reqs.Classes, err = QueryClasses(tx, modelKey)
		if err != nil {
			return err
		}

		// Load all data types.
		dataTypes, err := LoadTopLevelDataTypes(tx, modelKey)
		if err != nil {
			return err
		}

		reqs.Attributes, err = QueryAttributes(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Associations, err = QueryAssociations(tx, modelKey)
		if err != nil {
			return err
		}

		// State machines.
		reqs.States, err = QueryStates(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Events, err = QueryEvents(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Guards, err = QueryGuards(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Actions, err = QueryActions(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.Transitions, err = QueryTransitions(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.StateActions, err = QueryStateActions(tx, modelKey)
		if err != nil {
			return err
		}

		// Use cases.
		subdomainKeys, useCases, err := QueryUseCases(tx, modelKey)
		if err != nil {
			return err
		}
		reqs.UseCases = make(map[string][]model_use_case.UseCase)
		for _, uc := range useCases {
			subdomainKey := subdomainKeys[uc.Key]
			reqs.UseCases[subdomainKey] = append(reqs.UseCases[subdomainKey], uc)
		}
		reqs.UseCaseActors, err = QueryUseCaseActors(tx, modelKey)
		if err != nil {
			return err
		}

		reqs.Scenarios, err = QueryScenarios(tx, modelKey)
		if err != nil {
			return err
		}

		reqs.Objects, err = QueryObjects(tx, modelKey)
		if err != nil {
			return err
		}

		// Attach data types to attributes.
		for classKey, attributes := range reqs.Attributes {
			for i, attribute := range attributes {
				// Attribute key is the same for the data type.
				if dt, ok := dataTypes[attribute.Key]; ok {
					reqs.Attributes[classKey][i].DataType = &dt
				}
			}
		}

		return nil
	})
	if err != nil {
		return requirements.Requirements{}, err
	}

	return reqs, nil
}
