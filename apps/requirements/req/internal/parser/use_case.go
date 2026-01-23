package parser

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Note: sort is still used in generateUseCaseContent for actor keys

func parseUseCase(subdomainKey identity.Key, useCaseSubKey, filename, contents string) (useCase model_use_case.UseCase, err error) {

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

	// Construct the identity key for this use case.
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, useCaseSubKey)
	if err != nil {
		return model_use_case.UseCase{}, errors.WithStack(err)
	}

	useCase, err = model_use_case.NewUseCase(useCaseKey, parsedFile.Title, parsedFile.Markdown, level, readOnly, parsedFile.UmlComment)
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	// Parse actors.
	actorsAny, found := yamlData["actors"]
	if found {
		useCase.Actors = map[identity.Key]model_use_case.Actor{}
		actorsMap := actorsAny.(map[string]any)
		for actorKeyStr, commentAny := range actorsMap {
			comment := ""
			if commentStr, ok := commentAny.(string); ok {
				comment = commentStr
			}
			actor, err := model_use_case.NewActor(comment)
			if err != nil {
				return model_use_case.UseCase{}, err
			}
			// Construct the actor key from the string.
			actorKey, err := identity.NewClassKey(subdomainKey, actorKeyStr)
			if err != nil {
				return model_use_case.UseCase{}, errors.WithStack(err)
			}
			useCase.Actors[actorKey] = actor
		}
	}

	// Parse scenarios.
	scenariosAny, found := yamlData["scenarios"]
	if found {
		useCase.Scenarios = make(map[identity.Key]model_scenario.Scenario)
		scenariosMap := scenariosAny.(map[string]any)
		for scenarioSubKey, scenarioData := range scenariosMap {
			// Construct the scenario key.
			scenarioKey, err := identity.NewScenarioKey(useCaseKey, strings.ToLower(scenarioSubKey))
			if err != nil {
				return model_use_case.UseCase{}, errors.WithStack(err)
			}
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
				scenario.Objects = make(map[identity.Key]model_scenario.Object)
				objectsSlice := objectsAny.([]any)
				for i, objAny := range objectsSlice {
					object, err := objectFromYamlData(scenarioKey, i, objAny)
					if err != nil {
						return model_use_case.UseCase{}, err
					}
					scenario.Objects[object.Key] = object
				}
			}

			// Parse steps for this scenario.
			stepsAny, found := scenarioData["steps"]
			if found {
				stepsData := stepsAny.([]any)

				// Wrap in outer sequence node.
				nodeData := map[string]any{
					"statements": stepsData,
				}

				// Scope object and attribute keys before parsing into Node objects.
				// This ensures Node objects are always well-formed with complete keys.
				if err = scopeObjectKeys(scenarioKey, subdomainKey, nodeData); err != nil {
					return model_use_case.UseCase{}, err
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

				scenario.Steps = &node
			}

			// Add scenario to use case.
			useCase.Scenarios[scenario.Key] = scenario
		}
	}

	return useCase, nil
}

func objectFromYamlData(scenarioKey identity.Key, objectI int, objectAny any) (object model_scenario.Object, err error) {
	objectNum := uint(objectI + 1)

	objectSubKey := ""
	name := ""
	nameStyle := "unnamed"
	classKeyStr := ""
	multi := false
	umlComment := ""

	objectData, ok := objectAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		keyAny, found := objectData["key"]
		if found {
			objectSubKey = keyAny.(string)
		}
		objectSubKey = strings.ToLower(objectSubKey)

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
			classKeyStr = classKeyAny.(string)
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

	// Construct the object key.
	objectKey, err := identity.NewScenarioObjectKey(scenarioKey, objectSubKey)
	if err != nil {
		return model_scenario.Object{}, errors.WithStack(err)
	}

	// Parse the class key from the string.
	// If it's a simple key (no slashes), construct it as a class key in the same subdomain.
	// Otherwise, parse it as a full key.
	var classKey identity.Key
	if !strings.Contains(classKeyStr, "/") {
		// Simple key - construct full class key using the scenario's subdomain.
		// Extract subdomain key from scenario key (scenario key has format: domain/.../subdomain/.../usecase/.../scenario/...)
		scenarioKeyStr := scenarioKey.String()
		parts := strings.Split(scenarioKeyStr, "/usecase/")
		if len(parts) < 2 {
			return model_scenario.Object{}, errors.Errorf("invalid scenario key format: %s", scenarioKeyStr)
		}
		subdomainKeyStr := parts[0]
		subdomainKey, err := identity.ParseKey(subdomainKeyStr)
		if err != nil {
			return model_scenario.Object{}, errors.WithStack(err)
		}
		classKey, err = identity.NewClassKey(subdomainKey, classKeyStr)
		if err != nil {
			return model_scenario.Object{}, errors.WithStack(err)
		}
	} else {
		classKey, err = identity.ParseKey(classKeyStr)
		if err != nil {
			return model_scenario.Object{}, errors.WithStack(err)
		}
	}

	object, err = model_scenario.NewObject(
		objectKey,
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
				actors[actorKey.SubKey()] = actor.UmlComment
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
		// Sort scenarios by key for deterministic output.
		scenarios := make([]model_scenario.Scenario, 0, len(useCase.Scenarios))
		for _, scenario := range useCase.Scenarios {
			scenarios = append(scenarios, scenario)
		}
		sort.Slice(scenarios, func(i, j int) bool {
			return scenarios[i].Key.String() < scenarios[j].Key.String()
		})
		for _, scenario := range scenarios {
			name := scenario.Key.SubKey()
			yaml += "\n    " + name + ":\n"
			yaml += "        name: " + scenario.Name + "\n"
			if scenario.Details != "" {
				yaml += "        details: " + scenario.Details + "\n"
			}
			if len(scenario.Objects) > 0 {
				yaml += "        objects:\n"
				// Sort objects by ObjectNumber for deterministic output.
				objects := make([]model_scenario.Object, 0, len(scenario.Objects))
				for _, obj := range scenario.Objects {
					objects = append(objects, obj)
				}
				sort.Slice(objects, func(i, j int) bool {
					return objects[i].ObjectNumber < objects[j].ObjectNumber
				})
				for _, obj := range objects {
					objName := obj.Key.SubKey()
					yaml += "            - key: " + objName + "\n"
					if obj.Name != "" {
						yaml += "              name: " + obj.Name + "\n"
					}
					if obj.NameStyle != "" && obj.NameStyle != "unnamed" {
						yaml += "              style: " + obj.NameStyle + "\n"
					}
					if obj.ClassKey.String() != "" {
						// Output only the subkey for backwards compatibility with the md format.
						yaml += "              class_key: " + obj.ClassKey.SubKey() + "\n"
					}
					if obj.Multi {
						yaml += "              multi: true\n"
					}
					if obj.UmlComment != "" {
						yaml += "              uml_comment: " + obj.UmlComment + "\n"
					}
				}
			}
			if scenario.Steps != nil && len(scenario.Steps.Statements) > 0 {
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

// shortEventKey converts a full event key to its short form for output.
// Full: domain/test_domain/subdomain/test_subdomain/class/class_key/event/processlog
// Short: class_key/event/processlog
func shortEventKey(key *identity.Key) string {
	if key == nil {
		return ""
	}
	// Parse the key to extract class subkey and event subkey.
	// Event key: parentKey="domain/.../subdomain/.../class/class_key", keyType="event", subKey="processlog"
	// The parent of the event key is a class key.
	parentKey, err := identity.ParseKey(key.ParentKey())
	if err != nil {
		// Fallback to full string if parsing fails.
		return key.String()
	}
	// parentKey is the class key; its subKey is the class subkey.
	classSubKey := parentKey.SubKey()
	eventSubKey := key.SubKey()
	return classSubKey + "/event/" + eventSubKey
}

// shortScenarioKey converts a full scenario key to its short form for output.
// Full: domain/test_domain/subdomain/test_subdomain/usecase/use_case_key/scenario/scenario_b_key
// Short: use_case_key/scenario/scenario_b_key
func shortScenarioKey(key *identity.Key) string {
	if key == nil {
		return ""
	}
	// Parse the key to extract use case subkey and scenario subkey.
	// Scenario key: parentKey="domain/.../subdomain/.../usecase/use_case_key", keyType="scenario", subKey="scenario_b_key"
	// The parent of the scenario key is a use case key.
	parentKey, err := identity.ParseKey(key.ParentKey())
	if err != nil {
		// Fallback to full string if parsing fails.
		return key.String()
	}
	// parentKey is the use case key; its subKey is the use case subkey.
	useCaseSubKey := parentKey.SubKey()
	scenarioSubKey := key.SubKey()
	return useCaseSubKey + "/scenario/" + scenarioSubKey
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
		if node.FromObjectKey != nil {
			addField("from_object_key", node.FromObjectKey.SubKey())
		}
		if node.ToObjectKey != nil {
			addField("to_object_key", node.ToObjectKey.SubKey())
		}
		if node.EventKey != nil {
			addField("event_key", shortEventKey(node.EventKey))
		}
		if node.ScenarioKey != nil {
			addField("scenario_key", shortScenarioKey(node.ScenarioKey))
		}
		if node.IsDelete {
			addField("is_delete", "true        ")
		}
	}
	return s
}
