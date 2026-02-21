package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStepSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(StepSuite))
}

type StepSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	useCase   model_use_case.UseCase
	scenario  model_scenario.Scenario
	fromObj   model_scenario.Object
	toObj     model_scenario.Object
	event     model_state.Event
	query     model_state.Query
	// Pre-created step keys.
	stepKeys []identity.Key
}

func (suite *StepSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add prerequisite objects.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "use_case_key")))
	suite.scenario = t_AddScenario(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewScenarioKey(suite.useCase.Key, "scenario_key")), suite.useCase.Key)

	// Add objects for from/to references.
	suite.fromObj = t_AddObject(suite.T(), suite.db, suite.model.Key, suite.scenario.Key, helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "from_obj")), 0, suite.class.Key)
	suite.toObj = t_AddObject(suite.T(), suite.db, suite.model.Key, suite.scenario.Key, helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "to_obj")), 1, suite.class.Key)

	// Add event and query for FK references.
	suite.event = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewEventKey(suite.class.Key, "event_key")))
	suite.query = t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewQueryKey(suite.class.Key, "query_key")))

	// Pre-create step keys.
	suite.stepKeys = nil
	for i := 0; i < 10; i++ {
		k := helper.Must(identity.NewScenarioStepKey(suite.scenario.Key, fmt.Sprintf("%d", i)))
		suite.stepKeys = append(suite.stepKeys, k)
	}
}

// stepKey returns the i-th pre-created step key.
func (suite *StepSuite) stepKey(i int) identity.Key {
	return suite.stepKeys[i]
}

// t_strPtr returns a pointer to a string.
func t_strPtr(s string) *string { return &s }

func (suite *StepSuite) TestLoad() {

	// Nothing in database yet.
	_, _, _, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert a step directly with raw SQL.
	stepKey0 := suite.stepKey(0)
	_, err = dbExec(suite.db, `
		INSERT INTO scenario_step
			(
				model_key,
				scenario_step_key,
				scenario_key,
				parent_step_key,
				sort_order,
				step_type,
				leaf_type,
				condition,
				description,
				from_object_key,
				to_object_key,
				event_key,
				query_key,
				scenario_ref_key
			)
		VALUES
			(
				$1, $2, $3, NULL, 0, 'leaf', 'event', NULL, 'Step description',
				$4, $5, $6, NULL, NULL
			)`,
		suite.model.Key,
		stepKey0.String(),
		suite.scenario.Key.String(),
		suite.fromObj.Key.String(),
		suite.toObj.Key.String(),
		suite.event.Key.String(),
	)
	assert.Nil(suite.T(), err)

	scenarioKey, parentStepKey, step, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Nil(suite.T(), parentStepKey)
	assert.Equal(suite.T(), model_scenario.Step{
		Key:           suite.stepKey(0),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Step description",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}, step)
}

func (suite *StepSuite) TestAdd() {

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}

	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, step)
	assert.Nil(suite.T(), err)

	scenarioKey, parentStepKey, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Nil(suite.T(), parentStepKey)
	assert.Equal(suite.T(), step, loaded)
}

func (suite *StepSuite) TestAddWithParent() {

	// Add root step.
	rootStep := model_scenario.Step{
		Key:       suite.stepKey(0),
		SortOrder: 0,
		StepType:  model_scenario.STEP_TYPE_SEQUENCE,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, rootStep)
	assert.Nil(suite.T(), err)

	// Add child step with parent.
	parentKey := suite.stepKey(0)
	childStep := model_scenario.Step{
		Key:           suite.stepKey(1),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Child event",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err = AddStep(suite.db, suite.model.Key, suite.scenario.Key, &parentKey, childStep)
	assert.Nil(suite.T(), err)

	scenarioKey, loadedParent, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(1))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.NotNil(suite.T(), loadedParent)
	assert.Equal(suite.T(), suite.stepKey(0), *loadedParent)
	assert.Equal(suite.T(), childStep, loaded)
}

func (suite *StepSuite) TestUpdate() {

	// Add initial step.
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Original",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, step)
	assert.Nil(suite.T(), err)

	// Update to a query leaf.
	updated := model_scenario.Step{
		Key:           suite.stepKey(0),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
		Description:   "Updated",
		FromObjectKey: &suite.toObj.Key,
		ToObjectKey:   &suite.fromObj.Key,
		QueryKey:      &suite.query.Key,
	}
	err = UpdateStep(suite.db, suite.model.Key, updated)
	assert.Nil(suite.T(), err)

	_, _, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), updated, loaded)
}

func (suite *StepSuite) TestRemove() {

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		SortOrder:     0,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "To be removed",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, step)
	assert.Nil(suite.T(), err)

	err = RemoveStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)

	_, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *StepSuite) TestQuerySteps() {

	// Build a tree: sequence > [event leaf, query leaf]
	rootStep := model_scenario.Step{
		Key:       suite.stepKey(0),
		SortOrder: 0,
		StepType:  model_scenario.STEP_TYPE_SEQUENCE,
		Statements: []model_scenario.Step{
			{
				Key:           suite.stepKey(1),
				SortOrder:     0,
				StepType:      model_scenario.STEP_TYPE_LEAF,
				LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
				Description:   "Event step",
				FromObjectKey: &suite.fromObj.Key,
				ToObjectKey:   &suite.toObj.Key,
				EventKey:      &suite.event.Key,
			},
			{
				Key:           suite.stepKey(2),
				SortOrder:     1,
				StepType:      model_scenario.STEP_TYPE_LEAF,
				LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
				Description:   "Query step",
				FromObjectKey: &suite.toObj.Key,
				ToObjectKey:   &suite.fromObj.Key,
				QueryKey:      &suite.query.Key,
			},
		},
	}

	// Flatten and insert.
	rows := flattenSteps(suite.scenario.Key, &rootStep)
	err := AddSteps(suite.db, suite.model.Key, rows)
	assert.Nil(suite.T(), err)

	// Query and reconstruct.
	stepsMap, err := QuerySteps(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), stepsMap, 1)

	reconstructed := stepsMap[suite.scenario.Key]
	assert.NotNil(suite.T(), reconstructed)
	assert.Equal(suite.T(), rootStep, *reconstructed)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddSteps(t *testing.T, dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, root *model_scenario.Step) {
	rows := flattenSteps(scenarioKey, root)
	err := AddSteps(dbOrTx, modelKey, rows)
	assert.Nil(t, err)
}
