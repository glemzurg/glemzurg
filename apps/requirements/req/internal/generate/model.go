package generate

import (
	"path/filepath"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

func generateModelFiles(debug bool, outputPath string, reqs *req_flat.Requirements) (err error) {

	// The data we want.
	model := reqs.Model

	// Get actors as a sorted slice.
	var actors []model_actor.Actor
	for _, actor := range reqs.Actors {
		actors = append(actors, actor)
	}
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Key.String() < actors[j].Key.String()
	})

	// Get domains as a sorted slice.
	var domains []model_domain.Domain
	for _, domain := range reqs.Domains {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Key.String() < domains[j].Key.String()
	})

	// Get domain associations as a sorted slice.
	var associations []model_domain.Association
	for _, assoc := range reqs.DomainAssociations {
		associations = append(associations, assoc)
	}
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key.String() < associations[j].Key.String()
	})

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

func generateModelMdContents(reqs *req_flat.Requirements, model req_model.Model, actors []model_actor.Actor, domains []model_domain.Domain) (contents string, err error) {

	contents, err = generateFromTemplate(_modelMdTemplate, struct {
		Reqs    *req_flat.Requirements
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
