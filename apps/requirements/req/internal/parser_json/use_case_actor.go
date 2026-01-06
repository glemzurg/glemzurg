package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/use_case"

// useCaseActorInOut is an actor who acts in a user story.
type useCaseActorInOut struct {
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the useCaseActorInOut to use_case.UseCaseActor.
func (u useCaseActorInOut) ToRequirements() use_case.UseCaseActor {
	return use_case.UseCaseActor{
		UmlComment: u.UmlComment,
	}
}

// FromRequirements creates a useCaseActorInOut from use_case.UseCaseActor.
func FromRequirementsUseCaseActor(u use_case.UseCaseActor) useCaseActorInOut {
	return useCaseActorInOut{
		UmlComment: u.UmlComment,
	}
}
