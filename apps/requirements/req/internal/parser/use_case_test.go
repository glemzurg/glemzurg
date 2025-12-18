package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_USE_CASE_PATH_OK  = "test_files/use_case"
	t_USE_CASE_PATH_ERR = t_USE_CASE_PATH_OK + "/err"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseFileSuite))
}

type UseCaseFileSuite struct {
	suite.Suite
}

func (suite *UseCaseFileSuite) TestParseUseCaseFiles() {

	key := "use_case_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_USE_CASE_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual requirements.UseCase

		actual, err := parseUseCase(key, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateUseCaseContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
