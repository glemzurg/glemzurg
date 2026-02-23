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

	// Parse optional superclass/subclass generalization keys.
	var superclassOfKey *identity.Key
	if s, ok := yamlData["superclass_of_key"]; ok {
		k, err := identity.NewUseCaseGeneralizationKey(subdomainKey, s.(string))
		if err != nil {
			return model_use_case.UseCase{}, errors.WithStack(err)
		}
		superclassOfKey = &k
	}
	var subclassOfKey *identity.Key
	if s, ok := yamlData["subclass_of_key"]; ok {
		k, err := identity.NewUseCaseGeneralizationKey(subdomainKey, s.(string))
		if err != nil {
			return model_use_case.UseCase{}, errors.WithStack(err)
		}
		subclassOfKey = &k
	}

	// Construct the identity key for this use case.
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, useCaseSubKey)
	if err != nil {
		return model_use_case.UseCase{}, errors.WithStack(err)
	}

	useCase, err = model_use_case.NewUseCase(useCaseKey, parsedFile.Title, stripMarkdownTitle(parsedFile.Markdown), level, readOnly, superclassOfKey, subclassOfKey, parsedFile.UmlComment)
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

				// Wrap in outer sequence step.
				nodeData := map[string]any{
					"step_type":  "sequence",
					"statements": stepsData,
				}

				// Scope compact keys to fully qualified keys before parsing into Step objects.
				if err = scopeObjectKeys(scenarioKey, subdomainKey, nodeData); err != nil {
					return model_use_case.UseCase{}, err
				}

				// Turn into yaml.
				nodeYaml, err := yaml.Marshal(nodeData)
				if err != nil {
					return model_use_case.UseCase{}, err
				}

				var node model_scenario.Step
				if err = node.FromYAML(string(nodeYaml)); err != nil {
					return model_use_case.UseCase{}, err
				}

				// Auto-assign step keys from tree position.
				assignStepKeys(&node, scenarioKey)

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
	if useCase.SuperclassOfKey != nil {
		yaml += "superclass_of_key: " + useCase.SuperclassOfKey.SubKey + "\n"
	}
	if useCase.SubclassOfKey != nil {
		yaml += "subclass_of_key: " + useCase.SubclassOfKey.SubKey + "\n"
	}

	if len(useCase.Actors) > 0 {
		actors := make(map[string]string)
		for actorKey, actor := range useCase.Actors {
			if actor.UmlComment != "" {
				actors[actorKey.SubKey] = actor.UmlComment
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
			name := scenario.Key.SubKey
			yaml += "\n    " + name + ":\n"
			yaml += "        name: " + scenario.Name + "\n"
			yaml += formatYamlField("details", scenario.Details, 8)
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
					objName := obj.Key.SubKey
					yaml += "            - key: " + objName + "\n"
					yaml += formatYamlField("name", obj.Name, 14)
					if obj.NameStyle != "" && obj.NameStyle != "unnamed" {
						yaml += "              style: " + obj.NameStyle + "\n"
					}
					if obj.ClassKey.String() != "" {
						// Output only the subkey for backwards compatibility with the md format.
						yaml += "              class_key: " + obj.ClassKey.SubKey + "\n"
					}
					if obj.Multi {
						yaml += "              multi: true\n"
					}
					yaml += formatYamlField("uml_comment", obj.UmlComment, 14)
				}
			}
			if scenario.Steps != nil && len(scenario.Steps.Statements) > 0 {
				yaml += "        steps:\n"
				yaml += generateSteps(scenario.Steps.Statements, "            ", useCase.Key)
			}
		}
	}

	yamlStr := strings.TrimSpace(yaml)
	if yamlStr == "" {
		yamlStr = "\n"
	}
	content := prependMarkdownTitle(useCase.Name, useCase.Details) + "\n\n◆\n\n" + useCase.UmlComment + "\n\n◇"
	if yamlStr != "" {
		content += "\n\n" + yamlStr
	}
	return strings.TrimSpace(content)
}

func generateSteps(steps []model_scenario.Step, indent string, useCaseKey identity.Key) string {
	s := ""
	for _, step := range steps {
		s += generateStep(step, indent, useCaseKey)
	}
	return s
}

func generateStep(step model_scenario.Step, indent string, useCaseKey identity.Key) string {
	s := indent + "- step_type: " + step.StepType + "\n"

	if step.LeafType != nil {
		s += indent + "  leaf_type: " + *step.LeafType + "\n"
	}
	if step.Condition != "" {
		s += indent + "  condition: " + step.Condition + "\n"
	}
	if step.Description != "" {
		s += indent + "  description: " + step.Description + "\n"
	}
	if step.FromObjectKey != nil {
		s += indent + "  from_object_key: " + step.FromObjectKey.SubKey + "\n"
	}
	if step.ToObjectKey != nil {
		s += indent + "  to_object_key: " + step.ToObjectKey.SubKey + "\n"
	}
	if step.EventKey != nil {
		s += indent + "  event_key: " + shortEventKey(step.EventKey) + "\n"
	}
	if step.QueryKey != nil {
		s += indent + "  query_key: " + shortQueryKey(step.QueryKey) + "\n"
	}
	if step.ScenarioKey != nil {
		s += indent + "  scenario_key: " + shortScenarioKey(step.ScenarioKey, useCaseKey) + "\n"
	}
	if len(step.Statements) > 0 {
		s += indent + "  statements:\n"
		s += generateSteps(step.Statements, indent+"      ", useCaseKey)
	}
	return s
}

// assignStepKeys walks the Step tree and assigns keys using a sequential counter.
func assignStepKeys(step *model_scenario.Step, scenarioKey identity.Key) {
	counter := 0
	assignStepKeysRecursive(step, scenarioKey, &counter)
}

func assignStepKeysRecursive(step *model_scenario.Step, scenarioKey identity.Key, counter *int) {
	key, _ := identity.NewScenarioStepKey(scenarioKey, strconv.Itoa(*counter))
	step.Key = key
	*counter++
	for i := range step.Statements {
		assignStepKeysRecursive(&step.Statements[i], scenarioKey, counter)
	}
}

// shortEventKey returns the compact form of an event key: "classSubKey/eventSubKey".
func shortEventKey(key *identity.Key) string {
	parentKey, err := identity.ParseKey(key.ParentKey)
	if err != nil {
		return key.String()
	}
	return parentKey.SubKey + "/" + key.SubKey
}

// shortQueryKey returns the compact form of a query key: "classSubKey/querySubKey".
func shortQueryKey(key *identity.Key) string {
	parentKey, err := identity.ParseKey(key.ParentKey)
	if err != nil {
		return key.String()
	}
	return parentKey.SubKey + "/" + key.SubKey
}

// shortScenarioKey returns the compact form of a scenario key.
// If the scenario is in the same use case as the step's scenario, returns just "scenarioSubKey".
// Otherwise returns "useCaseSubKey/scenario/scenarioSubKey".
func shortScenarioKey(key *identity.Key, useCaseKey identity.Key) string {
	parentKey, err := identity.ParseKey(key.ParentKey)
	if err != nil {
		return key.String()
	}
	if parentKey == useCaseKey {
		return key.SubKey
	}
	return parentKey.SubKey + "/scenario/" + key.SubKey
}
