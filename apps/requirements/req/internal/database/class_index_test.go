package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassIndexSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ClassIndexSuite))
}

type ClassIndexSuite struct {
	suite.Suite
	db         *sql.DB
	model      req_model.Model
	domain     model_domain.Domain
	subdomain  model_domain.Subdomain
	class      model_class.Class
	attribute  model_class.Attribute
	attributeB model_class.Attribute
}

func (suite *ClassIndexSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
	suite.attribute = t_AddAttribute(suite.T(), suite.db, suite.model.Key, suite.class.Key, "attribute_key")
	suite.attributeB = t_AddAttribute(suite.T(), suite.db, suite.model.Key, suite.class.Key, "attribute_key_b")
}

func (suite *ClassIndexSuite) TestLoad() {

	// Nothing in database yet.
	indexes, err := LoadClassAttributeIndexes(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key))
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), indexes)

	_, err = dbExec(suite.db, `
		INSERT INTO class_index
			(
				model_key,
				class_key,
				index_num,
				attribute_key
			)
		VALUES
			(
				'model_key',
				'class_key',
				2,
				'attribute_key'
			),
			(
				'model_key',
				'class_key',
				1,
				'attribute_key'
			)
	`)
	assert.Nil(suite.T(), err)

	indexes, err = LoadClassAttributeIndexes(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []uint{1, 2}, indexes)
}

func (suite *ClassIndexSuite) TestAdd() {

	err := AddClassIndex(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key), 1)
	assert.Nil(suite.T(), err)

	err = AddClassIndex(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key), 2)
	assert.Nil(suite.T(), err)

	indexes, err := LoadClassAttributeIndexes(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []uint{1, 2}, indexes)
}

func (suite *ClassIndexSuite) TestRemove() {

	err := AddClassIndex(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key), 1)
	assert.Nil(suite.T(), err)

	err = AddClassIndex(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key), 2)
	assert.Nil(suite.T(), err)

	err = RemoveClassIndex(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key), 1)
	assert.Nil(suite.T(), err)

	indexes, err := LoadClassAttributeIndexes(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper(suite.attribute.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []uint{2}, indexes)
}
