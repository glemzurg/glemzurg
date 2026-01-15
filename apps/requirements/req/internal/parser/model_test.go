package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"

	"github.com/stretchr/testify/assert"
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
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual req_model.Model

		actual, err := parseModel(key, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateModelContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
