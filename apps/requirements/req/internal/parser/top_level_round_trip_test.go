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

	// Build the model tree. For now, only the model-level fields.
	// The parser initializes these maps when reading, so we must match.
	input := req_model.Model{
		Key:                   "test_model",
		Name:                  "Test Model",
		Details:               "# Test Model\n\nTest model details in markdown.",
		Actors:                map[identity.Key]model_actor.Actor{},
		ActorGeneralizations:  map[identity.Key]model_actor.Generalization{},
		Domains:               map[identity.Key]model_domain.Domain{},
		DomainAssociations:    map[identity.Key]model_domain.Association{},
		ClassAssociations:     map[identity.Key]model_class.Association{},
	}

	// Validate the model before writing.
	err := input.Validate()
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

func (suite *RoundTripSuite) TestRoundTripWithActorGeneralizations() {

	// Build actor generalization keys.
	genKeyA, err := identity.NewActorGeneralizationKey("gen_a")
	assert.Nil(suite.T(), err)
	genKeyB, err := identity.NewActorGeneralizationKey("gen_b")
	assert.Nil(suite.T(), err)

	genA, err := model_actor.NewGeneralization(genKeyA, "Generalization A", "## Generalization A\n\nDetails for gen A.", false, true, "")
	assert.Nil(suite.T(), err)
	genB, err := model_actor.NewGeneralization(genKeyB, "Generalization B", "## Generalization B\n\nDetails for gen B.", true, false, "uml comment for gen B")
	assert.Nil(suite.T(), err)

	input := req_model.Model{
		Key:     "test_model",
		Name:    "Test Model",
		Details: "# Test Model\n\nTest model details.",
		Actors:  map[identity.Key]model_actor.Actor{},
		ActorGeneralizations: map[identity.Key]model_actor.Generalization{
			genKeyA: genA,
			genKeyB: genB,
		},
		Domains:            map[identity.Key]model_domain.Domain{},
		DomainAssociations: map[identity.Key]model_domain.Association{},
		ClassAssociations:  map[identity.Key]model_class.Association{},
	}

	err = input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	tempDir := suite.T().TempDir()
	err = Write(input, tempDir)
	assert.Nil(suite.T(), err, "writing model should succeed")

	output, err := Parse(tempDir)
	assert.Nil(suite.T(), err, "parsing model should succeed")

	output.Key = input.Key

	assert.Equal(suite.T(), input, output)
}
