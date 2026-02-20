package database

import (
	"database/sql"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	// "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
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

		// Collect subdomains, generalizations, classes, and attributes into bulk structures.
		subdomainsMap := make(map[identity.Key][]model_domain.Subdomain)
		generalizationsMap := make(map[identity.Key][]model_class.Generalization)
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

		// Bulk insert classes.
		if err = AddClasses(tx, modelKey, classesMap); err != nil {
			return err
		}

		// Collect data types from attributes (must be inserted before attributes due to FK).
		dataTypes := make(map[string]model_data_type.DataType)
		for _, attrs := range attributesMap {
			for _, attr := range attrs {
				if attr.DataType != nil {
					dataTypes[attr.DataType.Key] = *attr.DataType
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

		// Attributes grouped by class key.
		attributesMap, err := QueryAttributes(tx, modelKey)
		if err != nil {
			return err
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

		// Stitch parameters onto queries.
		for classKey, queries := range queriesMap {
			for i, query := range queries {
				if params, ok := queryParamsMap[query.Key]; ok {
					queries[i].Parameters = params
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
								classKey := class.Key

								// Attach attributes to class.
								if attributes, ok := attributesMap[classKey]; ok {
									class.Attributes = make(map[identity.Key]model_class.Attribute)
									for _, attr := range attributes {
										class.Attributes[attr.Key] = attr
									}
								}

								// Attach queries to class.
								if queries, ok := queriesMap[classKey]; ok {
									class.Queries = make(map[identity.Key]model_state.Query)
									for _, query := range queries {
										class.Queries[query.Key] = query
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
