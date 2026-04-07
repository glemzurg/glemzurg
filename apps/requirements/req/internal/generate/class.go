package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"

	"github.com/pkg/errors"
)

func generateClassMdContents(reqs *req_flat.Requirements, class model_class.Class, classesDiagram, stateDiagram string) (contents string, err error) {
	contents, err = generateFromTemplate(_classMdTemplate, struct {
		Reqs           *req_flat.Requirements
		Class          model_class.Class
		ClassesDiagram string
		StateDiagram   string
	}{
		Reqs:           reqs,
		Class:          class,
		ClassesDiagram: classesDiagram,
		StateDiagram:   stateDiagram,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

func generateClassStateMermaidContents(reqs *req_flat.Requirements, class model_class.Class) (contents string, err error) {
	eventNameLookup := map[string]string{}
	for _, event := range class.Events {
		eventNameLookup[event.Key.String()] = event.Name
	}
	guardDetailsLookup := map[string]string{}
	for _, guard := range class.Guards {
		guardDetailsLookup[guard.Key.String()] = guard.Logic.Description
	}
	actionNameLookup := map[string]string{}
	for _, action := range class.Actions {
		actionNameLookup[action.Key.String()] = action.Name
	}

	contents, err = generateFromTemplate(_classStateMermaidTemplate, struct {
		Reqs               *req_flat.Requirements
		Class              model_class.Class
		EventNameLookup    map[string]string
		GuardDetailsLookup map[string]string
		ActionNameLookup   map[string]string
	}{
		Reqs:               reqs,
		Class:              class,
		EventNameLookup:    eventNameLookup,
		GuardDetailsLookup: guardDetailsLookup,
		ActionNameLookup:   actionNameLookup,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

// generateClassesMermaidContents generates Mermaid class diagram markup.
func generateClassesMermaidContents(reqs *req_flat.Requirements, generalizations []model_class.Generalization, classes []model_class.Class, associations []model_class.Association) (contents string, err error) {
	contents, err = generateFromTemplate(_classesMermaidTemplate, struct {
		Reqs            *req_flat.Requirements
		Generalizations []model_class.Generalization
		Classes         []model_class.Class
		Associations    []model_class.Association
	}{
		Reqs:            reqs,
		Generalizations: generalizations,
		Classes:         classes,
		Associations:    associations,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}
