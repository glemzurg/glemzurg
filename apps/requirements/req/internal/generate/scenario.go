package generate

import (
	"path/filepath"

	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/pkg/errors"
)

func generateScenarioFiles(outputPath string, reqs requirements.Requirements) (err error) {

	// Get all the data we want for these files.
	scenarioLookup := reqs.ScenarioLookup()

	// Generate file for each actor.
	for _, scenario := range scenarioLookup {

		// Generate a diagram.
		svgFilename := convertKeyToFilename("scenario", scenario.Key, "", ".svg")
		svgFilenameAbs := filepath.Join(outputPath, svgFilename)
		svgContents, err := generateScenarioSvgContents(reqs, scenario)
		if err != nil {
			return err
		}
		if err = writeFile(svgFilenameAbs, svgContents); err != nil {
			return err
		}
	}

	return nil
}

func generateScenarioSvgContents(reqs requirements.Requirements, scenario model_scenario.Scenario) (contents string, err error) {

	eventLookup := reqs.EventLookup()
	scenarioLookup := reqs.ScenarioLookup()
	objectLookup := reqs.ObjectLookup()

	s := svgsequence.NewSequence()

	// Add the actors in order.
	var actors []string
	if len(scenario.Objects) > 0 {
		for _, obj := range scenario.Objects {

			// Get fully populated object for proper name construction.
			object, found := objectLookup[obj.Key]
			if !found {
				return "", errors.Errorf("unknown object key: '%s'", obj.Key)
			}

			actors = append(actors, object.GetName())
		}
	} else {
		// No objects defined, so add placeholder actor for the placard.
		actors = append(actors, "No actors defined")
	}
	s.AddActors(actors...)

	// Add the steps.

	if len(scenario.Steps.Statements) == 0 {
		// No steps, so just add an informative placard.
		s.AddStep(svgsequence.Step{Source: actors[0], Target: actors[0], Text: "No operations defined"})
	} else {
		err = addSteps(eventLookup, s, scenario.Steps.Statements, scenarioLookup, objectLookup)
		if err != nil {
			return "", err
		}
	}

	contents, err = s.Generate()

	return contents, nil
}

func addSteps(eventLookup map[string]model_state.Event, s *svgsequence.Sequence, statements []model_scenario.Node, scenarioLookup map[string]model_scenario.Scenario, objectLookup map[string]model_scenario.Object) error {
	for _, stmt := range statements {
		switch stmt.Inferredtype() {
		case model_scenario.NODE_TYPE_LEAF:

			switch {

			case stmt.EventKey != "", stmt.AttributeKey != "":

				fromObject, found := objectLookup[stmt.FromObjectKey]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey)
				}
				toObject, found := objectLookup[stmt.ToObjectKey]
				if !found {
					return errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey)
				}

				text := stmt.Description

				// Events can be fully described.
				if stmt.EventKey != "" {
					event, found := eventLookup[stmt.EventKey]
					if !found {
						return errors.Errorf("unknown event key: '%s'", stmt.EventKey)
					}
					text += event.Name
					if len(event.Parameters) > 0 {
						text += "("
						for i, param := range event.Parameters {
							if i > 0 {
								text += ", "
							}
							text += param.Name
						}
						text += ")"
					}
				}

				s.AddStep(svgsequence.Step{
					Source: fromObject.GetName(),
					Target: toObject.GetName(),
					Text:   text,
				})

			case stmt.ScenarioKey != "":

				fromObject, found := objectLookup[stmt.FromObjectKey]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey)
				}
				toObject, found := objectLookup[stmt.ToObjectKey]
				if !found {
					return errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey)
				}

				// This is a call to another scenario.
				calledScenario, found := scenarioLookup[stmt.ScenarioKey]
				if !found {
					return errors.Errorf("unknown called scenario object key: '%s'", stmt.ToObjectKey)
				}
				s.AddStep(svgsequence.Step{
					Source: fromObject.GetName(),
					Target: toObject.GetName(),
					Text:   "Scenario: " + calledScenario.Name,
				})

			case stmt.IsDelete:
				// This is a delete operation.
				fromObject, found := objectLookup[stmt.FromObjectKey]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey)
				}
				s.AddStep(svgsequence.Step{
					Source: fromObject.GetName(),
					Target: fromObject.GetName(),
					Text:   "(delete)",
				})

			default:
				return errors.Errorf("leaf node must have one of event_key, scenario_key, attribute_key, or is_delete: '%+v'", stmt)
			}

		case model_scenario.NODE_TYPE_SEQUENCE:
			err := addSteps(eventLookup, s, stmt.Statements, scenarioLookup, objectLookup)
			if err != nil {
				return err
			}

		case model_scenario.NODE_TYPE_SWITCH:

			sectionLabel := "(Opt)"
			if len(stmt.Cases) > 1 {
				sectionLabel = "(Alt)"
			}

			for _, c := range stmt.Cases {
				s.OpenSection(sectionLabel+" ["+c.Condition+"]", "")

				err := addSteps(eventLookup, s, c.Statements, scenarioLookup, objectLookup)
				if err != nil {
					return err
				}

				s.CloseSection()
			}

		case model_scenario.NODE_TYPE_LOOP:
			s.OpenSection("(Loop) ["+stmt.Loop+"]", "")

			err := addSteps(eventLookup, s, stmt.Statements, scenarioLookup, objectLookup)
			if err != nil {
				return err
			}

			s.CloseSection()

		default:
			return errors.Errorf("unsupported node type in scenario SVG generation: '%s'", stmt.Inferredtype())
		}
	}
	return nil
}
