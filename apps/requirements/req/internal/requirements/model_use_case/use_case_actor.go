package model_use_case

// Actor is an actor who acts in a user story.
type Actor struct {
	UmlComment string
}

func NewActor(umlComment string) (actor Actor, err error) {

	actor = Actor{
		UmlComment: umlComment,
	}

	return actor, nil
}
