package convert

import (
	"testing"

	"github.com/stretchr/testify/suite"

	met "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression_type"
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
	result, err := raiseType(&met.BooleanType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("BOOLEAN", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeIntegerType() {
	result, err := raiseType(&met.IntegerType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Int", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeRationalType() {
	result, err := raiseType(&met.RationalType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Real", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeStringType() {
	result, err := raiseType(&met.StringType{}, s.ctx)
	s.Require().NoError(err)
	s.Equal("STRING", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeEnumType() {
	result, err := raiseType(&met.EnumType{Values: []string{"a", "b", "c"}}, s.ctx)
	s.Require().NoError(err)
	s.Equal(`{"a", "b", "c"}`, result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceType() {
	result, err := raiseType(&met.SequenceType{ElementType: &met.StringType{}, Unique: false}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Seq!Seq(STRING)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceTypeUnique() {
	result, err := raiseType(&met.SequenceType{ElementType: &met.StringType{}, Unique: true}, s.ctx)
	s.Require().NoError(err)
	s.Equal("_Seq!SeqUnique(STRING)", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeTupleType() {
	result, err := raiseType(&met.TupleType{
		ElementTypes: []met.ExpressionType{
			&met.IntegerType{},
			&met.StringType{},
		},
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("Int × STRING", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeRecordType() {
	result, err := raiseType(&met.RecordType{
		Fields: []met.RecordFieldType{
			{Name: "name", Type: &met.StringType{}},
			{Name: "age", Type: &met.IntegerType{}},
		},
	}, s.ctx)
	s.Require().NoError(err)
	s.Equal("[name: STRING, age: Int]", result)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeFunctionTypeError() {
	_, err := raiseType(&met.FunctionType{
		Params: []met.ExpressionType{&met.IntegerType{}},
		Return: &met.BooleanType{},
	}, s.ctx)
	s.Require().Error(err)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeObjectType() {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "Account")

	result, err := raiseType(&met.ObjectType{ClassKey: classKey}, s.ctx)
	s.Require().NoError(err)
	s.Equal("account", result) // NewClassKey lowercases the SubKey
}

func (s *RaiseTypeTestSuite) TestRaiseTypeNilError() {
	_, err := raiseType(nil, s.ctx)
	s.Require().Error(err)
}

func (s *RaiseTypeTestSuite) TestRaiseTypeSequenceOfTuples() {
	result, err := raiseType(&met.SequenceType{
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
