package parser

import (
	"github.com/glemzurg/futz/apps/req/requirements"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseUseCase(key, filename, contents string) (useCase requirements.UseCase, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.UseCase{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return requirements.UseCase{}, errors.WithStack(err)
	}

	level := "sea"
	levelAny, found := yamlData["level"]
	if found {
		level = levelAny.(string)
	}

	// If the title of the use case ends with "?" it is read-only.
	readOnly := strings.HasSuffix(parsedFile.Title, "?")

	useCase, err = requirements.NewUseCase(key, parsedFile.Title, parsedFile.Markdown, level, readOnly, parsedFile.UmlComment)
	if err != nil {
		return requirements.UseCase{}, err
	}

	// Parse actors.
	actorsAny, found := yamlData["actors"]
	if found {
		useCase.Actors = map[string]requirements.UseCaseActor{}
		actorsMap := actorsAny.(map[string]any)
		for actorKey, commentAny := range actorsMap {
			comment := ""
			if commentStr, ok := commentAny.(string); ok {
				comment = commentStr
			}
			useCaseActor, err := requirements.NewUseCaseActor(comment)
			if err != nil {
				return requirements.UseCase{}, err
			}
			useCase.Actors[actorKey] = useCaseActor
		}
	}

	// Parse scenarios.
	scenariosAny, found := yamlData["scenarios"]
	if found {
		useCase.Scenarios = map[string]requirements.Scenario{}
		scenariosMap := scenariosAny.(map[string]any)
		for scenarioKey, scenarioData := range scenariosMap {
			scenarioKey = key + "/scenario/" + strings.ToLower(scenarioKey)
			scenarioData := scenarioData.(map[string]any)

			name := ""
			details := ""

			nameAny, found := scenarioData["name"]
			if found {
				name = nameAny.(string)
			}

			detailsAny, found := scenarioData["details"]
			if found {
				details = detailsAny.(string)
			}

			scenario, err := requirements.NewScenario(scenarioKey, name, details)
			if err != nil {
				return requirements.UseCase{}, err
			}

			// Parse objects for this scenario.
			objectsAny, found := scenarioData["objects"]
			if found {
				objectsSlice := objectsAny.([]any)
				for i, objAny := range objectsSlice {
					scenarioObject, err := objectFromYamlData(scenarioKey, i, objAny)
					if err != nil {
						return requirements.UseCase{}, err
					}
					scenario.Objects = append(scenario.Objects, scenarioObject)
				}
			}

			// Parse stpes for this scenario.
			stepsAny, found := scenarioData["steps"]
			if found {
				stepsData := stepsAny.([]any)

				// Wrap in outer sequence node.
				nodeData := map[string]any{
					"statements": stepsData,
				}

				// Turn into yaml.
				nodeYaml, err := yaml.Marshal(nodeData)
				if err != nil {
					return requirements.UseCase{}, err
				}

				var node requirements.Node
				if err = node.FromYAML(string(nodeYaml)); err != nil {
					return requirements.UseCase{}, err
				}

				// Scope object keys to model-wide uniqueness.
				if err = node.ScopeObjects(scenarioKey); err != nil {
					return requirements.UseCase{}, err
				}

				scenario.Steps = node
			}

			// Add scenario to use case.
			useCase.Scenarios[scenarioKey] = scenario
		}
	}

	return useCase, nil
}

func objectFromYamlData(scenarioKey string, objectI int, objectAny any) (object requirements.ScenarioObject, err error) {
	objectNum := uint(objectI + 1)

	key := ""
	name := ""
	nameStyle := "unnamed"
	classKey := ""
	multi := false
	umlComment := ""

	objectData, ok := objectAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		keyAny, found := objectData["key"]
		if found {
			key = keyAny.(string)
		}
		key = scenarioKey + "/object/" + strings.ToLower(key)

		nameAny, found := objectData["name"]
		if found {
			switch v := nameAny.(type) {
			case string:
				name = v
			case int:
				name = strconv.Itoa(v)
			}
		}

		nameStyleAny, found := objectData["style"]
		if found {
			nameStyle = nameStyleAny.(string)
		}

		classKeyAny, found := objectData["class_key"]
		if found {
			classKey = classKeyAny.(string)
		}

		multiAny, found := objectData["multi"]
		if found {
			multi = multiAny.(bool)
		}

		umlCommentAny, found := objectData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}
	}

	object, err = requirements.NewScenarioObject(
		key,
		objectNum,
		name,
		nameStyle,
		classKey,
		multi,
		umlComment)
	if err != nil {
		return requirements.ScenarioObject{}, err
	}

	return object, nil
}
