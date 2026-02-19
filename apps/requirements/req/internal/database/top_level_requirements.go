package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
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

		// Collect subdomains, generalizations, and classes into bulk structures.
		subdomainsMap := make(map[identity.Key][]model_domain.Subdomain)
		generalizationsMap := make(map[identity.Key][]model_class.Generalization)
		classesMap := make(map[identity.Key][]model_class.Class)

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

				// Collect classes.
				for _, class := range subdomain.Classes {
					classesMap[subdomainKey] = append(classesMap[subdomainKey], class)
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

		// Bulk insert classes.
		if err = AddClasses(tx, modelKey, classesMap); err != nil {
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
				gf.Specification = logicsByKey[gf.Key]
				model.GlobalFunctions[gf.Key] = gf
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

		// Classes grouped by subdomain key.
		classesMap, err := QueryClasses(tx, modelKey)
		if err != nil {
			return err
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

						// Attach classes to subdomain.
						if classes, ok := classesMap[subdomainKey]; ok {
							subdomain.Classes = make(map[identity.Key]model_class.Class)
							for _, class := range classes {
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

		return nil
	})
	if err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
