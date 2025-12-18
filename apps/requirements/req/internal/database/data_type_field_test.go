package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDataTypeFieldSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(DataTypeFieldSuite))
}

type DataTypeFieldSuite struct {
	suite.Suite
	db         *sql.DB
	model      requirements.Model
	dataType   data_type.DataType
	dataTypeB  data_type.DataType
	fieldType  data_type.DataType
	fieldTypeB data_type.DataType
}

func (suite *DataTypeFieldSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.dataType = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key")
	suite.dataTypeB = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key_b")
	suite.fieldType = t_AddDataType(suite.T(), suite.db, suite.model.Key, "field_data_type_key")
	suite.fieldTypeB = t_AddDataType(suite.T(), suite.db, suite.model.Key, "field_data_type_key_b")
}

func (suite *DataTypeFieldSuite) TestLoad() {

	// Nothing in database yet.
	fields, err := LoadDataTypeFields(suite.db, strings.ToUpper(suite.model.Key), "data_type_key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), fields)

	_, err = dbExec(suite.db, `
		INSERT INTO data_type_field
			(
				model_key,
				data_type_key,
				name,
				field_data_type_key
			)
		VALUES
			(
				'model_key',
				'data_type_key',
				'NameA',
				'field_data_type_key'
			),
			(
				'model_key',
				'data_type_key',
				'NameB',
				'field_data_type_key'
			)
	`)
	assert.Nil(suite.T(), err)

	fields, err = LoadDataTypeFields(suite.db, strings.ToUpper(suite.model.Key), "data_TYPE_Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
			{
				Name:          "NameB",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
		},
	}, fields)
}

func (suite *DataTypeFieldSuite) TestAdd() {

	err := AddField(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), data_type.Field{
		Name:          "NameA",
		FieldDataType: &data_type.DataType{Key: "field_DATA_type_key"}, // Test case-insensitive..
	})
	assert.Nil(suite.T(), err)

	fields, err := LoadDataTypeFields(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
		},
	}, fields)
}

func (suite *DataTypeFieldSuite) TestUpdate() {

	err := AddField(suite.db, suite.model.Key, suite.dataType.Key, data_type.Field{
		Name:          "NameA",
		FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
	})
	assert.Nil(suite.T(), err)

	err = UpdateField(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), data_type.Field{
		Name:          "NameA",
		FieldDataType: &data_type.DataType{Key: "field_data_TYPE_key_b"}, // Test case-insensitive..
	})
	assert.Nil(suite.T(), err)

	fields, err := LoadDataTypeFields(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
	}, fields)
}

func (suite *DataTypeFieldSuite) TestRemove() {

	err := AddField(suite.db, suite.model.Key, suite.dataType.Key, data_type.Field{
		Name:          "NameA",
		FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
	})
	assert.Nil(suite.T(), err)

	err = RemoveField(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), "NameA")
	assert.Nil(suite.T(), err)

	fields, err := LoadDataTypeFields(suite.db, suite.model.Key, suite.dataType.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), fields)
}

func (suite *DataTypeFieldSuite) TestQuery() {

	err := AddField(suite.db, suite.model.Key, suite.dataType.Key, data_type.Field{
		Name:          "NameB",
		FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
	})
	assert.Nil(suite.T(), err)

	// Add another data type and field
	err = AddField(suite.db, suite.model.Key, suite.dataType.Key, data_type.Field{
		Name:          "NameA",
		FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
	})
	assert.Nil(suite.T(), err)

	fields, err := QueryFields(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
			{
				Name:          "NameB",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
	}, fields)
}

func (suite *DataTypeFieldSuite) TestBulkInsertFields() {
	err := BulkInsertFields(suite.db, strings.ToUpper(suite.model.Key), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
			{
				Name:          "NameB",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
		"data_type_key_b": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
	})
	assert.NoError(suite.T(), err)

	fields, err := QueryFields(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]data_type.Field{
		"data_type_key": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key"},
			},
			{
				Name:          "NameB",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
		"data_type_key_b": {
			{
				Name:          "NameA",
				FieldDataType: &data_type.DataType{Key: "field_data_type_key_b"},
			},
		},
	}, fields)
}
