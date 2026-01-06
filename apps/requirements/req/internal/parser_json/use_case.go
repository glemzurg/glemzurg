package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// useCaseInOut is a user story for the system.
type useCaseInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Level      string `json:"level"`   // How high cocept or tightly focused the user case is.
	ReadOnly   bool   `json:"read_only"`
	UmlComment string `json:"uml_comment"`
	// Nested.
	Actors    map[string]useCaseActorInOut `json:"actors"`
	Scenarios []scenarioInOut              `json:"scenarios"`
}

// ToRequirements converts the useCaseInOut to requirements.UseCase.
func (u useCaseInOut) ToRequirements() requirements.UseCase {

	useCase := requirements.UseCase{
		Key:        u.Key,
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
		Actors:     nil,
		Scenarios:  nil,
	}

	for k, v := range u.Actors {
		if useCase.Actors == nil {
			useCase.Actors = make(map[string]requirements.UseCaseActor)
		}
		useCase.Actors[k] = v.ToRequirements()
	}
	for _, s := range u.Scenarios {
		useCase.Scenarios = append(useCase.Scenarios, s.ToRequirements())
	}
	return useCase
}

// FromRequirementsUseCase creates a useCaseInOut from requirements.UseCase.
func FromRequirementsUseCase(u requirements.UseCase) useCaseInOut {

	useCase := useCaseInOut{
		Key:        u.Key,
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
		Actors:     nil,
		Scenarios:  nil,
	}

	for k, v := range u.Actors {
		if useCase.Actors == nil {
			useCase.Actors = make(map[string]useCaseActorInOut)
		}
		useCase.Actors[k] = FromRequirementsUseCaseActor(v)
	}
	for _, s := range u.Scenarios {
		useCase.Scenarios = append(useCase.Scenarios, FromRequirementsScenario(s))
	}
	return useCase
}
