package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func parseSubdomain(key, filename, contents string) (subdomain requirements.Subdomain, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.Subdomain{}, err
	}

	// There is no data for a "subdomain" entity. Just add to the markdown
	// so that it makes it into the output.
	markdown := parsedFile.Markdown

	if parsedFile.Data != "" {
		markdown += "\n\n" + parsedFile.Data
	}

	subdomain, err = requirements.NewSubdomain(key, parsedFile.Title, markdown, parsedFile.UmlComment)
	if err != nil {
		return requirements.Subdomain{}, err
	}
	return subdomain, nil
}

func generateSubdomainContent(subdomain requirements.Subdomain) string {
	return generateFileContent(subdomain.Details, subdomain.UmlComment, "")
}
