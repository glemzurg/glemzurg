package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ScenarioSuite))
}

type ScenarioSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	useCase   model_use_case.UseCase
}

func (suite *ScenarioSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "use_case_key")
}

func (suite *ScenarioSuite) TestLoad() {

	// Nothing in database yet.
	useCaseKey, scenario, err := LoadScenario(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseKey)
	assert.Empty(suite.T(), scenario)

	// Create expected steps
	expectedSteps := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "test step",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "test_event",
			},
		},
	}

	_, err = dbExec(suite.db, `
		INSERT INTO scenario
			(
				model_key,
				scenario_key,
				name,
				use_case_key,
				details,
				steps
			)
		VALUES
			(
				'model_key',
				'key',
				'Name',
				'use_case_key',
				'Details',
				'{"type":"sequence","statements":[{"type":"leaf","description":"test step","from_object_key":"obj1","to_object_key":"obj2","event_key":"test_event"}]}'
			)
	`)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err = LoadScenario(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
		Steps:   expectedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestAdd() {

	scenarioToAdd := model_scenario.Scenario{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
		Steps: model_scenario.Node{
			Statements: []model_scenario.Node{
				{
					Description:   "add test step",
					FromObjectKey: "obj1",
					ToObjectKey:   "obj2",
					EventKey:      "add_event",
				},
			},
		},
	}

	err := AddScenario(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.useCase.Key), scenarioToAdd)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
		Steps:   scenarioToAdd.Steps,
	}, scenario)
}

func (suite *ScenarioSuite) TestUpdate() {

	originalSteps := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "original step",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "orig_event",
			},
		},
	}

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, model_scenario.Scenario{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
		Steps:   originalSteps,
	})
	assert.Nil(suite.T(), err)

	updatedSteps := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "updated step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				EventKey:      "updated_event",
			},
		},
	}

	err = UpdateScenario(suite.db, strings.ToUpper(suite.model.Key), model_scenario.Scenario{
		Key:     "KeY", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	})
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     "key",
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestRemove() {

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, model_scenario.Scenario{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveScenario(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseKey)
	assert.Empty(suite.T(), scenario)
}

func (suite *ScenarioSuite) TestQueryScenarios() {

	stepsX := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "step X",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "event_x",
			},
		},
	}

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, model_scenario.Scenario{
		Key:     "scenario_key_x",
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   stepsX,
	})
	assert.Nil(suite.T(), err)

	steps := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				EventKey:      "event",
			},
		},
	}

	err = AddScenario(suite.db, suite.model.Key, suite.useCase.Key, model_scenario.Scenario{
		Key:     "scenario_key",
		Name:    "Name",
		Details: "Details",
		Steps:   steps,
	})
	assert.Nil(suite.T(), err)

	scenarios, err := QueryScenarios(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_scenario.Scenario{
		suite.useCase.Key: {
			{
				Key:     "scenario_key",
				Name:    "Name",
				Details: "Details",
				Steps:   steps,
			},
			{
				Key:     "scenario_key_x",
				Name:    "NameX",
				Details: "DetailsX",
				Steps:   stepsX,
			},
		},
	}, scenarios)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddScenario(t *testing.T, dbOrTx DbOrTx, modelKey, scenarioKey, useCaseKey string) (scenario model_scenario.Scenario) {

	err := AddScenario(dbOrTx, modelKey, useCaseKey, model_scenario.Scenario{
		Key:     scenarioKey,
		Name:    scenarioKey,
		Details: "",
		Steps: model_scenario.Node{
			Statements: []model_scenario.Node{
				{
					Description:   "helper step",
					FromObjectKey: "helper_from",
					ToObjectKey:   "helper_to",
					EventKey:      "helper_event",
				},
			},
		},
	})
	assert.Nil(t, err)

	_, scenario, err = LoadScenario(dbOrTx, modelKey, scenarioKey)
	assert.Nil(t, err)

	return scenario
}
