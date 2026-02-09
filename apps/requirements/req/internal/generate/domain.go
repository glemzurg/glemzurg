package generate

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

func generateDomainMdContents(reqs *req_flat.Requirements, model req_model.Model, domain model_domain.Domain) (contents string, err error) {

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
		Reqs       *req_flat.Requirements
		Model      req_model.Model
		Domain     model_domain.Domain
		Classes    []model_class.Class
		Subdomains []model_domain.Subdomain
	}{
		Reqs:       reqs,
		Model:      model,
		Domain:     domain,
		Classes:    allClasses,
		Subdomains: subdomains,
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
