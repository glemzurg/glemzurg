package model_actor

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
)

const (
	_USER_TYPE_PERSON = "person"
	_USER_TYPE_SYSTEM = "system"
)

// An actor is a external user of this sytem, either a person or another system.
type Actor struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	Type       string // "person" or "system"
	UmlComment string
	// Helpful data.
	ClassKeys []identity.Key // Classes that implement this actor.
}

func NewActor(key identity.Key, name, details, userType, umlComment string) (actor Actor, err error) {

	actor = Actor{
		Key:        key,
		Name:       name,
		Details:    details,
		Type:       userType,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&actor,
		validation.Field(&actor.Key, validation.Required),
		validation.Field(&actor.Name, validation.Required),
		validation.Field(&actor.Type, validation.Required, validation.In(_USER_TYPE_PERSON, _USER_TYPE_SYSTEM)),
	)
	if err != nil {
		return Actor{}, errors.WithStack(err)
	}

	return actor, nil
}

func CreateKeyActorLookup(domainClasses map[string][]model_class.Class, items []Actor) (lookup map[string]Actor) {

	// All the classes that are actors.
	actorClassKeyLookup := map[string][]string{}
	for _, classes := range domainClasses {
		for _, class := range classes {
			if class.ActorKey != "" {
				actorClasses := actorClassKeyLookup[class.ActorKey]
				actorClasses = append(actorClasses, class.Key)
				actorClassKeyLookup[class.ActorKey] = actorClasses
			}
		}
	}

	lookup = map[string]Actor{}
	for _, item := range items {

		item.ClassKeys = actorClassKeyLookup[item.Key]
		sort.Strings(item.ClassKeys)

		lookup[item.Key] = item
	}
	return lookup
}
