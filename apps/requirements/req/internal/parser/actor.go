package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseActor(key, filename, contents string) (actor requirements.Actor, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.Actor{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return requirements.Actor{}, errors.WithStack(err)
	}

	userType := ""
	userTypeAny, found := yamlData["type"]
	if found {
		userType = userTypeAny.(string)
	}

	actor, err = requirements.NewActor(key, parsedFile.Title, parsedFile.Markdown, userType, parsedFile.UmlComment)
	if err != nil {
		return requirements.Actor{}, err
	}
	return actor, nil
}

func generateActorContent(actor requirements.Actor) string {
	yaml := "type: " + actor.Type
	return generateFileContent(actor.Details, actor.UmlComment, yaml)
}
