package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

func generateClassMdContents(reqs *req_flat.Requirements, class model_class.Class) (contents string, err error) {

	// Create the lookups of keys to meaningful values.

	contents, err = generateFromTemplate(_classMdTemplate, struct {
		Reqs  *req_flat.Requirements
		Class model_class.Class
	}{
		Reqs:  reqs,
		Class: class,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

func generateClassStateSvgContents(reqs *req_flat.Requirements, class model_class.Class) (svgContents string, dotContents string, err error) {

	// Create the lookups of keys to meaningful values.
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

	dotContents, err = generateFromTemplate(_classStateDotTemplate, struct {
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
		return "", "", errors.WithStack(err)
	}

	svgContents, err = graphvizDotToSvg(dotContents)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return svgContents, dotContents, nil
}

// This is the class graph on a domain and class pages.
func generateClassesSvgContents(reqs *req_flat.Requirements, generalizations []model_class.Generalization, classes []model_class.Class, associations []model_class.Association) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_classesDotTemplate, struct {
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
		return "", "", errors.WithStack(err)
	}

	svgContents, err = graphvizDotToSvg(dotContents)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return svgContents, dotContents, nil
}
