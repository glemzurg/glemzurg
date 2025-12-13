package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioObjectSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ScenarioObjectSuite))
}

type ScenarioObjectSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    requirements.Domain
	subdomain requirements.Subdomain
	class     requirements.Class
	classB    requirements.Class
	useCase   requirements.UseCase
	scenario  requirements.Scenario
}

func (suite *ScenarioObjectSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
	suite.classB = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key_b")
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "use_case_key")
	suite.scenario = t_AddScenario(suite.T(), suite.db, suite.model.Key, "scenario_key", suite.useCase.Key)
}

func (suite *ScenarioObjectSuite) TestLoad() {

	// Nothing in database yet.
	scenarioKey, scenarioObject, err := LoadScenarioObject(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), scenarioKey)
	assert.Empty(suite.T(), scenarioObject)

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
				'key',
				'scenario_key',
				1,
				'Name',
				'name',
				'class_key',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	scenarioKey, scenarioObject, err = LoadScenarioObject(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "scenario_key", scenarioKey)
	assert.Equal(suite.T(), requirements.ScenarioObject{
		Key:          "key", // Test case-insensitive.
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_key",
		Multi:        true,
		UmlComment:   "UmlComment",
	}, scenarioObject)
}

func (suite *ScenarioObjectSuite) TestAdd() {

	err := AddScenarioObject(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper("scenario_key"), requirements.ScenarioObject{
		Key:          "KeY", // Test case-insensitive.
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_KEy", // Test case-insensitive.
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	scenarioKey, scenarioObject, err := LoadScenarioObject(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "scenario_key", scenarioKey)
	assert.Equal(suite.T(), requirements.ScenarioObject{
		Key:          "key",
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_key",
		Multi:        true,
		UmlComment:   "UmlComment",
	}, scenarioObject)
}

func (suite *ScenarioObjectSuite) TestUpdate() {

	err := AddScenarioObject(suite.db, suite.model.Key, "scenario_key", requirements.ScenarioObject{
		Key:          "key",
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_key",
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateScenarioObject(suite.db, strings.ToUpper(suite.model.Key), requirements.ScenarioObject{
		Key:          "kEy", // Test case-insensitive.
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     "class_KEY_b", // Test case-insensitive.
		Multi:        false,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	scenarioKey, scenarioObject, err := LoadScenarioObject(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "scenario_key", scenarioKey)
	assert.Equal(suite.T(), requirements.ScenarioObject{
		Key:          "key",
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     "class_key_b",
		Multi:        false,
		UmlComment:   "UmlCommentX",
	}, scenarioObject)
}

func (suite *ScenarioObjectSuite) TestRemove() {

	err := AddScenarioObject(suite.db, suite.model.Key, suite.scenario.Key, requirements.ScenarioObject{
		Key:          "key",
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_key",
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveScenarioObject(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	scenarioKey, scenarioObject, err := LoadScenarioObject(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), scenarioKey)
	assert.Empty(suite.T(), scenarioObject)
}

func (suite *ScenarioObjectSuite) TestQuery() {

	err := AddScenarioObject(suite.db, suite.model.Key, "scenario_key", requirements.ScenarioObject{
		Key:          "keyx",
		ObjectNumber: 2,
		Name:         "NameX",
		NameStyle:    "id",
		ClassKey:     "class_key_b",
		Multi:        true,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddScenarioObject(suite.db, suite.model.Key, "scenario_key", requirements.ScenarioObject{
		Key:          "key",
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    "name",
		ClassKey:     "class_key",
		Multi:        false,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	scenarioObjects, err := QueryScenarioObjects(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	expected := map[string][]requirements.ScenarioObject{
		"scenario_key": {
			{
				Key:          "key",
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "name",
				ClassKey:     "class_key",
				Multi:        false,
				UmlComment:   "UmlComment",
			},
			{
				Key:          "keyx",
				ObjectNumber: 2,
				Name:         "NameX",
				NameStyle:    "id",
				ClassKey:     "class_key_b",
				Multi:        true,
				UmlComment:   "UmlCommentX",
			},
		},
	}
	assert.Equal(suite.T(), expected, scenarioObjects)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddScenarioObject(t *testing.T, dbOrTx DbOrTx, modelKey, scenarioKey, scenarioObjectKey string, objectNumber uint, classKey string) (scenarioObject requirements.ScenarioObject) {

	err := AddScenarioObject(dbOrTx, modelKey, scenarioKey, requirements.ScenarioObject{
		Key:          scenarioObjectKey,
		ObjectNumber: objectNumber,
		Name:         scenarioObjectKey,
		NameStyle:    "name",
		ClassKey:     classKey,
		Multi:        true,
		UmlComment:   "UmlComment",
	})
	assert.Nil(t, err)

	_, scenarioObject, err = LoadScenarioObject(dbOrTx, modelKey, scenarioObjectKey)
	assert.Nil(t, err)

	return scenarioObject
}
