package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(suite.T(), err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "test_subdomain")
	assert.Nil(suite.T(), err)

	generalizationSubKey := "generalization_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_GENERALIZATION_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual model_class.Generalization

		actual, err := parseGeneralization(subdomainKey, generalizationSubKey, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)

		// Test round-trip: generate content from parsed object and compare to original.
		generated := generateGeneralizationContent(actual)
		assert.Equal(suite.T(), testData.Contents, generated, testName)
	}
}
