package parser

import (
	"glemzurg/reqmodel/requirements"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseGeneralization(key, filename, contents string) (generalization requirements.Generalization, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.Generalization{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return requirements.Generalization{}, errors.WithStack(err)
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

	generalization, err = requirements.NewGeneralization(key, parsedFile.Title, parsedFile.Markdown, isComplete, isStatic, parsedFile.UmlComment)
	if err != nil {
		return requirements.Generalization{}, err
	}
	return generalization, nil
}
