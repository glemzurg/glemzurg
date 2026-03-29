package parser_human

import (
	"encoding/json"
	"fmt"
	"testing"

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
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual File

		actual, err := parseFile(testData.Filename, testData.Contents)
		suite.Require().NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.Require().NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateFileContent(actual.Markdown, actual.UmlComment, actual.Data)
		suite.Equal(testData.Contents, generated, testName)
	}
}

func (suite *FileSuite) TestParseFilesErr() {
	testDataFiles, err := t_ContentsForAllMdFiles(t_GENERIC_PATH_ERR)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename

		// For errors the "JSON" is really just an error text string.
		errstr := testData.Json

		actual, err := parseFile(testData.Filename, testData.Contents)
		suite.Require().ErrorContains(err, errstr, testName)
		suite.Empty(actual, testName)
	}
}

func (suite *FileSuite) TestPrependMarkdownTitle() {
	tests := []struct {
		testName string
		title    string
		markdown string
		expected string
	}{
		{
			testName: "empty markdown",
			title:    "My Title",
			markdown: "",
			expected: "# My Title",
		},
		{
			testName: "plain text details",
			title:    "My Title",
			markdown: "Some details here.",
			expected: "# My Title\n\nSome details here.",
		},
		{
			testName: "details starting with section heading",
			title:    "Bet Limit",
			markdown: "## Cross-Entity Constraints\n\nSome constraints.",
			expected: "# Bet Limit\n\n## Cross-Entity Constraints\n\nSome constraints.",
		},
		{
			testName: "details starting with h1 heading",
			title:    "My Title",
			markdown: "# Existing Heading\n\nContent.",
			expected: "# My Title\n\n# Existing Heading\n\nContent.",
		},
	}
	for _, test := range tests {
		suite.Run(test.testName, func() {
			result := prependMarkdownTitle(test.title, test.markdown)
			suite.Equal(test.expected, result)
		})
	}
}

func (suite *FileSuite) TestPrependMarkdownSubtitle() {
	tests := []struct {
		testName string
		title    string
		markdown string
		expected string
	}{
		{
			testName: "empty markdown",
			title:    "My Subtitle",
			markdown: "",
			expected: "## My Subtitle",
		},
		{
			testName: "details starting with section heading",
			title:    "My Subtitle",
			markdown: "### Section\n\nContent.",
			expected: "## My Subtitle\n\n### Section\n\nContent.",
		},
	}
	for _, test := range tests {
		suite.Run(test.testName, func() {
			result := prependMarkdownSubtitle(test.title, test.markdown)
			suite.Equal(test.expected, result)
		})
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
		suite.Equal(test.title, title, testName)
	}
}
