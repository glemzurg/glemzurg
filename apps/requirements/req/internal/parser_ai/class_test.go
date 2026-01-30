package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.T().Run(testName, func(t *testing.T) {
			var expected inputClass

			actual, err := parseClass([]byte(testData.InputJSON), testData.Filename)
			assert.Nil(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			assert.Nil(t, err, testName)

			assert.Equal(t, expected.Name, actual.Name, testName+" name")
			assert.Equal(t, expected.Details, actual.Details, testName+" details")
			assert.Equal(t, expected.ActorKey, actual.ActorKey, testName+" actor_key")
			assert.Equal(t, expected.UMLComment, actual.UMLComment, testName+" uml_comment")
			assert.Equal(t, expected.Indexes, actual.Indexes, testName+" indexes")

			// Compare attributes map
			assert.Equal(t, len(expected.Attributes), len(actual.Attributes), testName+" attributes count")
			for key, expectedAttr := range expected.Attributes {
				actualAttr, exists := actual.Attributes[key]
				assert.True(t, exists, testName+" attribute '"+key+"' should exist")
				if exists {
					assert.Equal(t, expectedAttr.Name, actualAttr.Name, testName+" attribute '"+key+"' name")
					assert.Equal(t, expectedAttr.DataTypeRules, actualAttr.DataTypeRules, testName+" attribute '"+key+"' data_type_rules")
					assert.Equal(t, expectedAttr.Details, actualAttr.Details, testName+" attribute '"+key+"' details")
					assert.Equal(t, expectedAttr.DerivationPolicy, actualAttr.DerivationPolicy, testName+" attribute '"+key+"' derivation_policy")
					assert.Equal(t, expectedAttr.Nullable, actualAttr.Nullable, testName+" attribute '"+key+"' nullable")
					assert.Equal(t, expectedAttr.UMLComment, actualAttr.UMLComment, testName+" attribute '"+key+"' uml_comment")
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
		suite.T().Run(testName, func(t *testing.T) {
			_, err := parseClass([]byte(testData.InputJSON), testData.Filename)
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
