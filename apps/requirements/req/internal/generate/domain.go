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

func generateDomainFiles(debug bool, outputPath string, reqs *req_flat.Requirements) (err error) {

	// Get all the data we want for these files.
	domainLookup, _ := reqs.DomainLookup()

	// Generate file for each domain.
	for _, domain := range domainLookup {

		// Check if domain has multiple subdomains.
		hasMultipleSubdomains := len(domain.Subdomains) > 1

		// Generate domain markdown page.
		modelFilename := convertKeyToFilename("domain", domain.Key.String(), "", ".md")
		modelFilenameAbs := filepath.Join(outputPath, modelFilename)
		mdContents, err := generateDomainMdContents(reqs, reqs.Model, domain, hasMultipleSubdomains)
		if err != nil {
			return err
		}
		if err = writeFile(modelFilenameAbs, mdContents); err != nil {
			return err
		}

		if hasMultipleSubdomains {
			// Generate subdomains diagram (graph of subdomains).
			subdomainsFilename := convertKeyToFilename("domain", domain.Key.String(), "subdomains", ".svg")
			subdomainsFilenameAbs := filepath.Join(outputPath, subdomainsFilename)
			svgContents, dotContents, err := generateSubdomainsSvgContents(reqs, domain)
			if err != nil {
				return err
			}
			if err = writeFile(subdomainsFilenameAbs, svgContents); err != nil {
				return err
			}
			if err := debugWriteDotFile(debug, outputPath, subdomainsFilename, dotContents); err != nil {
				return err
			}
		} else {
			// Single subdomain: generate use cases and classes diagrams at domain level.

			// Gather all classes from all subdomains for this domain.
			var domainClasses []model_class.Class
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					domainClasses = append(domainClasses, class)
				}
			}

			// Generate use cases diagram.
			useCasesFilename := convertKeyToFilename("domain", domain.Key.String(), "use-cases", ".svg")
			useCasesFilenameAbs := filepath.Join(outputPath, useCasesFilename)

			// Gather all use cases from all subdomains for this domain.
			var domainUseCases []model_use_case.UseCase
			for _, subdomain := range domain.Subdomains {
				for _, useCase := range subdomain.UseCases {
					domainUseCases = append(domainUseCases, useCase)
				}
			}

			relevantUseCases, relevantActors, err := reqs.RegardingUseCases(domainUseCases)
			if err != nil {
				return err
			}
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
			generalizations, classes, associations := reqs.RegardingClasses(domainClasses)

			// Generate classes diagram.
			classesFilename := convertKeyToFilename("domain", domain.Key.String(), "classes", ".svg")
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

func generateDomainMdContents(reqs *req_flat.Requirements, model req_model.Model, domain model_domain.Domain, hasMultipleSubdomains bool) (contents string, err error) {

	// Gather all classes from all subdomains for sorting.
	var allClasses []model_class.Class
	for _, subdomain := range domain.Subdomains {
		for _, class := range subdomain.Classes {
			allClasses = append(allClasses, class)
		}
	}

	sort.Slice(allClasses, func(i, j int) bool {
		return allClasses[i].Name < allClasses[j].Name
	})

	// Get sorted subdomains for multi-subdomain mode.
	var subdomains []model_domain.Subdomain
	for _, subdomain := range domain.Subdomains {
		subdomains = append(subdomains, subdomain)
	}
	sort.Slice(subdomains, func(i, j int) bool {
		return subdomains[i].Key.String() < subdomains[j].Key.String()
	})

	contents, err = generateFromTemplate(_domainMdTemplate, struct {
		Reqs                  *req_flat.Requirements
		Model                 req_model.Model
		Domain                model_domain.Domain
		Classes               []model_class.Class
		HasMultipleSubdomains bool
		Subdomains            []model_domain.Subdomain
	}{
		Reqs:                  reqs,
		Model:                 model,
		Domain:                domain,
		Classes:               allClasses,
		HasMultipleSubdomains: hasMultipleSubdomains,
		Subdomains:            subdomains,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// This is the domain graph on the model page.
func generateDomainsSvgContents(reqs *req_flat.Requirements, domains []model_domain.Domain, associations []model_domain.Association) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_domainsDotTemplate, struct {
		Reqs         *req_flat.Requirements
		Domains      []model_domain.Domain
		Associations []model_domain.Association
	}{
		Reqs:         reqs,
		Domains:      domains,
		Associations: associations,
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

// generateSubdomainsSvgContents generates the SVG graph of subdomains for a domain.
func generateSubdomainsSvgContents(reqs *req_flat.Requirements, domain model_domain.Domain) (svgContents string, dotContents string, err error) {

	// Get sorted subdomains.
	var subdomains []model_domain.Subdomain
	for _, subdomain := range domain.Subdomains {
		subdomains = append(subdomains, subdomain)
	}
	sort.Slice(subdomains, func(i, j int) bool {
		return subdomains[i].Key.String() < subdomains[j].Key.String()
	})

	dotContents, err = generateFromTemplate(_subdomainsDotTemplate, struct {
		Reqs       *req_flat.Requirements
		Domain     model_domain.Domain
		Subdomains []model_domain.Subdomain
	}{
		Reqs:       reqs,
		Domain:     domain,
		Subdomains: subdomains,
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
