package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	_USE_CASE_LEVEL_SKY = "sky" // A high-level organizational user story.
	_USE_CASE_LEVEL_SEA = "sea" // A straight forward user story.
	_USE_CASE_LEVEL_MUD = "mud" // A small, reusable part of another user story (e.g. a login flow).
)

// UseCase is a user story for the system.
type UseCase struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Level      string // How high cocept or tightly focused the user case is.
	ReadOnly   bool   // This is a user story that does not change the state of the system.
	UmlComment string
	// Part of the data in a parsed file.
	Actors    map[string]UseCaseActor
	Scenarios []Scenario
	// Helpful data.
	DomainKey string
}

func NewUseCase(key, name, details, level string, readOnly bool, umlComment string) (useCase UseCase, err error) {

	useCase = UseCase{
		Key:        key,
		Name:       name,
		Details:    details,
		Level:      level,
		ReadOnly:   readOnly,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&useCase,
		validation.Field(&useCase.Key, validation.Required),
		validation.Field(&useCase.Name, validation.Required),
		validation.Field(&useCase.Level, validation.Required, validation.In(_USE_CASE_LEVEL_SKY, _USE_CASE_LEVEL_SEA, _USE_CASE_LEVEL_MUD)),
	)
	if err != nil {
		return UseCase{}, errors.WithStack(err)
	}

	return useCase, nil
}

func (uc *UseCase) SetDomainKey(domainKey string) {
	uc.DomainKey = domainKey
}

func (uc *UseCase) SetActors(actors map[string]UseCaseActor) {
	uc.Actors = actors
}

func (uc *UseCase) SetScenarios(scenarios []Scenario) {
	uc.Scenarios = scenarios
}

func createKeyUseCaseLookup(
	byCategory map[string][]UseCase,
	useCaseActors map[string]map[string]UseCaseActor,
	scenarios map[string][]Scenario,
) (lookup map[string]UseCase) {

	lookup = map[string]UseCase{}
	for domainKey, items := range byCategory {
		for _, item := range items {

			item.SetDomainKey(domainKey)
			item.SetActors(useCaseActors[item.Key])
			item.SetScenarios(scenarios[item.Key])

			lookup[item.Key] = item
		}
	}
	return lookup
}
