package generate

import (
	"path/filepath"

	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

	s := svgsequence.NewSequence()

	if len(scenario.Steps.Statements) == 0 {
		// No steps, so just add an informative plackard.
		s.AddStep(svgsequence.Step{Source: "Unknown", Target: "Unknown", Text: "No operations defined."})
	} else {
		for _, stmt := range scenario.Steps.Statements {
			source := "Unknown"
			if stmt.FromObject != nil {
				source = stmt.FromObject.Name
			}
			target := "Unknown"
			if stmt.ToObject != nil {
				target = stmt.ToObject.Name
			}
			text := stmt.Description
			if text == "" && stmt.Event != nil {
				text = stmt.Event.Name
			}
			s.AddStep(svgsequence.Step{
				Source: source,
				Target: target,
				Text:   text,
			})
		}
	}

	contents, err = s.Generate()

	return contents, nil
}
