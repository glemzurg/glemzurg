package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
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
	assert.Contains(t, body, unfinishedNotesMarker(model.UnfinishedNotes))
	assert.Contains(t, body, `class="unfinished-notes-glyph"`)

	product, ok := findClassByName(model, "Product")
	require.True(t, ok, "test model should include Product class")
	require.NotEmpty(t, product.UnfinishedNotes)

	productFile := convertKeyToFilename("class", product.Key.String(), "", ".md")
	productMD, ok := writer.md[productFile]
	require.True(t, ok, "expected class page for product (%s)", productFile)

	productBody := string(productMD)
	assert.Contains(t, productBody, unfinishedNotesBlock(product.UnfinishedNotes))
	assert.Contains(t, productBody, `class="unfinished-notes-glyph"`)

	require.NotNil(t, product.ActorKey)
	actor, ok := model.Actors[*product.ActorKey]
	require.True(t, ok, "product actor should exist in model")
	require.NotEmpty(t, actor.UnfinishedNotes)
	assert.Contains(t, productBody, unfinishedNotesMarker(actor.UnfinishedNotes),
		"actor bullet should show marker when linked actor has unfinished notes")
}

func findSingleSubdomainDomain(model core.Model) (model_domain.Domain, model_domain.Subdomain, bool) {
	for _, domain := range model.Domains {
		if len(domain.Subdomains) != 1 {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			return domain, subdomain, true
		}
	}
	return model_domain.Domain{}, model_domain.Subdomain{}, false
}

func TestGenerateSingleSubdomainUnfinishedNotesOnDomainPage(t *testing.T) {
	model := test_helper.GetTestModel()
	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	domain, subdomain, ok := findSingleSubdomainDomain(model)
	require.True(t, ok, "test model should include a single-subdomain domain")
	require.NotEmpty(t, subdomain.UnfinishedNotes)

	domainFile := convertKeyToFilename("domain", domain.Key.String(), "", ".md")
	domainMD, ok := writer.md[domainFile]
	require.True(t, ok, "expected domain page (%s)", domainFile)

	body := string(domainMD)
	assert.Contains(t, body, unfinishedNotesBlock(subdomain.UnfinishedNotes))
	assert.Contains(t, body, subdomain.Name)
}

func TestGenerateActorGenParticipantUnfinishedNotesMarkers(t *testing.T) {
	model := test_helper.GetTestModel()
	require.NotEmpty(t, model.ActorGeneralizations)

	var participantNotes string
	for _, actor := range model.Actors {
		if actor.Name == "Another Customer" {
			require.NotEmpty(t, actor.UnfinishedNotes)
			participantNotes = actor.UnfinishedNotes
			break
		}
	}
	require.NotEmpty(t, participantNotes)

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	body := string(writer.md["model.md"])
	assert.Contains(t, body, unfinishedNotesMarker(participantNotes))
}
