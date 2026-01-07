package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

// useCaseActorInOut is an actor who acts in a user story.
type useCaseActorInOut struct {
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the useCaseActorInOut to model_use_case.UseCaseActor.
func (u useCaseActorInOut) ToRequirements() model_use_case.UseCaseActor {
	return model_use_case.UseCaseActor{
		UmlComment: u.UmlComment,
	}
}

// FromRequirements creates a useCaseActorInOut from model_use_case.UseCaseActor.
func FromRequirementsUseCaseActor(u model_use_case.UseCaseActor) useCaseActorInOut {
	return useCaseActorInOut{
		UmlComment: u.UmlComment,
	}
}
