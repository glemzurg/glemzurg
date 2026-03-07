package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/stretchr/testify/suite"
)

const (
	t_ACTOR_PATH_OK  = "test_files/actor"
	t_ACTOR_PATH_ERR = t_ACTOR_PATH_OK + "/err"
)

func TestActorSuite(t *testing.T) {
	suite.Run(t, new(ActorFileSuite))
}

type ActorFileSuite struct {
	suite.Suite
}

func (suite *ActorFileSuite) TestParseActorFiles() {
	key := "actor_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_ACTOR_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_actor.Actor

		actual, err := parseActor(key, testData.Filename, testData.Contents)
		suite.Require().NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.Require().NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateActorContent(actual)
		suite.Equal(testData.Contents, generated, testName)
	}
}
