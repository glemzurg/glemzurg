package test_helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTestModel(t *testing.T) {
	// Should not panic and should return a valid model.
	model := GetTestModel()

	assert.Equal(t, "test_model", model.Key)
	assert.Equal(t, "Test Model", model.Name)

	// Verify all top-level collections are populated.
	assert.NotEmpty(t, model.Actors, "should have actors")
	assert.NotEmpty(t, model.ActorGeneralizations, "should have actor generalizations")
	assert.NotEmpty(t, model.Domains, "should have domains")
	assert.NotEmpty(t, model.DomainAssociations, "should have domain associations")
	assert.NotEmpty(t, model.Invariants, "should have invariants")
	assert.NotEmpty(t, model.GlobalFunctions, "should have global functions")

	// Verify the model validates.
	err := model.Validate()
	assert.Nil(t, err)
}

func TestGetStrictTestModel(t *testing.T) {
	// Should not panic and should return a valid model.
	model := GetStrictTestModel()

	assert.Equal(t, "test_model", model.Key)
	assert.Equal(t, "Test Model", model.Name)

	// Verify all top-level collections are populated.
	assert.NotEmpty(t, model.Actors, "should have actors")
	assert.NotEmpty(t, model.ActorGeneralizations, "should have actor generalizations")
	assert.NotEmpty(t, model.Domains, "should have domains")
	assert.NotEmpty(t, model.DomainAssociations, "should have domain associations")
	assert.NotEmpty(t, model.Invariants, "should have invariants")
	assert.NotEmpty(t, model.GlobalFunctions, "should have global functions")

	// Verify the model validates.
	err := model.Validate()
	assert.Nil(t, err)
}
