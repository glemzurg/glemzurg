package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model"

	"github.com/pkg/errors"
)

func generateModelFiles(debug bool, outputPath string, reqs requirements.Requirements) (err error) {

	// The data we want.
	model := reqs.Model
	actors := reqs.Actors
	domains := reqs.Domains
	associations := reqs.DomainAssociations

	// Generate model summary.
	modelFilename := "model.md"
	modelFilenameAbs := filepath.Join(outputPath, modelFilename)
	mdContents, err := generateModelMdContents(reqs, model, actors, domains)
	if err != nil {
		return err
	}
	if err = writeFile(modelFilenameAbs, mdContents); err != nil {
		return err
	}

	// Generate domains diagram.
	domainsFilename := "domains.svg"
	domainsFilenameAbs := filepath.Join(outputPath, domainsFilename)
	svgContents, dotContents, err := generateDomainsSvgContents(reqs, domains, associations)
	if err != nil {
		return err
	}
	if err = writeFile(domainsFilenameAbs, svgContents); err != nil {
		return err
	}
	if err := debugWriteDotFile(debug, outputPath, domainsFilename, dotContents); err != nil {
		return err
	}

	return nil
}

func generateModelMdContents(reqs requirements.Requirements, model model.Model, actors []actor.Actor, domains []domain.Domain) (contents string, err error) {

	contents, err = generateFromTemplate(_modelMdTemplate, struct {
		Reqs    requirements.Requirements
		Model   model.Model
		Actors  []actor.Actor
		Domains []domain.Domain
	}{
		Reqs:    reqs,
		Model:   model,
		Actors:  actors,
		Domains: domains,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
