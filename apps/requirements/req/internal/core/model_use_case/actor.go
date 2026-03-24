package model_use_case

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// Actor is an actor who acts in a user story.
type Actor struct {
	UmlComment string
}

func NewActor(umlComment string) Actor {
	return Actor{
		UmlComment: umlComment,
	}
}

// Validate validates the Actor struct.
func (a *Actor) Validate(_ *coreerr.ValidationContext) error {
	return nil
}

// ValidateWithParent validates the Actor.
// Actor does not have a key, so it does not validate parent relationships.
func (a *Actor) ValidateWithParent(ctx *coreerr.ValidationContext) error {
	return a.Validate(ctx)
}
