package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/actor"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseActor(key, filename, contents string) (parsedActor actor.Actor, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return actor.Actor{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return actor.Actor{}, errors.WithStack(err)
	}

	userType := ""
	userTypeAny, found := yamlData["type"]
	if found {
		userType = userTypeAny.(string)
	}

	parsedActor, err = actor.NewActor(key, parsedFile.Title, parsedFile.Markdown, userType, parsedFile.UmlComment)
	if err != nil {
		return actor.Actor{}, err
	}
	return parsedActor, nil
}

func generateActorContent(actor actor.Actor) string {
	yaml := "type: " + actor.Type
	return generateFileContent(actor.Details, actor.UmlComment, yaml)
}
