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
	t_CLASS_PATH_OK  = "test_files/class"
	t_CLASS_PATH_ERR = t_CLASS_PATH_OK + "/err"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

func (suite *ClassSuite) TestParseClassFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_CLASS_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			var expected inputClass

			actual, err := parseClass([]byte(testData.InputJSON), testData.Filename)
			suite.Require().NoError(err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			suite.Require().NoError(err, testName)

			suite.Equal(expected.Name, actual.Name, testName+" name")
			suite.Equal(expected.Details, actual.Details, testName+" details")
			suite.Equal(expected.ActorKey, actual.ActorKey, testName+" actor_key")
			suite.Equal(expected.UMLComment, actual.UMLComment, testName+" uml_comment")
			suite.Equal(expected.Indexes, actual.Indexes, testName+" indexes")

			// Compare attributes map
			suite.Len(actual.Attributes, len(expected.Attributes), testName+" attributes count")
			for key, expectedAttr := range expected.Attributes {
				actualAttr, exists := actual.Attributes[key]
				suite.True(exists, testName+" attribute '"+key+"' should exist")
				if exists {
					suite.Equal(expectedAttr.Name, actualAttr.Name, testName+" attribute '"+key+"' name")
					suite.Equal(expectedAttr.DataTypeRules, actualAttr.DataTypeRules, testName+" attribute '"+key+"' data_type_rules")
					suite.Equal(expectedAttr.Details, actualAttr.Details, testName+" attribute '"+key+"' details")
					suite.Equal(expectedAttr.DerivationPolicy, actualAttr.DerivationPolicy, testName+" attribute '"+key+"' derivation_policy")
					suite.Equal(expectedAttr.Nullable, actualAttr.Nullable, testName+" attribute '"+key+"' nullable")
					suite.Equal(expectedAttr.UMLComment, actualAttr.UMLComment, testName+" attribute '"+key+"' uml_comment")
				}
			}
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func (suite *ClassSuite) TestParseClassErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_CLASS_PATH_ERR)
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
			t := suite.T()
			_, err := parseClass([]byte(testData.InputJSON), testData.Filename)
			require.Error(t, err, testName+" should return an error")

			// Verify it's a ParseError with the expected values.
			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			assert.True(t, ok, testName+" should return a ParseError")
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
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
