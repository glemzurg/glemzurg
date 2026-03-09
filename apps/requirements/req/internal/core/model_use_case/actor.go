package model_use_case

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
func (a *Actor) Validate() error {
	return nil
}

// ValidateWithParent validates the Actor.
// Actor does not have a key, so it does not validate parent relationships.
func (a *Actor) ValidateWithParent() error {
	return a.Validate()
}
