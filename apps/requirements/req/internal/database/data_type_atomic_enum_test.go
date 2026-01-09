package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_data_type"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAtomicEnumSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AtomicEnumSuite))
}

type AtomicEnumSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	dataType  model_data_type.DataType
	dataTypeB model_data_type.DataType
	atomic    model_data_type.Atomic
	atomicB   model_data_type.Atomic
}

func (suite *AtomicEnumSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.dataType = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key")
	suite.dataTypeB = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key_b")
	suite.atomic = t_AddAtomic(suite.T(), suite.db, suite.model.Key, suite.dataType.Key, "enumeration", nil, t_BoolPtr(true), nil)
	suite.atomicB = t_AddAtomic(suite.T(), suite.db, suite.model.Key, suite.dataTypeB.Key, "enumeration", nil, t_BoolPtr(true), nil)
}

func (suite *AtomicEnumSuite) TestLoad() {

	// Nothing in database yet.
	atomicEnums, err := LoadAtomicEnums(suite.db, strings.ToUpper(suite.model.Key), "data_type_key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), atomicEnums)

	_, err = dbExec(suite.db, `
		INSERT INTO data_type_atomic_enum_value
			(
				model_key,
				data_type_key,
				value,
				sort_order
			)
		VALUES
			(
				'model_key',
				'data_type_key',
				'Value1',
				1
			),
			(
				'model_key',
				'data_type_key',
				'Value2',
				2
			)
	`)
	assert.Nil(suite.T(), err)

	atomicEnums, err = LoadAtomicEnums(suite.db, strings.ToUpper(suite.model.Key), "data_Type_Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_data_type.AtomicEnum{
		"data_type_key": {
			{Value: "Value1", SortOrder: 1},
			{Value: "Value2", SortOrder: 2},
		},
	}, atomicEnums)
}

func (suite *AtomicEnumSuite) TestAdd() {

	err := AddAtomicEnum(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.AtomicEnum{
		Value:     "Value1",
		SortOrder: 1,
	})
	assert.Nil(suite.T(), err)

	atomicEnums, err := LoadAtomicEnums(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_data_type.AtomicEnum{
		suite.dataType.Key: {
			{Value: "Value1", SortOrder: 1},
		},
	}, atomicEnums)
}

func (suite *AtomicEnumSuite) TestUpdate() {

	err := AddAtomicEnum(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicEnum{
		Value:     "Value1",
		SortOrder: 1,
	})
	assert.Nil(suite.T(), err)

	err = UpdateAtomicEnum(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), "Value1", model_data_type.AtomicEnum{
		Value:     "Value1Updated",
		SortOrder: 10,
	})
	assert.Nil(suite.T(), err)

	atomicEnums, err := LoadAtomicEnums(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_data_type.AtomicEnum{
		suite.dataType.Key: {
			{Value: "Value1Updated", SortOrder: 10},
		},
	}, atomicEnums)
}

func (suite *AtomicEnumSuite) TestRemove() {

	err := AddAtomicEnum(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicEnum{
		Value:     "Value1",
		SortOrder: 1,
	})
	assert.Nil(suite.T(), err)

	err = RemoveAtomicEnum(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), "Value1") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	atomicEnums, err := LoadAtomicEnums(suite.db, suite.model.Key, suite.dataType.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), atomicEnums)
}

func (suite *AtomicEnumSuite) TestQuery() {

	err := AddAtomicEnum(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.AtomicEnum{
		Value:     "Value",
		SortOrder: 1,
	})
	assert.Nil(suite.T(), err)

	// Add another data type and atomic enum
	err = AddAtomicEnum(suite.db, suite.model.Key, suite.dataTypeB.Key, model_data_type.AtomicEnum{
		Value:     "Value",
		SortOrder: 10,
	})
	assert.Nil(suite.T(), err)

	atomicEnums, err := QueryAtomicEnums(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_data_type.AtomicEnum{
		suite.dataType.Key: {
			{Value: "Value", SortOrder: 1},
		},
		suite.dataTypeB.Key: {
			{Value: "Value", SortOrder: 10},
		},
	}, atomicEnums)
}

func (suite *AtomicEnumSuite) TestBulkInsertAtomicEnums() {

	err := BulkInsertAtomicEnums(suite.db, strings.ToUpper(suite.model.Key), map[string][]model_data_type.AtomicEnum{
		"data_type_key": {
			{Value: "Value1", SortOrder: 1},
			{Value: "Value2", SortOrder: 2},
		},
		"data_type_key_b": {
			{Value: "Value1", SortOrder: 10},
		},
	})
	assert.NoError(suite.T(), err)

	atomicEnums, err := QueryAtomicEnums(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_data_type.AtomicEnum{
		"data_type_key": {
			{Value: "Value1", SortOrder: 1},
			{Value: "Value2", SortOrder: 2},
		},
		"data_type_key_b": {
			{Value: "Value1", SortOrder: 10},
		},
	}, atomicEnums)
}
