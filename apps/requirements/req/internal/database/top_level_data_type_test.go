package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_data_type"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTopLevelDataTypeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(TopLevelDataTypeSuite))
}

type TopLevelDataTypeSuite struct {
	suite.Suite
	db    *sql.DB
	model requirements.Model
}

func (suite *TopLevelDataTypeSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
}

func (suite *TopLevelDataTypeSuite) TestAddAndLoadTopLevelDataTypes() {

	// The nested child and grandchild keys are created on the insertion.

	// Add to database
	err := AddTopLevelDataTypes(suite.db, suite.model.Key, map[string]model_data_type.DataType{

		"enum_type": model_data_type.DataType{
			Key:            "enum_type",
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				Enums: []model_data_type.AtomicEnum{
					{Value: "value1"},
					{Value: "value2"},
				},
			},
		},

		"root1": model_data_type.DataType{
			Key:            "root1",
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		"root2": model_data_type.DataType{
			Key:            "root2",
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		"span_type": model_data_type.DataType{
			Key:            "span_type",
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:  "unconstrained",
					HigherType: "unconstrained",
					Precision:  1.0,
				},
			},
		},
	})
	assert.NoError(suite.T(), err)

	// Load from database
	loaded, err := LoadTopLevelDataTypes(suite.db, suite.model.Key)
	assert.NoError(suite.T(), err)

	// Verify that loaded matches original
	assert.Equal(suite.T(), map[string]model_data_type.DataType{

		"enum_type": model_data_type.DataType{
			Key:            "enum_type",
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				Enums: []model_data_type.AtomicEnum{
					{Value: "value1"},
					{Value: "value2"},
				},
			},
		},

		"root1": model_data_type.DataType{
			Key:            "root1",
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						Key:            "root1/child_field",
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									Key:            "root1/child_field/grandchild_field",
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		"root2": model_data_type.DataType{
			Key:            "root2",
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						Key:            "root2/child_field",
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									Key:            "root2/child_field/grandchild_field",
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		"span_type": model_data_type.DataType{
			Key:            "span_type",
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:  "unconstrained",
					HigherType: "unconstrained",
					Precision:  1.0,
				},
			},
		},
	}, loaded)
}
