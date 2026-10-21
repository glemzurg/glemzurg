// Package parser_json reads a requirements model stored as one JSON file per entity.
// Files ending in .annot.json are not model data. Callers that understand a
// particular annotation dialect read those files themselves.
package parser_json

import (
	ijson "github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_json"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

// ReadModel loads the model directory written by the requirements JSON writer.
func ReadModel(inputModelPath string) (core.Model, error) {
	return ijson.ReadModel(inputModelPath)
}
