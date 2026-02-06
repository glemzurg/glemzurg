package generate

import (
	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/pkg/errors"
)

func generateScenarioSvgContents(reqs *req_flat.Requirements, scenario model_scenario.Scenario) (contents string, err error) {

	eventLookup := reqs.EventLookup()
	scenarioLookup := reqs.ScenarioLookup()
	objectLookup := reqs.ObjectLookup()
	classLookup, _ := reqs.ClassLookup()

	s := svgsequence.NewSequence()

	// Add the actors in order.
	var actors []string
	if len(scenario.Objects) > 0 {
		for _, obj := range scenario.Objects {

			// Get fully populated object for proper name construction.
			object, found := objectLookup[obj.Key.String()]
			if !found {
				return "", errors.Errorf("unknown object key: '%s'", obj.Key.String())
			}

			// Get the class for the object to build the name.
			class, found := classLookup[object.ClassKey.String()]
			if !found {
				return "", errors.Errorf("unknown class key: '%s'", object.ClassKey.String())
			}

			actors = append(actors, object.GetName(class))
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
		err = addSteps(eventLookup, s, scenario.Steps.Statements, scenarioLookup, objectLookup, classLookup)
		if err != nil {
			return "", err
		}
	}

	contents, err = s.Generate()

	return contents, nil
}

func addSteps(eventLookup map[string]model_state.Event, s *svgsequence.Sequence, statements []model_scenario.Node, scenarioLookup map[string]model_scenario.Scenario, objectLookup map[string]model_scenario.Object, classLookup map[string]model_class.Class) error {
	for _, stmt := range statements {
		switch stmt.Inferredtype() {
		case model_scenario.NODE_TYPE_LEAF:

			switch {

			case stmt.EventKey != nil, stmt.AttributeKey != nil:

				fromObject, found := objectLookup[stmt.FromObjectKey.String()]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
				}
				toObject, found := objectLookup[stmt.ToObjectKey.String()]
				if !found {
					return errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey.String())
				}

				// Get the classes for the objects.
				fromClass, found := classLookup[fromObject.ClassKey.String()]
				if !found {
					return errors.Errorf("unknown from class key: '%s'", fromObject.ClassKey.String())
				}
				toClass, found := classLookup[toObject.ClassKey.String()]
				if !found {
					return errors.Errorf("unknown to class key: '%s'", toObject.ClassKey.String())
				}

				text := stmt.Description

				// Events can be fully described.
				if stmt.EventKey != nil {
					event, found := eventLookup[stmt.EventKey.String()]
					if !found {
						return errors.Errorf("unknown event key: '%s'", stmt.EventKey.String())
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
					Source: fromObject.GetName(fromClass),
					Target: toObject.GetName(toClass),
					Text:   text,
				})

			case stmt.ScenarioKey != nil:

				fromObject, found := objectLookup[stmt.FromObjectKey.String()]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
				}
				toObject, found := objectLookup[stmt.ToObjectKey.String()]
				if !found {
					return errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey.String())
				}

				// Get the classes for the objects.
				fromClass, found := classLookup[fromObject.ClassKey.String()]
				if !found {
					return errors.Errorf("unknown from class key: '%s'", fromObject.ClassKey.String())
				}
				toClass, found := classLookup[toObject.ClassKey.String()]
				if !found {
					return errors.Errorf("unknown to class key: '%s'", toObject.ClassKey.String())
				}

				// This is a call to another scenario.
				calledScenario, found := scenarioLookup[stmt.ScenarioKey.String()]
				if !found {
					return errors.Errorf("unknown called scenario object key: '%s'", stmt.ToObjectKey.String())
				}
				s.AddStep(svgsequence.Step{
					Source: fromObject.GetName(fromClass),
					Target: toObject.GetName(toClass),
					Text:   "Scenario: " + calledScenario.Name,
				})

			case stmt.IsDelete:
				// This is a delete operation.
				fromObject, found := objectLookup[stmt.FromObjectKey.String()]
				if !found {
					return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
				}

				// Get the class for the object.
				fromClass, found := classLookup[fromObject.ClassKey.String()]
				if !found {
					return errors.Errorf("unknown from class key: '%s'", fromObject.ClassKey.String())
				}

				s.AddStep(svgsequence.Step{
					Source: fromObject.GetName(fromClass),
					Target: fromObject.GetName(fromClass),
					Text:   "(delete)",
				})

			default:
				return errors.Errorf("leaf node must have one of event_key, scenario_key, attribute_key, or is_delete: '%+v'", stmt)
			}

		case model_scenario.NODE_TYPE_SEQUENCE:
			err := addSteps(eventLookup, s, stmt.Statements, scenarioLookup, objectLookup, classLookup)
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

				err := addSteps(eventLookup, s, c.Statements, scenarioLookup, objectLookup, classLookup)
				if err != nil {
					return err
				}

				s.CloseSection()
			}

		case model_scenario.NODE_TYPE_LOOP:
			s.OpenSection("(Loop) ["+stmt.Loop+"]", "")

			err := addSteps(eventLookup, s, stmt.Statements, scenarioLookup, objectLookup, classLookup)
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
