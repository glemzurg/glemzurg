package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// useCaseActorInOut is an actor who acts in a user story.
type useCaseActorInOut struct {
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the useCaseActorInOut to requirements.UseCaseActor.
func (u useCaseActorInOut) ToRequirements() requirements.UseCaseActor {
	return requirements.UseCaseActor{
		UmlComment: u.UmlComment,
	}
}

// FromRequirements creates a useCaseActorInOut from requirements.UseCaseActor.
func FromRequirementsUseCaseActor(u requirements.UseCaseActor) useCaseActorInOut {
	return useCaseActorInOut{
		UmlComment: u.UmlComment,
	}
}
