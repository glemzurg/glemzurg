package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestObjectSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ObjectSuite))
}

type ObjectSuite struct {
	suite.Suite
	db         *sql.DB
	model      req_model.Model
	domain     model_domain.Domain
	subdomain  model_domain.Subdomain
	class      model_class.Class
	classB     model_class.Class
	useCase    model_use_case.UseCase
	scenario   model_scenario.Scenario
	objectKey  identity.Key
	objectKeyB identity.Key
}

func (suite *ObjectSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.classB = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key_b")))
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "use_case_key")))
	suite.scenario = t_AddScenario(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewScenarioKey(suite.useCase.Key, "scenario_key")), suite.useCase.Key)

	// Create the object keys for reuse.
	suite.objectKey = helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "object_key"))
	suite.objectKeyB = helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "object_key_b"))
}

func (suite *ObjectSuite) TestLoad() {

	// Nothing in database yet.
	scenarioKey, object, err := LoadObject(suite.db, suite.model.Key, suite.objectKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), scenarioKey)
	assert.Empty(suite.T(), object)

	_, err = dbExec(suite.db, `
		INSERT INTO scenario_object
			(
				model_key,
				scenario_object_key,
				scenario_key,
				object_number,
				name,
				name_style,
				class_key,
				multi,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key/sobject/object_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key',
				1,
				'Name',
				'name',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	scenarioKey, object, err = LoadObject(suite.db, suite.model.Key, suite.objectKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Equal(suite.T(), model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        true,
		UmlComment:   "UmlComment",
	}, object)
}

func (suite *ObjectSuite) TestAdd() {

	err := AddObject(suite.db, suite.model.Key, suite.scenario.Key, model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	scenarioKey, object, err := LoadObject(suite.db, suite.model.Key, suite.objectKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Equal(suite.T(), model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        true,
		UmlComment:   "UmlComment",
	}, object)
}

func (suite *ObjectSuite) TestUpdate() {

	err := AddObject(suite.db, suite.model.Key, suite.scenario.Key, model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateObject(suite.db, suite.model.Key, model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     suite.classB.Key,
		Multi:        false,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	scenarioKey, object, err := LoadObject(suite.db, suite.model.Key, suite.objectKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Equal(suite.T(), model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     suite.classB.Key,
		Multi:        false,
		UmlComment:   "UmlCommentX",
	}, object)
}

func (suite *ObjectSuite) TestRemove() {

	err := AddObject(suite.db, suite.model.Key, suite.scenario.Key, model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveObject(suite.db, suite.model.Key, suite.objectKey)
	assert.Nil(suite.T(), err)

	scenarioKey, object, err := LoadObject(suite.db, suite.model.Key, suite.objectKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), scenarioKey)
	assert.Empty(suite.T(), object)
}

func (suite *ObjectSuite) TestQuery() {

	err := AddObject(suite.db, suite.model.Key, suite.scenario.Key, model_scenario.Object{
		Key:          suite.objectKeyB,
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     suite.classB.Key,
		Multi:        true,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddObject(suite.db, suite.model.Key, suite.scenario.Key, model_scenario.Object{
		Key:          suite.objectKey,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     suite.class.Key,
		Multi:        false,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	objects, err := QueryObjects(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	expected := map[identity.Key][]model_scenario.Object{
		suite.scenario.Key: {
			{
				Key:          suite.objectKey,
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "name",
				ClassKey:     suite.class.Key,
				Multi:        false,
				UmlComment:   "UmlComment",
			},
			{
				Key:          suite.objectKeyB,
				ObjectNumber: 2,
				Name:         "NameX",
				NameStyle:    "id",
				ClassKey:     suite.classB.Key,
				Multi:        true,
				UmlComment:   "UmlCommentX",
			},
		},
	}
	assert.Equal(suite.T(), expected, objects)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddObject(t *testing.T, dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, objectKey identity.Key, objectNumber uint, classKey identity.Key) (object model_scenario.Object) {

	err := AddObject(dbOrTx, modelKey, scenarioKey, model_scenario.Object{
		Key:          objectKey,
		ObjectNumber: objectNumber,
		Name:         objectKey.String(),
		NameStyle:    "name",
		ClassKey:     classKey,
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(t, err)

	_, object, err = LoadObject(dbOrTx, modelKey, objectKey)
	assert.Nil(t, err)

	return object
}
