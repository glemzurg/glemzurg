package generate

import (
	"path/filepath"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

func generateSubdomainFiles(debug bool, outputPath string, reqs *req_flat.Requirements) (err error) {

	// Get all domains.
	domainLookup, _ := reqs.DomainLookup()

	// Generate subdomain files only for domains with multiple subdomains.
	for _, domain := range domainLookup {
		// Skip if only one subdomain (use domain-level rendering instead).
		if len(domain.Subdomains) <= 1 {
			continue
		}

		// Generate file for each subdomain.
		for _, subdomain := range domain.Subdomains {

			// Generate subdomain markdown page.
			subdomainFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")
			subdomainFilenameAbs := filepath.Join(outputPath, subdomainFilename)
			mdContents, err := generateSubdomainMdContents(reqs, reqs.Model, domain, subdomain)
			if err != nil {
				return err
			}
			if err = writeFile(subdomainFilenameAbs, mdContents); err != nil {
				return err
			}

			// Gather classes for this subdomain.
			var subdomainClasses []model_class.Class
			for _, class := range subdomain.Classes {
				subdomainClasses = append(subdomainClasses, class)
			}

			// Generate use cases diagram for subdomain.
			useCasesFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "use-cases", ".svg")
			useCasesFilenameAbs := filepath.Join(outputPath, useCasesFilename)

			// Gather use cases for this subdomain.
			var subdomainUseCases []model_use_case.UseCase
			for _, useCase := range subdomain.UseCases {
				subdomainUseCases = append(subdomainUseCases, useCase)
			}

			relevantUseCases, relevantActors, err := reqs.RegardingUseCases(subdomainUseCases)
			if err != nil {
				return err
			}

			// Reuse domain's use case SVG generation (domain name shown in cluster label provides context).
			useCasesSvgContents, useCasesDotContents, err := generateUseCasesSvgContents(reqs, domain, relevantUseCases, relevantActors)
			if err != nil {
				return err
			}
			if err := debugWriteDotFile(debug, outputPath, useCasesFilename, useCasesDotContents); err != nil {
				return err
			}
			if err = writeFile(useCasesFilenameAbs, useCasesSvgContents); err != nil {
				return err
			}

			// Get the data that is important for this class diagram.
			generalizations, classes, associations := reqs.RegardingClasses(subdomainClasses)

			// Generate classes diagram.
			classesFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "classes", ".svg")
			classesFilenameAbs := filepath.Join(outputPath, classesFilename)
			classesSvgContents, classesDotContents, err := generateClassesSvgContents(reqs, generalizations, classes, associations)
			if err != nil {
				return err
			}
			if err = writeFile(classesFilenameAbs, classesSvgContents); err != nil {
				return err
			}
			if err = debugWriteDotFile(debug, outputPath, classesFilename, classesDotContents); err != nil {
				return err
			}
		}
	}

	return nil
}

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
