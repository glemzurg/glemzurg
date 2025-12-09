package generate

import (
	"path/filepath"

	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

func generateScenarioSvgContents(reqs requirements.Requirements, scenario requirements.Scenario) (contents string, err error) {

	scenarioObjectLookup := reqs.ScenarioObjectLookup()

	s := svgsequence.NewSequence()

	// Add the actors in order.
	var actors []string
	if len(scenario.Objects) > 0 {
		for _, obj := range scenario.Objects {

			// Get fully populated object for proper name construction.
			object, found := scenarioObjectLookup[obj.Key]
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
		for _, stmt := range scenario.Steps.Statements {
			switch stmt.Inferredtype() {
			case requirements.NODE_TYPE_LEAF:

				switch {

				case stmt.AttributeKey != "", stmt.EventKey != "":

					fromObject, found := scenarioObjectLookup[stmt.FromObjectKey]
					if !found {
						return "", errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey)
					}
					toObject, found := scenarioObjectLookup[stmt.ToObjectKey]
					if !found {
						return "", errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey)
					}

					text := stmt.Description

					s.AddStep(svgsequence.Step{
						Source: fromObject.GetName(),
						Target: toObject.GetName(),
						Text:   text,
					})

				default:
					return "", errors.Errorf("leaf node must have one of event_key, scenario_key, or attribute_key: '%+v'", stmt)
				}

			case requirements.NODE_TYPE_SWITCH:

				for _, c := range stmt.Cases {
					s.OpenSection("Opt ["+c.Condition+"]", "")

					// ... statement handling ...
					s.AddStep(svgsequence.Step{
						Source: "Maria", Target: "Maria",
						Text:  "*Thinks*\nLong time no see...",
						Color: "#36bbbbff",
					})

					s.CloseSection()
				}

			default:
				return "", errors.Errorf("unsupported node type in scenario SVG generation: '%s'", stmt.Inferredtype())
			}
		}
	}

	contents, err = s.Generate()

	return contents, nil
}
