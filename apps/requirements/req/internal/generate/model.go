package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"

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

func generateModelMdContents(reqs requirements.Requirements, model req_model.Model, actors []model_actor.Actor, domains []model_domain.Domain) (contents string, err error) {

	contents, err = generateFromTemplate(_modelMdTemplate, struct {
		Reqs    requirements.Requirements
		Model   req_model.Model
		Actors  []model_actor.Actor
		Domains []model_domain.Domain
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
