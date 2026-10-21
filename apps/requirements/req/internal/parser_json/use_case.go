package parser_json

import (
	"cmp"
	"slices"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeUseCaseGeneralization struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details,omitempty"`
	UnfinishedNotes string `json:"unfinished_notes,omitempty"`
	IsComplete      bool   `json:"is_complete,omitempty"`
	IsStatic        bool   `json:"is_static,omitempty"`
	UmlComment      string `json:"uml_comment,omitempty"`
}

type pipeUseCase struct {
	Key             string             `json:"key"`
	Name            string             `json:"name"`
	Details         string             `json:"details,omitempty"`
	UnfinishedNotes string             `json:"unfinished_notes,omitempty"`
	Level           string             `json:"level"`
	ReadOnly        bool               `json:"read_only,omitempty"`
	SuperclassOfKey string             `json:"superclass_of_key,omitempty"`
	SubclassOfKey   string             `json:"subclass_of_key,omitempty"`
	UmlComment      string             `json:"uml_comment,omitempty"`
	Actors          []pipeUseCaseActor `json:"actors,omitempty"`
	Scenarios       []pipeScenario     `json:"scenarios,omitempty"`
}

type pipeUseCaseActor struct {
	ClassKey   string `json:"class_key"`
	UmlComment string `json:"uml_comment,omitempty"`
}

type pipeUseCaseShare struct {
	SeaUseCaseKey string `json:"sea_use_case_key"`
	MudUseCaseKey string `json:"mud_use_case_key"`
	ShareType     string `json:"share_type"`
	UmlComment    string `json:"uml_comment,omitempty"`
}

type pipeScenario struct {
	Key     string               `json:"key"`
	Name    string               `json:"name"`
	Details string               `json:"details,omitempty"`
	Objects []pipeScenarioObject `json:"objects,omitempty"`
	Steps   *pipeStep            `json:"steps,omitempty"`
}

type pipeScenarioObject struct {
	Key          string `json:"key"`
	ObjectNumber uint   `json:"object_number,omitempty"`
	Name         string `json:"name,omitempty"`
	NameStyle    string `json:"name_style"`
	ClassKey     string `json:"class_key"`
	Multi        bool   `json:"multi,omitempty"`
	UmlComment   string `json:"uml_comment,omitempty"`
}

type pipeStep struct {
	Key           string     `json:"key"`
	StepType      string     `json:"step_type"`
	LeafType      string     `json:"leaf_type,omitempty"`
	Statements    []pipeStep `json:"statements,omitempty"`
	Condition     string     `json:"condition,omitempty"`
	Description   string     `json:"description,omitempty"`
	FromObjectKey string     `json:"from_object_key,omitempty"`
	ToObjectKey   string     `json:"to_object_key,omitempty"`
	EventKey      string     `json:"event_key,omitempty"`
	QueryKey      string     `json:"query_key,omitempty"`
	ScenarioKey   string     `json:"scenario_key,omitempty"`
}

func useCaseGeneralizationsFromCore(items map[identity.Key]model_use_case.Generalization) []pipeUseCaseGeneralization {
	if len(items) == 0 {
		return nil
	}
	out := make([]pipeUseCaseGeneralization, 0, len(items))
	for _, item := range items {
		out = append(out, pipeUseCaseGeneralization{
			Key:             keyString(item.Key),
			Name:            item.Name,
			Details:         item.Details,
			UnfinishedNotes: item.UnfinishedNotes,
			IsComplete:      item.IsComplete,
			IsStatic:        item.IsStatic,
			UmlComment:      item.UmlComment,
		})
	}
	sortByKey(out, func(item pipeUseCaseGeneralization) string { return item.Key })
	return out
}

func useCaseGeneralizationsToCore(items []pipeUseCaseGeneralization) (map[identity.Key]model_use_case.Generalization, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_use_case.Generalization, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "use_case_generalizations[%d]", i)
		}
		generalization := model_use_case.NewGeneralization(key, model_use_case.GeneralizationDetails{
			Name:    item.Name,
			Details: item.Details,
		}, item.UnfinishedNotes, model_use_case.GeneralizationTraits{
			IsComplete: item.IsComplete,
			IsStatic:   item.IsStatic,
		}, item.UmlComment)
		if err := putUnique(out, generalization.Key, generalization, i); err != nil {
			return nil, errors.Wrap(err, "use_case_generalizations")
		}
	}
	return out, nil
}

func useCasesFromCore(useCases map[identity.Key]model_use_case.UseCase) []pipeUseCase {
	if len(useCases) == 0 {
		return nil
	}
	out := make([]pipeUseCase, 0, len(useCases))
	for _, useCase := range useCases {
		out = append(out, useCaseFromCore(useCase))
	}
	sortByKey(out, func(item pipeUseCase) string { return item.Key })
	return out
}

func useCaseFromCore(useCase model_use_case.UseCase) pipeUseCase {
	return pipeUseCase{
		Key:             keyString(useCase.Key),
		Name:            useCase.Name,
		Details:         useCase.Details,
		UnfinishedNotes: useCase.UnfinishedNotes,
		Level:           useCase.Level,
		ReadOnly:        useCase.ReadOnly,
		SuperclassOfKey: optionalKeyString(useCase.SuperclassOfKey),
		SubclassOfKey:   optionalKeyString(useCase.SubclassOfKey),
		UmlComment:      useCase.UmlComment,
		Actors:          useCaseActorsFromCore(useCase.Actors),
		Scenarios:       scenariosFromCore(useCase.Scenarios),
	}
}

func useCasesToCore(items []pipeUseCase) (map[identity.Key]model_use_case.UseCase, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_use_case.UseCase, len(items))
	for i, item := range items {
		useCase, err := useCaseToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "use_cases[%d]", i)
		}
		if err := putUnique(out, useCase.Key, useCase, i); err != nil {
			return nil, errors.Wrap(err, "use_cases")
		}
	}
	return out, nil
}

func useCaseToCore(item pipeUseCase) (model_use_case.UseCase, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	superKey, err := optionalKeyToCore("superclass_of_key", item.SuperclassOfKey)
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	subKey, err := optionalKeyToCore("subclass_of_key", item.SubclassOfKey)
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	actors, err := useCaseActorsToCore(item.Actors)
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	scenarios, err := scenariosToCore(item.Scenarios)
	if err != nil {
		return model_use_case.UseCase{}, err
	}
	useCase := model_use_case.NewUseCase(key, model_use_case.UseCaseTraits{
		Level:    item.Level,
		ReadOnly: item.ReadOnly,
	}, model_use_case.GeneralizationRefs{
		SuperclassOfKey: superKey,
		SubclassOfKey:   subKey,
	}, model_use_case.UseCaseDetails{
		Name:            item.Name,
		Details:         item.Details,
		UnfinishedNotes: item.UnfinishedNotes,
		UmlComment:      item.UmlComment,
	})
	useCase.SetActors(actors)
	useCase.SetScenarios(scenarios)
	return useCase, nil
}

func useCaseActorsFromCore(actors map[identity.Key]model_use_case.Actor) []pipeUseCaseActor {
	if len(actors) == 0 {
		return nil
	}
	out := make([]pipeUseCaseActor, 0, len(actors))
	for classKey, actor := range actors {
		out = append(out, pipeUseCaseActor{
			ClassKey:   keyString(classKey),
			UmlComment: actor.UmlComment,
		})
	}
	sortByKey(out, func(item pipeUseCaseActor) string { return item.ClassKey })
	return out
}

func useCaseActorsToCore(items []pipeUseCaseActor) (map[identity.Key]model_use_case.Actor, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_use_case.Actor, len(items))
	for i, item := range items {
		classKey, err := parseKey("class_key", item.ClassKey)
		if err != nil {
			return nil, errors.Wrapf(err, "actors[%d]", i)
		}
		if err := putUnique(out, classKey, model_use_case.NewActor(item.UmlComment), i); err != nil {
			return nil, errors.Wrap(err, "actors")
		}
	}
	return out, nil
}

func useCaseSharesFromCore(shares map[identity.Key]map[identity.Key]model_use_case.UseCaseShared) []pipeUseCaseShare {
	if len(shares) == 0 {
		return nil
	}
	var out []pipeUseCaseShare
	for seaKey, mudShares := range shares {
		for mudKey, share := range mudShares {
			out = append(out, pipeUseCaseShare{
				SeaUseCaseKey: keyString(seaKey),
				MudUseCaseKey: keyString(mudKey),
				ShareType:     share.ShareType,
				UmlComment:    share.UmlComment,
			})
		}
	}
	if len(out) == 0 {
		return nil
	}
	slices.SortFunc(out, func(a, b pipeUseCaseShare) int {
		if sea := cmp.Compare(a.SeaUseCaseKey, b.SeaUseCaseKey); sea != 0 {
			return sea
		}
		return cmp.Compare(a.MudUseCaseKey, b.MudUseCaseKey)
	})
	return out
}

func useCaseSharesToCore(items []pipeUseCaseShare) (map[identity.Key]map[identity.Key]model_use_case.UseCaseShared, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)
	for i, item := range items {
		seaKey, err := parseKey("sea_use_case_key", item.SeaUseCaseKey)
		if err != nil {
			return nil, errors.Wrapf(err, "use_case_shares[%d]", i)
		}
		mudKey, err := parseKey("mud_use_case_key", item.MudUseCaseKey)
		if err != nil {
			return nil, errors.Wrapf(err, "use_case_shares[%d]", i)
		}
		if out[seaKey] == nil {
			out[seaKey] = make(map[identity.Key]model_use_case.UseCaseShared)
		}
		if _, exists := out[seaKey][mudKey]; exists {
			return nil, errors.Errorf("use_case_shares[%d]: duplicate key %s %s", i, seaKey.String(), mudKey.String())
		}
		out[seaKey][mudKey] = model_use_case.NewUseCaseShared(item.ShareType, item.UmlComment)
	}
	return out, nil
}

func scenariosFromCore(scenarios map[identity.Key]model_scenario.Scenario) []pipeScenario {
	if len(scenarios) == 0 {
		return nil
	}
	out := make([]pipeScenario, 0, len(scenarios))
	for _, scenario := range scenarios {
		item := pipeScenario{
			Key:     keyString(scenario.Key),
			Name:    scenario.Name,
			Details: scenario.Details,
			Objects: scenarioObjectsFromCore(scenario.Objects),
		}
		if scenario.Steps != nil {
			steps := stepFromCore(*scenario.Steps)
			item.Steps = &steps
		}
		out = append(out, item)
	}
	sortByKey(out, func(item pipeScenario) string { return item.Key })
	return out
}

func scenariosToCore(items []pipeScenario) (map[identity.Key]model_scenario.Scenario, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_scenario.Scenario, len(items))
	for i, item := range items {
		scenario, err := scenarioToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "scenarios[%d]", i)
		}
		if err := putUnique(out, scenario.Key, scenario, i); err != nil {
			return nil, errors.Wrap(err, "scenarios")
		}
	}
	return out, nil
}

func scenarioToCore(item pipeScenario) (model_scenario.Scenario, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_scenario.Scenario{}, err
	}
	objects, err := scenarioObjectsToCore(item.Objects)
	if err != nil {
		return model_scenario.Scenario{}, err
	}
	scenario := model_scenario.NewScenario(key, item.Name, item.Details)
	scenario.SetObjects(objects)
	if item.Steps != nil {
		steps, err := stepToCore(*item.Steps)
		if err != nil {
			return model_scenario.Scenario{}, errors.Wrap(err, "steps")
		}
		scenario.Steps = &steps
	}
	return scenario, nil
}

func scenarioObjectsFromCore(objects map[identity.Key]model_scenario.Object) []pipeScenarioObject {
	if len(objects) == 0 {
		return nil
	}
	out := make([]pipeScenarioObject, 0, len(objects))
	for _, object := range objects {
		out = append(out, pipeScenarioObject{
			Key:          keyString(object.Key),
			ObjectNumber: object.ObjectNumber,
			Name:         object.Name,
			NameStyle:    object.NameStyle,
			ClassKey:     keyString(object.ClassKey),
			Multi:        object.Multi,
			UmlComment:   object.UmlComment,
		})
	}
	sortByKey(out, func(item pipeScenarioObject) string { return item.Key })
	return out
}

func scenarioObjectsToCore(items []pipeScenarioObject) (map[identity.Key]model_scenario.Object, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_scenario.Object, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "objects[%d]", i)
		}
		classKey, err := parseKey("class_key", item.ClassKey)
		if err != nil {
			return nil, errors.Wrapf(err, "objects[%d]", i)
		}
		object := model_scenario.NewObject(key, item.ObjectNumber, model_scenario.ObjectDiagramName{
			Name:      item.Name,
			NameStyle: item.NameStyle,
		}, classKey, item.Multi, item.UmlComment)
		if err := putUnique(out, object.Key, object, i); err != nil {
			return nil, errors.Wrap(err, "objects")
		}
	}
	return out, nil
}

func stepFromCore(step model_scenario.Step) pipeStep {
	item := pipeStep{
		Key:           keyString(step.Key),
		StepType:      step.StepType,
		Condition:     step.Condition,
		Description:   step.Description,
		FromObjectKey: optionalKeyString(step.FromObjectKey),
		ToObjectKey:   optionalKeyString(step.ToObjectKey),
		EventKey:      optionalKeyString(step.EventKey),
		QueryKey:      optionalKeyString(step.QueryKey),
		ScenarioKey:   optionalKeyString(step.ScenarioKey),
	}
	if step.LeafType != nil {
		item.LeafType = *step.LeafType
	}
	for _, statement := range step.Statements {
		item.Statements = append(item.Statements, stepFromCore(statement))
	}
	return item
}

// stepToCore rebuilds a scenario step. Step has no constructor; its tree is the value.
func stepToCore(item pipeStep) (model_scenario.Step, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_scenario.Step{}, err
	}
	from, err := optionalKeyToCore("from_object_key", item.FromObjectKey)
	if err != nil {
		return model_scenario.Step{}, err
	}
	to, err := optionalKeyToCore("to_object_key", item.ToObjectKey)
	if err != nil {
		return model_scenario.Step{}, err
	}
	eventKey, err := optionalKeyToCore("event_key", item.EventKey)
	if err != nil {
		return model_scenario.Step{}, err
	}
	queryKey, err := optionalKeyToCore("query_key", item.QueryKey)
	if err != nil {
		return model_scenario.Step{}, err
	}
	scenarioKey, err := optionalKeyToCore("scenario_key", item.ScenarioKey)
	if err != nil {
		return model_scenario.Step{}, err
	}
	statements, err := stepsToCore(item.Statements)
	if err != nil {
		return model_scenario.Step{}, err
	}
	var leafType *string
	if item.LeafType != "" {
		value := item.LeafType
		leafType = &value
	}
	return model_scenario.Step{
		Key:           key,
		StepType:      item.StepType,
		LeafType:      leafType,
		Statements:    statements,
		Condition:     item.Condition,
		Description:   item.Description,
		FromObjectKey: from,
		ToObjectKey:   to,
		EventKey:      eventKey,
		QueryKey:      queryKey,
		ScenarioKey:   scenarioKey,
	}, nil
}

func stepsToCore(items []pipeStep) ([]model_scenario.Step, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_scenario.Step, 0, len(items))
	for i, item := range items {
		step, err := stepToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "statements[%d]", i)
		}
		out = append(out, step)
	}
	return out, nil
}
