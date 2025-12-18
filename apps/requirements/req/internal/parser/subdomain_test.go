package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_SUBDOMAIN_PATH_OK  = "test_files/subdomain"
	t_SUBDOMAIN_PATH_ERR = t_SUBDOMAIN_PATH_OK + "/err"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainFileSuite))
}

type SubdomainFileSuite struct {
	suite.Suite
}

func (suite *SubdomainFileSuite) TestParseSubdomainFiles() {

	key := "subdomain_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_SUBDOMAIN_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual requirements.Subdomain

		actual, err := parseSubdomain(key, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateSubdomainContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
