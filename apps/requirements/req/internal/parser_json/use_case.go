package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

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

// ToRequirements converts the useCaseInOut to model_use_case.UseCase.
func (u useCaseInOut) ToRequirements() (model_use_case.UseCase, error) {
	key, err := identity.ParseKey(u.Key)
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	useCase := model_use_case.UseCase{
		Key:        key,
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
	}

	for k, v := range u.Actors {
		if useCase.Actors == nil {
			useCase.Actors = make(map[identity.Key]model_use_case.Actor)
		}
		actorKey, err := identity.ParseKey(k)
		if err != nil {
			return model_use_case.UseCase{}, err
		}
		useCase.Actors[actorKey] = v.ToRequirements()
	}
	for _, s := range u.Scenarios {
		scenario, err := s.ToRequirements()
		if err != nil {
			return model_use_case.UseCase{}, err
		}
		if useCase.Scenarios == nil {
			useCase.Scenarios = make(map[identity.Key]model_scenario.Scenario)
		}
		useCase.Scenarios[scenario.Key] = scenario
	}
	return useCase, nil
}

// FromRequirementsUseCase creates a useCaseInOut from model_use_case.UseCase.
func FromRequirementsUseCase(u model_use_case.UseCase) useCaseInOut {

	useCase := useCaseInOut{
		Key:        u.Key.String(),
		Name:       u.Name,
		Details:    u.Details,
		Level:      u.Level,
		ReadOnly:   u.ReadOnly,
		UmlComment: u.UmlComment,
	}

	for k, v := range u.Actors {
		if useCase.Actors == nil {
			useCase.Actors = make(map[string]useCaseActorInOut)
		}
		useCase.Actors[k.String()] = FromRequirementsUseCaseActor(v)
	}
	for _, s := range u.Scenarios {
		useCase.Scenarios = append(useCase.Scenarios, FromRequirementsScenario(s))
	}
	return useCase
}
