package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model"
)

func parseModel(key, filename, contents string) (parsedModel model.Model, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model.Model{}, err
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

	parsedModel, err = model.NewModel(key, parsedFile.Title, markdown)
	if err != nil {
		return model.Model{}, err
	}

	return parsedModel, nil
}

func generateModelContent(m model.Model) string {
	return generateFileContent(m.Details, "", "")
}
