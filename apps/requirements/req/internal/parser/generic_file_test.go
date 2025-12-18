package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_GENERIC_PATH_OK  = "test_files/generic"
	t_GENERIC_PATH_ERR = t_GENERIC_PATH_OK + "/err"
)

func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileSuite))
}

type FileSuite struct {
	suite.Suite
}

func (suite *FileSuite) TestParseFiles() {

	testDataFiles, err := t_ContentsForAllMdFiles(t_GENERIC_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual File

		actual, err := parseFile(testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateFileContent(actual.Markdown, actual.UmlComment, actual.Data)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}

func (suite *FileSuite) TestParseFilesErr() {

	testDataFiles, err := t_ContentsForAllMdFiles(t_GENERIC_PATH_ERR)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename

		// For errors the "JSON" is really just an error text string.
		errstr := testData.Json

		actual, err := parseFile(testData.Filename, testData.Contents)
		assert.ErrorContains(suite.T(), err, errstr, testName)
		assert.Empty(suite.T(), actual, testName)
	}
}

func (suite *FileSuite) TestNew() {
	tests := []struct {
		markdown string
		title    string
	}{
		{
			markdown: `#    A Title 1`,
			title:    "A Title 1",
		},

		{
			markdown: `#A Title 1`,
			title:    "A Title 1",
		},

		{
			markdown: `
			
			##    A Title 1  

			And other content.
			`,
			title: "A Title 1",
		},

		{
			markdown: `A Title 1  

			And other content.
			`,
			title: "",
		},

		{
			markdown: ``,
			title:    "",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		title := extractMarkdownTitle(test.markdown)
		assert.Equal(suite.T(), test.title, title, testName)
	}
}
