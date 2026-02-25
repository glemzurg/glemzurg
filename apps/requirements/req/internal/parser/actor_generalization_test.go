package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_ACTOR_GENERALIZATION_PATH_OK  = "test_files/actor_generalization"
	t_ACTOR_GENERALIZATION_PATH_ERR = t_ACTOR_GENERALIZATION_PATH_OK + "/err"
)

func TestActorGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(ActorGeneralizationFileSuite))
}

type ActorGeneralizationFileSuite struct {
	suite.Suite
}

func (suite *ActorGeneralizationFileSuite) TestParseActorGeneralizationFiles() {

	generalizationSubKey := "generalization_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_ACTOR_GENERALIZATION_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_actor.Generalization

		actual, err := parseActorGeneralization(generalizationSubKey, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateActorGeneralizationContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
