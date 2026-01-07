package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
)

func parseSubdomain(key, filename, contents string) (subdomain model_domain.Subdomain, err error) {

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

	subdomain, err = model_domain.NewSubdomain(key, parsedFile.Title, markdown, parsedFile.UmlComment)
	if err != nil {
		return model_domain.Subdomain{}, err
	}
	return subdomain, nil
}

func generateSubdomainContent(subdomain model_domain.Subdomain) string {
	return generateFileContent(subdomain.Details, subdomain.UmlComment, "")
}
