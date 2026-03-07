package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
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
	suite.NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_domain.Domain

		actual, associations, err := parseDomain(key, testData.Filename, testData.Contents)
		suite.NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test associations if expected data exists (via _children.json file).
		if testData.JsonChildren != "" {
			var expectedAssociations []model_domain.Association
			err = json.Unmarshal([]byte(testData.JsonChildren), &expectedAssociations)
			suite.NoError(err, testName+" associations json")
			suite.Equal(expectedAssociations, associations, testName+" associations")
		}

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateDomainContent(actual, associations)
		suite.Equal(testData.Contents, generated, testName)
	}
}
