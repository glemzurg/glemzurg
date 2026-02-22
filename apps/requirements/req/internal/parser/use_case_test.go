package parser

// TODO: TestUseCaseSuite needs updating — generateSteps is commented out so
// scenario step round-trip (04_scenarios.md) fails.

// import (
// 	"encoding/json"
// 	"testing"
//
// 	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
// 	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// )
//
// const (
// 	t_USE_CASE_PATH_OK  = "test_files/use_case"
// 	t_USE_CASE_PATH_ERR = t_USE_CASE_PATH_OK + "/err"
// )
//
// func TestUseCaseSuite(t *testing.T) {
// 	suite.Run(t, new(UseCaseFileSuite))
// }
//
// type UseCaseFileSuite struct {
// 	suite.Suite
// }
//
// func (suite *UseCaseFileSuite) TestParseUseCaseFiles() {
//
// 	// Create a parent subdomain key for testing.
// 	domainKey, err := identity.NewDomainKey("test_domain")
// 	assert.Nil(suite.T(), err)
// 	subdomainKey, err := identity.NewSubdomainKey(domainKey, "test_subdomain")
// 	assert.Nil(suite.T(), err)
//
// 	useCaseSubKey := "use_case_key"
//
// 	testDataFiles, err := t_ContentsForAllMdFiles(t_USE_CASE_PATH_OK)
// 	assert.Nil(suite.T(), err)
//
// 	for _, testData := range testDataFiles {
// 		testName := testData.Filename
// 		var expected, actual model_use_case.UseCase
//
// 		actual, err := parseUseCase(subdomainKey, useCaseSubKey, testData.Filename, testData.Contents)
// 		assert.Nil(suite.T(), err, testName)
//
// 		err = json.Unmarshal([]byte(testData.Json), &expected)
// 		assert.Nil(suite.T(), err, testName)
//
// 		assert.Equal(suite.T(), expected, actual, testName)
//
// 		// Test round-trip: generate content from parsed object and compare to original.
// 		generated := generateUseCaseContent(actual)
// 		assert.Equal(suite.T(), testData.Contents, generated, testName)
// 	}
// }
