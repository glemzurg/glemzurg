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
	t_CLASS_PATH_OK  = "test_files/class"
	t_CLASS_PATH_ERR = t_CLASS_PATH_OK + "/err"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassFileSuite))
}

type ClassFileSuite struct {
	suite.Suite
}

func (suite *ClassFileSuite) TestParseClassFiles() {

	// Create a parent subdomain key for testing.
	domainKey, err := identity.NewDomainKey("test_domain")
	assert.Nil(suite.T(), err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "test_subdomain")
	assert.Nil(suite.T(), err)

	classSubKey := "class_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_CLASS_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.T().Run(testName, func(t *testing.T) {
			var expected, actual model_class.Class

			actual, _, err := parseClass(subdomainKey, classSubKey, testData.Filename, testData.Contents)
			assert.Nil(t, err, testName)

			err = json.Unmarshal([]byte(testData.Json), &expected)
			assert.Nil(t, err, testName)

			assert.Equal(t, expected, actual, testName)

			// Test round-trip: generate content from parsed object and compare to original.
			generated := generateClassContent(actual)
			assert.Equal(t, testData.Contents, generated, testName)
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}
