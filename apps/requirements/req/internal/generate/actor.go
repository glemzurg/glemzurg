package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

func generateActorFiles(outputPath string, reqs requirements.Requirements) (err error) {

	// Get all the data we want for these files.
	actorLookup := reqs.ActorLookup()

	// Generate file for each actor.
	for _, actor := range actorLookup {

		// Generate model summary.
		modelFilename := convertKeyToFilename("actor", actor.Key, "", ".md")
		modelFilenameAbs := filepath.Join(outputPath, modelFilename)
		mdContents, err := generateActorMdContents(reqs, actor)
		if err != nil {
			return err
		}
		if err = writeFile(modelFilenameAbs, mdContents); err != nil {
			return err
		}
	}

	return nil
}

func generateActorMdContents(reqs requirements.Requirements, actor requirements.Actor) (contents string, err error) {

	contents, err = generateFromTemplate(_actorMdTemplate, struct {
		Reqs  requirements.Requirements
		Actor requirements.Actor
	}{
		Reqs:  reqs,
		Actor: actor,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
