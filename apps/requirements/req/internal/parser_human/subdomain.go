package parser_human

import (
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseSubdomain(domainKey identity.Key, subdomainSubKey, filename, contents string) (subdomain model_domain.Subdomain, associations []model_domain.SubdomainAssociation, err error) {
	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}

	markdown := parsedFile.Markdown

	yamlData := map[string]any{}
	if parsedFile.Data != "" {
		if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
			return model_domain.Subdomain{}, nil, errors.WithStack(err)
		}
	}

	subdomainKey, err := identity.NewSubdomainKey(domainKey, subdomainSubKey)
	if err != nil {
		return model_domain.Subdomain{}, nil, errors.WithStack(err)
	}

	subdomain = model_domain.NewSubdomain(subdomainKey, parsedFile.Title, stripMarkdownTitle(markdown), parsedFile.UnfinishedNotes, parsedFile.UmlComment)

	var associationsData []any
	associationsAny, found := yamlData["associations"]
	if found {
		associationsData = associationsAny.([]any)
	}
	for i, associationAny := range associationsData {
		association, err := subdomainAssociationFromYamlData(domainKey, subdomainKey, i, associationAny)
		if err != nil {
			return model_domain.Subdomain{}, nil, err
		}
		associations = append(associations, association)
	}

	return subdomain, associations, nil
}

func subdomainAssociationFromYamlData(domainKey, problemSubdomainKey identity.Key, index int, associationAny any) (association model_domain.SubdomainAssociation, err error) {
	associationData, ok := associationAny.(map[string]any)
	if ok {
		_ = strconv.Itoa(index + 1)

		solutionSubdomainKeyStr := ""
		solutionSubdomainKeyAny, found := associationData["solution_subdomain_key"]
		if found {
			solutionSubdomainKeyStr = solutionSubdomainKeyAny.(string)
		}

		umlComment := ""
		umlCommentAny, found := associationData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		solutionSubdomainKey, err := identity.NewSubdomainKey(domainKey, solutionSubdomainKeyStr)
		if err != nil {
			return model_domain.SubdomainAssociation{}, errors.WithStack(err)
		}

		assocKey, err := identity.NewSubdomainAssociationKey(domainKey, problemSubdomainKey, solutionSubdomainKey)
		if err != nil {
			return model_domain.SubdomainAssociation{}, errors.WithStack(err)
		}

		association = model_domain.NewSubdomainAssociation(
			assocKey,
			problemSubdomainKey,
			solutionSubdomainKey,
			umlComment)
	}

	return association, nil
}

func generateSubdomainContent(subdomain model_domain.Subdomain, associations []model_domain.SubdomainAssociation) string {
	var yb strings.Builder
	if len(associations) > 0 {
		yb.WriteString("associations:\n")
		for _, assoc := range associations {
			yb.WriteString("\n    - solution_subdomain_key: " + assoc.SolutionSubdomainKey.SubKey + "\n")
			yb.WriteString(formatYamlField("uml_comment", assoc.UmlComment, 6))
		}
	}

	yamlStr := strings.TrimSpace(yb.String())
	return generateFileContent(prependMarkdownTitle(subdomain.Name, subdomain.Details), subdomain.UnfinishedNotes, subdomain.UmlComment, yamlStr)
}
