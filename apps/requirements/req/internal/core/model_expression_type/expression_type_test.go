package model_expression_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type ExpressionTypeTestSuite struct {
	suite.Suite
}

func TestExpressionTypeSuite(t *testing.T) {
	suite.Run(t, new(ExpressionTypeTestSuite))
}

// validClassKey returns a valid identity.Key for testing.
func validClassKey() identity.Key {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "c")
	return classKey
}

func (s *ExpressionTypeTestSuite) TestValidateScalars() {
	tests := []struct {
		testName string
		et       ExpressionType
		errstr   string
	}{
		{testName: "valid boolean", et: &BooleanType{}},
		{testName: "valid integer", et: &IntegerType{}},
		{testName: "valid rational", et: &RationalType{}},
		{testName: "valid string", et: &StringType{}},
		{testName: "valid enum", et: &EnumType{Values: []string{"a", "b"}}},
		{testName: "error enum empty", et: &EnumType{}, errstr: "Values"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.et.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTypeTestSuite) TestValidateCollections() {
	tests := []struct {
		testName string
		et       ExpressionType
		errstr   string
	}{
		{testName: "valid set", et: &SetType{ElementType: &IntegerType{}}},
		{testName: "error set nil element", et: &SetType{}, errstr: "SetType.ElementType: is required"},
		{testName: "valid sequence", et: &SequenceType{ElementType: &StringType{}}},
		{testName: "valid sequence unique", et: &SequenceType{ElementType: &StringType{}, Unique: true}},
		{testName: "error sequence nil element", et: &SequenceType{}, errstr: "SequenceType.ElementType: is required"},
		{testName: "valid bag", et: &BagType{ElementType: &IntegerType{}}},
		{testName: "error bag nil element", et: &BagType{}, errstr: "BagType.ElementType: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.et.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTypeTestSuite) TestValidateCompound() {
	tests := []struct {
		testName string
		et       ExpressionType
		errstr   string
	}{
		{testName: "valid tuple", et: &TupleType{ElementTypes: []ExpressionType{&IntegerType{}, &StringType{}}}},
		{testName: "error tuple empty", et: &TupleType{}, errstr: "ElementTypes"},
		{testName: "error tuple nil element", et: &TupleType{ElementTypes: []ExpressionType{nil}}, errstr: "TupleType.ElementTypes[0]: is required"},
		{testName: "valid record", et: &RecordType{Fields: []RecordFieldType{{Name: "x", Type: &IntegerType{}}}}},
		{testName: "error record empty", et: &RecordType{}, errstr: "Fields"},
		{testName: "error record missing name", et: &RecordType{Fields: []RecordFieldType{{Type: &IntegerType{}}}}, errstr: "RecordType.Fields[0].Name: is required"},
		{testName: "error record nil type", et: &RecordType{Fields: []RecordFieldType{{Name: "x"}}}, errstr: "RecordType.Fields[0].Type: is required"},
		{testName: "valid function no params", et: &FunctionType{Return: &IntegerType{}}},
		{testName: "valid function with params", et: &FunctionType{Params: []ExpressionType{&IntegerType{}, &IntegerType{}}, Return: &IntegerType{}}},
		{testName: "error function nil return", et: &FunctionType{}, errstr: "FunctionType.Return: is required"},
		{testName: "error function nil param", et: &FunctionType{Params: []ExpressionType{nil}, Return: &IntegerType{}}, errstr: "FunctionType.Params[0]: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.et.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTypeTestSuite) TestValidateReferences() {
	tests := []struct {
		testName string
		et       ExpressionType
		errstr   string
	}{
		{testName: "valid object", et: &ObjectType{ClassKey: validClassKey()}},
		{testName: "error object empty key", et: &ObjectType{}, errstr: "ObjectType.ClassKey"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.et.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTypeTestSuite) TestValidateNested() {
	// Deep nesting: Set of Sequence of Record with Tuple field.
	et := &SetType{
		ElementType: &SequenceType{
			ElementType: &RecordType{
				Fields: []RecordFieldType{
					{Name: "coords", Type: &TupleType{ElementTypes: []ExpressionType{&RationalType{}, &RationalType{}}}},
					{Name: "label", Type: &StringType{}},
				},
			},
		},
	}
	s.NoError(et.Validate())

	// Nested with error deep inside.
	etBad := &SetType{
		ElementType: &SequenceType{
			ElementType: &RecordType{
				Fields: []RecordFieldType{
					{Name: "coords", Type: &TupleType{}}, // Empty TupleType.
				},
			},
		},
	}
	err := etBad.Validate()
	s.Error(err)
	s.Contains(err.Error(), "ElementTypes")
}

func (s *ExpressionTypeTestSuite) TestTypeName() {
	s.Equal(TypeBoolean, (&BooleanType{}).TypeName())
	s.Equal(TypeInteger, (&IntegerType{}).TypeName())
	s.Equal(TypeRational, (&RationalType{}).TypeName())
	s.Equal(TypeString, (&StringType{}).TypeName())
	s.Equal(TypeEnum, (&EnumType{Values: []string{"a"}}).TypeName())
	s.Equal(TypeSet, (&SetType{ElementType: &IntegerType{}}).TypeName())
	s.Equal(TypeSequence, (&SequenceType{ElementType: &IntegerType{}}).TypeName())
	s.Equal(TypeBag, (&BagType{ElementType: &IntegerType{}}).TypeName())
	s.Equal(TypeTuple, (&TupleType{ElementTypes: []ExpressionType{&IntegerType{}}}).TypeName())
	s.Equal(TypeRecord, (&RecordType{Fields: []RecordFieldType{{Name: "x", Type: &IntegerType{}}}}).TypeName())
	s.Equal(TypeFunction, (&FunctionType{Return: &IntegerType{}}).TypeName())
	s.Equal(TypeObject, (&ObjectType{ClassKey: validClassKey()}).TypeName())
}

func (s *ExpressionTypeTestSuite) TestValidateExpressionTypeHelper() {
	// Nil is valid.
	s.NoError(ValidateExpressionType(nil))
	// Valid type.
	s.NoError(ValidateExpressionType(&BooleanType{}))
	// Invalid type.
	err := ValidateExpressionType(&SetType{})
	s.Error(err)
}
