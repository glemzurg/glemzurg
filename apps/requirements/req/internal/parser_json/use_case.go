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
	actors := make(map[string]requirements.UseCaseActor)
	for k, v := range u.Actors {
		actors[k] = v.ToRequirements()
	}
	scenarios := make([]requirements.Scenario, len(u.Scenarios))
	for i, s := range u.Scenarios {
		scenarios[i] = s.ToRequirements()
	}
	return requirements.UseCase{
		Key:        u.Key,
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
		Actors:     actors,
		Scenarios:  scenarios,
	}
}

// FromRequirementsUseCase creates a useCaseInOut from requirements.UseCase.
func FromRequirementsUseCase(u requirements.UseCase) useCaseInOut {
	actors := make(map[string]useCaseActorInOut)
	for k, v := range u.Actors {
		actors[k] = FromRequirementsUseCaseActor(v)
	}
	scenarios := make([]scenarioInOut, len(u.Scenarios))
	for i, s := range u.Scenarios {
		scenarios[i] = FromRequirementsScenario(s)
	}
	return useCaseInOut{
		Key:        u.Key,
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
		Actors:     actors,
		Scenarios:  scenarios,
	}
}
