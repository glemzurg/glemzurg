package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseActor(key, filename, contents string) (actor model_actor.Actor, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_actor.Actor{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_actor.Actor{}, errors.WithStack(err)
	}

	userType := ""
	userTypeAny, found := yamlData["type"]
	if found {
		userType = userTypeAny.(string)
	}

	actor, err = model_actor.NewActor(key, parsedFile.Title, parsedFile.Markdown, userType, parsedFile.UmlComment)
	if err != nil {
		return model_actor.Actor{}, err
	}
	return actor, nil
}

func generateActorContent(actor model_actor.Actor) string {
	yaml := "type: " + actor.Type
	return generateFileContent(actor.Details, actor.UmlComment, yaml)
}
