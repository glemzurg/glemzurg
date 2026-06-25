package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/modelfacts"

	"github.com/pkg/errors"
)

func generateSubdomainFactsMdContents(reqs *req_flat.Requirements, model core.Model, domain model_domain.Domain, subdomain model_domain.Subdomain) (contents string, err error) {
	facts := modelfacts.FactsForSubdomain(subdomain)

	contents, err = generateFromTemplate(_factsMdTemplate, struct {
		Reqs      *req_flat.Requirements
		Model     core.Model
		Domain    model_domain.Domain
		Subdomain model_domain.Subdomain
		Facts     modelfacts.SubdomainFacts
	}{
		Reqs:      reqs,
		Model:     model,
		Domain:    domain,
		Subdomain: subdomain,
		Facts:     facts,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
