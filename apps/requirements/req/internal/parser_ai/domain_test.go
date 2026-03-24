package parser_ai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	t_DOMAIN_PATH_OK  = "test_files/domain"
	t_DOMAIN_PATH_ERR = t_DOMAIN_PATH_OK + "/err"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

func (suite *DomainSuite) TestParseDomainFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_DOMAIN_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			var expected inputDomain

			actual, err := parseDomain([]byte(testData.InputJSON), testData.Filename)
			suite.Require().NoError(err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			suite.Require().NoError(err, testName)

			suite.Equal(expected, *actual, testName)
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func (suite *DomainSuite) TestParseDomainErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_DOMAIN_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	// If there are no error test files, skip this test.
	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.Run(testName, func() {
			_, err := parseDomain([]byte(testData.InputJSON), testData.Filename)
			suite.Require().Error(err, testName+" should return an error")

			// Verify it's a ParseError with the expected values.
			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			suite.Equal(expected.Code, parseErr.Code, testName+" error code")

			// Test error file name separately from message content.
			suite.Equal(expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			// Test message string explicitly.
			// For dynamic messages (like schema validation), use MessagePrefix to match the start.
			if expected.Message != "" {
				suite.Equal(expected.Message, parseErr.Message, testName+" error message")
			} else if expected.MessagePrefix != "" {
				suite.True(len(parseErr.Message) >= len(expected.MessagePrefix) &&
					parseErr.Message[:len(expected.MessagePrefix)] == expected.MessagePrefix,
					testName+" error message should start with '"+expected.MessagePrefix+"', got '"+parseErr.Message+"'")
			}

			// Docs are always attached to all errors

			// File is always set to the input filename
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
