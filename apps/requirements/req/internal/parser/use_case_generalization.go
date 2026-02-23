package parser

import (
	"strconv"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseUseCaseGeneralization(subdomainKey identity.Key, generalizationSubKey, filename, contents string) (generalization model_use_case.Generalization, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_use_case.Generalization{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_use_case.Generalization{}, errors.WithStack(err)
	}

	isComplete := true
	isCompleteAny, found := yamlData["is_complete"]
	if found {
		isComplete = isCompleteAny.(bool)
	}

	isStatic := true
	isStaticAny, found := yamlData["is_static"]
	if found {
		isStatic = isStaticAny.(bool)
	}

	// Construct the identity key for this generalization.
	generalizationKey, err := identity.NewUseCaseGeneralizationKey(subdomainKey, generalizationSubKey)
	if err != nil {
		return model_use_case.Generalization{}, errors.WithStack(err)
	}

	generalization, err = model_use_case.NewGeneralization(generalizationKey, parsedFile.Title, parsedFile.Markdown, isComplete, isStatic, parsedFile.UmlComment)
	if err != nil {
		return model_use_case.Generalization{}, err
	}
	return generalization, nil
}

func generateUseCaseGeneralizationContent(generalization model_use_case.Generalization) string {
	yamlStr := ""
	if generalization.IsComplete != true {
		yamlStr += "is_complete: " + strconv.FormatBool(generalization.IsComplete) + "\n"
	}
	if generalization.IsStatic != true {
		yamlStr += "is_static: " + strconv.FormatBool(generalization.IsStatic) + "\n"
	}
	return generateFileContent(generalization.Details, generalization.UmlComment, yamlStr)
}
