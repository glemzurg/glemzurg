package parser_ai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	t_ACTION_PATH_OK  = "test_files/action"
	t_ACTION_PATH_ERR = t_ACTION_PATH_OK + "/err"
)

func TestActionSuite(t *testing.T) {
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
}

func (suite *ActionSuite) TestParseActionFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_ACTION_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			t := suite.T() //nolint:testifylint // captures subtest result
			var expected inputAction

			actual, err := parseAction([]byte(testData.InputJSON), testData.Filename)
			require.NoError(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			require.NoError(t, err, testName)

			assert.Equal(t, expected.Name, actual.Name, testName+" name")
			assert.Equal(t, expected.Details, actual.Details, testName+" details")
			assert.Equal(t, expected.Requires, actual.Requires, testName+" requires")
			assert.Equal(t, expected.Guarantees, actual.Guarantees, testName+" guarantees")
		})
		if !pass {
			break
		}
	}
}

func (suite *ActionSuite) TestParseActionErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_ACTION_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.Run(testName, func() {
			_, err := parseAction([]byte(testData.InputJSON), testData.Filename)
			suite.Error(err, testName+" should return an error")

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			suite.Equal(expected.Code, parseErr.Code, testName+" error code")
			suite.Equal(expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			if expected.Message != "" {
				suite.Equal(expected.Message, parseErr.Message, testName+" error message")
			} else if expected.MessagePrefix != "" {
				suite.True(len(parseErr.Message) >= len(expected.MessagePrefix) &&
					parseErr.Message[:len(expected.MessagePrefix)] == expected.MessagePrefix,
					testName+" error message should start with '"+expected.MessagePrefix+"', got '"+parseErr.Message+"'")
			}

			if expected.HasSchema {
				suite.NotEmpty(parseErr.Schema, testName+" should have schema content")
			} else {
				suite.Empty(parseErr.Schema, testName+" should not have schema content")
			}

			suite.NotEmpty(parseErr.Docs, testName+" should have docs content")
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
