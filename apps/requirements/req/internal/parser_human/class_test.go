package parser_human

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
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
	domainKey := helper.Must(identity.NewDomainKey("test_domain"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "test_subdomain"))

	classSubKey := "class_key"

	testDataFiles := helper.Must(t_ContentsForAllMdFiles(t_CLASS_PATH_OK))

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			t := suite.T() //nolint:testifylint // captures subtest result
			var expected, actual model_class.Class

			actual, associations, err := parseClass(subdomainKey, classSubKey, testData.Filename, testData.Contents)
			require.NoError(t, err, testName)

			err = json.Unmarshal([]byte(testData.Json), &expected)
			require.NoError(t, err, testName)

			suite.Equal(expected, actual, testName)

			// Test associations if expected data exists (via _children.json file).
			if testData.JsonChildren != "" {
				var expectedAssociations []model_class.Association
				err = json.Unmarshal([]byte(testData.JsonChildren), &expectedAssociations)
				require.NoError(t, err, testName+" associations json")
				suite.Equal(expectedAssociations, associations, testName+" associations")
			}

			// Test round-trip: generate content from parsed object and compare to original.
			generated := generateClassContent(actual, associations)
			suite.Equal(testData.Contents, generated, testName)
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}
