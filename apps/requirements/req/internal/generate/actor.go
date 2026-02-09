package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
)

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
