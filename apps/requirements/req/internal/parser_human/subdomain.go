package parser_human

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

func parseSubdomain(domainKey identity.Key, subdomainSubKey, filename, contents string) (subdomain model_domain.Subdomain, err error) {
	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_domain.Subdomain{}, err
	}

	// There is no data for a "subdomain" entity. Just add to the markdown
	// so that it makes it into the output.
	markdown := parsedFile.Markdown

	if parsedFile.Data != "" {
		markdown += "\n\n" + parsedFile.Data
	}

	// Construct the identity key for this subdomain.
	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomainSubKey)
	if err != nil {
		return model_domain.Subdomain{}, errors.WithStack(err)
	}

	subdomain = model_domain.NewSubdomain(subdomainKey, parsedFile.Title, stripMarkdownTitle(markdown), parsedFile.UmlComment)
	return subdomain, nil
}

func generateSubdomainContent(subdomain model_domain.Subdomain) string {
	return generateFileContent(prependMarkdownTitle(subdomain.Name, subdomain.Details), subdomain.UmlComment, "")
}
