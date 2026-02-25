package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_LOGIC_PATH_OK  = "test_files/logic"
	t_LOGIC_PATH_ERR = t_LOGIC_PATH_OK + "/err"
)

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
}

func (suite *LogicSuite) TestParseLogicFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_LOGIC_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.T().Run(testName, func(t *testing.T) {
			var expected inputLogic

			actual, err := parseLogic([]byte(testData.InputJSON), testData.Filename)
			assert.Nil(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			assert.Nil(t, err, testName)

			assert.Equal(t, expected, *actual, testName)
		})
		if !pass {
			break
		}
	}
}

func (suite *LogicSuite) TestParseLogicErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_LOGIC_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.T().Run(testName, func(t *testing.T) {
			_, err := parseLogic([]byte(testData.InputJSON), testData.Filename)
			assert.NotNil(t, err, testName+" should return an error")

			parseErr, ok := err.(*ParseError)
			assert.True(t, ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			assert.Equal(t, expected.Code, parseErr.Code, testName+" error code")
			assert.Equal(t, expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			if expected.Message != "" {
				assert.Equal(t, expected.Message, parseErr.Message, testName+" error message")
			} else if expected.MessagePrefix != "" {
				assert.True(t, len(parseErr.Message) >= len(expected.MessagePrefix) &&
					parseErr.Message[:len(expected.MessagePrefix)] == expected.MessagePrefix,
					testName+" error message should start with '"+expected.MessagePrefix+"', got '"+parseErr.Message+"'")
			}

			if expected.HasSchema {
				assert.NotEmpty(t, parseErr.Schema, testName+" should have schema content")
			} else {
				assert.Empty(t, parseErr.Schema, testName+" should not have schema content")
			}

			assert.NotEmpty(t, parseErr.Docs, testName+" should have docs content")
			assert.Equal(t, testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				assert.Equal(t, expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
