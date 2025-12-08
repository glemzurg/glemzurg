package parser

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/futz/apps/req/requirements"

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

	key := "class_key"

	testDataFiles, err := t_ContentsForAllMdFiles(t_CLASS_PATH_OK)
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		var expected, actual requirements.Class

		actual, err := parseClass(key, testData.Filename, testData.Contents)
		assert.Nil(suite.T(), err, testName)

		err = json.Unmarshal([]byte(testData.Json), &expected)
		assert.Nil(suite.T(), err, testName)

		assert.Equal(suite.T(), expected, actual, testName)
	}
}
