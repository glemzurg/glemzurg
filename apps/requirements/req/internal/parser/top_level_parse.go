package parser

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

func Parse(modelPath string) (model req_model.Model, err error) {
	log.Printf("Parse files in '%s'", modelPath)

	// First, gather every thing we expect we need to parse, and what kind of entity it is.

	toParseFiles, err := findFilesToParse(modelPath)
	if err != nil {
		return req_model.Model{}, errors.WithStack(err)
	}

	for _, toParseFile := range toParseFiles {
		log.Println("   walk:", toParseFile.String())
	}

	model, err = parseForDatabase(modelPath, toParseFiles)
	if err != nil {
		return req_model.Model{}, errors.WithStack(err)
	}

	// Verify the model is well-formed after the parse.
	if err = model.Validate(); err != nil {
		return req_model.Model{}, errors.WithStack(err)
	}

	return model, nil
}

func findFilesToParse(modelPath string) (toParseFiles []fileToParse, err error) {

	// We need to walk from the root to deeper into the file tree to find everything.
	root, err := filepath.Abs(modelPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {

		// There was an error walking to this path.
		if err != nil {
			return errors.WithStack(err)
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return errors.WithStack(err)
		}

		// Everything that goes in the database is a file, not a directory.
		if d.IsDir() {
			return nil
		}

		toParseFile, err := newFileToParse(modelPath, relPath, path)
		if err != nil {
			return err
		}

		toParseFiles = append(toParseFiles, toParseFile)

		return nil
	})
	if err != nil {
		return nil, err
	}

	sortFilesToParse(toParseFiles)

	return toParseFiles, nil
}

func parseForDatabase(modelKey string, filesToParse []fileToParse) (model req_model.Model, err error) {

	// Ensure are sorted in the order by which they may need information from prior objects.
	// This is for foreign keys.
	sortFilesToParse(filesToParse)

	// Track domains by their key so we can look them up when parsing classes/use cases.
	// We store them directly in the model map and track by string key for lookup.
	domainKeysBySubKey := map[string]identity.Key{}

	// Track subdomains by their composite path key (domain/subdomain) for lookup.
	subdomainKeysByPath := map[string]identity.Key{}

	// Collect all class associations from all classes, then distribute after all classes are parsed.
	allClassAssociations := map[identity.Key]model_class.Association{}

	// Now, parse each file according to its type.

	for _, toParseFile := range filesToParse {

		// Load the file data.
		contentBytes, err := os.ReadFile(toParseFile.PathAbs)
		if err != nil {
			return req_model.Model{}, errors.WithStack(err)
		}

		// Trim any space on it.
		contents := string(contentBytes)

		// Handle each kind of file differently.
		switch toParseFile.FileType {

		case _EXT_MODEL:
			model, err = parseModel(modelKey, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}
			// Initialize maps.
			model.Actors = make(map[identity.Key]model_actor.Actor)
			model.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization)
			model.Domains = make(map[identity.Key]model_domain.Domain)
			model.DomainAssociations = make(map[identity.Key]model_domain.Association)
			model.ClassAssociations = make(map[identity.Key]model_class.Association)

		case _EXT_ACTOR:
			actor, err := parseActor(toParseFile.Actor, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}
			model.Actors[actor.Key] = actor

		case _EXT_GENERALIZATION:
			if toParseFile.Domain == "" {
				// Actor generalization (under actors/ directory).
				actorGen, err := parseActorGeneralization(toParseFile.Generalization, toParseFile.PathRel, contents)
				if err != nil {
					return req_model.Model{}, err
				}
				if model.ActorGeneralizations == nil {
					model.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization)
				}
				model.ActorGeneralizations[actorGen.Key] = actorGen
			} else {
				// Class generalization (under domain/subdomain directory).
				domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
				if !ok {
					return req_model.Model{}, errors.Errorf("domain '%s' not found for generalization '%s'", toParseFile.Domain, toParseFile.Generalization)
				}
				domain := model.Domains[domainKey]

				// Determine which subdomain to use (explicit or default).
				subdomainName := toParseFile.Subdomain
				if subdomainName == "" {
					subdomainName = "default"
				}
				subdomainKey, ok := subdomainKeysByPath[toParseFile.Domain+"/"+subdomainName]
				if !ok {
					return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for generalization '%s'", subdomainName, toParseFile.Domain, toParseFile.Generalization)
				}
				subdomain, ok := domain.Subdomains[subdomainKey]
				if !ok {
					return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for generalization '%s'", subdomainName, toParseFile.Domain, toParseFile.Generalization)
				}

				generalization, err := parseClassGeneralization(subdomainKey, toParseFile.Generalization, toParseFile.PathRel, contents)
				if err != nil {
					return req_model.Model{}, err
				}
				if subdomain.Generalizations == nil {
					subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
				}
				subdomain.Generalizations[generalization.Key] = generalization
				domain.Subdomains[subdomainKey] = subdomain
				model.Domains[domainKey] = domain
			}

		case _EXT_DOMAIN:
			domain, associations, err := parseDomain(toParseFile.Domain, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			// Add associations to model level.
			for _, assoc := range associations {
				model.DomainAssociations[assoc.Key] = assoc
			}

			// Give each domain a default subdomain.
			defaultSubdomainKey, err := identity.NewSubdomainKey(domain.Key, "default")
			if err != nil {
				return req_model.Model{}, errors.WithStack(err)
			}
			subdomain, err := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "", "")
			if err != nil {
				return req_model.Model{}, err
			}
			domain.Subdomains = map[identity.Key]model_domain.Subdomain{
				defaultSubdomainKey: subdomain,
			}

			model.Domains[domain.Key] = domain
			domainKeysBySubKey[toParseFile.Domain] = domain.Key
			// Track default subdomain by path for lookup.
			subdomainKeysByPath[toParseFile.Domain+"/default"] = defaultSubdomainKey

		case _EXT_SUBDOMAIN:
			// Find the domain for this subdomain.
			domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for subdomain '%s'", toParseFile.Domain, toParseFile.Subdomain)
			}
			domain := model.Domains[domainKey]

			subdomain, err := parseSubdomain(domainKey, toParseFile.Subdomain, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			domain.Subdomains[subdomain.Key] = subdomain
			model.Domains[domainKey] = domain
			// Track subdomain by path for lookup.
			subdomainKeysByPath[toParseFile.Domain+"/"+toParseFile.Subdomain] = subdomain.Key

		case _EXT_CLASS:
			// Need to find the domain for this class.
			domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for class '%s'", toParseFile.Domain, toParseFile.Class)
			}
			domain := model.Domains[domainKey]

			// Determine which subdomain to use (explicit or default).
			subdomainName := toParseFile.Subdomain
			if subdomainName == "" {
				subdomainName = "default"
			}
			subdomainKey, ok := subdomainKeysByPath[toParseFile.Domain+"/"+subdomainName]
			if !ok {
				return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for class '%s'", subdomainName, toParseFile.Domain, toParseFile.Class)
			}
			subdomain, ok := domain.Subdomains[subdomainKey]
			if !ok {
				return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for class '%s'", subdomainName, toParseFile.Domain, toParseFile.Class)
			}

			// Extract just the class subkey from the full class path (domain/classname -> classname).
			classSubKey := toParseFile.Class
			if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.Class) {
				classSubKey = toParseFile.Class[idx:]
			}

			class, associations, err := parseClass(subdomainKey, classSubKey, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			// Collect associations for distribution after all classes are parsed.
			for _, assoc := range associations {
				allClassAssociations[assoc.Key] = assoc
			}

			// Add the class to the subdomain.
			if subdomain.Classes == nil {
				subdomain.Classes = make(map[identity.Key]model_class.Class)
			}
			subdomain.Classes[class.Key] = class
			domain.Subdomains[subdomainKey] = subdomain
			model.Domains[domainKey] = domain

		case _EXT_USE_CASE:
			// Need to find the domain for this use case.
			domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for use case '%s'", toParseFile.Domain, toParseFile.UseCase)
			}
			domain := model.Domains[domainKey]

			// Determine which subdomain to use (explicit or default).
			subdomainName := toParseFile.Subdomain
			if subdomainName == "" {
				subdomainName = "default"
			}
			subdomainKey, ok := subdomainKeysByPath[toParseFile.Domain+"/"+subdomainName]
			if !ok {
				return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for use case '%s'", subdomainName, toParseFile.Domain, toParseFile.UseCase)
			}
			subdomain, ok := domain.Subdomains[subdomainKey]
			if !ok {
				return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for use case '%s'", subdomainName, toParseFile.Domain, toParseFile.UseCase)
			}

			// Extract just the use case subkey from the full use case path (domain/usecasename -> usecasename).
			useCaseSubKey := toParseFile.UseCase
			if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.UseCase) {
				useCaseSubKey = toParseFile.UseCase[idx:]
			}

			useCase, err := parseUseCase(subdomainKey, useCaseSubKey, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			// Add the use case to the subdomain.
			if subdomain.UseCases == nil {
				subdomain.UseCases = make(map[identity.Key]model_use_case.UseCase)
			}
			subdomain.UseCases[useCase.Key] = useCase
			domain.Subdomains[subdomainKey] = subdomain
			model.Domains[domainKey] = domain

		default:
			return req_model.Model{}, errors.WithStack(errors.Errorf(`unknown filetype: '%s'`, toParseFile.FileType))
		}
	}

	// Distribute class associations to the correct level (model, domain, subdomain).
	if len(allClassAssociations) > 0 {
		if err := model.SetClassAssociations(allClassAssociations); err != nil {
			return req_model.Model{}, errors.Wrap(err, "failed to set class associations")
		}
	}

	return model, nil
}
