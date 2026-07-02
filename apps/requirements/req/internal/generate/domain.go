package generate

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"

	"github.com/pkg/errors"
)

func generateDomainMdContents(reqs *req_flat.Requirements, model core.Model, domain model_domain.Domain, diagrams domainDiagrams) (contents string, err error) {
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
		Reqs              *req_flat.Requirements
		Model             core.Model
		Domain            model_domain.Domain
		Classes           []model_class.Class
		Subdomains        []model_domain.Subdomain
		SubdomainsDiagram string
		ClassesDiagram    string
		UseCasesDiagram   string
	}{
		Reqs:              reqs,
		Model:             model,
		Domain:            domain,
		Classes:           allClasses,
		Subdomains:        subdomains,
		SubdomainsDiagram: diagrams.SubdomainsDiagram,
		ClassesDiagram:    diagrams.ClassesDiagram,
		UseCasesDiagram:   diagrams.UseCasesDiagram,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// domainDiagrams holds the Mermaid diagram strings for a domain page.
type domainDiagrams struct {
	SubdomainsDiagram string
	ClassesDiagram    string
	UseCasesDiagram   string
}

// generateDomainsMermaidContents generates Mermaid markup for the domains overview diagram.
func generateDomainsMermaidContents(reqs *req_flat.Requirements, domains []model_domain.Domain, associations []model_domain.Association) (contents string, err error) {
	contents, err = generateFromTemplate(_domainsMermaidTemplate, struct {
		Reqs         *req_flat.Requirements
		Domains      []model_domain.Domain
		Associations []model_domain.Association
	}{
		Reqs:         reqs,
		Domains:      domains,
		Associations: associations,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// generateSubdomainsMermaidContents generates Mermaid markup for the subdomains diagram.
func generateSubdomainsMermaidContents(reqs *req_flat.Requirements, domain model_domain.Domain) (contents string, err error) {
	// Get sorted subdomains.
	var subdomains []model_domain.Subdomain
	for _, subdomain := range domain.Subdomains {
		subdomains = append(subdomains, subdomain)
	}
	sort.Slice(subdomains, func(i, j int) bool {
		return subdomains[i].Key.String() < subdomains[j].Key.String()
	})

	contents, err = generateFromTemplate(_subdomainsMermaidTemplate, struct {
		Reqs       *req_flat.Requirements
		Domain     model_domain.Domain
		Subdomains []model_domain.Subdomain
	}{
		Reqs:       reqs,
		Domain:     domain,
		Subdomains: subdomains,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
