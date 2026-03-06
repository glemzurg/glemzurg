package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"

	"github.com/pkg/errors"
)

func generateModelMdContents(reqs *req_flat.Requirements, model core.Model, actors []model_actor.Actor, domains []model_domain.Domain) (contents string, err error) {

	contents, err = generateFromTemplate(_modelMdTemplate, struct {
		Reqs    *req_flat.Requirements
		Model   core.Model
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
