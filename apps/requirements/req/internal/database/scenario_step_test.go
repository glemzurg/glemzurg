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
	_, _, _, _, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
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

	scenarioKey, parentStepKey, sortOrder, step, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Nil(suite.T(), parentStepKey)
	assert.Equal(suite.T(), 0, sortOrder)
	assert.Equal(suite.T(), model_scenario.Step{
		Key:           suite.stepKey(0),
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
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}

	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	scenarioKey, parentStepKey, sortOrder, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.Nil(suite.T(), parentStepKey)
	assert.Equal(suite.T(), 0, sortOrder)
	assert.Equal(suite.T(), step, loaded)
}

func (suite *StepSuite) TestAddWithParent() {

	// Add root step.
	rootStep := model_scenario.Step{
		Key:      suite.stepKey(0),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, rootStep)
	assert.Nil(suite.T(), err)

	// Add child step with parent.
	parentKey := suite.stepKey(0)
	childStep := model_scenario.Step{
		Key:           suite.stepKey(1),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Child event",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err = AddStep(suite.db, suite.model.Key, suite.scenario.Key, &parentKey, 0, childStep)
	assert.Nil(suite.T(), err)

	scenarioKey, loadedParent, sortOrder, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(1))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.scenario.Key, scenarioKey)
	assert.NotNil(suite.T(), loadedParent)
	assert.Equal(suite.T(), suite.stepKey(0), *loadedParent)
	assert.Equal(suite.T(), 0, sortOrder)
	assert.Equal(suite.T(), childStep, loaded)
}

func (suite *StepSuite) TestUpdate() {

	// Add initial step.
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Original",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Update to a query leaf.
	updated := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
		Description:   "Updated",
		FromObjectKey: &suite.toObj.Key,
		ToObjectKey:   &suite.fromObj.Key,
		QueryKey:      &suite.query.Key,
	}
	err = UpdateStep(suite.db, suite.model.Key, 0, updated)
	assert.Nil(suite.T(), err)

	_, _, _, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), updated, loaded)
}

func (suite *StepSuite) TestRemove() {

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "To be removed",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	err = RemoveStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)

	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *StepSuite) TestQuerySteps() {

	// Build a tree: sequence > [event leaf, query leaf]
	rootStep := model_scenario.Step{
		Key:      suite.stepKey(0),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
		Statements: []model_scenario.Step{
			{
				Key:           suite.stepKey(1),
				StepType:      model_scenario.STEP_TYPE_LEAF,
				LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
				Description:   "Event step",
				FromObjectKey: &suite.fromObj.Key,
				ToObjectKey:   &suite.toObj.Key,
				EventKey:      &suite.event.Key,
			},
			{
				Key:           suite.stepKey(2),
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

// ===== Foreign Key Tests =====

// TestFKScenario tests fk_step_scenario: scenario_key must reference an existing scenario.
func (suite *StepSuite) TestFKScenario() {
	bogusScenarioKey := helper.Must(identity.NewScenarioKey(suite.useCase.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}

	// Insert with non-existent scenario_key should fail.
	err := AddStep(suite.db, suite.model.Key, bogusScenarioKey, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKScenarioCascade tests fk_step_scenario ON DELETE CASCADE: deleting the scenario deletes its steps.
func (suite *StepSuite) TestFKScenarioCascade() {
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the scenario.
	err = RemoveScenario(suite.db, suite.model.Key, suite.scenario.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKParent tests fk_step_parent: parent_step_key must reference an existing step.
func (suite *StepSuite) TestFKParent() {
	bogusParentKey := helper.Must(identity.NewScenarioStepKey(suite.scenario.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}

	// Insert with non-existent parent_step_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, &bogusParentKey, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKParentCascade tests fk_step_parent ON DELETE CASCADE: deleting a parent step deletes its children.
func (suite *StepSuite) TestFKParentCascade() {
	// Add root step.
	rootStep := model_scenario.Step{
		Key:      suite.stepKey(0),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, rootStep)
	assert.Nil(suite.T(), err)

	// Add child step.
	parentKey := suite.stepKey(0)
	childStep := model_scenario.Step{
		Key:           suite.stepKey(1),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Child event",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err = AddStep(suite.db, suite.model.Key, suite.scenario.Key, &parentKey, 0, childStep)
	assert.Nil(suite.T(), err)

	// Delete the parent step.
	err = RemoveStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.Nil(suite.T(), err)

	// Child should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(1))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKFromObject tests fk_step_from_object: from_object_key must reference an existing scenario_object.
func (suite *StepSuite) TestFKFromObject() {
	bogusObjectKey := helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &bogusObjectKey,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}

	// Insert with non-existent from_object_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKFromObjectCascade tests fk_step_from_object ON DELETE CASCADE: deleting the from_object deletes the step.
func (suite *StepSuite) TestFKFromObjectCascade() {
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the from_object.
	err = RemoveObject(suite.db, suite.model.Key, suite.fromObj.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKToObject tests fk_step_to_object: to_object_key must reference an existing scenario_object.
func (suite *StepSuite) TestFKToObject() {
	bogusObjectKey := helper.Must(identity.NewScenarioObjectKey(suite.scenario.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &bogusObjectKey,
		EventKey:      &suite.event.Key,
	}

	// Insert with non-existent to_object_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKToObjectCascade tests fk_step_to_object ON DELETE CASCADE: deleting the to_object deletes the step.
func (suite *StepSuite) TestFKToObjectCascade() {
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the to_object.
	err = RemoveObject(suite.db, suite.model.Key, suite.toObj.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKEvent tests fk_step_event: event_key must reference an existing event.
func (suite *StepSuite) TestFKEvent() {
	bogusEventKey := helper.Must(identity.NewEventKey(suite.class.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &bogusEventKey,
	}

	// Insert with non-existent event_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKEventCascade tests fk_step_event ON DELETE CASCADE: deleting the event deletes the step.
func (suite *StepSuite) TestFKEventCascade() {
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
		Description:   "Event step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		EventKey:      &suite.event.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the event.
	err = RemoveEvent(suite.db, suite.model.Key, suite.class.Key, suite.event.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKQuery tests fk_step_query: query_key must reference an existing query.
func (suite *StepSuite) TestFKQuery() {
	bogusQueryKey := helper.Must(identity.NewQueryKey(suite.class.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
		Description:   "Query step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		QueryKey:      &bogusQueryKey,
	}

	// Insert with non-existent query_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKQueryCascade tests fk_step_query ON DELETE CASCADE: deleting the query deletes the step.
func (suite *StepSuite) TestFKQueryCascade() {
	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
		Description:   "Query step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		QueryKey:      &suite.query.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the query.
	err = RemoveQuery(suite.db, suite.model.Key, suite.class.Key, suite.query.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

// TestFKScenarioRef tests fk_step_scenario_ref: scenario_ref_key must reference an existing scenario.
func (suite *StepSuite) TestFKScenarioRef() {
	bogusScenarioKey := helper.Must(identity.NewScenarioKey(suite.useCase.Key, "bogus"))

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_SCENARIO),
		Description:   "Scenario step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		ScenarioKey:   &bogusScenarioKey,
	}

	// Insert with non-existent scenario_ref_key should fail.
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.NotNil(suite.T(), err)
}

// TestFKScenarioRefCascade tests fk_step_scenario_ref ON DELETE CASCADE: deleting the referenced scenario deletes the step.
func (suite *StepSuite) TestFKScenarioRefCascade() {
	// Create a second scenario to reference.
	scenarioKeyB := helper.Must(identity.NewScenarioKey(suite.useCase.Key, "scenario_b"))
	scenarioB := t_AddScenario(suite.T(), suite.db, suite.model.Key, scenarioKeyB, suite.useCase.Key)

	step := model_scenario.Step{
		Key:           suite.stepKey(0),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      t_strPtr(model_scenario.LEAF_TYPE_SCENARIO),
		Description:   "Scenario step",
		FromObjectKey: &suite.fromObj.Key,
		ToObjectKey:   &suite.toObj.Key,
		ScenarioKey:   &scenarioB.Key,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, step)
	assert.Nil(suite.T(), err)

	// Delete the referenced scenario.
	err = RemoveScenario(suite.db, suite.model.Key, scenarioB.Key)
	assert.Nil(suite.T(), err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddSteps(t *testing.T, dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, root *model_scenario.Step) {
	rows := flattenSteps(scenarioKey, root)
	err := AddSteps(dbOrTx, modelKey, rows)
	assert.Nil(t, err)
}
