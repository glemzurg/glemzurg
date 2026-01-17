package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAtomicSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AtomicSuite))
}

type AtomicSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	dataType  model_data_type.DataType
	dataTypeB model_data_type.DataType
}

func (suite *AtomicSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.dataType = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key")
	suite.dataTypeB = t_AddDataType(suite.T(), suite.db, suite.model.Key, "data_type_key_b")
}

func (suite *AtomicSuite) TestLoad() {

	// Nothing in database yet.
	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, strings.ToUpper(suite.model.Key), "data_type_key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), parentDataTypeKey)
	assert.Empty(suite.T(), atomic)

	_, err = dbExec(suite.db, `
		INSERT INTO data_type_atomic
			(
				model_key,
				data_type_key,
				constraint_type,
				reference,
				enum_ordered,
				object_class_key
			)
		VALUES
			(
				'model_key',
				'data_type_key',
				'reference',
				'Reference',
				true,
				'ObjectClassKey'
			)
	`)
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err = LoadAtomic(suite.db, strings.ToUpper(suite.model.Key), "data_Type_Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "data_type_key", parentDataTypeKey)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	}, atomic)
}

func (suite *AtomicSuite) TestAdd() {

	err := AddAtomic(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	})
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "data_type_key", parentDataTypeKey)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	}, atomic)
}

func (suite *AtomicSuite) TestAddNulls() {

	err := AddAtomic(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      nil,
		EnumOrdered:    nil,
		ObjectClassKey: nil,
	})
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "data_type_key", parentDataTypeKey)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      nil,
		EnumOrdered:    nil,
		ObjectClassKey: nil,
	}, atomic)
}

func (suite *AtomicSuite) TestUpdate() {

	err := AddAtomic(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	})
	assert.Nil(suite.T(), err)

	err = UpdateAtomic(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.Atomic{
		ConstraintType: "object",
		Reference:      t_StrPtr("ReferenceX"),
		EnumOrdered:    t_BoolPtr(false),
		ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
	})
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "data_type_key", parentDataTypeKey)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "object",
		Reference:      t_StrPtr("ReferenceX"),
		EnumOrdered:    t_BoolPtr(false),
		ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
	}, atomic)
}

func (suite *AtomicSuite) TestUpdateNulls() {

	err := AddAtomic(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	})
	assert.Nil(suite.T(), err)

	err = UpdateAtomic(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key), model_data_type.Atomic{
		ConstraintType: "object",
		Reference:      nil,
		EnumOrdered:    nil,
		ObjectClassKey: nil,
	})
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, suite.model.Key, suite.dataType.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "data_type_key", parentDataTypeKey)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "object",
		Reference:      nil,
		EnumOrdered:    nil,
		ObjectClassKey: nil,
	}, atomic)
}

func (suite *AtomicSuite) TestRemove() {

	err := AddAtomic(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	})
	assert.Nil(suite.T(), err)

	err = RemoveAtomic(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.dataType.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	parentDataTypeKey, atomic, err := LoadAtomic(suite.db, suite.model.Key, suite.dataType.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), parentDataTypeKey)
	assert.Empty(suite.T(), atomic)
}

func (suite *AtomicSuite) TestQuery() {

	err := AddAtomic(suite.db, suite.model.Key, suite.dataTypeB.Key, model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("ReferenceX"),
		EnumOrdered:    t_BoolPtr(false),
		ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
	})

	assert.Nil(suite.T(), err)
	err = AddAtomic(suite.db, suite.model.Key, suite.dataType.Key, model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("Reference"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: t_StrPtr("ObjectClassKey"),
	})
	assert.Nil(suite.T(), err)

	atomics, err := QueryAtomics(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]model_data_type.Atomic{
		"data_type_key": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("Reference"),
			EnumOrdered:    t_BoolPtr(true),
			ObjectClassKey: t_StrPtr("ObjectClassKey"),
		},
		"data_type_key_b": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("ReferenceX"),
			EnumOrdered:    t_BoolPtr(false),
			ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
		},
	}, atomics)
}

func (suite *AtomicSuite) TestBulkInsertAtomics() {

	err := BulkInsertAtomics(suite.db, strings.ToUpper(suite.model.Key), map[string]model_data_type.Atomic{
		"data_type_key": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("Reference"),
			EnumOrdered:    t_BoolPtr(true),
			ObjectClassKey: t_StrPtr("ObjectClassKey"),
		},
		"data_type_key_b": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("ReferenceX"),
			EnumOrdered:    t_BoolPtr(false),
			ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
		},
	})
	assert.NoError(suite.T(), err)

	atomics, err := QueryAtomics(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]model_data_type.Atomic{
		"data_type_key": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("Reference"),
			EnumOrdered:    t_BoolPtr(true),
			ObjectClassKey: t_StrPtr("ObjectClassKey"),
		},
		"data_type_key_b": model_data_type.Atomic{
			ConstraintType: "reference",
			Reference:      t_StrPtr("ReferenceX"),
			EnumOrdered:    t_BoolPtr(false),
			ObjectClassKey: t_StrPtr("ObjectClassKeyX"),
		},
	}, atomics)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddAtomic(t *testing.T, dbOrTx DbOrTx, modelKey, dataTypeKey string, constraintType string, reference *string, enumOrdered *bool, objectClassKey *string) (atomic model_data_type.Atomic) {

	atomic = model_data_type.Atomic{
		ConstraintType: constraintType,
		Reference:      reference,
		EnumOrdered:    enumOrdered,
		ObjectClassKey: objectClassKey,
	}
	err := AddAtomic(dbOrTx, modelKey, dataTypeKey, atomic)
	assert.Nil(t, err)

	_, atomic, err = LoadAtomic(dbOrTx, modelKey, dataTypeKey)
	assert.Nil(t, err)

	return atomic
}

func (suite *AtomicSuite) TestVerifyTestObjects() {

	atomic := t_AddAtomic(suite.T(), suite.db, suite.model.Key, suite.dataType.Key, "reference", t_StrPtr("some_ref"), t_BoolPtr(true), nil)
	assert.Equal(suite.T(), model_data_type.Atomic{
		ConstraintType: "reference",
		Reference:      t_StrPtr("some_ref"),
		EnumOrdered:    t_BoolPtr(true),
		ObjectClassKey: nil,
	}, atomic)

}
