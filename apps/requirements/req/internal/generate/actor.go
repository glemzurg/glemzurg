package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
)

func generateActorFiles(outputPath string, reqs *req_flat.Requirements) (err error) {

	// Get all the data we want for these files.
	actorLookup := reqs.ActorLookup()

	// Generate file for each actor.
	for _, actor := range actorLookup {

		// Generate model summary.
		modelFilename := convertKeyToFilename("actor", actor.Key.String(), "", ".md")
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

func generateActorMdContents(reqs *req_flat.Requirements, actor model_actor.Actor) (contents string, err error) {

	contents, err = generateFromTemplate(_actorMdTemplate, struct {
		Reqs  *req_flat.Requirements
		Actor model_actor.Actor
	}{
		Reqs:  reqs,
		Actor: actor,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
