package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"

	"github.com/pkg/errors"
)

func generateUseCaseMdContents(reqs *req_flat.Requirements, useCase model_use_case.UseCase) (contents string, err error) {
	contents, err = generateFromTemplate(_useCaseMdTemplate, struct {
		Reqs    *req_flat.Requirements
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

// generateUseCasesMermaidContents generates Mermaid use case diagram markup.
func generateUseCasesMermaidContents(reqs *req_flat.Requirements, domain model_domain.Domain, useCases []model_use_case.UseCase, actors []model_actor.Actor) (contents string, err error) {
	contents, err = generateFromTemplate(_useCasesMermaidTemplate, struct {
		Reqs     *req_flat.Requirements
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
		return "", errors.WithStack(err)
	}

	return contents, nil
}
