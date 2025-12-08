package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/futz/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./database/... -dbtests`")
	}
	suite.Run(t, new(ScenarioSuite))
}

type ScenarioSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    requirements.Domain
	subdomain requirements.Subdomain
	useCase   requirements.UseCase
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
	expectedSteps := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "test step",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				ActionKey:     "test_action",
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
				'{"type":"sequence","statements":[{"type":"leaf","description":"test step","from_object_key":"obj1","to_object_key":"obj2","action_key":"test_action"}]}'
			)
	`)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err = LoadScenario(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), requirements.Scenario{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
		Steps:   expectedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestAdd() {

	scenarioToAdd := requirements.Scenario{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
		Steps: requirements.Node{
			Statements: []requirements.Node{
				{
					Description:   "add test step",
					FromObjectKey: "obj1",
					ToObjectKey:   "obj2",
					ActionKey:     "add_action",
				},
			},
		},
	}

	err := AddScenario(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.useCase.Key), scenarioToAdd)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), requirements.Scenario{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
		Steps:   scenarioToAdd.Steps,
	}, scenario)
}

func (suite *ScenarioSuite) TestUpdate() {

	originalSteps := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "original step",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				ActionKey:     "orig_action",
			},
		},
	}

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, requirements.Scenario{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
		Steps:   originalSteps,
	})
	assert.Nil(suite.T(), err)

	updatedSteps := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "updated step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				ActionKey:     "updated_action",
			},
		},
	}

	err = UpdateScenario(suite.db, strings.ToUpper(suite.model.Key), requirements.Scenario{
		Key:     "KeY", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	})
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "use_case_key", useCaseKey)
	assert.Equal(suite.T(), requirements.Scenario{
		Key:     "key",
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestRemove() {

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, requirements.Scenario{
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

	stepsX := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "step X",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				ActionKey:     "action_x",
			},
		},
	}

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, requirements.Scenario{
		Key:     "scenario_key_x",
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   stepsX,
	})
	assert.Nil(suite.T(), err)

	steps := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				ActionKey:     "action",
			},
		},
	}

	err = AddScenario(suite.db, suite.model.Key, suite.useCase.Key, requirements.Scenario{
		Key:     "scenario_key",
		Name:    "Name",
		Details: "Details",
		Steps:   steps,
	})
	assert.Nil(suite.T(), err)

	scenarios, err := QueryScenarios(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]requirements.Scenario{
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

func t_AddScenario(t *testing.T, dbOrTx DbOrTx, modelKey, scenarioKey, useCaseKey string) (scenario requirements.Scenario) {

	err := AddScenario(dbOrTx, modelKey, useCaseKey, requirements.Scenario{
		Key:     scenarioKey,
		Name:    scenarioKey,
		Details: "",
		Steps: requirements.Node{
			Statements: []requirements.Node{
				{
					Description:   "helper step",
					FromObjectKey: "helper_from",
					ToObjectKey:   "helper_to",
					ActionKey:     "helper_action",
				},
			},
		},
	})
	assert.Nil(t, err)

	_, scenario, err = LoadScenario(dbOrTx, modelKey, scenarioKey)
	assert.Nil(t, err)

	return scenario
}
