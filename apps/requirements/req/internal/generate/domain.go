package generate

import (
	"path/filepath"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"

	"github.com/pkg/errors"
)

func generateDomainFiles(debug bool, outputPath string, reqs requirements.Requirements) (err error) {

	// Get all the data we want for these files.
	domainLookup, _ := reqs.DomainLookup()

	// Generate file for each domain.
	for _, domain := range domainLookup {

		// Generate model summary.
		modelFilename := convertKeyToFilename("domain", domain.Key, "", ".md")
		modelFilenameAbs := filepath.Join(outputPath, modelFilename)
		mdContents, err := generateDomainMdContents(reqs, reqs.Model, domain)
		if err != nil {
			return err
		}
		if err = writeFile(modelFilenameAbs, mdContents); err != nil {
			return err
		}

		// Generate use cases diagram.
		useCasesFilename := convertKeyToFilename("domain", domain.Key, "use-cases", ".svg")
		useCasesFilenameAbs := filepath.Join(outputPath, useCasesFilename)
		relevantUseCases, relevantActors, err := reqs.RegardingUseCases(domain.UseCases)
		if err != nil {
			return err
		}
		useCasesSvgContents, useCasesDotContents, err := generateUseCasesSvgContents(reqs, domain, relevantUseCases, relevantActors)
		if err != nil {
			return err
		}
		if err := debugWriteDotFile(debug, outputPath, useCasesFilename, useCasesDotContents); err != nil {
			return err
		}
		if err = writeFile(useCasesFilenameAbs, useCasesSvgContents); err != nil {
			return err
		}

		// Get the data that is important for this class diagram.
		generalizations, classes, associations := reqs.RegardingClasses(domain.Classes)

		// Generate classes diagram.
		classesFilename := convertKeyToFilename("domain", domain.Key, "classes", ".svg")
		classesFilenameAbs := filepath.Join(outputPath, classesFilename)
		classesSvgContents, classesDotContents, err := generateClassesSvgContents(reqs, generalizations, classes, associations)
		if err != nil {
			return err
		}
		if err = writeFile(classesFilenameAbs, classesSvgContents); err != nil {
			return err
		}
		if err = debugWriteDotFile(debug, outputPath, classesFilename, classesDotContents); err != nil {
			return err
		}
	}

	return nil
}

func generateDomainMdContents(reqs requirements.Requirements, model requirements.Model, domain model_domain.Domain) (contents string, err error) {

	sort.Slice(domain.Classes, func(i, j int) bool {
		return domain.Classes[i].Name < domain.Classes[j].Name
	})

	contents, err = generateFromTemplate(_domainMdTemplate, struct {
		Reqs   requirements.Requirements
		Model  requirements.Model
		Domain model_domain.Domain
	}{
		Reqs:   reqs,
		Model:  model,
		Domain: domain,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// This is the domain graph on the model page.
func generateDomainsSvgContents(reqs requirements.Requirements, domains []model_domain.Domain, associations []model_domain.Association) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_domainsDotTemplate, struct {
		Reqs         requirements.Requirements
		Domains      []model_domain.Domain
		Associations []model_domain.Association
	}{
		Reqs:         reqs,
		Domains:      domains,
		Associations: associations,
	})
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	svgContents, err = graphvizDotToSvg(dotContents)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return svgContents, dotContents, nil
}
