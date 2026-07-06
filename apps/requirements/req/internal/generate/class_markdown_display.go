package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func classMarkdownDisplayName(reqs *req_flat.Requirements, viewerSubdomainKey identity.Key, targetClass model_class.Class) string {
	domainLookup := reqs.ClassDomainLookup()
	subdomainLookup := reqs.ClassSubdomainLookup()
	targetDomain := domainLookup[targetClass.Key.String()]
	targetSubdomain := subdomainLookup[targetClass.Key.String()]
	name, err := model_class.FormatClassMarkdownDisplayName(
		viewerSubdomainKey,
		targetClass,
		targetDomain.Name,
		targetSubdomain.Name,
	)
	if err != nil {
		return targetClass.Name
	}
	return name
}
