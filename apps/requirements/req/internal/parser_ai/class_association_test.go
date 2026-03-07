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
	t_ASSOCIATION_PATH_OK  = "test_files/association"
	t_ASSOCIATION_PATH_ERR = t_ASSOCIATION_PATH_OK + "/err"
)

func TestAssociationSuite(t *testing.T) {
	suite.Run(t, new(AssociationSuite))
}

type AssociationSuite struct {
	suite.Suite
}

func (suite *AssociationSuite) TestParseAssociationFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_ASSOCIATION_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			t := suite.T()
			var expected inputClassAssociation

			actual, err := parseAssociation([]byte(testData.InputJSON), testData.Filename)
			require.NoError(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			require.NoError(t, err, testName)

			suite.Equal(expected.Name, actual.Name, testName+" name")
			suite.Equal(expected.Details, actual.Details, testName+" details")
			suite.Equal(expected.FromClassKey, actual.FromClassKey, testName+" from_class_key")
			suite.Equal(expected.FromMultiplicity, actual.FromMultiplicity, testName+" from_multiplicity")
			suite.Equal(expected.ToClassKey, actual.ToClassKey, testName+" to_class_key")
			suite.Equal(expected.ToMultiplicity, actual.ToMultiplicity, testName+" to_multiplicity")
			suite.Equal(expected.AssociationClassKey, actual.AssociationClassKey, testName+" association_class_key")
			suite.Equal(expected.UmlComment, actual.UmlComment, testName+" uml_comment")
		})
		if !pass {
			break
		}
	}
}

func (suite *AssociationSuite) TestParseAssociationErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_ASSOCIATION_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.Run(testName, func() {
			t := suite.T()
			_, err := parseAssociation([]byte(testData.InputJSON), testData.Filename)
			require.Error(t, err, testName+" should return an error")

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			assert.True(t, ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			suite.Equal(expected.Code, parseErr.Code, testName+" error code")
			suite.Equal(expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			if expected.Message != "" {
				suite.Equal(expected.Message, parseErr.Message, testName+" error message")
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
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
