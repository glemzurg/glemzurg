package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

const (
	t_GENERALIZATION_PATH_OK  = "test_files/generalization"
	t_GENERALIZATION_PATH_ERR = t_GENERALIZATION_PATH_OK + "/err"
)

func TestGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(GeneralizationFileSuite))
}

type GeneralizationFileSuite struct {
	suite.Suite
}

func (suite *GeneralizationFileSuite) TestParseGeneralizationFiles() {
	// Create a parent subdomain key for testing.
	domainKey, err := identity.NewDomainKey("test_domain")
	suite.Require().NoError(err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "test_subdomain")
	suite.Require().NoError(err)

	generalizationSubKey := "generalization_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_GENERALIZATION_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_class.Generalization

		actual, err := parseClassGeneralization(subdomainKey, generalizationSubKey, testData.Filename, testData.Contents)
		suite.NoError(err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		suite.NoError(err, testName)

		suite.Equal(expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateGeneralizationContent(actual)
		suite.Equal(testData.Contents, generated, testName)
	}
}
