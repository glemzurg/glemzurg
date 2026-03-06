package model_use_case

// Actor is an actor who acts in a user story.
type Actor struct {
	UmlComment string
}

func NewActor(umlComment string) (actor Actor, err error) {

	actor = Actor{
		UmlComment: umlComment,
	}

	if err = actor.Validate(); err != nil {
		return Actor{}, err
	}

	return actor, nil
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
