package model_expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type ExpressionTestSuite struct {
	suite.Suite
}

func TestExpressionSuite(t *testing.T) {
	suite.Run(t, new(ExpressionTestSuite))
}

// validAttributeKey returns a valid identity.Key for testing.
func validAttributeKey() identity.Key {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "c")
	attrKey, _ := identity.NewAttributeKey(classKey, "a")
	return attrKey
}

// validActionKey returns a valid action identity.Key for testing.
func validActionKey() identity.Key {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "c")
	actionKey, _ := identity.NewActionKey(classKey, "act")
	return actionKey
}

// validGlobalFunctionKey returns a valid global function identity.Key for testing.
func validGlobalFunctionKey() identity.Key {
	key, _ := identity.NewGlobalFunctionKey("_Func")
	return key
}

func (s *ExpressionTestSuite) TestValidateLiterals() {
	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid bool literal true", expr: &BoolLiteral{Value: true}},
		{testName: "valid bool literal false", expr: &BoolLiteral{Value: false}},
		{testName: "valid int literal", expr: &IntLiteral{Value: 42}},
		{testName: "valid int literal zero", expr: &IntLiteral{Value: 0}},
		{testName: "valid string literal", expr: &StringLiteral{Value: "hello"}},
		{testName: "valid string literal empty", expr: &StringLiteral{Value: ""}},
		{testName: "valid rational", expr: &RationalLiteral{Numerator: 1, Denominator: 3}},
		{testName: "error rational zero denominator", expr: &RationalLiteral{Numerator: 1, Denominator: 0}, errstr: "denominator cannot be zero"},
		{testName: "valid set literal empty", expr: &SetLiteral{}},
		{testName: "valid set literal with elements", expr: &SetLiteral{Elements: []Expression{&IntLiteral{Value: 1}, &IntLiteral{Value: 2}}}},
		{testName: "error set literal nil element", expr: &SetLiteral{Elements: []Expression{nil}}, errstr: "Elements[0]: is required"},
		{testName: "valid tuple literal", expr: &TupleLiteral{Elements: []Expression{&IntLiteral{Value: 1}}}},
		{testName: "error tuple literal empty", expr: &TupleLiteral{}, errstr: "Elements"},
		{testName: "valid record literal", expr: &RecordLiteral{Fields: []RecordField{{Name: "x", Value: &IntLiteral{Value: 1}}}}},
		{testName: "error record literal empty", expr: &RecordLiteral{}, errstr: "Fields"},
		{testName: "error record literal missing name", expr: &RecordLiteral{Fields: []RecordField{{Value: &IntLiteral{Value: 1}}}}, errstr: "Name: is required"},
		{testName: "error record literal nil value", expr: &RecordLiteral{Fields: []RecordField{{Name: "x"}}}, errstr: "Value: is required"},
		{testName: "valid set constant nat", expr: &SetConstant{Kind: SetConstantNat}},
		{testName: "valid set constant boolean", expr: &SetConstant{Kind: SetConstantBoolean}},
		{testName: "error set constant invalid", expr: &SetConstant{Kind: "invalid"}, errstr: "Kind"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateReferences() {
	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid self ref", expr: &SelfRef{}},
		{testName: "valid attribute ref", expr: &AttributeRef{AttributeKey: validAttributeKey()}},
		{testName: "error attribute ref empty key", expr: &AttributeRef{}, errstr: "AttributeKey"},
		{testName: "valid local var", expr: &LocalVar{Name: "x"}},
		{testName: "error local var empty", expr: &LocalVar{}, errstr: "Name"},
		{testName: "valid prior field value", expr: &PriorFieldValue{Field: "count"}},
		{testName: "error prior field value empty", expr: &PriorFieldValue{}, errstr: "Field"},
		{testName: "valid next state", expr: &NextState{Expr: &SelfRef{}}},
		{testName: "error next state nil", expr: &NextState{}, errstr: "Expr: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateBinaryOps() {
	left := &IntLiteral{Value: 1}
	right := &IntLiteral{Value: 2}
	boolLeft := &BoolLiteral{Value: true}
	boolRight := &BoolLiteral{Value: false}

	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid binary arith", expr: &BinaryArith{Op: ArithAdd, Left: left, Right: right}},
		{testName: "error binary arith invalid op", expr: &BinaryArith{Op: "bad", Left: left, Right: right}, errstr: "Op"},
		{testName: "error binary arith nil left", expr: &BinaryArith{Op: ArithAdd, Right: right}, errstr: "Left: is required"},
		{testName: "error binary arith nil right", expr: &BinaryArith{Op: ArithAdd, Left: left}, errstr: "Right: is required"},
		{testName: "valid binary logic", expr: &BinaryLogic{Op: LogicAnd, Left: boolLeft, Right: boolRight}},
		{testName: "error binary logic invalid op", expr: &BinaryLogic{Op: "bad", Left: boolLeft, Right: boolRight}, errstr: "Op"},
		{testName: "valid compare", expr: &Compare{Op: CompareGt, Left: left, Right: right}},
		{testName: "error compare invalid op", expr: &Compare{Op: "bad", Left: left, Right: right}, errstr: "Op"},
		{testName: "valid set op", expr: &SetOp{Op: SetUnion, Left: &SetLiteral{}, Right: &SetLiteral{}}},
		{testName: "valid set compare", expr: &SetCompare{Op: SetCompareSubset, Left: &SetLiteral{}, Right: &SetLiteral{}}},
		{testName: "valid bag op", expr: &BagOp{Op: BagSum, Left: &SetLiteral{}, Right: &SetLiteral{}}},
		{testName: "valid bag compare", expr: &BagCompare{Op: BagCompareSubBag, Left: &SetLiteral{}, Right: &SetLiteral{}}},
		{testName: "valid membership", expr: &Membership{Element: left, Set: &SetLiteral{}}},
		{testName: "valid membership negated", expr: &Membership{Element: left, Set: &SetLiteral{}, Negated: true}},
		{testName: "error membership nil element", expr: &Membership{Set: &SetLiteral{}}, errstr: "Element: is required"},
		{testName: "error membership nil set", expr: &Membership{Element: left}, errstr: "Set: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateUnaryOps() {
	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid negate", expr: &Negate{Expr: &IntLiteral{Value: 1}}},
		{testName: "error negate nil", expr: &Negate{}, errstr: "Expr: is required"},
		{testName: "valid not", expr: &Not{Expr: &BoolLiteral{Value: true}}},
		{testName: "error not nil", expr: &Not{}, errstr: "Expr: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateCollections() {
	base := &SelfRef{}
	idx := &IntLiteral{Value: 1}

	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid field access", expr: &FieldAccess{Base: base, Field: "name"}},
		{testName: "error field access nil base", expr: &FieldAccess{Field: "name"}, errstr: "Base: is required"},
		{testName: "error field access empty field", expr: &FieldAccess{Base: base}, errstr: "Field"},
		{testName: "valid tuple index", expr: &TupleIndex{Tuple: base, Index: idx}},
		{testName: "error tuple index nil tuple", expr: &TupleIndex{Index: idx}, errstr: "Tuple: is required"},
		{testName: "valid record update", expr: &RecordUpdate{Base: base, Alterations: []FieldAlteration{{Field: "x", Value: idx}}}},
		{testName: "error record update no alterations", expr: &RecordUpdate{Base: base}, errstr: "Alterations"},
		{testName: "valid string index", expr: &StringIndex{Str: &StringLiteral{Value: "abc"}, Index: idx}},
		{testName: "valid string concat", expr: &StringConcat{Operands: []Expression{&StringLiteral{Value: "a"}, &StringLiteral{Value: "b"}}}},
		{testName: "error string concat too few", expr: &StringConcat{Operands: []Expression{&StringLiteral{Value: "a"}}}, errstr: "Operands"},
		{testName: "valid tuple concat", expr: &TupleConcat{Operands: []Expression{&TupleLiteral{Elements: []Expression{idx}}, &TupleLiteral{Elements: []Expression{idx}}}}},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateControlFlow() {
	cond := &BoolLiteral{Value: true}
	val := &IntLiteral{Value: 1}

	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid if then else", expr: &IfThenElse{Condition: cond, Then: val, Else: val}},
		{testName: "error if nil condition", expr: &IfThenElse{Then: val, Else: val}, errstr: "Condition: is required"},
		{testName: "error if nil then", expr: &IfThenElse{Condition: cond, Else: val}, errstr: "Then: is required"},
		{testName: "error if nil else", expr: &IfThenElse{Condition: cond, Then: val}, errstr: "Else: is required"},
		{testName: "valid case", expr: &Case{Branches: []CaseBranch{{Condition: cond, Result: val}}}},
		{testName: "valid case with otherwise", expr: &Case{Branches: []CaseBranch{{Condition: cond, Result: val}}, Otherwise: val}},
		{testName: "error case no branches", expr: &Case{}, errstr: "Branches"},
		{testName: "error case nil condition", expr: &Case{Branches: []CaseBranch{{Result: val}}}, errstr: "Condition: is required"},
		{testName: "error case nil result", expr: &Case{Branches: []CaseBranch{{Condition: cond}}}, errstr: "Result: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateQuantifiers() {
	domain := &SetConstant{Kind: SetConstantNat}
	pred := &BoolLiteral{Value: true}
	start := &IntLiteral{Value: 1}
	end := &IntLiteral{Value: 10}

	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid forall", expr: &Quantifier{Kind: QuantifierForall, Variable: "x", Domain: domain, Predicate: pred}},
		{testName: "valid exists", expr: &Quantifier{Kind: QuantifierExists, Variable: "x", Domain: domain, Predicate: pred}},
		{testName: "error quantifier invalid kind", expr: &Quantifier{Kind: "bad", Variable: "x", Domain: domain, Predicate: pred}, errstr: "Kind"},
		{testName: "error quantifier no variable", expr: &Quantifier{Kind: QuantifierForall, Domain: domain, Predicate: pred}, errstr: "Variable"},
		{testName: "error quantifier nil domain", expr: &Quantifier{Kind: QuantifierForall, Variable: "x", Predicate: pred}, errstr: "Domain: is required"},
		{testName: "error quantifier nil predicate", expr: &Quantifier{Kind: QuantifierForall, Variable: "x", Domain: domain}, errstr: "Predicate: is required"},
		{testName: "valid set filter", expr: &SetFilter{Variable: "x", Set: domain, Predicate: pred}},
		{testName: "error set filter no variable", expr: &SetFilter{Set: domain, Predicate: pred}, errstr: "Variable"},
		{testName: "error set filter nil set", expr: &SetFilter{Variable: "x", Predicate: pred}, errstr: "Set: is required"},
		{testName: "valid set range", expr: &SetRange{Start: start, End: end}},
		{testName: "error set range nil start", expr: &SetRange{End: end}, errstr: "Start: is required"},
		{testName: "error set range nil end", expr: &SetRange{Start: start}, errstr: "End: is required"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateCalls() {
	arg := &IntLiteral{Value: 1}

	tests := []struct {
		testName string
		expr     Expression
		errstr   string
	}{
		{testName: "valid action call", expr: &ActionCall{ActionKey: validActionKey(), Args: []Expression{arg}}},
		{testName: "valid action call no args", expr: &ActionCall{ActionKey: validActionKey()}},
		{testName: "error action call empty key", expr: &ActionCall{}, errstr: "ActionKey"},
		{testName: "error action call nil arg", expr: &ActionCall{ActionKey: validActionKey(), Args: []Expression{nil}}, errstr: "Args[0]: is required"},
		{testName: "valid global call", expr: &GlobalCall{FunctionKey: validGlobalFunctionKey(), Args: []Expression{arg}}},
		{testName: "error global call empty key", expr: &GlobalCall{}, errstr: "FunctionKey"},
		{testName: "valid builtin call", expr: &BuiltinCall{Module: "_Seq", Function: "Len", Args: []Expression{arg}}},
		{testName: "error builtin call no module", expr: &BuiltinCall{Function: "Len"}, errstr: "Module"},
		{testName: "error builtin call no function", expr: &BuiltinCall{Module: "_Seq"}, errstr: "Function"},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionTestSuite) TestValidateExpression() {
	// Test the ValidateExpression helper function.
	s.NoError(ValidateExpression(nil))
	s.NoError(ValidateExpression(&BoolLiteral{Value: true}))
	s.Error(ValidateExpression(&RationalLiteral{Numerator: 1, Denominator: 0}))
}

func (s *ExpressionTestSuite) TestNodeType() {
	// Verify a sampling of NodeType() returns.
	s.Equal(NodeBoolLiteral, (&BoolLiteral{}).NodeType())
	s.Equal(NodeIntLiteral, (&IntLiteral{}).NodeType())
	s.Equal(NodeSelfRef, (&SelfRef{}).NodeType())
	s.Equal(NodeBinaryArith, (&BinaryArith{}).NodeType())
	s.Equal(NodeQuantifier, (&Quantifier{}).NodeType())
	s.Equal(NodeActionCall, (&ActionCall{}).NodeType())
	s.Equal(NodeBuiltinCall, (&BuiltinCall{}).NodeType())
	s.Equal(NodeIfThenElse, (&IfThenElse{}).NodeType())
	s.Equal(NodeRecordUpdate, (&RecordUpdate{}).NodeType())
}

func (s *ExpressionTestSuite) TestRecursiveValidation() {
	// Ensure validation propagates through nested expressions.
	// Build a tree with an invalid leaf deep inside.
	invalidLeaf := &RationalLiteral{Numerator: 1, Denominator: 0}
	tree := &BinaryArith{
		Op:   ArithAdd,
		Left: &IntLiteral{Value: 1},
		Right: &Negate{
			Expr: invalidLeaf,
		},
	}
	err := tree.Validate()
	s.Error(err)
	s.Contains(err.Error(), "denominator cannot be zero")
}
