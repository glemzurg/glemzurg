package parser_ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
)

func WriteModel(model core.Model, outputModelPath string) error {
	inputModel, err := ConvertFromModel(&model)
	if err != nil {
		return err
	}

	if err := writeModelTree(inputModel, outputModelPath); err != nil {
		return err
	}

	return nil
}

// writeModelTree writes a complete model tree to the filesystem.
// The modelDir is the root directory where the model will be written.
func writeModelTree(model *inputModel, modelDir string) error {
	// Create model directory
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return err
	}

	// Write model.json
	if err := writeJSON(filepath.Join(modelDir, "model.json"), model); err != nil {
		return err
	}

	if err := writeModelInvariants(model, modelDir); err != nil {
		return err
	}
	if err := writeModelActorsAndGeneralizations(model, modelDir); err != nil {
		return err
	}
	if err := writeModelCollections(model, modelDir); err != nil {
		return err
	}
	if err := writeModelAssociationsAndDomains(model, modelDir); err != nil {
		return err
	}

	return nil
}

func writeAssociationInvariants(assocDir, key string, assoc *inputClassAssociation) error {
	if len(assoc.Invariants) == 0 {
		return nil
	}
	invariantsDir := filepath.Join(assocDir, key, "invariants")
	if err := os.MkdirAll(invariantsDir, 0755); err != nil {
		return err
	}
	for i, inv := range assoc.Invariants {
		filename := fmt.Sprintf("%03d.invariant.json", i+1)
		if err := writeJSON(filepath.Join(invariantsDir, filename), inv); err != nil {
			return err
		}
	}
	return nil
}

// writeModelInvariants writes model-level invariants to the filesystem.
func writeModelInvariants(model *inputModel, modelDir string) error {
	if len(model.Invariants) == 0 {
		return nil
	}
	invariantsDir := filepath.Join(modelDir, "invariants")
	if err := os.MkdirAll(invariantsDir, 0755); err != nil {
		return err
	}
	for i, inv := range model.Invariants {
		filename := fmt.Sprintf("%03d.invariant.json", i+1)
		if err := writeJSON(filepath.Join(invariantsDir, filename), inv); err != nil {
			return err
		}
	}
	return nil
}

// writeModelActorsAndGeneralizations writes actors and actor generalizations.
func writeModelActorsAndGeneralizations(model *inputModel, modelDir string) error {
	if len(model.Actors) > 0 {
		actorsDir := filepath.Join(modelDir, "actors")
		if err := os.MkdirAll(actorsDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.Actors) {
			actor := model.Actors[key]
			if err := writeJSON(filepath.Join(actorsDir, key+".actor.json"), actor); err != nil {
				return err
			}
		}
	}
	if len(model.ActorGeneralizations) > 0 {
		agDir := filepath.Join(modelDir, "actor_generalizations")
		if err := os.MkdirAll(agDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.ActorGeneralizations) {
			gen := model.ActorGeneralizations[key]
			if err := writeJSON(filepath.Join(agDir, key+".agen.json"), gen); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeModelCollections writes global functions and named sets.
func writeModelCollections(model *inputModel, modelDir string) error {
	if len(model.GlobalFunctions) > 0 {
		gfDir := filepath.Join(modelDir, "global_functions")
		if err := os.MkdirAll(gfDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.GlobalFunctions) {
			gf := model.GlobalFunctions[key]
			if err := writeJSON(filepath.Join(gfDir, key+".json"), gf); err != nil {
				return err
			}
		}
	}
	if len(model.NamedSets) > 0 {
		nsDir := filepath.Join(modelDir, "named_sets")
		if err := os.MkdirAll(nsDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.NamedSets) {
			ns := model.NamedSets[key]
			if err := writeJSON(filepath.Join(nsDir, key+".nset.json"), ns); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeModelAssociationsAndDomains writes class associations, domain associations, and domains.
func writeModelAssociationsAndDomains(model *inputModel, modelDir string) error {
	if len(model.ClassAssociations) > 0 {
		assocDir := filepath.Join(modelDir, "class_associations")
		if err := os.MkdirAll(assocDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.ClassAssociations) {
			assoc := model.ClassAssociations[key]
			if err := writeJSON(filepath.Join(assocDir, key+".assoc.json"), assoc); err != nil {
				return err
			}
			if err := writeAssociationInvariants(assocDir, key, assoc); err != nil {
				return err
			}
		}
	}
	if len(model.DomainAssociations) > 0 {
		daDir := filepath.Join(modelDir, "domain_associations")
		if err := os.MkdirAll(daDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(model.DomainAssociations) {
			da := model.DomainAssociations[key]
			if err := writeJSON(filepath.Join(daDir, key+".domain_assoc.json"), da); err != nil {
				return err
			}
		}
	}
	if len(model.Domains) > 0 {
		domainsDir := filepath.Join(modelDir, "domains")
		if err := os.MkdirAll(domainsDir, 0755); err != nil {
			return err
		}
		for _, domainKey := range sortedKeys(model.Domains) {
			domain := model.Domains[domainKey]
			if err := writeDomainTree(domain, filepath.Join(domainsDir, domainKey)); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeDomainTree writes a domain and its children to the filesystem.
func writeDomainTree(domain *inputDomain, domainDir string) error {
	// Create domain directory
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return err
	}

	// Write domain.json
	if err := writeJSON(filepath.Join(domainDir, "domain.json"), domain); err != nil {
		return err
	}

	if len(domain.SubdomainAssociations) > 0 {
		saDir := filepath.Join(domainDir, "subdomain_associations")
		if err := os.MkdirAll(saDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(domain.SubdomainAssociations) {
			sa := domain.SubdomainAssociations[key]
			if err := writeJSON(filepath.Join(saDir, key+".subdomain_assoc.json"), sa); err != nil {
				return err
			}
		}
	}

	// Write domain-level class associations
	if len(domain.ClassAssociations) > 0 {
		assocDir := filepath.Join(domainDir, "class_associations")
		if err := os.MkdirAll(assocDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(domain.ClassAssociations) {
			assoc := domain.ClassAssociations[key]
			if err := writeJSON(filepath.Join(assocDir, key+".assoc.json"), assoc); err != nil {
				return err
			}
			if err := writeAssociationInvariants(assocDir, key, assoc); err != nil {
				return err
			}
		}
	}

	// Write subdomains
	if len(domain.Subdomains) > 0 {
		subdomainsDir := filepath.Join(domainDir, "subdomains")
		if err := os.MkdirAll(subdomainsDir, 0755); err != nil {
			return err
		}
		for _, subdomainKey := range sortedKeys(domain.Subdomains) {
			subdomain := domain.Subdomains[subdomainKey]
			if err := writeSubdomainTree(subdomain, filepath.Join(subdomainsDir, subdomainKey)); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeSubdomainTree writes a subdomain and its children to the filesystem.
func writeSubdomainTree(subdomain *inputSubdomain, subdomainDir string) error {
	// Create subdomain directory
	if err := os.MkdirAll(subdomainDir, 0755); err != nil {
		return err
	}

	// Write subdomain.json
	if err := writeJSON(filepath.Join(subdomainDir, "subdomain.json"), subdomain); err != nil {
		return err
	}

	// Write subdomain-level class associations
	if len(subdomain.ClassAssociations) > 0 {
		assocDir := filepath.Join(subdomainDir, "class_associations")
		if err := os.MkdirAll(assocDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(subdomain.ClassAssociations) {
			assoc := subdomain.ClassAssociations[key]
			if err := writeJSON(filepath.Join(assocDir, key+".assoc.json"), assoc); err != nil {
				return err
			}
			if err := writeAssociationInvariants(assocDir, key, assoc); err != nil {
				return err
			}
		}
	}

	// Write class generalizations
	if len(subdomain.ClassGeneralizations) > 0 {
		genDir := filepath.Join(subdomainDir, "class_generalizations")
		if err := os.MkdirAll(genDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(subdomain.ClassGeneralizations) {
			gen := subdomain.ClassGeneralizations[key]
			if err := writeJSON(filepath.Join(genDir, key+".cgen.json"), gen); err != nil {
				return err
			}
		}
	}

	// Write classes
	if len(subdomain.Classes) > 0 {
		classesDir := filepath.Join(subdomainDir, "classes")
		if err := os.MkdirAll(classesDir, 0755); err != nil {
			return err
		}
		for _, classKey := range sortedKeys(subdomain.Classes) {
			class := subdomain.Classes[classKey]
			if err := writeClassTree(class, filepath.Join(classesDir, classKey)); err != nil {
				return err
			}
		}
	}

	// Write use case generalizations
	if len(subdomain.UseCaseGeneralizations) > 0 {
		genDir := filepath.Join(subdomainDir, "use_case_generalizations")
		if err := os.MkdirAll(genDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(subdomain.UseCaseGeneralizations) {
			gen := subdomain.UseCaseGeneralizations[key]
			if err := writeJSON(filepath.Join(genDir, key+".ucgen.json"), gen); err != nil {
				return err
			}
		}
	}

	// Write use cases
	if len(subdomain.UseCases) > 0 {
		useCasesDir := filepath.Join(subdomainDir, "use_cases")
		if err := os.MkdirAll(useCasesDir, 0755); err != nil {
			return err
		}
		for _, useCaseKey := range sortedKeys(subdomain.UseCases) {
			useCase := subdomain.UseCases[useCaseKey]
			if err := writeUseCaseTree(useCase, filepath.Join(useCasesDir, useCaseKey)); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeClassTree writes a class and its children to the filesystem.
func writeClassTree(class *inputClass, classDir string) error {
	// Create class directory
	if err := os.MkdirAll(classDir, 0755); err != nil {
		return err
	}

	// Write class.json
	if err := writeJSON(filepath.Join(classDir, "class.json"), class); err != nil {
		return err
	}

	// Write invariants
	if len(class.Invariants) > 0 {
		invariantsDir := filepath.Join(classDir, "invariants")
		if err := os.MkdirAll(invariantsDir, 0755); err != nil {
			return err
		}
		for i, inv := range class.Invariants {
			filename := fmt.Sprintf("%03d.invariant.json", i+1)
			if err := writeJSON(filepath.Join(invariantsDir, filename), inv); err != nil {
				return err
			}
		}
	}

	// Write attribute invariants (per attribute subdirectory)
	for _, attr := range class.Attributes {
		attrKey := attr.Key
		if len(attr.Invariants) > 0 {
			attrInvariantsDir := filepath.Join(classDir, "attributes", attrKey, "invariants")
			if err := os.MkdirAll(attrInvariantsDir, 0755); err != nil {
				return err
			}
			for i, inv := range attr.Invariants {
				filename := fmt.Sprintf("%03d.invariant.json", i+1)
				if err := writeJSON(filepath.Join(attrInvariantsDir, filename), inv); err != nil {
					return err
				}
			}
		}
	}

	// Write state_machine.json if present
	if class.StateMachine != nil {
		if err := writeJSON(filepath.Join(classDir, "state_machine.json"), class.StateMachine); err != nil {
			return err
		}
	}

	// Write actions
	if len(class.Actions) > 0 {
		actionsDir := filepath.Join(classDir, "actions")
		if err := os.MkdirAll(actionsDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(class.Actions) {
			action := class.Actions[key]
			if err := writeJSON(filepath.Join(actionsDir, key+".json"), action); err != nil {
				return err
			}
			if err := writeOwnerParameterInvariants(classDir, "actions", key, action.Parameters); err != nil {
				return err
			}
		}
	}

	// Write queries
	if len(class.Queries) > 0 {
		queriesDir := filepath.Join(classDir, "queries")
		if err := os.MkdirAll(queriesDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(class.Queries) {
			query := class.Queries[key]
			if err := writeJSON(filepath.Join(queriesDir, key+".json"), query); err != nil {
				return err
			}
			if err := writeOwnerParameterInvariants(classDir, "queries", key, query.Parameters); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeUseCaseTree writes a use case and its children to the filesystem.
func writeUseCaseTree(useCase *inputUseCase, useCaseDir string) error {
	// Create use case directory
	if err := os.MkdirAll(useCaseDir, 0755); err != nil {
		return err
	}

	// Write use_case.json
	if err := writeJSON(filepath.Join(useCaseDir, "use_case.json"), useCase); err != nil {
		return err
	}

	// Write scenarios
	if len(useCase.Scenarios) > 0 {
		scenariosDir := filepath.Join(useCaseDir, "scenarios")
		if err := os.MkdirAll(scenariosDir, 0755); err != nil {
			return err
		}
		for _, key := range sortedKeys(useCase.Scenarios) {
			scenario := useCase.Scenarios[key]
			if err := writeJSON(filepath.Join(scenariosDir, key+".scenario.json"), scenario); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeOwnerParameterInvariants writes per-parameter invariant files under an action or query.
func writeOwnerParameterInvariants(classDir, ownerKind, ownerKey string, params []inputParameter) error {
	for _, param := range params {
		if len(param.Invariants) == 0 {
			continue
		}
		paramDirKey, err := safeParameterDirKey(param.Name)
		if err != nil {
			return err
		}
		paramInvariantsDir := filepath.Join(classDir, ownerKind, ownerKey, "parameters", paramDirKey, "invariants")
		if err := os.MkdirAll(paramInvariantsDir, 0755); err != nil {
			return err
		}
		for i, inv := range param.Invariants {
			filename := fmt.Sprintf("%03d.invariant.json", i+1)
			if err := writeJSON(filepath.Join(paramInvariantsDir, filename), inv); err != nil {
				return err
			}
		}
	}
	return nil
}

// classAssociationMapKey builds the map key and filename stem for a class association.
func classAssociationMapKey(assoc *inputClassAssociation, nameKey string) string {
	from := strings.ReplaceAll(assoc.FromClassKey, "/", ".")
	to := strings.ReplaceAll(assoc.ToClassKey, "/", ".")
	return fmt.Sprintf("%s--%s--%s", from, to, nameKey)
}

// sortedKeys returns sorted keys from a string-keyed map.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// writeJSON writes a struct as JSON to a file.
func writeJSON(filename string, v any) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}
