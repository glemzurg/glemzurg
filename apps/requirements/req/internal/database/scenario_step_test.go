package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

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
	model     core.Model
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
	for i := range 10 {
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
	suite.ErrorIs(err, ErrNotFound)

	// Insert a step directly with raw SQL.
	err = dbExec(suite.db, `
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
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key/sstep/0',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key',
				NULL,
				0,
				'leaf',
				'event',
				NULL,
				'Step description',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key/sobject/from_obj',
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key/scenario/scenario_key/sobject/to_obj',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/event/event_key',
				NULL,
				NULL
			)
	`)
	suite.Require().NoError(err)

	scenarioKey, parentStepKey, sortOrder, step, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.Require().NoError(err)
	suite.Equal(suite.scenario.Key, scenarioKey)
	suite.Nil(parentStepKey)
	suite.Equal(0, sortOrder)
	suite.Equal(model_scenario.Step{
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
	suite.Require().NoError(err)

	scenarioKey, parentStepKey, sortOrder, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.Require().NoError(err)
	suite.Equal(suite.scenario.Key, scenarioKey)
	suite.Nil(parentStepKey)
	suite.Equal(0, sortOrder)
	suite.Equal(step, loaded)
}

func (suite *StepSuite) TestAddWithParent() {
	// Add root step.
	rootStep := model_scenario.Step{
		Key:      suite.stepKey(0),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, rootStep)
	suite.Require().NoError(err)

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
	suite.Require().NoError(err)

	scenarioKey, loadedParent, sortOrder, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(1))
	suite.Require().NoError(err)
	suite.Equal(suite.scenario.Key, scenarioKey)
	suite.NotNil(loadedParent)
	suite.Equal(suite.stepKey(0), *loadedParent)
	suite.Equal(0, sortOrder)
	suite.Equal(childStep, loaded)
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
	suite.Require().NoError(err)

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
	suite.Require().NoError(err)

	_, _, _, loaded, err := LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.Require().NoError(err)
	suite.Equal(updated, loaded)
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
	suite.Require().NoError(err)

	err = RemoveStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.Require().NoError(err)

	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().NoError(err)

	// Query and reconstruct.
	stepsMap, err := QuerySteps(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Len(stepsMap, 1)

	reconstructed := stepsMap[suite.scenario.Key]
	suite.NotNil(reconstructed)
	suite.Equal(rootStep, *reconstructed)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the scenario.
	err = RemoveScenario(suite.db, suite.model.Key, suite.scenario.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
}

// TestFKParentCascade tests fk_step_parent ON DELETE CASCADE: deleting a parent step deletes its children.
func (suite *StepSuite) TestFKParentCascade() {
	// Add root step.
	rootStep := model_scenario.Step{
		Key:      suite.stepKey(0),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
	}
	err := AddStep(suite.db, suite.model.Key, suite.scenario.Key, nil, 0, rootStep)
	suite.Require().NoError(err)

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
	suite.Require().NoError(err)

	// Delete the parent step.
	err = RemoveStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.Require().NoError(err)

	// Child should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(1))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the from_object.
	err = RemoveObject(suite.db, suite.model.Key, suite.fromObj.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the to_object.
	err = RemoveObject(suite.db, suite.model.Key, suite.toObj.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the event.
	err = RemoveEvent(suite.db, suite.model.Key, suite.class.Key, suite.event.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the query.
	err = RemoveQuery(suite.db, suite.model.Key, suite.class.Key, suite.query.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
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
	suite.Require().Error(err)
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
	suite.Require().NoError(err)

	// Delete the referenced scenario.
	err = RemoveScenario(suite.db, suite.model.Key, scenarioB.Key)
	suite.Require().NoError(err)

	// Step should be gone.
	_, _, _, _, err = LoadStep(suite.db, suite.model.Key, suite.stepKey(0))
	suite.ErrorIs(err, ErrNotFound)
}
