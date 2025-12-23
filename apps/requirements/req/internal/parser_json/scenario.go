package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// scenarioInOut is a documented scenario for a use case, such as a sequence diagram.
type scenarioInOut struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Details string    `json:"details"` // Markdown.
	Steps   nodeInOut `json:"steps"`   // The "abstract syntax tree" of the scenario.
	// Nested.
	Objects []scenarioObjectInOut `json:"objects"`
}

// ToRequirements converts the scenarioInOut to requirements.Scenario.
func (s scenarioInOut) ToRequirements() requirements.Scenario {

	scenario := requirements.Scenario{
		Key:     s.Key,
		Name:    s.Name,
		Details: s.Details,
		Steps:   s.Steps.ToRequirements(),
		Objects: nil,
	}

	for _, o := range s.Objects {
		scenario.Objects = append(scenario.Objects, o.ToRequirements())
	}

	return scenario
}

// FromRequirementsScenario creates a scenarioInOut from requirements.Scenario.
func FromRequirementsScenario(s requirements.Scenario) scenarioInOut {

	scenario := scenarioInOut{
		Key:     s.Key,
		Name:    s.Name,
		Details: s.Details,
		Steps:   FromRequirementsNode(s.Steps),
		Objects: nil,
	}

	for _, o := range s.Objects {
		scenario.Objects = append(scenario.Objects, FromRequirementsScenarioObject(o))
	}
	return scenario
}
