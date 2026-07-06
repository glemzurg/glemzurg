package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
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
	testDataFiles, err := t_ContentsForAllMdFiles(t_SUBDOMAIN_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		domainKey, subdomainSubKey := subdomainTestKeys(testData.Filename)

		var expected model_domain.Subdomain
		actual, associations, err := parseSubdomain(domainKey, subdomainSubKey, testData.Filename, testData.Contents)
		suite.Require().NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.Require().NoError(err, testName)

		suite.Equal(expected, actual, testName)

		if testData.JsonChildren != "" {
			var expectedAssociations []model_domain.SubdomainAssociation
			err = json.Unmarshal([]byte(testData.JsonChildren), &expectedAssociations)
			suite.Require().NoError(err, testName+" associations json")
			suite.Equal(expectedAssociations, associations, testName+" associations")
		} else {
			suite.Empty(associations, testName)
		}

		generated := generateSubdomainContent(actual, associations)
		suite.Equal(testData.Contents, generated, testName)
	}
}

func subdomainTestKeys(filename string) (domainKey identity.Key, subdomainSubKey string) {
	domainKey = helper.Must(identity.NewDomainKey("test_domain"))
	if filename == "test_files/subdomain/02_associations.md" {
		return domainKey, "billing"
	}
	return domainKey, "subdomain_key"
}