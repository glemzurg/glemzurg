package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	t_DOMAIN_PATH_OK  = "test_files/domain"
	t_DOMAIN_PATH_ERR = t_DOMAIN_PATH_OK + "/err"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainFileSuite))
}

type DomainFileSuite struct {
	suite.Suite
}

func (suite *DomainFileSuite) TestParseDomainFiles() {

	key := "domain_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_DOMAIN_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_domain.Domain

		actual, err := parseDomain(key, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateDomainContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
