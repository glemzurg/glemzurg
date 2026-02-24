package parser_ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

func WriteModel(model req_model.Model, outputModelPath string) error {

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

	// Write invariants
	if len(model.Invariants) > 0 {
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
	}

	// Write actors
	if len(model.Actors) > 0 {
		actorsDir := filepath.Join(modelDir, "actors")
		if err := os.MkdirAll(actorsDir, 0755); err != nil {
			return err
		}
		for key, actor := range model.Actors {
			if err := writeJSON(filepath.Join(actorsDir, key+".actor.json"), actor); err != nil {
				return err
			}
		}
	}

	// Write actor generalizations
	if len(model.ActorGeneralizations) > 0 {
		agDir := filepath.Join(modelDir, "actor_generalizations")
		if err := os.MkdirAll(agDir, 0755); err != nil {
			return err
		}
		for key, gen := range model.ActorGeneralizations {
			if err := writeJSON(filepath.Join(agDir, key+".agen.json"), gen); err != nil {
				return err
			}
		}
	}

	// Write global functions
	if len(model.GlobalFunctions) > 0 {
		gfDir := filepath.Join(modelDir, "global_functions")
		if err := os.MkdirAll(gfDir, 0755); err != nil {
			return err
		}
		for key, gf := range model.GlobalFunctions {
			if err := writeJSON(filepath.Join(gfDir, key+".json"), gf); err != nil {
				return err
			}
		}
	}

	// Write model-level class associations
	if len(model.ClassAssociations) > 0 {
		assocDir := filepath.Join(modelDir, "class_associations")
		if err := os.MkdirAll(assocDir, 0755); err != nil {
			return err
		}
		for _, assoc := range model.ClassAssociations {
			filename := classAssociationFilename(assoc, AssocLevelModel)
			if err := writeJSON(filepath.Join(assocDir, filename), assoc); err != nil {
				return err
			}
		}
	}

	// Write domain associations
	if len(model.DomainAssociations) > 0 {
		daDir := filepath.Join(modelDir, "domain_associations")
		if err := os.MkdirAll(daDir, 0755); err != nil {
			return err
		}
		for key, da := range model.DomainAssociations {
			if err := writeJSON(filepath.Join(daDir, key+".domain_assoc.json"), da); err != nil {
				return err
			}
		}
	}

	// Write domains
	if len(model.Domains) > 0 {
		domainsDir := filepath.Join(modelDir, "domains")
		if err := os.MkdirAll(domainsDir, 0755); err != nil {
			return err
		}
		for domainKey, domain := range model.Domains {
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

	// Write domain-level class associations
	if len(domain.ClassAssociations) > 0 {
		assocDir := filepath.Join(domainDir, "class_associations")
		if err := os.MkdirAll(assocDir, 0755); err != nil {
			return err
		}
		for _, assoc := range domain.ClassAssociations {
			filename := classAssociationFilename(assoc, AssocLevelDomain)
			if err := writeJSON(filepath.Join(assocDir, filename), assoc); err != nil {
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
		for subdomainKey, subdomain := range domain.Subdomains {
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
		for _, assoc := range subdomain.ClassAssociations {
			filename := classAssociationFilename(assoc, AssocLevelSubdomain)
			if err := writeJSON(filepath.Join(assocDir, filename), assoc); err != nil {
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
		for key, gen := range subdomain.ClassGeneralizations {
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
		for classKey, class := range subdomain.Classes {
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
		for key, gen := range subdomain.UseCaseGeneralizations {
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
		for useCaseKey, useCase := range subdomain.UseCases {
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
		for key, action := range class.Actions {
			if err := writeJSON(filepath.Join(actionsDir, key+".json"), action); err != nil {
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
		for key, query := range class.Queries {
			if err := writeJSON(filepath.Join(queriesDir, key+".json"), query); err != nil {
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
		for key, scenario := range useCase.Scenarios {
			if err := writeJSON(filepath.Join(scenariosDir, key+".scenario.json"), scenario); err != nil {
				return err
			}
		}
	}

	return nil
}

func classAssociationFilename(assoc *inputClassAssociation, level AssociationLevel) string {
	from := strings.ReplaceAll(assoc.FromClassKey, "/", ".")
	to := strings.ReplaceAll(assoc.ToClassKey, "/", ".")
	name := strings.ToLower(strings.ReplaceAll(assoc.Name, " ", "_"))
	return fmt.Sprintf("%s--%s--%s.assoc.json", from, to, name)
}

// writeJSON writes a struct as JSON to a file.
func writeJSON(filename string, v any) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
