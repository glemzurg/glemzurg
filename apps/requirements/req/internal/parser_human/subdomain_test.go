package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	// Create a parent domain key for testing.
	domainKey, err := identity.NewDomainKey("test_domain")
	suite.Require().NoError(err)

	subdomainSubKey := "subdomain_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_SUBDOMAIN_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_domain.Subdomain

		actual, err := parseSubdomain(domainKey, subdomainSubKey, testData.Filename, testData.Contents)
		suite.Require().NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.Require().NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateSubdomainContent(actual)
		suite.Equal(testData.Contents, generated, testName)
	}
}
