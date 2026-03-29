package parser_human

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Note: sort is still used in generateUseCaseContent for actor keys

func parseUseCase(subdomainKey identity.Key, useCaseSubKey, filename, contents string) (useCase model_use_case.UseCase, err error) {
	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_use_case.UseCase{}, errors.WithStack(err)
	}

	level := "sea"
	if levelAny, found := yamlData["level"]; found {
		level = levelAny.(string)
	}

	readOnly := strings.HasSuffix(parsedFile.Title, "?")

	superclassOfKey, err := parseUseCaseGenRefKey(subdomainKey, yamlData, "superclass_of_key")
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	subclassOfKey, err := parseUseCaseGenRefKey(subdomainKey, yamlData, "subclass_of_key")
	if err != nil {
		return model_use_case.UseCase{}, err
	}

	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, useCaseSubKey)
	if err != nil {
		return model_use_case.UseCase{}, errors.WithStack(err)
	}

	useCase = model_use_case.NewUseCase(useCaseKey, parsedFile.Title, stripMarkdownTitle(parsedFile.Markdown), level, readOnly, model_use_case.GeneralizationRefs{SuperclassOfKey: superclassOfKey, SubclassOfKey: subclassOfKey}, parsedFile.UmlComment)

	// Parse actors.
	if err := parseUseCaseActors(&useCase, subdomainKey, yamlData); err != nil {
		return model_use_case.UseCase{}, err
	}

	// Parse scenarios.
	if err := parseUseCaseScenarios(&useCase, subdomainKey, useCaseKey, yamlData); err != nil {
		return model_use_case.UseCase{}, err
	}

	return useCase, nil
}

// parseUseCaseGenRefKey extracts an optional generalization reference key from use case YAML data.
func parseUseCaseGenRefKey(subdomainKey identity.Key, yamlData map[string]any, field string) (*identity.Key, error) {
	s, ok := yamlData[field]
	if !ok {
		return nil, nil //nolint:nilnil // optional field, absence is not an error
	}
	k, err := identity.NewUseCaseGeneralizationKey(subdomainKey, s.(string))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &k, nil
}

// parseUseCaseActors parses the actors section from YAML data and sets them on the use case.
func parseUseCaseActors(useCase *model_use_case.UseCase, subdomainKey identity.Key, yamlData map[string]any) error {
	actorsAny, found := yamlData["actors"]
	if !found {
		return nil
	}
	useCase.Actors = map[identity.Key]model_use_case.Actor{}
	actorsMap := actorsAny.(map[string]any)
	for actorKeyStr, commentAny := range actorsMap {
		comment := ""
		if commentStr, ok := commentAny.(string); ok {
			comment = commentStr
		}
		actor := model_use_case.NewActor(comment)
		actorKey, err := identity.NewClassKey(subdomainKey, actorKeyStr)
		if err != nil {
			return errors.WithStack(err)
		}
		useCase.Actors[actorKey] = actor
	}
	return nil
}

// parseUseCaseScenarios parses the scenarios section from YAML data and sets them on the use case.
func parseUseCaseScenarios(useCase *model_use_case.UseCase, subdomainKey, useCaseKey identity.Key, yamlData map[string]any) error {
	scenariosAny, found := yamlData["scenarios"]
	if !found {
		return nil
	}
	useCase.Scenarios = make(map[identity.Key]model_scenario.Scenario)
	scenariosMap := scenariosAny.(map[string]any)
	for scenarioSubKey, scenarioDataAny := range scenariosMap {
		scenarioKey, err := identity.NewScenarioKey(useCaseKey, identity.NormalizeSubKey(scenarioSubKey))
		if err != nil {
			return errors.WithStack(err)
		}
		scenario, err := parseOneScenario(scenarioKey, subdomainKey, scenarioDataAny)
		if err != nil {
			return err
		}
		useCase.Scenarios[scenario.Key] = scenario
	}
	return nil
}

// parseOneScenario parses a single scenario from YAML data.
func parseOneScenario(scenarioKey, subdomainKey identity.Key, scenarioDataAny any) (model_scenario.Scenario, error) {
	scenarioData := scenarioDataAny.(map[string]any)

	name := ""
	if nameAny, found := scenarioData["name"]; found {
		name = nameAny.(string)
	}
	details := ""
	if detailsAny, found := scenarioData["details"]; found {
		details = detailsAny.(string)
	}

	scenario := model_scenario.NewScenario(scenarioKey, name, details)

	// Parse objects.
	if objectsAny, found := scenarioData["objects"]; found {
		scenario.Objects = make(map[identity.Key]model_scenario.Object)
		objectsSlice := objectsAny.([]any)
		for i, objAny := range objectsSlice {
			object, err := objectFromYamlData(scenarioKey, i, objAny)
			if err != nil {
				return model_scenario.Scenario{}, err
			}
			scenario.Objects[object.Key] = object
		}
	}

	// Parse steps.
	if err := parseScenarioSteps(&scenario, scenarioKey, subdomainKey, scenarioData); err != nil {
		return model_scenario.Scenario{}, err
	}

	return scenario, nil
}

// parseScenarioSteps parses the steps section for a scenario.
func parseScenarioSteps(scenario *model_scenario.Scenario, scenarioKey, subdomainKey identity.Key, scenarioData map[string]any) error {
	stepsAny, found := scenarioData["steps"]
	if !found {
		return nil
	}
	stepsData := stepsAny.([]any)

	nodeData := map[string]any{
		"step_type":  "sequence",
		"statements": stepsData,
	}

	if err := scopeObjectKeys(scenarioKey, subdomainKey, nodeData); err != nil {
		return err
	}

	nodeYaml, err := yaml.Marshal(nodeData)
	if err != nil {
		return err
	}

	var node model_scenario.Step
	if err = node.FromYAML(string(nodeYaml)); err != nil {
		return err
	}

	assignStepKeys(&node, scenarioKey)
	scenario.Steps = &node
	return nil
}

func objectFromYamlData(scenarioKey identity.Key, objectI int, objectAny any) (object model_scenario.Object, err error) {
	objectNum := uint(objectI + 1) //nolint:gosec // objectI is a small slice index from parsed YAML data, no overflow risk

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
		objectSubKey = identity.NormalizeSubKey(objectSubKey)

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

	object = model_scenario.NewObject(
		objectKey,
		objectNum,
		name,
		nameStyle,
		classKey,
		multi,
		umlComment)

	return object, nil
}

func generateUseCaseContent(useCase model_use_case.UseCase) string {
	var yb strings.Builder

	generateUseCaseTopFields(&yb, useCase)
	generateUseCaseActorsYaml(&yb, useCase)
	generateUseCaseScenariosYaml(&yb, useCase)

	yamlStr := strings.TrimSpace(yb.String())
	if yamlStr == "" {
		yamlStr = "\n"
	}
	content := prependMarkdownTitle(useCase.Name, useCase.Details) + "\n\n◆\n\n" + useCase.UmlComment + "\n\n◇"
	if yamlStr != "" {
		content += "\n\n" + yamlStr
	}
	return strings.TrimSpace(content)
}

// generateUseCaseTopFields writes the top-level YAML fields (level, superclass_of_key, subclass_of_key).
func generateUseCaseTopFields(yb *strings.Builder, useCase model_use_case.UseCase) {
	if useCase.Level != "sea" {
		yb.WriteString("level: " + useCase.Level + "\n")
	}
	if useCase.SuperclassOfKey != nil {
		yb.WriteString("superclass_of_key: " + useCase.SuperclassOfKey.SubKey + "\n")
	}
	if useCase.SubclassOfKey != nil {
		yb.WriteString("subclass_of_key: " + useCase.SubclassOfKey.SubKey + "\n")
	}
}

// generateUseCaseActorsYaml writes the actors YAML section.
func generateUseCaseActorsYaml(yb *strings.Builder, useCase model_use_case.UseCase) {
	if len(useCase.Actors) == 0 {
		return
	}
	actors := make(map[string]string)
	for actorKey, actor := range useCase.Actors {
		if actor.UmlComment != "" {
			actors[actorKey.SubKey] = actor.UmlComment
		}
	}
	if len(actors) == 0 {
		return
	}
	yb.WriteString("\nactors:\n")
	keys := make([]string, 0, len(actors))
	for k := range actors {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, k := range keys {
		yb.WriteString("    " + k + ": " + actors[k] + "\n")
	}
}

// generateUseCaseScenariosYaml writes the scenarios YAML section.
func generateUseCaseScenariosYaml(yb *strings.Builder, useCase model_use_case.UseCase) {
	if len(useCase.Scenarios) == 0 {
		return
	}
	yb.WriteString("\nscenarios:\n")
	scenarios := make([]model_scenario.Scenario, 0, len(useCase.Scenarios))
	for _, scenario := range useCase.Scenarios {
		scenarios = append(scenarios, scenario)
	}
	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].Key.String() < scenarios[j].Key.String()
	})
	for _, scenario := range scenarios {
		generateOneScenarioYaml(yb, scenario, useCase.Key)
	}
}

// generateOneScenarioYaml writes a single scenario's YAML content.
func generateOneScenarioYaml(yb *strings.Builder, scenario model_scenario.Scenario, useCaseKey identity.Key) {
	yb.WriteString("\n    " + scenario.Key.SubKey + ":\n")
	yb.WriteString("        name: " + scenario.Name + "\n")
	yb.WriteString(formatYamlField("details", scenario.Details, 8))
	generateScenarioObjectsYaml(yb, scenario)
	if scenario.Steps != nil && len(scenario.Steps.Statements) > 0 {
		yb.WriteString("        steps:\n")
		yb.WriteString(generateSteps(scenario.Steps.Statements, "            ", useCaseKey))
	}
}

// generateScenarioObjectsYaml writes the objects section within a scenario.
func generateScenarioObjectsYaml(yb *strings.Builder, scenario model_scenario.Scenario) {
	if len(scenario.Objects) == 0 {
		return
	}
	yb.WriteString("        objects:\n")
	objects := make([]model_scenario.Object, 0, len(scenario.Objects))
	for _, obj := range scenario.Objects {
		objects = append(objects, obj)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ObjectNumber < objects[j].ObjectNumber
	})
	for _, obj := range objects {
		yb.WriteString("            - key: " + obj.Key.SubKey + "\n")
		yb.WriteString(formatYamlField("name", obj.Name, 14))
		if obj.NameStyle != "" && obj.NameStyle != "unnamed" {
			yb.WriteString("              style: " + obj.NameStyle + "\n")
		}
		if obj.ClassKey.String() != "" {
			yb.WriteString("              class_key: " + obj.ClassKey.SubKey + "\n")
		}
		if obj.Multi {
			yb.WriteString("              multi: true\n")
		}
		yb.WriteString(formatYamlField("uml_comment", obj.UmlComment, 14))
	}
}

func generateSteps(steps []model_scenario.Step, indent string, useCaseKey identity.Key) string {
	var sb strings.Builder
	for _, step := range steps {
		sb.WriteString(generateStep(step, indent, useCaseKey))
	}
	return sb.String()
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
