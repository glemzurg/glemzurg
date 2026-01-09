package parser

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseUseCase(key, filename, contents string) (useCase model_use_case.UseCase, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_use_case.UseCase{}, errors.WithStack(err)
	}

	level := "sea"
	levelAny, found := yamlData["level"]
	if found {
		level = levelAny.(string)
	}

	// If the title of the use case ends with "?" it is read-only.
	readOnly := strings.HasSuffix(parsedFile.Title, "?")

	useCase, err = model_use_case.NewUseCase(key, parsedFile.Title, parsedFile.Markdown, level, readOnly, parsedFile.UmlComment)
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	// Parse actors.
	actorsAny, found := yamlData["actors"]
	if found {
		useCase.Actors = map[string]model_use_case.Actor{}
		actorsMap := actorsAny.(map[string]any)
		for actorKey, commentAny := range actorsMap {
			comment := ""
			if commentStr, ok := commentAny.(string); ok {
				comment = commentStr
			}
			actor, err := model_use_case.NewActor(comment)
			if err != nil {
				return model_use_case.UseCase{}, err
			}
			useCase.Actors[actorKey] = actor
		}
	}

	// Parse scenarios.
	scenariosAny, found := yamlData["scenarios"]
	if found {
		useCase.Scenarios = []model_scenario.Scenario{}
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

			scenario, err := model_scenario.NewScenario(scenarioKey, name, details)
			if err != nil {
				return model_use_case.UseCase{}, err
			}

			// Parse objects for this scenario.
			objectsAny, found := scenarioData["objects"]
			if found {
				objectsSlice := objectsAny.([]any)
				for i, objAny := range objectsSlice {
					object, err := objectFromYamlData(scenarioKey, i, objAny)
					if err != nil {
						return model_use_case.UseCase{}, err
					}
					scenario.Objects = append(scenario.Objects, object)
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
					return model_use_case.UseCase{}, err
				}

				var node model_scenario.Node
				if err = node.FromYAML(string(nodeYaml)); err != nil {
					return model_use_case.UseCase{}, err
				}

				// Scope object keys to model-wide uniqueness.
				if err = node.ScopeObjects(scenarioKey); err != nil {
					return model_use_case.UseCase{}, err
				}

				scenario.Steps = node
			}

			// Add scenario to use case.
			useCase.Scenarios = append(useCase.Scenarios, scenario)
		}
	}

	// Sort the scenarios.
	sort.Slice(useCase.Scenarios, func(i, j int) bool {
		return useCase.Scenarios[i].Key < useCase.Scenarios[j].Key
	})

	return useCase, nil
}

func objectFromYamlData(scenarioKey string, objectI int, objectAny any) (object model_scenario.Object, err error) {
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

	object, err = model_scenario.NewObject(
		key,
		objectNum,
		name,
		nameStyle,
		classKey,
		multi,
		umlComment)
	if err != nil {
		return model_scenario.Object{}, err
	}

	return object, nil
}

func generateUseCaseContent(useCase model_use_case.UseCase) string {
	yaml := ""
	if useCase.Level != "sea" {
		yaml += "level: " + useCase.Level + "\n"
	}

	if len(useCase.Actors) > 0 {
		actors := make(map[string]string)
		for actorKey, actor := range useCase.Actors {
			if actor.UmlComment != "" {
				actors[actorKey] = actor.UmlComment
			}
		}
		if len(actors) > 0 {
			yaml += "\nactors:\n"
			keys := make([]string, 0, len(actors))
			for k := range actors {
				keys = append(keys, k)
			}
			sort.Sort(sort.Reverse(sort.StringSlice(keys)))
			for _, k := range keys {
				yaml += "    " + k + ": " + actors[k] + "\n"
			}
		}
	}

	if len(useCase.Scenarios) > 0 {
		yaml += "\nscenarios:\n"
		for _, scenario := range useCase.Scenarios {
			name := strings.Split(scenario.Key, "/scenario/")[1]
			yaml += "\n    " + name + ":\n"
			yaml += "        name: " + scenario.Name + "\n"
			if scenario.Details != "" {
				yaml += "        details: " + scenario.Details + "\n"
			}
			if len(scenario.Objects) > 0 {
				yaml += "        objects:\n"
				for _, obj := range scenario.Objects {
					objName := strings.Split(obj.Key, "/object/")[1]
					yaml += "            - key: " + objName + "\n"
					if obj.Name != "" {
						yaml += "              name: " + obj.Name + "\n"
					}
					if obj.NameStyle != "" && obj.NameStyle != "unnamed" {
						yaml += "              style: " + obj.NameStyle + "\n"
					}
					if obj.ClassKey != "" {
						yaml += "              class_key: " + obj.ClassKey + "\n"
					}
					if obj.Multi {
						yaml += "              multi: true\n"
					}
					if obj.UmlComment != "" {
						yaml += "              uml_comment: " + obj.UmlComment + "\n"
					}
				}
			}
			if len(scenario.Steps.Statements) > 0 {
				yaml += "        steps:\n"
				yaml += generateSteps(scenario.Steps.Statements, "            ")
			}
		}
	}

	yamlStr := strings.TrimSpace(yaml)
	if yamlStr == "" {
		yamlStr = "\n"
	}
	content := useCase.Details + "\n\n◆\n\n" + useCase.UmlComment + "\n\n◇"
	if yamlStr != "" {
		content += "\n\n" + yamlStr
	}
	return strings.TrimSpace(content)
}

func generateSteps(nodes []model_scenario.Node, indent string) string {
	s := ""
	for _, node := range nodes {
		s += generateNode(node, indent)
	}
	return s
}

func generateNode(node model_scenario.Node, indent string) string {
	s := indent + "- "
	if node.Loop != "" {
		s += "loop: " + node.Loop + "\n"
		subIndent := indent + "  "
		if len(node.Statements) > 0 {
			s += subIndent + "statements:\n"
			s += generateSteps(node.Statements, subIndent+"  ")
		}
	} else if len(node.Cases) > 0 {
		s += "cases:\n"
		subIndent := indent + "    "
		for _, c := range node.Cases {
			s += subIndent + "- condition: " + c.Condition + "\n"
			if len(c.Statements) > 0 {
				s += subIndent + "  statements:\n"
				s += generateSteps(c.Statements, subIndent+"    ")
			}
		}
	} else {
		// Leaf node
		first := true
		subIndent := indent + "  "
		addField := func(key, value string) {
			if first {
				s += key + ": " + value + "\n"
				first = false
			} else {
				s += subIndent + key + ": " + value + "\n"
			}
		}
		if node.Description != "" {
			addField("description", node.Description)
		}
		if node.FromObjectKey != "" {
			addField("from_object_key", strings.Split(node.FromObjectKey, "/object/")[1])
		}
		if node.ToObjectKey != "" {
			addField("to_object_key", strings.Split(node.ToObjectKey, "/object/")[1])
		}
		if node.EventKey != "" {
			addField("event_key", node.EventKey)
		}
		if node.ScenarioKey != "" {
			addField("scenario_key", node.ScenarioKey)
		}
		if node.IsDelete {
			addField("is_delete", "true        ")
		}
	}
	return s
}
