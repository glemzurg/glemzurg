package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

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
