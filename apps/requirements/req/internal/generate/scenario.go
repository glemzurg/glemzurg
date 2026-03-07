package generate

import (
	"strings"

	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/pkg/errors"
)

// stepContext bundles the lookup maps needed when processing scenario steps.
type stepContext struct {
	eventLookup    map[string]model_state.Event
	scenarioLookup map[string]model_scenario.Scenario
	objectLookup   map[string]model_scenario.Object
	classLookup    map[string]model_class.Class
}

func generateScenarioSvgContents(reqs *req_flat.Requirements, scenario model_scenario.Scenario) (contents string, err error) {
	classLookup, _ := reqs.ClassLookup()
	ctx := stepContext{
		eventLookup:    reqs.EventLookup(),
		scenarioLookup: reqs.ScenarioLookup(),
		objectLookup:   reqs.ObjectLookup(),
		classLookup:    classLookup,
	}

	s := svgsequence.NewSequence()

	// Add the actors in order.
	actors, err := buildActorList(ctx, scenario)
	if err != nil {
		return "", err
	}
	s.AddActors(actors...)

	// Add the steps.
	if scenario.Steps == nil || len(scenario.Steps.Statements) == 0 {
		// No steps, so just add an informative placard.
		s.AddStep(svgsequence.Step{Source: actors[0], Target: actors[0], Text: "No operations defined"})
	} else {
		err = addSteps(ctx, s, scenario.Steps.Statements)
		if err != nil {
			return "", err
		}
	}

	contents, err = s.Generate()

	return contents, err
}

// buildActorList constructs the ordered list of actor display names for a scenario.
func buildActorList(ctx stepContext, scenario model_scenario.Scenario) ([]string, error) {
	if len(scenario.Objects) == 0 {
		return []string{"No actors defined"}, nil
	}

	var actors []string
	for _, obj := range scenario.Objects {
		object, found := ctx.objectLookup[obj.Key.String()]
		if !found {
			return nil, errors.Errorf("unknown object key: '%s'", obj.Key.String())
		}
		class, found := ctx.classLookup[object.ClassKey.String()]
		if !found {
			return nil, errors.Errorf("unknown class key: '%s'", object.ClassKey.String())
		}
		actors = append(actors, object.GetName(class))
	}
	return actors, nil
}

func addSteps(ctx stepContext, s *svgsequence.Sequence, statements []model_scenario.Step) error {
	for _, stmt := range statements {
		switch stmt.StepType {
		case model_scenario.STEP_TYPE_LEAF:
			if err := addLeafStep(ctx, s, stmt); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_SEQUENCE:
			if err := addSteps(ctx, s, stmt.Statements); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_SWITCH:
			if err := addSwitchStep(ctx, s, stmt); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_LOOP:
			if err := addLoopStep(ctx, s, stmt); err != nil {
				return err
			}

		default:
			return errors.Errorf("unsupported step type in scenario SVG generation: '%s'", stmt.StepType)
		}
	}
	return nil
}

// addLeafStep dispatches a leaf step to the appropriate handler based on its leaf type.
func addLeafStep(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	if stmt.LeafType == nil {
		return errors.Errorf("leaf step missing leaf_type: '%+v'", stmt)
	}

	switch *stmt.LeafType {
	case model_scenario.LEAF_TYPE_EVENT:
		return addEventLeaf(ctx, s, stmt)
	case model_scenario.LEAF_TYPE_QUERY:
		return addQueryLeaf(ctx, s, stmt)
	case model_scenario.LEAF_TYPE_SCENARIO:
		return addScenarioLeaf(ctx, s, stmt)
	case model_scenario.LEAF_TYPE_DELETE:
		return addDeleteLeaf(ctx, s, stmt)
	default:
		return errors.Errorf("unsupported leaf type in scenario SVG generation: '%s'", *stmt.LeafType)
	}
}

// addSwitchStep handles STEP_TYPE_SWITCH by opening sections for each case.
func addSwitchStep(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	sectionLabel := "(Opt)"
	if len(stmt.Statements) > 1 {
		sectionLabel = "(Alt)"
	}

	for _, c := range stmt.Statements {
		s.OpenSection(sectionLabel+" ["+c.Condition+"]", "")

		if err := addSteps(ctx, s, c.Statements); err != nil {
			return err
		}

		s.CloseSection()
	}
	return nil
}

// addLoopStep handles STEP_TYPE_LOOP by opening a loop section.
func addLoopStep(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	s.OpenSection("(Loop) ["+stmt.Condition+"]", "")

	if err := addSteps(ctx, s, stmt.Statements); err != nil {
		return err
	}

	s.CloseSection()
	return nil
}

// resolveFromToObjects resolves the from and to objects and their classes for a step.
func resolveFromToObjects(ctx stepContext, stmt model_scenario.Step) (fromName, toName string, err error) {
	fromObject, found := ctx.objectLookup[stmt.FromObjectKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
	}
	toObject, found := ctx.objectLookup[stmt.ToObjectKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey.String())
	}

	fromClass, found := ctx.classLookup[fromObject.ClassKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown from class key: '%s'", fromObject.ClassKey.String())
	}
	toClass, found := ctx.classLookup[toObject.ClassKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown to class key: '%s'", toObject.ClassKey.String())
	}

	return fromObject.GetName(fromClass), toObject.GetName(toClass), nil
}

// addEventLeaf handles a LEAF_TYPE_EVENT step.
func addEventLeaf(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	fromName, toName, err := resolveFromToObjects(ctx, stmt)
	if err != nil {
		return err
	}

	text := buildEventText(ctx, stmt)

	s.AddStep(svgsequence.Step{
		Source: fromName,
		Target: toName,
		Text:   text,
	})
	return nil
}

// buildEventText constructs the display text for an event leaf step.
func buildEventText(ctx stepContext, stmt model_scenario.Step) string {
	var textBuilder strings.Builder
	textBuilder.WriteString(stmt.Description)

	if stmt.EventKey != nil {
		event, found := ctx.eventLookup[stmt.EventKey.String()]
		if found {
			textBuilder.WriteString(event.Name)
			if len(event.Parameters) > 0 {
				textBuilder.WriteString("(")
				for i, param := range event.Parameters {
					if i > 0 {
						textBuilder.WriteString(", ")
					}
					textBuilder.WriteString(param.Name)
				}
				textBuilder.WriteString(")")
			}
		}
	}

	return textBuilder.String()
}

// addQueryLeaf handles a LEAF_TYPE_QUERY step.
func addQueryLeaf(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	fromName, toName, err := resolveFromToObjects(ctx, stmt)
	if err != nil {
		return err
	}

	s.AddStep(svgsequence.Step{
		Source: fromName,
		Target: toName,
		Text:   stmt.Description,
	})
	return nil
}

// addScenarioLeaf handles a LEAF_TYPE_SCENARIO step.
func addScenarioLeaf(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	fromName, toName, err := resolveFromToObjects(ctx, stmt)
	if err != nil {
		return err
	}

	calledScenario, found := ctx.scenarioLookup[stmt.ScenarioKey.String()]
	if !found {
		return errors.Errorf("unknown called scenario key: '%s'", stmt.ScenarioKey.String())
	}
	s.AddStep(svgsequence.Step{
		Source: fromName,
		Target: toName,
		Text:   "Scenario: " + calledScenario.Name,
	})
	return nil
}

// addDeleteLeaf handles a LEAF_TYPE_DELETE step.
func addDeleteLeaf(ctx stepContext, s *svgsequence.Sequence, stmt model_scenario.Step) error {
	fromObject, found := ctx.objectLookup[stmt.FromObjectKey.String()]
	if !found {
		return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
	}

	fromClass, found := ctx.classLookup[fromObject.ClassKey.String()]
	if !found {
		return errors.Errorf("unknown from class key: '%s'", fromObject.ClassKey.String())
	}

	s.AddStep(svgsequence.Step{
		Source: fromObject.GetName(fromClass),
		Target: fromObject.GetName(fromClass),
		Text:   "(delete)",
	})
	return nil
}
