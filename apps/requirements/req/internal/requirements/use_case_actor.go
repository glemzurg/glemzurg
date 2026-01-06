package requirements

// UseCaseActor is an actor who acts in a user story.
type UseCaseActor struct {
	UmlComment string
}

func NewUseCaseActor(umlComment string) (useCaseActor UseCaseActor, err error) {

	useCaseActor = UseCaseActor{
		UmlComment: umlComment,
	}

	return useCaseActor, nil
}
