package requirements

const (
	StakeholderNotLinked = "stakeholder-not-linked" // A stakeholder must link somewehere.

	ActorNotInUseCase = "actor-not-in-use-case" // Actors must have use cases.

	UseCaseNoBullets                                 = "use-case-no-bullets"              // A use case without bulleted lists.
	UseCaseNoActor                                   = "use-case-no-actor"                // A use case that has no actor tied to it.
	UseCaseStepMissingFunctionalRequirementOrUseCase = "step-missing-functional-use-case" // A use case step that has no functional requirement tied to it.

	FunctionalNotInUseCase = "functional-not-in-use-case" // Functional requirements must be in at least one use case.

	UnansweredQuestion     = "unanswered-question"      // There is an unanswered question in this requirement.
	UnknownRequirementLink = "unknown-requirement-link" // A well-formed requirement link is not going to a known requirement.
)

var _IncompleteWhys = map[string]string{
	StakeholderNotLinked: "Stakeholder does not link anywhere.",

	ActorNotInUseCase: "Actor not in any use cases.",

	UseCaseNoBullets: "Use case missing bullets.",
	UseCaseNoActor:   "Use case missing actors.",
	UseCaseStepMissingFunctionalRequirementOrUseCase: "Use case step missing a functional requirement, or sub-usecase link:",

	FunctionalNotInUseCase: "Functional requirements not in any use cases.",

	UnansweredQuestion:     "Has unanswered question:",
	UnknownRequirementLink: "Has link to unknown requirement:",
}

type Incomplete struct {
	Header  Header // What is incomplete.
	Why     string // Why is this incomplete.
	Details string // The details on what is incomplete.
}

func newIncomplete(header Header, why, details string) (incomplete Incomplete) {
	return Incomplete{
		Header:  header,
		Why:     why,
		Details: details,
	}
}

func (i *Incomplete) String() (value string) {
	value = _IncompleteWhys[i.Why]
	if i.Details != "" {
		value += " " + i.Details
	}
	return value
}
