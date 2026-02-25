package parser

import (
	"strconv"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseActorGeneralization(generalizationSubKey, filename, contents string) (generalization model_actor.Generalization, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_actor.Generalization{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_actor.Generalization{}, errors.WithStack(err)
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
	generalizationKey, err := identity.NewActorGeneralizationKey(generalizationSubKey)
	if err != nil {
		return model_actor.Generalization{}, errors.WithStack(err)
	}

	generalization, err = model_actor.NewGeneralization(generalizationKey, parsedFile.Title, stripMarkdownTitle(parsedFile.Markdown), isComplete, isStatic, parsedFile.UmlComment)
	if err != nil {
		return model_actor.Generalization{}, err
	}
	return generalization, nil
}

func generateActorGeneralizationContent(generalization model_actor.Generalization) string {
	yamlStr := ""
	if generalization.IsComplete != true {
		yamlStr += "is_complete: " + strconv.FormatBool(generalization.IsComplete) + "\n"
	}
	if generalization.IsStatic != true {
		yamlStr += "is_static: " + strconv.FormatBool(generalization.IsStatic) + "\n"
	}
	return generateFileContent(prependMarkdownSubtitle(generalization.Name, generalization.Details), generalization.UmlComment, yamlStr)
}
