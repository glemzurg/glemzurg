package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/suite"
)

func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

func TestTopLevelDataTypeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(TopLevelDataTypeSuite))
}

type TopLevelDataTypeSuite struct {
	suite.Suite
	db    *sql.DB
	model core.Model
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

		// Unordered enumeration (enum of value1, value2).
		t_rawDtKey("enum_type").String(): {
			Key:            t_rawDtKey("enum_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    boolPtr(false),
				Enums: []model_data_type.AtomicEnum{
					{Value: "value1", SortOrder: 0},
					{Value: "value2", SortOrder: 1},
				},
			},
		},

		// Ordered enumeration (ordered enum of low, medium, high, critical).
		t_rawDtKey("ordered_enum_type").String(): {
			Key:            t_rawDtKey("ordered_enum_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    boolPtr(true),
				Enums: []model_data_type.AtomicEnum{
					{Value: "low", SortOrder: 0},
					{Value: "medium", SortOrder: 1},
					{Value: "high", SortOrder: 2},
					{Value: "critical", SortOrder: 3},
				},
			},
		},

		// Reference (ref from domain_a>subdomain_a>product).
		t_rawDtKey("ref_type").String(): {
			Key:            t_rawDtKey("ref_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "reference",
				Reference:      strPtr("domain_a>subdomain_a>product"),
			},
		},

		// Atomic unconstrained.
		t_rawDtKey("unconstrained_type").String(): {
			Key:            t_rawDtKey("unconstrained_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},

		// Atomic datetime.
		t_rawDtKey("datetime_type").String(): {
			Key:            t_rawDtKey("datetime_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "datetime",
			},
		},

		// Nested records.
		t_rawDtKey("root1").String(): {
			Key:            t_rawDtKey("root1"),
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

		t_rawDtKey("root2").String(): {
			Key:            t_rawDtKey("root2"),
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

		// Span with unconstrained bounds.
		t_rawDtKey("span_type").String(): {
			Key:            t_rawDtKey("span_type"),
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

		// Span with closed numeric bounds ([1 .. 10000] at 1 unit).
		t_rawDtKey("span_closed_type").String(): {
			Key:            t_rawDtKey("span_closed_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "closed",
					LowerValue:        intPtr(1),
					LowerDenominator:  intPtr(1),
					HigherType:        "closed",
					HigherValue:       intPtr(10000),
					HigherDenominator: intPtr(1),
					Units:             "unit",
					Precision:         1.0,
				},
			},
		},

		// Span with unconstrained lower, closed higher ((unconstrained .. 100] at 1 unit).
		t_rawDtKey("span_mixed_type").String(): {
			Key:            t_rawDtKey("span_mixed_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "unconstrained",
					HigherType:        "closed",
					HigherValue:       intPtr(100),
					HigherDenominator: intPtr(1),
					Units:             "unit",
					Precision:         1.0,
				},
			},
		},

		// Span with open lower, precision=0.01 ((0 .. 1000000] at 0.01 dollar).
		t_rawDtKey("span_precision_type").String(): {
			Key:            t_rawDtKey("span_precision_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "open",
					LowerValue:        intPtr(0),
					LowerDenominator:  intPtr(1),
					HigherType:        "closed",
					HigherValue:       intPtr(1000000),
					HigherDenominator: intPtr(1),
					Units:             "dollar",
					Precision:         0.01,
				},
			},
		},

		// Unique unordered collection of unconstrained (unique unordered of unconstrained).
		t_rawDtKey("unordered_collection_type").String(): {
			Key:              t_rawDtKey("unordered_collection_type"),
			CollectionType:   "unordered",
			CollectionUnique: boolPtr(true),
			CollectionMin:    intPtr(0),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},

		// Ordered collection with min/max, object atomic (1-100 ordered of obj of some_class).
		t_rawDtKey("ordered_collection_type").String(): {
			Key:            t_rawDtKey("ordered_collection_type"),
			CollectionType: "ordered",
			CollectionMin:  intPtr(1),
			CollectionMax:  intPtr(100),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "object",
				ObjectClassKey: strPtr("some_class"),
			},
		},

		// Ordered collection with min-only (3+ ordered of unconstrained).
		t_rawDtKey("ordered_min_collection_type").String(): {
			Key:            t_rawDtKey("ordered_min_collection_type"),
			CollectionType: "ordered",
			CollectionMin:  intPtr(3),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},
	})
	suite.Require().NoError(err)

	// Load from database
	loaded, err := LoadTopLevelDataTypes(suite.db, suite.model.Key)
	suite.Require().NoError(err)

	// Verify that loaded matches original.
	suite.Equal(map[string]model_data_type.DataType{

		t_rawDtKey("enum_type").String(): {
			Key:            t_rawDtKey("enum_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    boolPtr(false),
				Enums: []model_data_type.AtomicEnum{
					{Value: "value1", SortOrder: 0},
					{Value: "value2", SortOrder: 1},
				},
			},
		},

		t_rawDtKey("ordered_enum_type").String(): {
			Key:            t_rawDtKey("ordered_enum_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    boolPtr(true),
				Enums: []model_data_type.AtomicEnum{
					{Value: "low", SortOrder: 0},
					{Value: "medium", SortOrder: 1},
					{Value: "high", SortOrder: 2},
					{Value: "critical", SortOrder: 3},
				},
			},
		},

		t_rawDtKey("ref_type").String(): {
			Key:            t_rawDtKey("ref_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "reference",
				Reference:      strPtr("domain_a>subdomain_a>product"),
			},
		},

		t_rawDtKey("unconstrained_type").String(): {
			Key:            t_rawDtKey("unconstrained_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},

		t_rawDtKey("datetime_type").String(): {
			Key:            t_rawDtKey("datetime_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "datetime",
			},
		},

		t_rawDtKey("root1").String(): {
			Key:            t_rawDtKey("root1"),
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						Key:            helper.Must(identity.NewDataTypeKey(t_rawDtKey("root1"), "child_field")),
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									Key:            helper.Must(identity.NewDataTypeKey(helper.Must(identity.NewDataTypeKey(t_rawDtKey("root1"), "child_field")), "grandchild_field")),
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		t_rawDtKey("root2").String(): {
			Key:            t_rawDtKey("root2"),
			CollectionType: "record",
			RecordFields: []model_data_type.Field{
				{
					Name: "child_field",
					FieldDataType: &model_data_type.DataType{
						Key:            helper.Must(identity.NewDataTypeKey(t_rawDtKey("root2"), "child_field")),
						CollectionType: "record",
						RecordFields: []model_data_type.Field{
							{
								Name: "grandchild_field",
								FieldDataType: &model_data_type.DataType{
									Key:            helper.Must(identity.NewDataTypeKey(helper.Must(identity.NewDataTypeKey(t_rawDtKey("root2"), "child_field")), "grandchild_field")),
									CollectionType: "atomic",
									Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
								},
							},
						},
					},
				},
			},
		},

		t_rawDtKey("span_type").String(): {
			Key:            t_rawDtKey("span_type"),
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

		t_rawDtKey("span_closed_type").String(): {
			Key:            t_rawDtKey("span_closed_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "closed",
					LowerValue:        intPtr(1),
					LowerDenominator:  intPtr(1),
					HigherType:        "closed",
					HigherValue:       intPtr(10000),
					HigherDenominator: intPtr(1),
					Units:             "unit",
					Precision:         1.0,
				},
			},
		},

		t_rawDtKey("span_mixed_type").String(): {
			Key:            t_rawDtKey("span_mixed_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "unconstrained",
					HigherType:        "closed",
					HigherValue:       intPtr(100),
					HigherDenominator: intPtr(1),
					Units:             "unit",
					Precision:         1.0,
				},
			},
		},

		t_rawDtKey("span_precision_type").String(): {
			Key:            t_rawDtKey("span_precision_type"),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "span",
				Span: &model_data_type.AtomicSpan{
					LowerType:         "open",
					LowerValue:        intPtr(0),
					LowerDenominator:  intPtr(1),
					HigherType:        "closed",
					HigherValue:       intPtr(1000000),
					HigherDenominator: intPtr(1),
					Units:             "dollar",
					Precision:         0.01,
				},
			},
		},

		// CollectionMin=0 is written as NULL (0 means "no minimum"), loaded back as nil.
		t_rawDtKey("unordered_collection_type").String(): {
			Key:              t_rawDtKey("unordered_collection_type"),
			CollectionType:   "unordered",
			CollectionUnique: boolPtr(true),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},

		t_rawDtKey("ordered_collection_type").String(): {
			Key:            t_rawDtKey("ordered_collection_type"),
			CollectionType: "ordered",
			CollectionMin:  intPtr(1),
			CollectionMax:  intPtr(100),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "object",
				ObjectClassKey: strPtr("some_class"),
			},
		},

		t_rawDtKey("ordered_min_collection_type").String(): {
			Key:            t_rawDtKey("ordered_min_collection_type"),
			CollectionType: "ordered",
			CollectionMin:  intPtr(3),
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},
	}, loaded)
}
