package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseActor(actorSubKey, filename, contents string) (actor model_actor.Actor, err error) {

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

	// Parse optional superclass/subclass generalization keys.
	var superclassOfKey *identity.Key
	if s, ok := yamlData["superclass_of_key"]; ok {
		k, err := identity.NewActorGeneralizationKey(s.(string))
		if err != nil {
			return model_actor.Actor{}, errors.WithStack(err)
		}
		superclassOfKey = &k
	}
	var subclassOfKey *identity.Key
	if s, ok := yamlData["subclass_of_key"]; ok {
		k, err := identity.NewActorGeneralizationKey(s.(string))
		if err != nil {
			return model_actor.Actor{}, errors.WithStack(err)
		}
		subclassOfKey = &k
	}

	// Construct the identity key for this actor.
	actorKey, err := identity.NewActorKey(actorSubKey)
	if err != nil {
		return model_actor.Actor{}, errors.WithStack(err)
	}

	actor, err = model_actor.NewActor(actorKey, parsedFile.Title, parsedFile.Markdown, userType, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
	if err != nil {
		return model_actor.Actor{}, err
	}
	return actor, nil
}

func generateActorContent(actor model_actor.Actor) string {
	yamlStr := "type: " + actor.Type + "\n"
	if actor.SuperclassOfKey != nil {
		yamlStr += "superclass_of_key: " + actor.SuperclassOfKey.SubKey + "\n"
	}
	if actor.SubclassOfKey != nil {
		yamlStr += "subclass_of_key: " + actor.SubclassOfKey.SubKey + "\n"
	}
	return generateFileContent(actor.Details, actor.UmlComment, yamlStr)
}
