package parser

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

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

	// Track domains by their string key so we can look them up when parsing classes/use cases.
	domainsByKey := map[string]*model_domain.Domain{}

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

		case _EXT_ACTOR:
			actor, err := parseActor(toParseFile.Actor, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}
			model.Actors = append(model.Actors, actor)

		case _EXT_GENERALIZATION:
			// Need to find the domain for this generalization.
			domain, ok := domainsByKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for generalization '%s'", toParseFile.Domain, toParseFile.Generalization)
			}
			// Use the default subdomain.
			if len(domain.Subdomains) == 0 {
				return req_model.Model{}, errors.Errorf("domain '%s' has no subdomains for generalization '%s'", toParseFile.Domain, toParseFile.Generalization)
			}
			subdomainKey := domain.Subdomains[0].Key

			generalization, err := parseGeneralization(subdomainKey, toParseFile.Generalization, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}
			domain.Subdomains[0].Generalizations = append(domain.Subdomains[0].Generalizations, generalization)

		case _EXT_DOMAIN:
			domain, err := parseDomain(toParseFile.Domain, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			// Move associations to model level.
			model.DomainAssociations = append(model.DomainAssociations, domain.Associations...)
			domain.Associations = nil

			// Give each domain a default subdomain.
			defaultSubdomainKey, err := identity.NewSubdomainKey(domain.Key, "default")
			if err != nil {
				return req_model.Model{}, errors.WithStack(err)
			}
			subdomain, err := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "", "")
			if err != nil {
				return req_model.Model{}, err
			}
			domain.Subdomains = []model_domain.Subdomain{subdomain}

			model.Domains = append(model.Domains, domain)
			// Store pointer to domain so we can modify it when adding classes/use cases.
			domainsByKey[toParseFile.Domain] = &model.Domains[len(model.Domains)-1]

		case _EXT_CLASS:
			// Need to find the domain for this class.
			domain, ok := domainsByKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for class '%s'", toParseFile.Domain, toParseFile.Class)
			}
			// Use the default subdomain.
			if len(domain.Subdomains) == 0 {
				return req_model.Model{}, errors.Errorf("domain '%s' has no subdomains for class '%s'", toParseFile.Domain, toParseFile.Class)
			}
			subdomainKey := domain.Subdomains[0].Key

			// Extract just the class subkey from the full class path (domain/classname -> classname).
			classSubKey := toParseFile.Class
			if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.Class) {
				classSubKey = toParseFile.Class[idx:]
			}

			class, err := parseClass(subdomainKey, classSubKey, toParseFile.PathRel, contents)
			if err != nil {
				return req_model.Model{}, err
			}

			// Move associations to subdomain level.
			domain.Subdomains[0].Associations = append(domain.Subdomains[0].Associations, class.Associations...)
			class.Associations = nil

			// Add the class to the subdomain.
			domain.Subdomains[0].Classes = append(domain.Subdomains[0].Classes, class)

		case _EXT_USE_CASE:
			// Need to find the domain for this use case.
			domain, ok := domainsByKey[toParseFile.Domain]
			if !ok {
				return req_model.Model{}, errors.Errorf("domain '%s' not found for use case '%s'", toParseFile.Domain, toParseFile.UseCase)
			}
			// Use the default subdomain.
			if len(domain.Subdomains) == 0 {
				return req_model.Model{}, errors.Errorf("domain '%s' has no subdomains for use case '%s'", toParseFile.Domain, toParseFile.UseCase)
			}
			subdomainKey := domain.Subdomains[0].Key

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
			domain.Subdomains[0].UseCases = append(domain.Subdomains[0].UseCases, useCase)

		default:
			return req_model.Model{}, errors.WithStack(errors.Errorf(`unknown filetype: '%s'`, toParseFile.FileType))
		}
	}

	return model, nil
}
