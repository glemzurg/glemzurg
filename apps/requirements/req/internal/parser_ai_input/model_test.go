package parser_ai_input

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_MODEL_PATH_OK  = "test_files/model"
	t_MODEL_PATH_ERR = t_MODEL_PATH_OK + "/err"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) TestParseModelFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_MODEL_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.T().Run(testName, func(t *testing.T) {
			var expected inputModel

			actual, err := parseModel([]byte(testData.InputJSON))
			assert.Nil(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			assert.Nil(t, err, testName)

			assert.Equal(t, expected, *actual, testName)
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func (suite *ModelSuite) TestParseModelErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_MODEL_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	// If there are no error test files, skip this test.
	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.T().Run(testName, func(t *testing.T) {
			_, err := parseModel([]byte(testData.InputJSON))
			assert.NotNil(t, err, testName+" should return an error")

			// Verify it's a ParseError with the expected code.
			parseErr, ok := err.(*errors.ParseError)
			assert.True(t, ok, testName+" should return a ParseError")
			if ok && testData.ErrorCode != 0 {
				assert.Equal(t, testData.ErrorCode, parseErr.Code, testName+" error code")
			}
			if ok && testData.ErrorField != "" {
				assert.Equal(t, testData.ErrorField, parseErr.Field, testName+" error field")
			}
		})
	}
}
