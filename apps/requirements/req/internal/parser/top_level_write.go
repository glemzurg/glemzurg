package parser

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

// Write writes a req_model.Model to the filesystem in the data/yaml format.
// This is the inverse operation of Parse.
// The outputPath is the root directory where the model will be written.
func Write(model req_model.Model, outputPath string) error {

	// Validate the model before writing.
	if err := model.Validate(); err != nil {
		return errors.Wrap(err, "model validation failed")
	}

	// Create the output directory.
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	// Write the model file (this.model).
	modelContent := generateModelContent(model)
	modelPath := filepath.Join(outputPath, "this"+_EXT_MODEL)
	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write model file")
	}

	// Write actors and actor generalizations.
	if len(model.Actors) > 0 || len(model.ActorGeneralizations) > 0 {
		actorsDir := filepath.Join(outputPath, _PATH_ACTORS)
		if err := os.MkdirAll(actorsDir, 0755); err != nil {
			return errors.Wrap(err, "failed to create actors directory")
		}

		for _, actor := range model.Actors {
			actorContent := generateActorContent(actor)
			actorPath := filepath.Join(actorsDir, actor.Key.SubKey+_EXT_ACTOR)
			if err := os.WriteFile(actorPath, []byte(actorContent), 0644); err != nil {
				return errors.Wrapf(err, "failed to write actor file: %s", actor.Key.SubKey)
			}
		}

		for _, actorGen := range model.ActorGeneralizations {
			genContent := generateActorGeneralizationContent(actorGen)
			genPath := filepath.Join(actorsDir, actorGen.Key.SubKey+_EXT_GENERALIZATION)
			if err := os.WriteFile(genPath, []byte(genContent), 0644); err != nil {
				return errors.Wrapf(err, "failed to write actor generalization file: %s", actorGen.Key.SubKey)
			}
		}
	}

	// Build a lookup of domain associations by domain key.
	domainAssocsByDomain := buildDomainAssociationsLookup(model.DomainAssociations)

	// Write domains.
	for _, domain := range model.Domains {
		if err := writeDomain(outputPath, domain, domainAssocsByDomain, model.ClassAssociations); err != nil {
			return err
		}
	}

	return nil
}

// buildDomainAssociationsLookup creates a map of domain associations grouped by problem domain key.
func buildDomainAssociationsLookup(associations map[identity.Key]model_domain.Association) map[string][]model_domain.Association {
	lookup := make(map[string][]model_domain.Association)
	for _, assoc := range associations {
		domainKeyStr := assoc.ProblemDomainKey.String()
		lookup[domainKeyStr] = append(lookup[domainKeyStr], assoc)
	}
	return lookup
}

// writeDomain writes a domain and its contents to the filesystem.
func writeDomain(outputPath string, domain model_domain.Domain, domainAssocsByDomain map[string][]model_domain.Association, modelClassAssociations map[identity.Key]model_class.Association) error {

	// Create the domain directory using the domain's subkey.
	domainDir := filepath.Join(outputPath, domain.Key.SubKey)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create domain directory: %s", domain.Key.SubKey)
	}

	// Get associations for this domain.
	associations := domainAssocsByDomain[domain.Key.String()]

	// Write the domain file (this.domain).
	domainContent := generateDomainContent(domain, associations)
	domainPath := filepath.Join(domainDir, "this"+_EXT_DOMAIN)
	if err := os.WriteFile(domainPath, []byte(domainContent), 0644); err != nil {
		return errors.Wrapf(err, "failed to write domain file: %s", domain.Key.SubKey)
	}

	// Merge domain-level and model-level class associations for writing into class files.
	// These are written alongside subdomain-level associations in each class's file.
	mergedHigherAssocs := mergeClassAssociations(domain.ClassAssociations, modelClassAssociations)

	// Process subdomains.
	for _, subdomain := range domain.Subdomains {
		if subdomain.Key.SubKey == "default" {
			// Default subdomain: write contents directly under domain directory (backward compatible).
			if err := writeSubdomainContents(domainDir, subdomain, mergedHigherAssocs); err != nil {
				return err
			}
		} else {
			// Explicit subdomain: create subdomain directory with this.subdomain file.
			if err := writeExplicitSubdomain(domainDir, subdomain, mergedHigherAssocs); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeExplicitSubdomain writes an explicit (non-default) subdomain as a separate directory
// with a this.subdomain file and its contents.
func writeExplicitSubdomain(domainDir string, subdomain model_domain.Subdomain, higherAssocs map[identity.Key]model_class.Association) error {
	// Create subdomain directory.
	subdomainDir := filepath.Join(domainDir, subdomain.Key.SubKey)
	if err := os.MkdirAll(subdomainDir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create subdomain directory: %s", subdomain.Key.SubKey)
	}

	// Write this.subdomain file.
	subdomainContent := generateSubdomainContent(subdomain)
	subdomainPath := filepath.Join(subdomainDir, "this"+_EXT_SUBDOMAIN)
	if err := os.WriteFile(subdomainPath, []byte(subdomainContent), 0644); err != nil {
		return errors.Wrapf(err, "failed to write subdomain file: %s", subdomain.Key.SubKey)
	}

	// Write contents under subdomain directory.
	if err := writeSubdomainContents(subdomainDir, subdomain, higherAssocs); err != nil {
		return err
	}

	return nil
}

// writeSubdomainContents writes the contents of a subdomain (classes, generalizations, use cases).
// For default subdomains, the baseDir is the domain directory.
// For explicit subdomains, the baseDir is the subdomain directory.
// higherAssocs contains domain-level and model-level class associations that may reference classes in this subdomain.
func writeSubdomainContents(baseDir string, subdomain model_domain.Subdomain, higherAssocs map[identity.Key]model_class.Association) error {

	// Build a lookup of class associations by from class key.
	// Include subdomain-level associations plus any higher-level associations
	// whose from-class is in this subdomain.
	classAssocsByClass := buildClassAssociationsLookup(subdomain.ClassAssociations)
	for _, assoc := range higherAssocs {
		for classKey := range subdomain.Classes {
			if assoc.FromClassKey == classKey {
				classKeyStr := classKey.String()
				classAssocsByClass[classKeyStr] = append(classAssocsByClass[classKeyStr], assoc)
			}
		}
	}

	// Write classes and generalizations to classes/ directory if there are any.
	if len(subdomain.Classes) > 0 || len(subdomain.Generalizations) > 0 {
		classesDir := filepath.Join(baseDir, "classes")
		if err := os.MkdirAll(classesDir, 0755); err != nil {
			return errors.Wrap(err, "failed to create classes directory")
		}

		// Write generalizations.
		for _, gen := range subdomain.Generalizations {
			genContent := generateGeneralizationContent(gen)
			genPath := filepath.Join(classesDir, gen.Key.SubKey+_EXT_GENERALIZATION)
			if err := os.WriteFile(genPath, []byte(genContent), 0644); err != nil {
				return errors.Wrapf(err, "failed to write generalization file: %s", gen.Key.SubKey)
			}
		}

		// Write classes.
		for _, class := range subdomain.Classes {
			associations := classAssocsByClass[class.Key.String()]
			classContent := generateClassContent(class, associations)
			classPath := filepath.Join(classesDir, class.Key.SubKey+_EXT_CLASS)
			if err := os.WriteFile(classPath, []byte(classContent), 0644); err != nil {
				return errors.Wrapf(err, "failed to write class file: %s", class.Key.SubKey)
			}
		}
	}

	// Write use cases to use_cases/ directory if there are any.
	if len(subdomain.UseCases) > 0 {
		useCasesDir := filepath.Join(baseDir, "use_cases")
		if err := os.MkdirAll(useCasesDir, 0755); err != nil {
			return errors.Wrap(err, "failed to create use_cases directory")
		}

		for _, useCase := range subdomain.UseCases {
			useCaseContent := generateUseCaseContent(useCase)
			useCasePath := filepath.Join(useCasesDir, useCase.Key.SubKey+_EXT_USE_CASE)
			if err := os.WriteFile(useCasePath, []byte(useCaseContent), 0644); err != nil {
				return errors.Wrapf(err, "failed to write use case file: %s", useCase.Key.SubKey)
			}
		}
	}

	return nil
}

// mergeClassAssociations combines multiple association maps into one.
func mergeClassAssociations(maps ...map[identity.Key]model_class.Association) map[identity.Key]model_class.Association {
	merged := make(map[identity.Key]model_class.Association)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// buildClassAssociationsLookup creates a map of class associations grouped by from class key.
func buildClassAssociationsLookup(associations map[identity.Key]model_class.Association) map[string][]model_class.Association {
	lookup := make(map[string][]model_class.Association)
	for _, assoc := range associations {
		classKeyStr := assoc.FromClassKey.String()
		lookup[classKeyStr] = append(lookup[classKeyStr], assoc)
	}
	// Sort each slice for deterministic output.
	for k := range lookup {
		sort.Slice(lookup[k], func(i, j int) bool {
			return lookup[k][i].Key.String() < lookup[k][j].Key.String()
		})
	}
	return lookup
}
