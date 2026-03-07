package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"

	"github.com/stretchr/testify/suite"
)

func TestAtomicSpanSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AtomicSpanSuite))
}

type AtomicSpanSuite struct {
	suite.Suite
	db        *sql.DB
	model     core.Model
	dataType  model_data_type.DataType
	dataTypeB model_data_type.DataType
	atomic    model_data_type.Atomic
	atomicB   model_data_type.Atomic
}

func (suite *AtomicSpanSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.dataType = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key")
	suite.dataTypeB = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key_b")
	suite.atomic = t_AddAtomic(suite.T(), suite.db, suite.model.Key, suite.dataType.Key, "span", nil, nil)
	suite.atomicB = t_AddAtomic(suite.T(), suite.db, suite.model.Key, suite.dataTypeB.Key, "span", nil, nil)
}

func (suite *AtomicSpanSuite) TestLoad() {
	// Nothing in database yet.
	parentDataTypeKey, span, err := LoadAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), "data_type_key")
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(parentDataTypeKey)
	suite.Empty(span)

	err = dbExec(suite.db, `
		INSERT INTO data_type_atomic_span
			(
				model_key,
				data_type_key,
				lower_type,
				lower_value,
				lower_denominator,
				higher_type,
				higher_value,
				higher_denominator,
				units,
				precision
			)
		VALUES
			(
				'model_key',
				'data_type_key',
				'closed',
				1,
				2,
				'open',
				3,
				4,
				'Units',
				0.01
			)
	`)
	suite.Require().NoError(err)

	parentDataTypeKey, span, err = LoadAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), "data_Type_Key") // Test case-insensitive.
	suite.Require().NoError(err)
	suite.Equal("data_type_key", parentDataTypeKey)
	suite.Equal(model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	}, span)
}

func (suite *AtomicSpanSuite) TestAdd() {
	err := AddAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.AtomicSpan{ // Test case-insensitive.
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	parentDataTypeKey, span, err := LoadAtomicSpan(suite.db, suite.model.Key, "data_type_key")
	suite.Require().NoError(err)
	suite.Equal("data_type_key", parentDataTypeKey)
	suite.Equal(model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	}, span)
}

func (suite *AtomicSpanSuite) TestAddNulls() {
	err := AddAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        nil,
		LowerDenominator:  nil,
		HigherType:        "open",
		HigherValue:       nil,
		HigherDenominator: nil,
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	parentDataTypeKey, span, err := LoadAtomicSpan(suite.db, suite.model.Key, "data_type_key")
	suite.Require().NoError(err)
	suite.Equal("data_type_key", parentDataTypeKey)
	suite.Equal(model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        nil,
		LowerDenominator:  nil,
		HigherType:        "open",
		HigherValue:       nil,
		HigherDenominator: nil,
		Units:             "Units",
		Precision:         0.01,
	}, span)
}

func (suite *AtomicSpanSuite) TestUpdate() {
	err := AddAtomicSpan(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	err = UpdateAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.AtomicSpan{
		LowerType:         "open",
		LowerValue:        t_IntPtr(10),
		LowerDenominator:  t_IntPtr(20),
		HigherType:        "closed",
		HigherValue:       t_IntPtr(30),
		HigherDenominator: t_IntPtr(40),
		Units:             "UnitsX",
		Precision:         0.001,
	})
	suite.Require().NoError(err)

	parentDataTypeKey, span, err := LoadAtomicSpan(suite.db, suite.model.Key, "data_type_key")
	suite.Require().NoError(err)
	suite.Equal("data_type_key", parentDataTypeKey)
	suite.Equal(model_data_type.AtomicSpan{
		LowerType:         "open",
		LowerValue:        t_IntPtr(10),
		LowerDenominator:  t_IntPtr(20),
		HigherType:        "closed",
		HigherValue:       t_IntPtr(30),
		HigherDenominator: t_IntPtr(40),
		Units:             "UnitsX",
		Precision:         0.001,
	}, span)
}

func (suite *AtomicSpanSuite) TestUpdateNulls() {
	err := AddAtomicSpan(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	updatedSpan := model_data_type.AtomicSpan{
		LowerType:         "open",
		LowerValue:        nil,
		LowerDenominator:  nil,
		HigherType:        "closed",
		HigherValue:       nil,
		HigherDenominator: nil,
		Units:             "UnitsX",
		Precision:         0.001,
	}
	err = UpdateAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), updatedSpan)
	suite.Require().NoError(err)

	parentDataTypeKey, span, err := LoadAtomicSpan(suite.db, suite.model.Key, "data_type_key")
	suite.Require().NoError(err)
	suite.Equal("data_type_key", parentDataTypeKey)
	suite.Equal(model_data_type.AtomicSpan{
		LowerType:         "open",
		LowerValue:        nil,
		LowerDenominator:  nil,
		HigherType:        "closed",
		HigherValue:       nil,
		HigherDenominator: nil,
		Units:             "UnitsX",
		Precision:         0.001,
	}, span)
}

func (suite *AtomicSpanSuite) TestRemove() {
	err := AddAtomicSpan(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	err = RemoveAtomicSpan(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key)) // Test case-insensitive.
	suite.Require().NoError(err)

	_, _, err = LoadAtomicSpan(suite.db, suite.model.Key, suite.dataType.Key)
	suite.Require().ErrorIs(err, ErrNotFound)
}

func (suite *AtomicSpanSuite) TestQuery() {
	err := AddAtomicSpan(suite.db, suite.model.Key, suite.dataTypeB.Key, model_data_type.AtomicSpan{
		LowerType:         "open",
		LowerValue:        t_IntPtr(10),
		LowerDenominator:  t_IntPtr(20),
		HigherType:        "closed",
		HigherValue:       t_IntPtr(30),
		HigherDenominator: t_IntPtr(40),
		Units:             "UnitsX",
		Precision:         0.001,
	})
	suite.Require().NoError(err)

	err = AddAtomicSpan(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(3),
		HigherDenominator: t_IntPtr(4),
		Units:             "Units",
		Precision:         0.01,
	})
	suite.Require().NoError(err)

	atomicSpans, err := QueryAtomicSpans(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	suite.Require().NoError(err)
	suite.Equal(map[string]model_data_type.AtomicSpan{
		"data_type_key": {
			LowerType:         "closed",
			LowerValue:        t_IntPtr(1),
			LowerDenominator:  t_IntPtr(2),
			HigherType:        "open",
			HigherValue:       t_IntPtr(3),
			HigherDenominator: t_IntPtr(4),
			Units:             "Units",
			Precision:         0.01,
		},
		"data_type_key_b": {
			LowerType:         "open",
			LowerValue:        t_IntPtr(10),
			LowerDenominator:  t_IntPtr(20),
			HigherType:        "closed",
			HigherValue:       t_IntPtr(30),
			HigherDenominator: t_IntPtr(40),
			Units:             "UnitsX",
			Precision:         0.001,
		},
	}, atomicSpans)
}

func (suite *AtomicSpanSuite) TestBulkInsertAtomicSpans() {
	err := BulkInsertAtomicSpans(suite.db, strings.ToUpper(suite.model.Key), map[string]model_data_type.AtomicSpan{
		"data_type_key": {
			LowerType:         "closed",
			LowerValue:        t_IntPtr(1),
			LowerDenominator:  t_IntPtr(2),
			HigherType:        "open",
			HigherValue:       t_IntPtr(3),
			HigherDenominator: t_IntPtr(4),
			Units:             "Units",
			Precision:         0.01,
		},
		"data_type_key_b": {
			LowerType:         "open",
			LowerValue:        t_IntPtr(10),
			LowerDenominator:  t_IntPtr(20),
			HigherType:        "closed",
			HigherValue:       t_IntPtr(30),
			HigherDenominator: t_IntPtr(40),
			Units:             "UnitsX",
			Precision:         0.001,
		},
	})
	suite.Require().NoError(err)

	atomicSpans, err := QueryAtomicSpans(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[string]model_data_type.AtomicSpan{
		"data_type_key": {
			LowerType:         "closed",
			LowerValue:        t_IntPtr(1),
			LowerDenominator:  t_IntPtr(2),
			HigherType:        "open",
			HigherValue:       t_IntPtr(3),
			HigherDenominator: t_IntPtr(4),
			Units:             "Units",
			Precision:         0.01,
		},
		"data_type_key_b": {
			LowerType:         "open",
			LowerValue:        t_IntPtr(10),
			LowerDenominator:  t_IntPtr(20),
			HigherType:        "closed",
			HigherValue:       t_IntPtr(30),
			HigherDenominator: t_IntPtr(40),
			Units:             "UnitsX",
			Precision:         0.001,
		},
	}, atomicSpans)
}
