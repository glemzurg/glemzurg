package convert

import (
	"testing"

	"github.com/stretchr/testify/suite"

	met "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type RaiseTypeTestSuite struct {
	suite.Suite
	ctx *RaiseContext
}

func TestRaiseTypeSuite(t *testing.T) {
	suite.Run(t, new(RaiseTypeTestSuite))
}

func (s *RaiseTypeTestSuite) SetupTest() {
	s.ctx = &RaiseContext{}
}

func (s *RaiseTypeTestSuite) TestRaiseTypeBooleanType() {
	result, err := RaiseType(&met.BooleanType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("BOOLEAN", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeIntegerType() {
	result, err := RaiseType(&met.IntegerType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Int", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeRationalType() {
	result, err := RaiseType(&met.RationalType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Real", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeStringType() {
	result, err := RaiseType(&met.StringType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("STRING", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeEnumType() {
	result, err := RaiseType(&met.EnumType{Values: []string{"a", "b", "c"}}, s.ctx)
	s.Require().NoError(err)
	s.Equal(`{"a", "b", "c"}`, result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSetType() {
	result, err := RaiseType(&met.SetType{ElementType: &met.IntegerType{}}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Set!_Set(Int)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceType() {
	result, err := RaiseType(&met.SequenceType{ElementType: &met.StringType{}, Unique: false}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Seq!Seq(STRING)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceTypeUnique() {
	result, err := RaiseType(&met.SequenceType{ElementType: &met.StringType{}, Unique: true}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Seq!SeqUnique(STRING)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeBagType() {
	result, err := RaiseType(&met.BagType{ElementType: &met.IntegerType{}}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Bags!_Bag(Int)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeTupleType() {
	result, err := RaiseType(&met.TupleType{
		ElementTypes: []met.ExpressionType{
			&met.IntegerType{},
			&met.StringType{},
		},
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Int × STRING", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeRecordType() {
	result, err := RaiseType(&met.RecordType{
		Fields: []met.RecordFieldType{
			{Name: "name", Type: &met.StringType{}},
			{Name: "age", Type: &met.IntegerType{}},
		},
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("[name: STRING, age: Int]", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeFunctionTypeError() {
	_, err := RaiseType(&met.FunctionType{
		Params: []met.ExpressionType{&met.IntegerType{}},
		Return: &met.BooleanType{},
	}, s.ctx)
	s.Require().Error(err)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeObjectType() {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "Account")

	result, err := RaiseType(&met.ObjectType{ClassKey: classKey}, s.ctx)
	s.Require().NoError(err)
	s.Equal("account", result) // NewClassKey lowercases the SubKey
}

func (s *RaiseTypeTestSuite) TestRaiseTypeNilError() {
	_, err := RaiseType(nil, s.ctx)
	s.Require().Error(err)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeNestedSetOfRecords() {
	result, err := RaiseType(&met.SetType{
		ElementType: &met.RecordType{
			Fields: []met.RecordFieldType{
				{Name: "id", Type: &met.IntegerType{}},
				{Name: "name", Type: &met.StringType{}},
			},
		},
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Set!_Set([id: Int, name: STRING])", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceOfTuples() {
	result, err := RaiseType(&met.SequenceType{
		ElementType: &met.TupleType{
			ElementTypes: []met.ExpressionType{
				&met.IntegerType{},
				&met.BooleanType{},
			},
		},
		Unique: false,
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Seq!Seq(Int × BOOLEAN)", result)
}
