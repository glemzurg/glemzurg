package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PrintTestSuite struct {
	suite.Suite
}

func TestPrintSuite(t *testing.T) {
	suite.Run(t, new(PrintTestSuite))
}

// --- Literal tests ---

func (s *PrintTestSuite) TestPrintBooleanLiteral() {
	s.Equal("TRUE", Print(&BooleanLiteral{Value: true}))
	s.Equal("FALSE", Print(&BooleanLiteral{Value: false}))
}

func (s *PrintTestSuite) TestPrintNumberLiteral() {
	s.Equal("42", Print(NewNumberLiteral("42")))
	s.Equal("3.14", Print(NewDecimalNumberLiteral("3", "14")))
}

func (s *PrintTestSuite) TestPrintStringLiteral() {
	s.Equal(`"hello"`, Print(&StringLiteral{Value: "hello"}))
	s.Equal(`"say \"hi\""`, Print(&StringLiteral{Value: `say "hi"`}))
}

func (s *PrintTestSuite) TestPrintSetLiteral() {
	s.Equal("{}", Print(&SetLiteral{Elements: []Expression{}}))
	s.Equal("{1, 2, 3}", Print(&SetLiteral{Elements: []Expression{
		NewNumberLiteral("1"),
		NewNumberLiteral("2"),
		NewNumberLiteral("3"),
	}}))
}

func (s *PrintTestSuite) TestPrintTupleLiteral() {
	s.Equal("⟨⟩", Print(&TupleLiteral{Elements: []Expression{}}))
	s.Equal("⟨1, 2⟩", Print(&TupleLiteral{Elements: []Expression{
		NewNumberLiteral("1"),
		NewNumberLiteral("2"),
	}}))
}

func (s *PrintTestSuite) TestPrintRecordInstance() {
	s.Equal("[a ↦ 1, b ↦ 2]", Print(&RecordInstance{
		Bindings: []*FieldBinding{
			{Field: &Identifier{Value: "a"}, Expression: NewNumberLiteral("1")},
			{Field: &Identifier{Value: "b"}, Expression: NewNumberLiteral("2")},
		},
	}))
}

func (s *PrintTestSuite) TestPrintSetConstant() {
	s.Equal("Nat", Print(&SetConstant{Value: "Nat"}))
	s.Equal("BOOLEAN", Print(&SetConstant{Value: "BOOLEAN"}))
}

func (s *PrintTestSuite) TestPrintIdentifier() {
	s.Equal("x", Print(&Identifier{Value: "x"}))
}

func (s *PrintTestSuite) TestPrintExistingValue() {
	s.Equal("@", Print(&ExistingValue{}))
}

// --- Binary logic operators ---

func (s *PrintTestSuite) TestPrintBinaryLogicAnd() {
	s.Equal("a ∧ b", Print(&BinaryLogic{
		Operator: "∧",
		Left:     &Identifier{Value: "a"},
		Right:    &Identifier{Value: "b"},
	}))
}

func (s *PrintTestSuite) TestPrintBinaryLogicOr() {
	s.Equal("a ∨ b", Print(&BinaryLogic{
		Operator: "∨",
		Left:     &Identifier{Value: "a"},
		Right:    &Identifier{Value: "b"},
	}))
}

func (s *PrintTestSuite) TestPrintBinaryLogicImplies() {
	s.Equal("a ⇒ b", Print(&BinaryLogic{
		Operator: "⇒",
		Left:     &Identifier{Value: "a"},
		Right:    &Identifier{Value: "b"},
	}))
}

func (s *PrintTestSuite) TestPrintBinaryLogicEquiv() {
	s.Equal("a ≡ b", Print(&BinaryLogic{
		Operator: "≡",
		Left:     &Identifier{Value: "a"},
		Right:    &Identifier{Value: "b"},
	}))
}

// --- Precedence: and/or ---

func (s *PrintTestSuite) TestPrintAndHigherThanOr_NoParens() {
	// a ∧ b ∨ c — and binds tighter, so no parens needed
	s.Equal("a ∧ b ∨ c", Print(&BinaryLogic{
		Operator: "∨",
		Left: &BinaryLogic{
			Operator: "∧",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

func (s *PrintTestSuite) TestPrintOrInsideAnd_NeedsParens() {
	// (a ∨ b) ∧ c — or inside and needs parens
	s.Equal("(a ∨ b) ∧ c", Print(&BinaryLogic{
		Operator: "∧",
		Left: &BinaryLogic{
			Operator: "∨",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

// --- Precedence: implies right-associativity ---

func (s *PrintTestSuite) TestPrintImpliesRightAssoc_NoParens() {
	// a ⇒ b ⇒ c — right-assoc, no parens needed
	s.Equal("a ⇒ b ⇒ c", Print(&BinaryLogic{
		Operator: "⇒",
		Left:     &Identifier{Value: "a"},
		Right: &BinaryLogic{
			Operator: "⇒",
			Left:     &Identifier{Value: "b"},
			Right:    &Identifier{Value: "c"},
		},
	}))
}

func (s *PrintTestSuite) TestPrintImpliesLeftGrouping_NeedsParens() {
	// (a ⇒ b) ⇒ c — left grouping needs parens for right-assoc operator
	s.Equal("(a ⇒ b) ⇒ c", Print(&BinaryLogic{
		Operator: "⇒",
		Left: &BinaryLogic{
			Operator: "⇒",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

// --- Precedence: arithmetic ---

func (s *PrintTestSuite) TestPrintMulHigherThanAdd_NoParens() {
	// a + b * c — mul binds tighter
	s.Equal("a + b * c", Print(&BinaryArithmetic{
		Operator: "+",
		Left:     &Identifier{Value: "a"},
		Right: &BinaryArithmetic{
			Operator: "*",
			Left:     &Identifier{Value: "b"},
			Right:    &Identifier{Value: "c"},
		},
	}))
}

func (s *PrintTestSuite) TestPrintAddInsideMul_NeedsParens() {
	// (a + b) * c — add inside mul needs parens
	s.Equal("(a + b) * c", Print(&BinaryArithmetic{
		Operator: "*",
		Left: &BinaryArithmetic{
			Operator: "+",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

func (s *PrintTestSuite) TestPrintPowerRightAssoc_NoParens() {
	// a ^ b ^ c — right-assoc
	s.Equal("a ^ b ^ c", Print(&BinaryArithmetic{
		Operator: "^",
		Left:     &Identifier{Value: "a"},
		Right: &BinaryArithmetic{
			Operator: "^",
			Left:     &Identifier{Value: "b"},
			Right:    &Identifier{Value: "c"},
		},
	}))
}

func (s *PrintTestSuite) TestPrintPowerLeftGrouping_NeedsParens() {
	// (a ^ b) ^ c
	s.Equal("(a ^ b) ^ c", Print(&BinaryArithmetic{
		Operator: "^",
		Left: &BinaryArithmetic{
			Operator: "^",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

// --- Precedence: comparison higher than logic ---

func (s *PrintTestSuite) TestPrintComparisonHigherThanLogic_NoParens() {
	// x > 5 ∧ y < 10 — comparisons higher than logic
	s.Equal("x > 5 ∧ y < 10", Print(&BinaryLogic{
		Operator: "∧",
		Left: &BinaryComparison{
			Operator: ">",
			Left:     &Identifier{Value: "x"},
			Right:    NewNumberLiteral("5"),
		},
		Right: &BinaryComparison{
			Operator: "<",
			Left:     &Identifier{Value: "y"},
			Right:    NewNumberLiteral("10"),
		},
	}))
}

// --- Negation ---

func (s *PrintTestSuite) TestPrintUnaryNegation() {
	s.Equal("-x", Print(NewNegation(&Identifier{Value: "x"})))
}

func (s *PrintTestSuite) TestPrintNotHigherThanAnd() {
	// ¬a ∧ b — not higher than and
	s.Equal("¬a ∧ b", Print(&BinaryLogic{
		Operator: "∧",
		Left: &UnaryLogic{
			Operator: "¬",
			Right:    &Identifier{Value: "a"},
		},
		Right: &Identifier{Value: "b"},
	}))
}

// --- Membership ---

func (s *PrintTestSuite) TestPrintMembership() {
	s.Equal("x ∈ S", Print(&Membership{
		Operator: "∈",
		Left:     &Identifier{Value: "x"},
		Right:    &Identifier{Value: "S"},
	}))
	s.Equal("x ∉ S", Print(&Membership{
		Operator: "∉",
		Left:     &Identifier{Value: "x"},
		Right:    &Identifier{Value: "S"},
	}))
}

// --- Set operations ---

func (s *PrintTestSuite) TestPrintSetOperations() {
	s.Equal("A ∪ B", Print(&BinarySetOperation{
		Operator: "∪",
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
	s.Equal("A ∩ B", Print(&BinarySetOperation{
		Operator: "∩",
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
	s.Equal(`A \ B`, Print(&BinarySetOperation{
		Operator: `\`,
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
}

// --- Set comparison ---

func (s *PrintTestSuite) TestPrintSetComparison() {
	s.Equal("A ⊆ B", Print(&BinarySetComparison{
		Operator: "⊆",
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
}

// --- Bag operations ---

func (s *PrintTestSuite) TestPrintBagOperations() {
	s.Equal("A ⊕ B", Print(&BinaryBagOperation{
		Operator: "⊕",
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
	s.Equal("A ⊖ B", Print(&BinaryBagOperation{
		Operator: "⊖",
		Left:     &Identifier{Value: "A"},
		Right:    &Identifier{Value: "B"},
	}))
}

// --- Cartesian product ---

func (s *PrintTestSuite) TestPrintCartesianProduct() {
	s.Equal("A × B × C", Print(&CartesianProduct{
		Operands: []Expression{
			&Identifier{Value: "A"},
			&Identifier{Value: "B"},
			&Identifier{Value: "C"},
		},
	}))
}

// --- Set range ---

func (s *PrintTestSuite) TestPrintSetRange() {
	s.Equal("1..10", Print(&SetRangeExpr{
		Start: NewNumberLiteral("1"),
		End:   NewNumberLiteral("10"),
	}))
}

// --- Concat ---

func (s *PrintTestSuite) TestPrintTupleConcat() {
	s.Equal("a ∘ b ∘ c", Print(&TupleConcat{
		Operator: "∘",
		Operands: []Expression{
			&Identifier{Value: "a"},
			&Identifier{Value: "b"},
			&Identifier{Value: "c"},
		},
	}))
}

// --- Fraction ---

func (s *PrintTestSuite) TestPrintFraction() {
	s.Equal("3 / 4", Print(NewFraction(
		NewNumberLiteral("3"),
		NewNumberLiteral("4"),
	)))
}

// --- Field access ---

func (s *PrintTestSuite) TestPrintFieldAccess() {
	s.Equal("r.x", Print(&FieldAccess{
		Base:   &Identifier{Value: "r"},
		Member: "x",
	}))
}

func (s *PrintTestSuite) TestPrintFieldAccessChained() {
	s.Equal("a.b.c", Print(&FieldAccess{
		Base: &FieldAccess{
			Base:   &Identifier{Value: "a"},
			Member: "b",
		},
		Member: "c",
	}))
}

// --- Tuple/string index ---

func (s *PrintTestSuite) TestPrintTupleIndex() {
	s.Equal("t[1]", Print(&TupleIndex{
		Tuple: &Identifier{Value: "t"},
		Index: NewNumberLiteral("1"),
	}))
}

// --- Prime ---

func (s *PrintTestSuite) TestPrintPrime() {
	s.Equal("x'", Print(&Primed{Base: &Identifier{Value: "x"}}))
}

// --- Quantifier ---

func (s *PrintTestSuite) TestPrintQuantifier() {
	s.Equal("∀ x ∈ S : x > 0", Print(&Quantifier{
		Quantifier: "∀",
		Membership: &Membership{
			Operator: "∈",
			Left:     &Identifier{Value: "x"},
			Right:    &Identifier{Value: "S"},
		},
		Predicate: &BinaryComparison{
			Operator: ">",
			Left:     &Identifier{Value: "x"},
			Right:    NewNumberLiteral("0"),
		},
	}))
}

// --- Set filter ---

func (s *PrintTestSuite) TestPrintSetFilter() {
	s.Equal("{x ∈ S : x > 0}", Print(&SetFilter{
		Membership: &Membership{
			Operator: "∈",
			Left:     &Identifier{Value: "x"},
			Right:    &Identifier{Value: "S"},
		},
		Predicate: &BinaryComparison{
			Operator: ">",
			Left:     &Identifier{Value: "x"},
			Right:    NewNumberLiteral("0"),
		},
	}))
}

// --- IF/THEN/ELSE ---

func (s *PrintTestSuite) TestPrintIfThenElse() {
	s.Equal("IF x > 0 THEN x ELSE 0", Print(&IfThenElse{
		Condition: &BinaryComparison{
			Operator: ">",
			Left:     &Identifier{Value: "x"},
			Right:    NewNumberLiteral("0"),
		},
		Then: &Identifier{Value: "x"},
		Else: NewNumberLiteral("0"),
	}))
}

// --- CASE ---

func (s *PrintTestSuite) TestPrintCaseExpr() {
	s.Equal("CASE x = 1 → 10 □ x = 2 → 20 □ OTHER → 0", Print(&CaseExpr{
		Branches: []*CaseBranch{
			{
				Condition: &BinaryEquality{Operator: "=", Left: &Identifier{Value: "x"}, Right: NewNumberLiteral("1")},
				Result:    NewNumberLiteral("10"),
			},
			{
				Condition: &BinaryEquality{Operator: "=", Left: &Identifier{Value: "x"}, Right: NewNumberLiteral("2")},
				Result:    NewNumberLiteral("20"),
			},
		},
		Other: NewNumberLiteral("0"),
	}))
}

func (s *PrintTestSuite) TestPrintCaseExprWrapsImplies() {
	// CASE condition that uses implies should be wrapped
	s.Equal("CASE (a ⇒ b) → 1", Print(&CaseExpr{
		Branches: []*CaseBranch{
			{
				Condition: &BinaryLogic{Operator: "⇒", Left: &Identifier{Value: "a"}, Right: &Identifier{Value: "b"}},
				Result:    NewNumberLiteral("1"),
			},
		},
	}))
}

// --- Record altered (EXCEPT) ---

func (s *PrintTestSuite) TestPrintRecordAltered() {
	s.Equal("[r EXCEPT !.x = 1, !.y = 2]", Print(&RecordAltered{
		Identifier: &Identifier{Value: "r"},
		Alterations: []*FieldAlteration{
			{Field: &FieldAccess{Member: "x"}, Expression: NewNumberLiteral("1")},
			{Field: &FieldAccess{Member: "y"}, Expression: NewNumberLiteral("2")},
		},
	}))
}

// --- Function call ---

func (s *PrintTestSuite) TestPrintFunctionCall() {
	s.Equal("Foo(1, 2)", Print(&FunctionCall{
		Name: &Identifier{Value: "Foo"},
		Args: []Expression{NewNumberLiteral("1"), NewNumberLiteral("2")},
	}))
}

func (s *PrintTestSuite) TestPrintFunctionCallScoped() {
	s.Equal("_Seq!Len(s)", Print(&FunctionCall{
		ScopePath: []*Identifier{{Value: "_Seq"}},
		Name:      &Identifier{Value: "Len"},
		Args:      []Expression{&Identifier{Value: "s"}},
	}))
}

// --- Equality ---

func (s *PrintTestSuite) TestPrintEquality() {
	s.Equal("x = y", Print(&BinaryEquality{
		Operator: "=",
		Left:     &Identifier{Value: "x"},
		Right:    &Identifier{Value: "y"},
	}))
	s.Equal("x ≠ y", Print(&BinaryEquality{
		Operator: "≠",
		Left:     &Identifier{Value: "x"},
		Right:    &Identifier{Value: "y"},
	}))
}

// --- Complex precedence ---

func (s *PrintTestSuite) TestPrintComplexArithmeticPrecedence() {
	// 1 + 2 * 3 — no parens needed
	s.Equal("1 + 2 * 3", Print(&BinaryArithmetic{
		Operator: "+",
		Left:     NewNumberLiteral("1"),
		Right: &BinaryArithmetic{
			Operator: "*",
			Left:     NewNumberLiteral("2"),
			Right:    NewNumberLiteral("3"),
		},
	}))
}

func (s *PrintTestSuite) TestPrintSubLeftAssoc_NoParens() {
	// a - b - c — left assoc, represented as (a-b) - c, no extra parens
	s.Equal("a - b - c", Print(&BinaryArithmetic{
		Operator: "-",
		Left: &BinaryArithmetic{
			Operator: "-",
			Left:     &Identifier{Value: "a"},
			Right:    &Identifier{Value: "b"},
		},
		Right: &Identifier{Value: "c"},
	}))
}

func (s *PrintTestSuite) TestPrintSubRightGrouping_NeedsParens() {
	// a - (b - c) — needs parens for non-default grouping
	s.Equal("a - (b - c)", Print(&BinaryArithmetic{
		Operator: "-",
		Left:     &Identifier{Value: "a"},
		Right: &BinaryArithmetic{
			Operator: "-",
			Left:     &Identifier{Value: "b"},
			Right:    &Identifier{Value: "c"},
		},
	}))
}

// --- RecordTypeExpr ---

func (s *PrintTestSuite) TestPrintRecordTypeExpr() {
	s.Equal("[name: STRING, age: Int]", Print(&RecordTypeExpr{
		Fields: []*RecordTypeField{
			{Name: &Identifier{Value: "name"}, Type: &Identifier{Value: "STRING"}},
			{Name: &Identifier{Value: "age"}, Type: &Identifier{Value: "Int"}},
		},
	}))
}

// --- SetLiteralEnum ---

func (s *PrintTestSuite) TestPrintSetLiteralEnum() {
	s.Equal(`{"a", "b", "c"}`, Print(&SetLiteralEnum{
		Values: []string{"a", "b", "c"},
	}))
}

// --- SetLiteralInt ---

func (s *PrintTestSuite) TestPrintSetLiteralInt() {
	s.Equal("{1, 2, 3}", Print(&SetLiteralInt{
		Values: []int{1, 2, 3},
	}))
}
