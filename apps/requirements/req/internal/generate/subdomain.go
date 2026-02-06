package generate

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

func generateSubdomainMdContents(reqs *req_flat.Requirements, model req_model.Model, domain model_domain.Domain, subdomain model_domain.Subdomain) (contents string, err error) {

	// Gather classes for sorting.
	var allClasses []model_class.Class
	for _, class := range subdomain.Classes {
		allClasses = append(allClasses, class)
	}
	sort.Slice(allClasses, func(i, j int) bool {
		return allClasses[i].Name < allClasses[j].Name
	})

	// Gather use cases for sorting.
	var allUseCases []model_use_case.UseCase
	for _, useCase := range subdomain.UseCases {
		allUseCases = append(allUseCases, useCase)
	}
	sort.Slice(allUseCases, func(i, j int) bool {
		return allUseCases[i].Key.String() < allUseCases[j].Key.String()
	})

	contents, err = generateFromTemplate(_subdomainMdTemplate, struct {
		Reqs      *req_flat.Requirements
		Model     req_model.Model
		Domain    model_domain.Domain
		Subdomain model_domain.Subdomain
		Classes   []model_class.Class
		UseCases  []model_use_case.UseCase
	}{
		Reqs:      reqs,
		Model:     model,
		Domain:    domain,
		Subdomain: subdomain,
		Classes:   allClasses,
		UseCases:  allUseCases,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
