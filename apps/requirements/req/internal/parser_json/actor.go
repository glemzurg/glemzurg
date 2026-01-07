package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"

// actorInOut is a external user of this system, either a person or another system.
type actorInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Type       string `json:"type"`    // "person" or "system"
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the actorInOut to model_actor.Actor.
func (a actorInOut) ToRequirements() model_actor.Actor {
	return model_actor.Actor{
		Key:        a.Key,
		Name:       a.Name,
		Details:    a.Details,
		Type:       a.Type,
		UmlComment: a.UmlComment,
		ClassKeys:  nil, // Not stored in JSON
	}
}

// FromRequirements creates a actorInOut from model_actor.Actor.
func FromRequirementsActor(a model_actor.Actor) actorInOut {
	return actorInOut{
		Key:        a.Key,
		Name:       a.Name,
		Details:    a.Details,
		Type:       a.Type,
		UmlComment: a.UmlComment,
	}
}
