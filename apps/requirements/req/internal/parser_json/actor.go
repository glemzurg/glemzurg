package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
)

// actorInOut is a external user of this system, either a person or another system.
type actorInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Type       string `json:"type"`    // "person" or "system"
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the actorInOut to model_actor.Actor.
func (a actorInOut) ToRequirements() (model_actor.Actor, error) {
	key, err := identity.ParseKey(a.Key)
	if err != nil {
		return model_actor.Actor{}, err
	}

	return model_actor.Actor{
		Key:        key,
		Name:       a.Name,
		Details:    a.Details,
		Type:       a.Type,
		UmlComment: a.UmlComment,
	}, nil
}

// FromRequirements creates a actorInOut from model_actor.Actor.
func FromRequirementsActor(a model_actor.Actor) actorInOut {
	return actorInOut{
		Key:        a.Key.String(),
		Name:       a.Name,
		Details:    a.Details,
		Type:       a.Type,
		UmlComment: a.UmlComment,
	}
}
