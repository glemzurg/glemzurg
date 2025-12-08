package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

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

func generateUseCaseMdContents(reqs requirements.Requirements, useCase requirements.UseCase) (contents string, err error) {

	contents, err = generateFromTemplate(_useCaseMdTemplate, struct {
		Reqs    requirements.Requirements
		UseCase requirements.UseCase
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
func generateUseCasesSvgContents(reqs requirements.Requirements, domain requirements.Domain, useCases []requirements.UseCase, actors []requirements.Actor) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_useCasesDotTemplate, struct {
		Reqs     requirements.Requirements
		Domain   requirements.Domain
		UseCases []requirements.UseCase
		Actors   []requirements.Actor
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
