package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLogicSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
	db        *sql.DB
	model     core.Model
	logicKey  identity.Key
	logicKeyB identity.Key
}

func (suite *LogicSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the keys for reuse.
	suite.logicKey = helper.Must(identity.NewInvariantKey("0"))
	suite.logicKeyB = helper.Must(identity.NewInvariantKey("1"))
}

func (suite *LogicSuite) TestLoad() {
	// Nothing in database yet.
	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(logic)

	err = dbExec(suite.db, `
		INSERT INTO logic
			(
				model_key,
				logic_key,
				logic_type,
				description,
				target,
				notation,
				specification,
				sort_order
			)
		VALUES
			(
				'model_key',
				'invariant/0',
				'assessment',
				'Description',
				'Target',
				'tla_plus',
				'Specification',
				1
			)
	`)
	suite.Require().NoError(err)

	logic, err = LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	}, logic)
}

func (suite *LogicSuite) TestAdd() {
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	})
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	}, logic)
}

func (suite *LogicSuite) TestAddNulls() {
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: ""},
	})
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: ""},
	}, logic)
}

func (suite *LogicSuite) TestUpdate() {
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	})
	suite.Require().NoError(err)

	err = UpdateLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "DescriptionX",
		Target:      "TargetX",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "SpecificationX"},
	}, 0)
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "DescriptionX",
		Target:      "TargetX",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "SpecificationX"},
	}, logic)
}

func (suite *LogicSuite) TestUpdateNulls() {
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	})
	suite.Require().NoError(err)

	err = UpdateLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "DescriptionX",
		Target:      "",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: ""},
	}, 0)
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "DescriptionX",
		Target:      "",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: ""},
	}, logic)
}

func (suite *LogicSuite) TestRemove() {
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:         suite.logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "Description",
		Target:      "Target",
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	})
	suite.Require().NoError(err)

	err = RemoveLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(logic)
}

func (suite *LogicSuite) TestQuery() {
	err := AddLogics(suite.db, suite.model.Key, []model_logic.Logic{
		{
			Key:         suite.logicKeyB,
			Type:        model_logic.LogicTypeAssessment,
			Description: "DescriptionX",
			Target:      "TargetX",
			Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "SpecificationX"},
		},
		{
			Key:         suite.logicKey,
			Type:        model_logic.LogicTypeAssessment,
			Description: "Description",
			Target:      "Target",
			Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
		},
	}, map[string]int{
		suite.logicKeyB.String(): 0,
		suite.logicKey.String():  1,
	})
	suite.Require().NoError(err)

	logics, err := QueryLogics(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal([]model_logic.Logic{
		{
			Key:         suite.logicKey,
			Type:        model_logic.LogicTypeAssessment,
			Description: "Description",
			Target:      "Target",
			Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
		},
		{
			Key:         suite.logicKeyB,
			Type:        model_logic.LogicTypeAssessment,
			Description: "DescriptionX",
			Target:      "TargetX",
			Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "SpecificationX"},
		},
	}, logics)
}

func (suite *LogicSuite) TestAddLetType() {
	ts := logic_spec.TypeSpec{Notation: "tla_plus", Specification: "Int"}
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:            suite.logicKey,
		Type:           model_logic.LogicTypeLet,
		Description:    "Compute threshold",
		Target:         "threshold",
		Spec:           logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "10"},
		TargetTypeSpec: &ts,
	})
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:            suite.logicKey,
		Type:           model_logic.LogicTypeLet,
		Description:    "Compute threshold",
		Target:         "threshold",
		Spec:           logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "10"},
		TargetTypeSpec: &logic_spec.TypeSpec{Notation: "tla_plus", Specification: "Int"},
	}, logic)
}

func (suite *LogicSuite) TestAddWithTargetTypeSpec() {
	ts := logic_spec.TypeSpec{Notation: "tla_plus", Specification: "STRING"}
	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:            suite.logicKey,
		Type:           model_logic.LogicTypeAssessment,
		Description:    "Description",
		Target:         "",
		Spec:           logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
		TargetTypeSpec: &ts,
	})
	suite.Require().NoError(err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(model_logic.Logic{
		Key:            suite.logicKey,
		Type:           model_logic.LogicTypeAssessment,
		Description:    "Description",
		Target:         "",
		Spec:           logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
		TargetTypeSpec: &logic_spec.TypeSpec{Notation: "tla_plus", Specification: "STRING"},
	}, logic)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddLogic(t *testing.T, dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (logic model_logic.Logic) {
	err := AddLogic(dbOrTx, modelKey, model_logic.Logic{
		Key:         logicKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: logicKey.String(),
		Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "Specification"},
	})
	require.NoError(t, err)

	logic, err = LoadLogic(dbOrTx, modelKey, logicKey)
	require.NoError(t, err)

	return logic
}
