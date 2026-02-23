package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

// useCaseActorInOut is an actor who acts in a user story.
type useCaseActorInOut struct {
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the useCaseActorInOut to model_use_case.Actor.
func (u useCaseActorInOut) ToRequirements() model_use_case.Actor {
	return model_use_case.Actor{
		UmlComment: u.UmlComment,
	}
}

// FromRequirements creates a useCaseActorInOut from model_use_case.Actor.
func FromRequirementsUseCaseActor(u model_use_case.Actor) useCaseActorInOut {
	return useCaseActorInOut{
		UmlComment: u.UmlComment,
	}
}
