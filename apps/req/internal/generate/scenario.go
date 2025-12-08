package generate

import (
	"path/filepath"

	svgsequence "github.com/aorith/svg-sequence"
	"github.com/glemzurg/futz/apps/req/internal/requirements"
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

	s.AddStep(svgsequence.Step{Source: "Bob", Target: "Maria", Text: "Hi! How are you doing?"})
	s.OpenSection("response", "")
	s.AddStep(svgsequence.Step{
		Source: "Maria", Target: "Maria",
		Text:  "*Thinks*\nLong time no see...",
		Color: "#667777",
	})
	s.AddStep(svgsequence.Step{Source: "Maria", Target: "Bob", Text: "Fine!"})
	s.CloseSection()
	contents, err = s.Generate()

	return contents, nil
}
