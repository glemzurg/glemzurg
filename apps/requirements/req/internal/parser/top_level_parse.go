package parser

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/pkg/errors"
)

func Parse(modelPath string) (reqs requirements.Requirements, err error) {
	log.Printf("Parse files in '%s'", modelPath)

	// First, gather every thing we expect we need to parse, and what kind of entity it is.

	toParseFiles, err := findFilesToParse(modelPath)
	if err != nil {
		return requirements.Requirements{}, errors.WithStack(err)
	}

	for _, toParseFile := range toParseFiles {
		log.Println("   walk:", toParseFile.String())
	}

	reqs, err = parseForDatabase(modelPath, toParseFiles)
	if err != nil {
		return requirements.Requirements{}, errors.WithStack(err)
	}

	return reqs, nil
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

func parseForDatabase(modelKey string, filesToParse []fileToParse) (reqs requirements.Requirements, err error) {

	// Ensure are sorted in the order by which they may need information from prior objects.
	// This is for foreign keys.
	sortFilesToParse(filesToParse)

	// Allocate memory for structures.
	if reqs.Subdomains == nil {
		reqs.Subdomains = map[string][]model_domain.Subdomain{}
	}
	if reqs.Classes == nil {
		reqs.Classes = map[string][]model_class.Class{}
	}
	if reqs.Attributes == nil {
		reqs.Attributes = map[string][]model_class.Attribute{}
	}
	if reqs.States == nil {
		reqs.States = map[string][]model_state.State{}
	}
	if reqs.Events == nil {
		reqs.Events = map[string][]model_state.Event{}
	}
	if reqs.Guards == nil {
		reqs.Guards = map[string][]model_state.Guard{}
	}
	if reqs.Actions == nil {
		reqs.Actions = map[string][]model_state.Action{}
	}
	if reqs.Transitions == nil {
		reqs.Transitions = map[string][]model_state.Transition{}
	}
	if reqs.StateActions == nil {
		reqs.StateActions = map[string][]model_state.StateAction{}
	}
	if reqs.UseCases == nil {
		reqs.UseCases = map[string][]model_use_case.UseCase{}
	}
	if reqs.UseCaseActors == nil {
		reqs.UseCaseActors = map[string]map[string]model_use_case.UseCaseActor{}
	}
	if reqs.Scenarios == nil {
		reqs.Scenarios = map[string][]model_scenario.Scenario{}
	}
	if reqs.ScenarioObjects == nil {
		reqs.ScenarioObjects = map[string][]model_scenario.ScenarioObject{}
	}

	// Now, parse each file according to its type.

	for _, toParseFile := range filesToParse {

		// Load the file data.
		contentBytes, err := os.ReadFile(toParseFile.PathAbs)
		if err != nil {
			return requirements.Requirements{}, errors.WithStack(err)
		}

		// Trim any space on it.
		contents := string(contentBytes)

		// Handle each kind of file differently.
		switch toParseFile.FileType {

		case _EXT_MODEL:
			model, err := parseModel(modelKey, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}
			reqs.Model = model

		case _EXT_ACTOR:
			actor, err := parseActor(toParseFile.Actor, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}
			reqs.Actors = append(reqs.Actors, actor)

		case _EXT_GENERALIZATION:
			generalization, err := parseGeneralization(toParseFile.Generalization, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}
			reqs.Generalizations = append(reqs.Generalizations, generalization)

		case _EXT_DOMAIN:
			domain, err := parseDomain(toParseFile.Domain, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}
			reqs.Domains = append(reqs.Domains, domain)

			// Migrate associations to greater structure.
			reqs.DomainAssociations = append(reqs.DomainAssociations, domain.Associations...)
			domain.Associations = nil

			// Give each domain a default subdomain.
			for _, domain := range reqs.Domains {
				subdomain, err := model_domain.NewSubdomain(defaultSubdomain(domain.Key), "Default", "", "")
				if err != nil {
					return requirements.Requirements{}, err
				}
				reqs.Subdomains[domain.Key] = []model_domain.Subdomain{subdomain}
			}

		case _EXT_CLASS:

			class, err := parseClass(toParseFile.Class, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}

			// Migrate attributes to greater structure.
			reqs.Attributes[toParseFile.Class] = class.Attributes
			class.Attributes = nil

			// Migrate associations to greater structure.
			reqs.Associations = append(reqs.Associations, class.Associations...)
			class.Associations = nil

			// Migrate state actions to greate structure.
			for _, state := range class.States {
				reqs.StateActions[state.Key] = state.Actions
			}

			// Migrate states to greater structure.
			reqs.States[toParseFile.Class] = class.States
			class.States = nil

			// Migrate events to greater structure.
			reqs.Events[toParseFile.Class] = class.Events
			class.Events = nil

			// Migrate gaurds to greater structure.
			reqs.Guards[toParseFile.Class] = class.Guards
			class.Guards = nil

			// Migrate actions to greater structure.
			reqs.Actions[toParseFile.Class] = class.Actions
			class.Actions = nil

			// Migrate transitions to greater structure.
			reqs.Transitions[toParseFile.Class] = class.Transitions
			class.Transitions = nil

			// Add the class itself.
			classes := reqs.Classes[defaultSubdomain(toParseFile.Domain)]
			classes = append(classes, class)
			reqs.Classes[defaultSubdomain(toParseFile.Domain)] = classes

		case _EXT_USE_CASE:

			useCase, err := parseUseCase(toParseFile.UseCase, toParseFile.PathRel, contents)
			if err != nil {
				return requirements.Requirements{}, err
			}

			// Add the use case itself.
			useCases := reqs.UseCases[defaultSubdomain(toParseFile.Domain)]
			useCases = append(useCases, useCase)
			reqs.UseCases[defaultSubdomain(toParseFile.Domain)] = useCases

			// Put the parts on the greater structure.
			reqs.UseCaseActors[useCase.Key] = useCase.Actors
			reqs.Scenarios[useCase.Key] = useCase.Scenarios
			for _, scenario := range useCase.Scenarios {
				reqs.ScenarioObjects[scenario.Key] = scenario.Objects
			}

		default:
			return requirements.Requirements{}, errors.WithStack(errors.Errorf(`unknown filetype: '%s'`, toParseFile.FileType))
		}
	}

	return reqs, nil
}

func defaultSubdomain(domainKey string) (subdomainKey string) {
	return domainKey
}
