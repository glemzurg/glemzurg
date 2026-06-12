package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findClassByName(model core.Model, name string) (model_class.Class, bool) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if class.Name == name {
					return class, true
				}
			}
		}
	}
	return model_class.Class{}, false
}

func TestGenerateUnfinishedNotesInMarkdown(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	modelMD, ok := writer.md["model.md"]
	require.True(t, ok, "expected model.md")

	body := string(modelMD)
	require.NotEmpty(t, model.UnfinishedNotes)
	assert.Contains(t, body, unfinishedNotesBlock(model.UnfinishedNotes))
	assert.Contains(t, body, `class="unfinished-notes-marker"`)
	assert.Contains(t, body, _unfinishedNotesGlyph)

	product, ok := findClassByName(model, "Product")
	require.True(t, ok, "test model should include Product class")
	require.NotEmpty(t, product.UnfinishedNotes)

	productFile := convertKeyToFilename("class", product.Key.String(), "", ".md")
	productMD, ok := writer.md[productFile]
	require.True(t, ok, "expected class page for product (%s)", productFile)

	productBody := string(productMD)
	assert.Contains(t, productBody, unfinishedNotesBlock(product.UnfinishedNotes))
	assert.Contains(t, productBody, `unfinished-notes-marker`)
	assert.Contains(t, productBody, _unfinishedNotesGlyph)

	require.NotNil(t, product.ActorKey)
	actor, ok := model.Actors[*product.ActorKey]
	require.True(t, ok, "product actor should exist in model")
	require.NotEmpty(t, actor.UnfinishedNotes)
	assert.Contains(t, productBody, unfinishedNotesMarker(actor.UnfinishedNotes),
		"actor bullet should show marker when linked actor has unfinished notes")
}
