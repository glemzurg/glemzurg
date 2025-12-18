package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func parseModel(key, filename, contents string) (model requirements.Model, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return requirements.Model{}, err
	}

	// There is no uml comment for a "model" entity (it is not displayed),
	// and no data for a model entity. Just add everything to the markdown
	// so that everything typed makes it into the output.
	markdown := parsedFile.Markdown

	if parsedFile.UmlComment != "" {
		markdown += "\n\n" + parsedFile.UmlComment
	}

	if parsedFile.Data != "" {
		markdown += "\n\n" + parsedFile.Data
	}

	model, err = requirements.NewModel(key, parsedFile.Title, markdown)
	if err != nil {
		return requirements.Model{}, err
	}

	return model, nil
}

func generateModelContent(model requirements.Model) string {
	return generateFileContent(model.Details, "", "")
}
