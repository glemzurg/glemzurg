package parser

import (
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/domain"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseDomain(key, filename, contents string) (parsedDomain domain.Domain, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return domain.Domain{}, err
	}

	// There is no data for a "domain" entity. Just add to the markdown
	// so that it makes it into the output.
	markdown := parsedFile.Markdown

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return domain.Domain{}, errors.WithStack(err)
	}

	realized := false
	realizedAny, found := yamlData["realized"]
	if found {
		realized = realizedAny.(bool)
	}

	parsedDomain, err = domain.NewDomain(key, parsedFile.Title, markdown, realized, parsedFile.UmlComment)
	if err != nil {
		return domain.Domain{}, err
	}

	// Add any associations we found.
	var associationsData []any
	associationsAny, found := yamlData["associations"]
	if found {
		associationsData = associationsAny.([]any)
	}

	var associations []domain.DomainAssociation
	for i, associationAny := range associationsData {
		association, err := domainAssociationFromYamlData(parsedDomain.Key, i, associationAny)
		if err != nil {
			return domain.Domain{}, err
		}
		associations = append(associations, association)
	}
	parsedDomain.Associations = associations

	return parsedDomain, nil
}

func domainAssociationFromYamlData(problemDomainKey string, index int, associationAny any) (association domain.DomainAssociation, err error) {

	associationData, ok := associationAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		key := problemDomainKey + "/association/" + strconv.Itoa(index+1) // Don't start at zero.

		solutionDomainKey := ""
		solutionDomainKeyAny, found := associationData["solution_domain_key"]
		if found {
			solutionDomainKey = solutionDomainKeyAny.(string)
		}

		umlComment := ""
		umlCommentAny, found := associationData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		association, err = domain.NewDomainAssociation(
			key,
			problemDomainKey,
			solutionDomainKey,
			umlComment)
		if err != nil {
			return domain.DomainAssociation{}, err
		}
	}

	return association, nil
}

func generateDomainContent(d domain.Domain) string {
	yaml := "realized: " + strconv.FormatBool(d.Realized)

	if len(d.Associations) > 0 {
		yaml += "\n\nassociations:\n"
		for _, assoc := range d.Associations {
			yaml += "\n    - solution_domain_key: " + assoc.SolutionDomainKey
			if assoc.UmlComment != "" {
				yaml += "\n      uml_comment: " + assoc.UmlComment
			}
			yaml += "\n"
		}
	}

	yamlStr := strings.TrimSpace(yaml)
	return generateFileContent(d.Details, d.UmlComment, yamlStr)
}
