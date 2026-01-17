package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

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
	db          *sql.DB
	model       req_model.Model
	domain      model_domain.Domain
	subdomain   model_domain.Subdomain
	useCase     model_use_case.UseCase
	scenarioKey identity.Key
	scenarioKeyB identity.Key
}

func (suite *ScenarioSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "use_case_key")))

	// Create the scenario keys for reuse.
	suite.scenarioKey = helper.Must(identity.NewScenarioKey(suite.useCase.Key, "scenario_key"))
	suite.scenarioKeyB = helper.Must(identity.NewScenarioKey(suite.useCase.Key, "scenario_key_b"))
}

func (suite *ScenarioSuite) TestLoad() {

	// Nothing in database yet.
	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseKey)
	assert.Empty(suite.T(), scenario)

	// Create expected steps
	expectedSteps := &model_scenario.Node{
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
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key',
				'Name',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key',
				'Details',
				'{"type":"sequence","statements":[{"type":"leaf","description":"test step","from_object_key":"obj1","to_object_key":"obj2","event_key":"test_event"}]}'
			)
	`)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err = LoadScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.useCase.Key, useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "Name",
		Details: "Details",
		Steps:   expectedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestAdd() {

	scenarioToAdd := model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "Name",
		Details: "Details",
		Steps: &model_scenario.Node{
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

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, scenarioToAdd)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.useCase.Key, useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "Name",
		Details: "Details",
		Steps:   scenarioToAdd.Steps,
	}, scenario)
}

func (suite *ScenarioSuite) TestUpdate() {

	originalSteps := &model_scenario.Node{
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
		Key:     suite.scenarioKey,
		Name:    "Name",
		Details: "Details",
		Steps:   originalSteps,
	})
	assert.Nil(suite.T(), err)

	updatedSteps := &model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "updated step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				EventKey:      "updated_event",
			},
		},
	}

	err = UpdateScenario(suite.db, suite.model.Key, model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	})
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.useCase.Key, useCaseKey)
	assert.Equal(suite.T(), model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "NameX",
		Details: "DetailsX",
		Steps:   updatedSteps,
	}, scenario)
}

func (suite *ScenarioSuite) TestRemove() {

	err := AddScenario(suite.db, suite.model.Key, suite.useCase.Key, model_scenario.Scenario{
		Key:     suite.scenarioKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.Nil(suite.T(), err)

	useCaseKey, scenario, err := LoadScenario(suite.db, suite.model.Key, suite.scenarioKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseKey)
	assert.Empty(suite.T(), scenario)
}

func (suite *ScenarioSuite) TestQueryScenarios() {

	stepsX := &model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "step X",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "event_x",
			},
		},
	}

	steps := &model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "step",
				FromObjectKey: "obj3",
				ToObjectKey:   "obj4",
				EventKey:      "event",
			},
		},
	}

	err := AddScenarios(suite.db, suite.model.Key, map[identity.Key][]model_scenario.Scenario{
		suite.useCase.Key: {
			{
				Key:     suite.scenarioKeyB,
				Name:    "NameX",
				Details: "DetailsX",
				Steps:   stepsX,
			},
			{
				Key:     suite.scenarioKey,
				Name:    "Name",
				Details: "Details",
				Steps:   steps,
			},
		},
	})
	assert.Nil(suite.T(), err)

	scenarios, err := QueryScenarios(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_scenario.Scenario{
		suite.useCase.Key: {
			{
				Key:     suite.scenarioKey,
				Name:    "Name",
				Details: "Details",
				Steps:   steps,
			},
			{
				Key:     suite.scenarioKeyB,
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

func t_AddScenario(t *testing.T, dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, useCaseKey identity.Key) (scenario model_scenario.Scenario) {

	err := AddScenario(dbOrTx, modelKey, useCaseKey, model_scenario.Scenario{
		Key:     scenarioKey,
		Name:    scenarioKey.String(),
		Details: "",
		Steps: &model_scenario.Node{
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
