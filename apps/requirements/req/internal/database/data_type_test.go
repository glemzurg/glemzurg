package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDataTypeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(DataTypeSuite))
}

type DataTypeSuite struct {
	suite.Suite
	db    *sql.DB
	model req_model.Model
}

func (suite *DataTypeSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
}

func (suite *DataTypeSuite) TestLoad() {

	// Nothing in database yet.
	dataType, err := LoadDataType(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), dataType)

	_, err = dbExec(suite.db, `
		INSERT INTO data_type
			(
				model_key,
				data_type_key,
				collection_type,
				collection_unique,
				collection_min,
				collection_max
			)
		VALUES
			(
				'model_key',
				'key',
				'atomic',
				true,
				5,
				10
			)
	`)
	assert.Nil(suite.T(), err)

	dataType, err = LoadDataType(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:              "key", // Test case-insensitive.
		CollectionType:   "atomic",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	}, dataType)
}

func (suite *DataTypeSuite) TestAdd() {

	err := AddDataType(suite.db, strings.ToUpper(suite.model.Key), model_data_type.DataType{
		Key:              "KeY", // Test case-insensitive.
		CollectionType:   "record",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	})
	assert.Nil(suite.T(), err)

	dataType, err := LoadDataType(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:              "key",
		CollectionType:   "record",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	}, dataType)
}

func (suite *DataTypeSuite) TestAddNulls() {

	err := AddDataType(suite.db, strings.ToUpper(suite.model.Key), model_data_type.DataType{
		Key:              "KeY", // Test case-insensitive.
		CollectionType:   "unordered",
		CollectionUnique: nil,
		CollectionMin:    nil,
		CollectionMax:    nil,
	})
	assert.Nil(suite.T(), err)

	dataType, err := LoadDataType(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:              "key",
		CollectionType:   "unordered",
		CollectionUnique: nil,
		CollectionMin:    nil,
		CollectionMax:    nil,
	}, dataType)
}

func (suite *DataTypeSuite) TestUpdate() {

	err := AddDataType(suite.db, suite.model.Key, model_data_type.DataType{
		Key:              "key",
		CollectionType:   "atomic",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	})
	assert.Nil(suite.T(), err)

	err = UpdateDataType(suite.db, strings.ToUpper(suite.model.Key), model_data_type.DataType{
		Key:              "kEy", // Test case-insensitive.
		CollectionType:   "stack",
		CollectionUnique: t_BoolPtr(false),
		CollectionMin:    t_IntPtr(15),
		CollectionMax:    t_IntPtr(20),
	})
	assert.Nil(suite.T(), err)

	dataType, err := LoadDataType(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:              "key",
		CollectionType:   "stack",
		CollectionUnique: t_BoolPtr(false),
		CollectionMin:    t_IntPtr(15),
		CollectionMax:    t_IntPtr(20),
	}, dataType)
}

func (suite *DataTypeSuite) TestUpdateNulls() {

	err := AddDataType(suite.db, suite.model.Key, model_data_type.DataType{
		Key:              "key",
		CollectionType:   "atomic",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	})
	assert.Nil(suite.T(), err)

	err = UpdateDataType(suite.db, strings.ToUpper(suite.model.Key), model_data_type.DataType{
		Key:              "kEy", // Test case-insensitive.
		CollectionType:   "queue",
		CollectionUnique: nil,
		CollectionMin:    nil,
		CollectionMax:    nil,
	})
	assert.Nil(suite.T(), err)

	dataType, err := LoadDataType(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:              "key",
		CollectionType:   "queue",
		CollectionUnique: nil,
		CollectionMin:    nil,
		CollectionMax:    nil,
	}, dataType)
}

func (suite *DataTypeSuite) TestDelete() {

	err := AddDataType(suite.db, suite.model.Key, model_data_type.DataType{
		Key:              "key",
		CollectionType:   "atomic",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	})
	assert.Nil(suite.T(), err)

	err = DeleteDataType(suite.db, strings.ToUpper(suite.model.Key), "KeY") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	dataType, err := LoadDataType(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), dataType)
}

func (suite *DataTypeSuite) TestQuery() {

	// Add some data types.
	err := AddDataType(suite.db, suite.model.Key, model_data_type.DataType{
		Key:              "key2",
		CollectionType:   "record",
		CollectionUnique: t_BoolPtr(false),
		CollectionMin:    t_IntPtr(15),
		CollectionMax:    t_IntPtr(20),
	})
	assert.Nil(suite.T(), err)

	err = AddDataType(suite.db, suite.model.Key, model_data_type.DataType{
		Key:              "key1",
		CollectionType:   "atomic",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(5),
		CollectionMax:    t_IntPtr(10),
	})
	assert.Nil(suite.T(), err)

	dataTypes, err := QueryDataTypes(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_data_type.DataType{
		{
			Key:              "key1",
			CollectionType:   "atomic",
			CollectionUnique: t_BoolPtr(true),
			CollectionMin:    t_IntPtr(5),
			CollectionMax:    t_IntPtr(10),
		},
		{
			Key:              "key2",
			CollectionType:   "record",
			CollectionUnique: t_BoolPtr(false),
			CollectionMin:    t_IntPtr(15),
			CollectionMax:    t_IntPtr(20),
		},
	}, dataTypes)
}

func (suite *DataTypeSuite) TestBulkInsertDataTypes() {

	err := BulkInsertDataTypes(suite.db, strings.ToUpper(suite.model.Key), []model_data_type.DataType{
		{
			Key:              "key1",
			CollectionType:   "atomic",
			CollectionUnique: t_BoolPtr(true),
			CollectionMin:    t_IntPtr(5),
			CollectionMax:    t_IntPtr(10),
		},
		{
			Key:              "key2",
			CollectionType:   "record",
			CollectionUnique: t_BoolPtr(false),
			CollectionMin:    t_IntPtr(15),
			CollectionMax:    t_IntPtr(20),
		},
	})
	assert.NoError(suite.T(), err)

	dataTypes, err := QueryDataTypes(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_data_type.DataType{
		{
			Key:              "key1",
			CollectionType:   "atomic",
			CollectionUnique: t_BoolPtr(true),
			CollectionMin:    t_IntPtr(5),
			CollectionMax:    t_IntPtr(10),
		},
		{
			Key:              "key2",
			CollectionType:   "record",
			CollectionUnique: t_BoolPtr(false),
			CollectionMin:    t_IntPtr(15),
			CollectionMax:    t_IntPtr(20),
		},
	}, dataTypes)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddDataType(t *testing.T, dbOrTx DbOrTx, modelKey, dataTypeKey string) (dataType model_data_type.DataType) {

	err := AddDataType(dbOrTx, modelKey, model_data_type.DataType{
		Key:            dataTypeKey,
		CollectionType: "atomic",
	})
	assert.Nil(t, err)

	dataType, err = LoadDataType(dbOrTx, modelKey, dataTypeKey)
	assert.Nil(t, err)

	return dataType
}

func (suite *DataTypeSuite) TestVerifyTestObjects() {

	dataType := t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key")
	assert.Equal(suite.T(), model_data_type.DataType{
		Key:            "data_type_key",
		CollectionType: "atomic",
	}, dataType)

}
