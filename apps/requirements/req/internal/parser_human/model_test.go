package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"

	"github.com/stretchr/testify/suite"
)

const (
	t_MODEL_PATH_OK  = "test_files/model"
	t_MODEL_PATH_ERR = t_MODEL_PATH_OK + "/err"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelFileSuite))
}

type ModelFileSuite struct {
	suite.Suite
}

func (suite *ModelFileSuite) TestParseModelFiles() {
	key := "model_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_MODEL_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual core.Model

		actual, err := parseModel(key, testData.Filename, testData.Contents)
		suite.Require().NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.Require().NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateModelContent(actual)
		suite.Equal(testData.Contents, generated, testName)
	}
}
