package generate

import (
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

func generateClassFiles(debug bool, outputPath string, reqs requirements.Requirements) (err error) {

	// The data we're interested in.
	classLookup, _ := reqs.ClassLookup()

	// Generate class data.
	for _, class := range classLookup {

		// Generate class summary.
		classFilename := convertKeyToFilename("class", class.Key, "", ".md")
		classFilenameAbs := filepath.Join(outputPath, classFilename)
		classMdContents, err := generateClassMdContents(reqs, class)
		if err != nil {
			return err
		}
		if err = writeFile(classFilenameAbs, classMdContents); err != nil {
			return err
		}

		// Get the data that is important for this class diagram.
		generalizations, classes, associations := reqs.RegardingClasses([]requirements.Class{class})

		// Generate classes diagram.
		classesSvgFilename := convertKeyToFilename("class", class.Key, "", ".svg")
		classesSvgFilenameAbs := filepath.Join(outputPath, classesSvgFilename)
		classesSvgContents, classesDotContents, err := generateClassesSvgContents(reqs, generalizations, classes, associations)
		if err != nil {
			return err
		}
		if err = writeFile(classesSvgFilenameAbs, classesSvgContents); err != nil {
			return err
		}
		if err := debugWriteDotFile(debug, outputPath, classesSvgFilename, classesDotContents); err != nil {
			return err
		}

		// State Machine diagram.
		if len(class.States) > 0 {

			statesSvgFilename := convertKeyToFilename("class", class.Key, "states", ".svg")
			statesSvgFilenameAbs := filepath.Join(outputPath, statesSvgFilename)
			statesSvgContents, statesDotContents, err := generateClassStateSvgContents(reqs, class)
			if err != nil {
				return err
			}
			if err = writeFile(statesSvgFilenameAbs, statesSvgContents); err != nil {
				return err
			}
			if err := debugWriteDotFile(debug, outputPath, statesSvgFilename, statesDotContents); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateClassMdContents(reqs requirements.Requirements, class requirements.Class) (contents string, err error) {

	// Create the lookups of keys to meaningful values.

	contents, err = generateFromTemplate(_classMdTemplate, struct {
		Reqs  requirements.Requirements
		Class requirements.Class
	}{
		Reqs:  reqs,
		Class: class,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return contents, nil
}

func generateClassStateSvgContents(reqs requirements.Requirements, class requirements.Class) (svgContents string, dotContents string, err error) {

	// Create the lookups of keys to meaningful values.
	eventNameLookup := map[string]string{}
	for _, event := range class.Events {
		eventNameLookup[event.Key] = event.Name
	}
	guardDetailsLookup := map[string]string{}
	for _, guard := range class.Guards {
		guardDetailsLookup[guard.Key] = guard.Details
	}
	actionNameLookup := map[string]string{}
	for _, action := range class.Actions {
		actionNameLookup[action.Key] = action.Name
	}

	dotContents, err = generateFromTemplate(_classStateDotTemplate, struct {
		Reqs               requirements.Requirements
		Class              requirements.Class
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
func generateClassesSvgContents(reqs requirements.Requirements, generalizations []requirements.Generalization, classes []requirements.Class, associations []requirements.Association) (svgContents string, dotContents string, err error) {

	dotContents, err = generateFromTemplate(_classesDotTemplate, struct {
		Reqs            requirements.Requirements
		Generalizations []requirements.Generalization
		Classes         []requirements.Class
		Associations    []requirements.Association
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
