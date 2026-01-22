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
	db           *sql.DB
	model        req_model.Model
	domain       model_domain.Domain
	subdomain    model_domain.Subdomain
	useCase      model_use_case.UseCase
	scenarioKey  identity.Key
	scenarioKeyB identity.Key
	// Class key for event parents.
	classKey identity.Key
	// Object keys for test nodes.
	obj1Key       identity.Key
	obj2Key       identity.Key
	obj3Key       identity.Key
	obj4Key       identity.Key
	helperFromKey identity.Key
	helperToKey   identity.Key
	// Event keys for test nodes.
	testEventKey    identity.Key
	addEventKey     identity.Key
	origEventKey    identity.Key
	updatedEventKey identity.Key
	eventXKey       identity.Key
	eventKey        identity.Key
	helperEventKey  identity.Key
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

	// Create class key for event parents.
	suite.classKey = helper.Must(identity.NewClassKey(suite.subdomain.Key, "test_class"))

	// Create object keys for test nodes.
	suite.obj1Key = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "obj1"))
	suite.obj2Key = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "obj2"))
	suite.obj3Key = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "obj3"))
	suite.obj4Key = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "obj4"))
	suite.helperFromKey = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "helper_from"))
	suite.helperToKey = helper.Must(identity.NewScenarioObjectKey(suite.scenarioKey, "helper_to"))

	// Create event keys for test nodes (events need a class parent).
	suite.testEventKey = helper.Must(identity.NewEventKey(suite.classKey, "test_event"))
	suite.addEventKey = helper.Must(identity.NewEventKey(suite.classKey, "add_event"))
	suite.origEventKey = helper.Must(identity.NewEventKey(suite.classKey, "orig_event"))
	suite.updatedEventKey = helper.Must(identity.NewEventKey(suite.classKey, "updated_event"))
	suite.eventXKey = helper.Must(identity.NewEventKey(suite.classKey, "event_x"))
	suite.eventKey = helper.Must(identity.NewEventKey(suite.classKey, "event"))
	suite.helperEventKey = helper.Must(identity.NewEventKey(suite.classKey, "helper_event"))
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
				FromObjectKey: suite.obj1Key,
				ToObjectKey:   suite.obj2Key,
				EventKey:      &suite.testEventKey,
			},
		},
	}

	// Build JSON with full key strings.
	stepsJSON := `{"statements":[{"description":"test step","from_object_key":"` + suite.obj1Key.String() + `","to_object_key":"` + suite.obj2Key.String() + `","event_key":"` + suite.testEventKey.String() + `"}]}`

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
				'`+stepsJSON+`'
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
					FromObjectKey: suite.obj1Key,
					ToObjectKey:   suite.obj2Key,
					EventKey:      &suite.addEventKey,
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
				FromObjectKey: suite.obj1Key,
				ToObjectKey:   suite.obj2Key,
				EventKey:      &suite.origEventKey,
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
				FromObjectKey: suite.obj3Key,
				ToObjectKey:   suite.obj4Key,
				EventKey:      &suite.updatedEventKey,
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
				FromObjectKey: suite.obj1Key,
				ToObjectKey:   suite.obj2Key,
				EventKey:      &suite.eventXKey,
			},
		},
	}

	steps := &model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description:   "step",
				FromObjectKey: suite.obj3Key,
				ToObjectKey:   suite.obj4Key,
				EventKey:      &suite.eventKey,
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
	// Create object keys for helper nodes.
	helperFromKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "helper_from"))
	helperToKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "helper_to"))

	// Extract subdomain key from use case key to create class key, then event key.
	useCaseKeyStr := useCaseKey.String()
	// Parse the subdomain key portion (everything before /usecase/).
	subdomainKey := helper.Must(identity.ParseKey(useCaseKeyStr[:len(useCaseKeyStr)-len("/usecase/"+useCaseKey.SubKey())]))
	helperClassKey := helper.Must(identity.NewClassKey(subdomainKey, "helper_class"))
	helperEventKey := helper.Must(identity.NewEventKey(helperClassKey, "helper_event"))

	err := AddScenario(dbOrTx, modelKey, useCaseKey, model_scenario.Scenario{
		Key:     scenarioKey,
		Name:    scenarioKey.String(),
		Details: "",
		Steps: &model_scenario.Node{
			Statements: []model_scenario.Node{
				{
					Description:   "helper step",
					FromObjectKey: helperFromKey,
					ToObjectKey:   helperToKey,
					EventKey:      &helperEventKey,
				},
			},
		},
	})
	assert.Nil(t, err)

	_, scenario, err = LoadScenario(dbOrTx, modelKey, scenarioKey)
	assert.Nil(t, err)

	return scenario
}
