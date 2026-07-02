package generate

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"

	"github.com/pkg/errors"
)

// useCaseScenarioSection pairs a scenario with its rendered diagram markdown.
type useCaseScenarioSection struct {
	Scenario     model_scenario.Scenario
	DiagramEmbed string
}

func generateUseCaseMdContents(reqs *req_flat.Requirements, writer ContentWriter, useCase model_use_case.UseCase) (contents string, err error) {
	scenarioSections, err := buildUseCaseScenarioSections(reqs, writer, useCase)
	if err != nil {
		return "", err
	}

	contents, err = generateFromTemplate(_useCaseMdTemplate, struct {
		Reqs             *req_flat.Requirements
		UseCase          model_use_case.UseCase
		ScenarioSections []useCaseScenarioSection
	}{
		Reqs:             reqs,
		UseCase:          useCase,
		ScenarioSections: scenarioSections,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

func buildUseCaseScenarioSections(
	reqs *req_flat.Requirements,
	writer ContentWriter,
	useCase model_use_case.UseCase,
) ([]useCaseScenarioSection, error) {
	if len(useCase.Scenarios) == 0 {
		return nil, nil
	}

	useCaseKey := useCase.Key.String()
	scenarios := make([]model_scenario.Scenario, 0, len(useCase.Scenarios))
	for _, scenario := range useCase.Scenarios {
		scenarios = append(scenarios, scenario)
	}
	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].Key.String() < scenarios[j].Key.String()
	})

	var sections []useCaseScenarioSection
	for _, scenario := range scenarios {
		body, err := generateScenarioMermaidContents(reqs, scenario)
		if err != nil {
			return nil, err
		}
		source := "sequenceDiagram\n" + body
		suffix := "scenario-" + strings.ReplaceAll(scenario.Key.String(), "/", ".")
		svgFilename := convertKeyToFilename("use_case", useCaseKey, suffix, ".svg")
		embed, err := embedDiagram(writer, source, svgFilename, scenario.Name)
		if err != nil {
			return nil, err
		}
		sections = append(sections, useCaseScenarioSection{
			Scenario:     scenario,
			DiagramEmbed: embed,
		})
	}

	return sections, nil
}

// generateUseCasesMermaidContents generates Mermaid use case diagram markup.
func generateUseCasesMermaidContents(reqs *req_flat.Requirements, domain model_domain.Domain, useCases []model_use_case.UseCase, actors []model_actor.Actor) (contents string, err error) {
	contents, err = generateFromTemplate(_useCasesMermaidTemplate, struct {
		Reqs     *req_flat.Requirements
		Domain   model_domain.Domain
		UseCases []model_use_case.UseCase
		Actors   []model_actor.Actor
	}{
		Reqs:     reqs,
		Domain:   domain,
		UseCases: useCases,
		Actors:   actors,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
