package parser_json

import (
	"cmp"
	"slices"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeModel struct {
	Format               string                    `json:"format"`
	Key                  string                    `json:"key"`
	Name                 string                    `json:"name"`
	Details              string                    `json:"details,omitempty"`
	UnfinishedNotes      string                    `json:"unfinished_notes,omitempty"`
	Invariants           []pipeLogic               `json:"invariants,omitempty"`
	GlobalFunctions      []pipeGlobalFunction      `json:"global_functions,omitempty"`
	NamedSets            []pipeNamedSet            `json:"named_sets,omitempty"`
	Actors               []pipeActor               `json:"actors,omitempty"`
	ActorGeneralizations []pipeActorGeneralization `json:"actor_generalizations,omitempty"`
	Domains              []pipeDomain              `json:"domains,omitempty"`
	DomainAssociations   []pipeDomainAssociation   `json:"domain_associations,omitempty"`
	ClassAssociations    []pipeClassAssociation    `json:"class_associations,omitempty"`
}

type pipeGlobalFunction struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	Parameters []string  `json:"parameters,omitempty"`
	Logic      pipeLogic `json:"logic"`
}

type pipeNamedSet struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Spec        *pipeSpec `json:"spec,omitempty"`
	TypeSpec    *pipeSpec `json:"type_spec,omitempty"`
}

func fromModel(model core.Model) (pipeModel, error) {
	doc := pipeModel{
		Format:          formatCorePipe,
		Key:             model.Key,
		Name:            model.Name,
		Details:         model.Details,
		UnfinishedNotes: model.UnfinishedNotes,
	}
	for _, inv := range model.Invariants {
		doc.Invariants = append(doc.Invariants, logicFromCore(inv))
	}
	for _, gf := range model.GlobalFunctions {
		doc.GlobalFunctions = append(doc.GlobalFunctions, globalFunctionFromCore(gf))
	}
	sortByKey(doc.GlobalFunctions, func(item pipeGlobalFunction) string { return item.Key })
	for _, ns := range model.NamedSets {
		doc.NamedSets = append(doc.NamedSets, namedSetFromCore(ns))
	}
	sortByKey(doc.NamedSets, func(item pipeNamedSet) string { return item.Key })
	for _, actor := range model.Actors {
		doc.Actors = append(doc.Actors, actorFromCore(actor))
	}
	sortByKey(doc.Actors, func(item pipeActor) string { return item.Key })
	for _, generalization := range model.ActorGeneralizations {
		doc.ActorGeneralizations = append(doc.ActorGeneralizations, actorGeneralizationFromCore(generalization))
	}
	sortByKey(doc.ActorGeneralizations, func(item pipeActorGeneralization) string { return item.Key })
	for _, domain := range model.Domains {
		item, err := domainFromCore(domain)
		if err != nil {
			return pipeModel{}, errors.Wrapf(err, "domain %s", domain.Key.String())
		}
		doc.Domains = append(doc.Domains, item)
	}
	sortByKey(doc.Domains, func(item pipeDomain) string { return item.Key })
	for _, assoc := range model.DomainAssociations {
		doc.DomainAssociations = append(doc.DomainAssociations, domainAssociationFromCore(assoc))
	}
	sortByKey(doc.DomainAssociations, func(item pipeDomainAssociation) string { return item.Key })
	for _, assoc := range model.ClassAssociations {
		item, err := classAssociationFromCore(assoc)
		if err != nil {
			return pipeModel{}, errors.Wrapf(err, "class association %s", assoc.Key.String())
		}
		doc.ClassAssociations = append(doc.ClassAssociations, item)
	}
	sortByKey(doc.ClassAssociations, func(item pipeClassAssociation) string { return item.Key })
	return doc, nil
}

func toModel(doc pipeModel) (core.Model, error) {
	if doc.Format != formatCorePipe {
		return core.Model{}, errors.Errorf("format must be %q", formatCorePipe)
	}

	invariants, err := logicsToCore(doc.Invariants)
	if err != nil {
		return core.Model{}, errors.Wrap(err, "invariants")
	}
	globalFunctions, err := globalFunctionsToCore(doc.GlobalFunctions)
	if err != nil {
		return core.Model{}, errors.Wrap(err, "global_functions")
	}
	namedSets, err := namedSetsToCore(doc.NamedSets)
	if err != nil {
		return core.Model{}, errors.Wrap(err, "named_sets")
	}

	model := core.NewModel(doc.Key, core.ModelDetails{Name: doc.Name, Details: doc.Details}, doc.UnfinishedNotes, invariants, globalFunctions, namedSets)
	if err := setModelChildren(&model, doc); err != nil {
		return core.Model{}, err
	}
	return model, nil
}

// setModelChildren attaches the tree NewModel does not take.
func setModelChildren(model *core.Model, doc pipeModel) error {
	actors, err := actorsToCore(doc.Actors)
	if err != nil {
		return errors.Wrap(err, "actors")
	}
	actorGeneralizations, err := actorGeneralizationsToCore(doc.ActorGeneralizations)
	if err != nil {
		return errors.Wrap(err, "actor_generalizations")
	}
	domains, err := domainsToCore(doc.Domains)
	if err != nil {
		return errors.Wrap(err, "domains")
	}
	domainAssociations, err := domainAssociationsToCore(doc.DomainAssociations)
	if err != nil {
		return errors.Wrap(err, "domain_associations")
	}
	classAssociations, err := classAssociationsToCore(doc.ClassAssociations)
	if err != nil {
		return errors.Wrap(err, "class_associations")
	}

	model.Actors = actors
	model.ActorGeneralizations = actorGeneralizations
	model.Domains = domains
	model.DomainAssociations = domainAssociations
	if len(classAssociations) > 0 {
		if err := model.SetClassAssociations(classAssociations); err != nil {
			return errors.Wrap(err, "class_associations")
		}
	}
	return nil
}

func globalFunctionFromCore(gf model_logic.GlobalFunction) pipeGlobalFunction {
	return pipeGlobalFunction{
		Key:        keyString(gf.Key),
		Name:       gf.Name,
		Parameters: gf.Parameters,
		Logic:      logicFromCore(gf.Logic),
	}
}

func globalFunctionsToCore(items []pipeGlobalFunction) (map[identity.Key]model_logic.GlobalFunction, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_logic.GlobalFunction, len(items))
	for i, item := range items {
		gf, err := globalFunctionToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, gf.Key, gf, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func globalFunctionToCore(item pipeGlobalFunction) (model_logic.GlobalFunction, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_logic.GlobalFunction{}, err
	}
	logic, err := logicToCore(item.Logic)
	if err != nil {
		return model_logic.GlobalFunction{}, errors.Wrap(err, "logic")
	}
	return model_logic.NewGlobalFunction(key, item.Name, item.Parameters, logic), nil
}

func namedSetFromCore(ns model_logic.NamedSet) pipeNamedSet {
	return pipeNamedSet{
		Key:         keyString(ns.Key),
		Name:        ns.Name,
		Description: ns.Description,
		Spec:        specFromCore(ns.Spec),
		TypeSpec:    typeSpecFromCore(ns.TypeSpec),
	}
}

func namedSetsToCore(items []pipeNamedSet) (map[identity.Key]model_logic.NamedSet, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_logic.NamedSet, len(items))
	for i, item := range items {
		ns, err := namedSetToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, ns.Key, ns, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func namedSetToCore(item pipeNamedSet) (model_logic.NamedSet, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_logic.NamedSet{}, err
	}
	spec, err := specToCore(item.Spec)
	if err != nil {
		return model_logic.NamedSet{}, errors.Wrap(err, "spec")
	}
	typeSpec, err := typeSpecToCore(item.TypeSpec)
	if err != nil {
		return model_logic.NamedSet{}, errors.Wrap(err, "type_spec")
	}
	return model_logic.NewNamedSet(key, item.Name, item.Description, spec, typeSpec), nil
}

// sortByKey orders map-backed collections so two dumps of one model match.
func sortByKey[T any](items []T, key func(T) string) {
	slices.SortFunc(items, func(a, b T) int {
		return cmp.Compare(key(a), key(b))
	})
}

func putUnique[V any](dst map[identity.Key]V, key identity.Key, value V, index int) error {
	if _, exists := dst[key]; exists {
		return errors.Errorf("[%d]: duplicate key %s", index, key.String())
	}
	dst[key] = value
	return nil
}
