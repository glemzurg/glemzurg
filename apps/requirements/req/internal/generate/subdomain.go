package generate

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"

	"github.com/pkg/errors"
)

func generateSubdomainMdContents(reqs *req_flat.Requirements, model core.Model, domain model_domain.Domain, subdomain model_domain.Subdomain, classesDiagram, useCasesDiagram string) (contents string, err error) {
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
		Reqs                   *req_flat.Requirements
		Model                  core.Model
		Domain                 model_domain.Domain
		Subdomain              model_domain.Subdomain
		Classes                []model_class.Class
		ExternalDiagramClasses []model_class.Class
		UseCases               []model_use_case.UseCase
		ClassesDiagram         string
		UseCasesDiagram        string
	}{
		Reqs:                   reqs,
		Model:                  model,
		Domain:                 domain,
		Subdomain:              subdomain,
		Classes:                allClasses,
		ExternalDiagramClasses: diagramClassesOutsideSubdomain(reqs, subdomain),
		UseCases:               allUseCases,
		ClassesDiagram:         classesDiagram,
		UseCasesDiagram:        useCasesDiagram,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// diagramClassesOutsideSubdomain returns classes in the subdomain class UML that live outside the subdomain.
func diagramClassesOutsideSubdomain(reqs *req_flat.Requirements, subdomain model_domain.Subdomain) []model_class.Class {
	var localClasses []model_class.Class
	for _, class := range subdomain.Classes {
		localClasses = append(localClasses, class)
	}
	_, diagramClasses, _ := reqs.RegardingClasses(localClasses)

	localKeys := make(map[string]struct{}, len(localClasses))
	for _, class := range localClasses {
		localKeys[class.Key.String()] = struct{}{}
	}

	var external []model_class.Class
	for _, class := range diagramClasses {
		if _, inSubdomain := localKeys[class.Key.String()]; inSubdomain {
			continue
		}
		external = append(external, class)
	}
	return classesSortedByName(external)
}
