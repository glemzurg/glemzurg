package parser_ai

import (
	"encoding/json"
	"testing"

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

			actual, err := parseModel([]byte(testData.InputJSON), testData.Filename)
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
			_, err := parseModel([]byte(testData.InputJSON), testData.Filename)
			assert.NotNil(t, err, testName+" should return an error")

			// Verify it's a ParseError with the expected values.
			parseErr, ok := err.(*ParseError)
			assert.True(t, ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			assert.Equal(t, expected.Code, parseErr.Code, testName+" error code")

			// Test error file name separately from message content.
			assert.Equal(t, expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			// Test message string explicitly.
			// For dynamic messages (like schema validation), use MessagePrefix to match the start.
			if expected.Message != "" {
				assert.Equal(t, expected.Message, parseErr.Message, testName+" error message")
			} else if expected.MessagePrefix != "" {
				assert.True(t, len(parseErr.Message) >= len(expected.MessagePrefix) &&
					parseErr.Message[:len(expected.MessagePrefix)] == expected.MessagePrefix,
					testName+" error message should start with '"+expected.MessagePrefix+"', got '"+parseErr.Message+"'")
			}

			// Check schema content presence
			if expected.HasSchema {
				assert.NotEmpty(t, parseErr.Schema, testName+" should have schema content")
			} else {
				assert.Empty(t, parseErr.Schema, testName+" should not have schema content")
			}

			// Docs are always attached to all errors
			assert.NotEmpty(t, parseErr.Docs, testName+" should have docs content")

			// File is always set to the input filename
			assert.Equal(t, testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				assert.Equal(t, expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
