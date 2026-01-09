package model_scenario

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
)

// Scenario is a documented scenario for a use case, such as a sequence diagram.
type Scenario struct {
	Key     string
	Name    string
	Details string // Markdown.
	Steps   Node   // The "abstract syntax tree" of the scenario.
	// Part of the data in a parsed file.
	Objects []Object
	// Steps represent the structured program steps of the scenario.
}

func NewScenario(key, name, details string) (scenario Scenario, err error) {

	scenario = Scenario{
		Key:     key,
		Name:    name,
		Details: details,
	}

	err = validation.ValidateStruct(&scenario,
		validation.Field(&scenario.Key, validation.Required),
		validation.Field(&scenario.Name, validation.Required),
	)
	if err != nil {
		return Scenario{}, errors.WithStack(err)
	}

	return scenario, nil
}

func (sc *Scenario) SetObjects(objects []Object) {
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ObjectNumber < objects[j].ObjectNumber
	})
	sc.Objects = objects
}

func CreateKeyScenarioLookup(
	byUseCase map[string][]Scenario,
	objectsByScenario map[string][]Object,
) (lookup map[string]Scenario) {

	lookup = map[string]Scenario{}
	for _, items := range byUseCase {
		for _, item := range items {
			item.SetObjects(objectsByScenario[item.Key])
			lookup[item.Key] = item
		}
	}
	return lookup
}

func PopulateScenarioStepReferences(
	scenarios map[string]Scenario,
	objects map[string]Object,
	attributes map[string]model_class.Attribute,
	events map[string]model_state.Event,
) (err error) {
	for key := range scenarios {
		scenario := scenarios[key]
		err = scenario.Steps.PopulateReferences(objects, events, attributes, scenarios)
		if err != nil {
			return err
		}
		scenarios[key] = scenario
	}
	return nil
}
