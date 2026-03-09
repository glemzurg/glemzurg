package parser_ai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	t_ACTOR_GENERALIZATION_PATH_OK  = "test_files/actor_generalization"
	t_ACTOR_GENERALIZATION_PATH_ERR = t_ACTOR_GENERALIZATION_PATH_OK + "/err"
)

func TestActorGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(ActorGeneralizationSuite))
}

type ActorGeneralizationSuite struct {
	suite.Suite
}

func (suite *ActorGeneralizationSuite) TestParseActorGeneralizationFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_ACTOR_GENERALIZATION_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			var expected inputActorGeneralization

			actual, err := parseActorGeneralization([]byte(testData.InputJSON), testData.Filename)
			suite.Require().NoError(err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			suite.Require().NoError(err, testName)

			suite.Equal(expected, *actual, testName)
		})
		if !pass {
			break
		}
	}
}

func (suite *ActorGeneralizationSuite) TestParseActorGeneralizationErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_ACTOR_GENERALIZATION_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.Run(testName, func() {
			_, err := parseActorGeneralization([]byte(testData.InputJSON), testData.Filename)
			suite.Require().Error(err, testName+" should return an error")

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
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
