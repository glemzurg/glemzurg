package parser

import (
	"strconv"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseGeneralization(key, filename, contents string) (generalization model_class.Generalization, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_class.Generalization{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_class.Generalization{}, errors.WithStack(err)
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

	generalization, err = model_class.NewGeneralization(key, parsedFile.Title, parsedFile.Markdown, isComplete, isStatic, parsedFile.UmlComment)
	if err != nil {
		return model_class.Generalization{}, err
	}
	return generalization, nil
}

func generateGeneralizationContent(generalization model_class.Generalization) string {
	yamlStr := ""
	if generalization.IsComplete != true {
		yamlStr += "is_complete: " + strconv.FormatBool(generalization.IsComplete) + "\n"
	}
	if generalization.IsStatic != true {
		yamlStr += "is_static: " + strconv.FormatBool(generalization.IsStatic) + "\n"
	}
	return generateFileContent(generalization.Details, generalization.UmlComment, yamlStr)
}
