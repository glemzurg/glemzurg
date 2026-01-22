package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
)

// scenarioInOut is a documented scenario for a use case, such as a sequence diagram.
type scenarioInOut struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Details string    `json:"details"` // Markdown.
	Steps   nodeInOut `json:"steps"`   // The "abstract syntax tree" of the scenario.
	// Nested.
	Objects []objectInOut `json:"objects"`
}

// ToRequirements converts the scenarioInOut to model_scenario.Scenario.
func (s scenarioInOut) ToRequirements() (model_scenario.Scenario, error) {
	key, err := identity.ParseKey(s.Key)
	if err != nil {
		return model_scenario.Scenario{}, err
	}

	var stepsPtr *model_scenario.Node
	if !s.Steps.isEmpty() {
		steps, err := s.Steps.ToRequirements()
		if err != nil {
			return model_scenario.Scenario{}, err
		}
		stepsPtr = &steps
	}
	scenario := model_scenario.Scenario{
		Key:     key,
		Name:    s.Name,
		Details: s.Details,
		Steps:   stepsPtr,
	}

	for _, o := range s.Objects {
		obj, err := o.ToRequirements()
		if err != nil {
			return model_scenario.Scenario{}, err
		}
		if scenario.Objects == nil {
			scenario.Objects = make(map[identity.Key]model_scenario.Object)
		}
		scenario.Objects[obj.Key] = obj
	}

	return scenario, nil
}

// FromRequirementsScenario creates a scenarioInOut from model_scenario.Scenario.
func FromRequirementsScenario(s model_scenario.Scenario) scenarioInOut {
	var steps nodeInOut
	if s.Steps != nil {
		steps = FromRequirementsNode(*s.Steps)
	}

	scenario := scenarioInOut{
		Key:     s.Key.String(),
		Name:    s.Name,
		Details: s.Details,
		Steps:   steps,
	}

	for _, o := range s.Objects {
		scenario.Objects = append(scenario.Objects, FromRequirementsObject(o))
	}
	return scenario
}
