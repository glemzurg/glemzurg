package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRoundTripSuite(t *testing.T) {
	suite.Run(t, new(RoundTripSuite))
}

type RoundTripSuite struct {
	suite.Suite
}

func (suite *RoundTripSuite) TestRoundTrip() {

	// -- Actor generalizations --
	genKeyA, err := identity.NewActorGeneralizationKey("gen_a")
	assert.Nil(suite.T(), err)
	genKeyB, err := identity.NewActorGeneralizationKey("gen_b")
	assert.Nil(suite.T(), err)

	genA, err := model_actor.NewGeneralization(genKeyA, "Generalization A", "## Generalization A\n\nDetails for gen A.", false, true, "")
	assert.Nil(suite.T(), err)
	genB, err := model_actor.NewGeneralization(genKeyB, "Generalization B", "## Generalization B\n\nDetails for gen B.", true, false, "uml comment for gen B")
	assert.Nil(suite.T(), err)

	// -- Actors --
	actorKeyA, err := identity.NewActorKey("alice")
	assert.Nil(suite.T(), err)
	actorKeyB, err := identity.NewActorKey("bob")
	assert.Nil(suite.T(), err)

	// Alice is the superclass of gen_a.
	actorA, err := model_actor.NewActor(actorKeyA, "Alice", "# Alice\n\nA person actor.", "person", &genKeyA, nil, "")
	assert.Nil(suite.T(), err)
	// Bob is the subclass of gen_a, and superclass of gen_b.
	actorB, err := model_actor.NewActor(actorKeyB, "Bob", "# Bob\n\nA system actor.", "system", &genKeyB, &genKeyA, "uml comment for bob")
	assert.Nil(suite.T(), err)

	// -- Model --
	input := req_model.Model{
		Key:     "test_model",
		Name:    "Test Model",
		Details: "# Test Model\n\nTest model details in markdown.",
		Actors: map[identity.Key]model_actor.Actor{
			actorKeyA: actorA,
			actorKeyB: actorB,
		},
		ActorGeneralizations: map[identity.Key]model_actor.Generalization{
			genKeyA: genA,
			genKeyB: genB,
		},
		Domains:            map[identity.Key]model_domain.Domain{},
		DomainAssociations: map[identity.Key]model_domain.Association{},
		ClassAssociations:  map[identity.Key]model_class.Association{},
	}

	// Validate the model before writing.
	err = input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Write to a temporary folder.
	tempDir := suite.T().TempDir()
	err = Write(input, tempDir)
	assert.Nil(suite.T(), err, "writing model should succeed")

	// Read from the temporary folder.
	output, err := Parse(tempDir)
	assert.Nil(suite.T(), err, "parsing model should succeed")

	// The parsed model's Key will be the tempDir path, not our original key.
	// Overwrite it for comparison since the parser uses the modelPath as the key.
	output.Key = input.Key

	// Compare the model values.
	assert.Equal(suite.T(), input, output)
}
