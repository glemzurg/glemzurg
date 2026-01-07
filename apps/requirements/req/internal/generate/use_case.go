package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/pkg/errors"
)

func generateUseCaseFiles(outputPath string, reqs requirements.Requirements) (err error) {

	// Get all the data we want for these files.
	useCaseLookup := reqs.UseCaseLookup()

	// Generate file for each actor.
	for _, useCase := range useCaseLookup {

		// Generate model summary.
		modelFilename := convertKeyToFilename("use_case", useCase.Key, "", ".md")
		modelFilenameAbs := filepath.Join(outputPath, modelFilename)
		mdContents, err := generateUseCaseMdContents(reqs, useCase)
		if err != nil {
			return err
		}
		if err = writeFile(modelFilenameAbs, mdContents); err != nil {
			return err
		}
	}

	return nil
}

func generateUseCaseMdContents(reqs requirements.Requirements, useCase model_use_case.UseCase) (contents string, err error) {

	contents, err = generateFromTemplate(_useCaseMdTemplate, struct {
		Reqs    requirements.Requirements
		UseCase model_use_case.UseCase
	}{
		Reqs:    reqs,
		UseCase: useCase,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// This is the class graph on a domain and class pages.
func generateUseCasesSvgContents(reqs requirements.Requirements, domain model_domain.Domain, useCases []model_use_case.UseCase, actors []model_actor.Actor) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_useCasesDotTemplate, struct {
		Reqs     requirements.Requirements
		Domain   model_domain.Domain
		UseCases []model_use_case.UseCase
		Actors   []model_actor.Actor
	}{
		Reqs:     reqs,
		Domain:   domain,
		UseCases: useCases,
		Actors:   actors,
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
