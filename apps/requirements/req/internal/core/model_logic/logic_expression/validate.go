package logic_expression

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// --- valid op sets ---

var validArithOps = map[ArithOp]bool{
	ArithAdd: true, ArithSub: true, ArithMul: true,
	ArithDiv: true, ArithMod: true, ArithPow: true,
}

var validLogicOps = map[LogicOp]bool{
	LogicAnd: true, LogicOr: true, LogicImplies: true, LogicEquiv: true,
}

var validCompareOps = map[CompareOp]bool{
	CompareLt: true, CompareGt: true, CompareLte: true,
	CompareGte: true, CompareEq: true, CompareNeq: true,
}

var validSetOpKinds = map[SetOpKind]bool{
	SetUnion: true, SetIntersect: true, SetDifference: true,
}

var validSetCompareOps = map[SetCompareOp]bool{
	SetCompareSubsetEq: true, SetCompareSubset: true,
	SetCompareSupersetEq: true, SetCompareSuperset: true,
}

var validBagOpKinds = map[BagOpKind]bool{
	BagSum: true, BagDifference: true,
}

var validBagCompareOps = map[BagCompareOp]bool{
	BagCompareProperSubBag: true, BagCompareSubBag: true,
	BagCompareProperSupBag: true, BagCompareSupBag: true,
}

var validQuantifierKinds = map[QuantifierKind]bool{
	QuantifierForall: true, QuantifierExists: true,
}

var validSetConstantKinds = map[SetConstantKind]bool{
	SetConstantNat: true, SetConstantInt: true,
	SetConstantReal: true, SetConstantBoolean: true,
}

// --- Literal validation ---

func (n *BoolLiteral) Validate() error { return nil }

func (n *IntLiteral) Validate() error {
	if n.Value == nil {
		return coreerr.New(coreerr.ExprIntValueRequired, "IntLiteral.Value: is required", "Value")
	}
	return nil
}

func (n *StringLiteral) Validate() error { return nil }

func (n *RationalLiteral) Validate() error {
	if n.Value == nil {
		return coreerr.New(coreerr.ExprRatValueRequired, "RationalLiteral.Value: is required", "Value")
	}
	return nil
}

func (n *SetLiteral) Validate() error {
	for i, elem := range n.Elements {
		if elem == nil {
			return coreerr.New(coreerr.ExprSetElemRequired, fmt.Sprintf("SetLiteral.Elements[%d]: is required", i), fmt.Sprintf("Elements[%d]", i))
		}
		if err := elem.Validate(); err != nil {
			return coreerr.New(coreerr.ExprSetElemInvalid, fmt.Sprintf("SetLiteral.Elements[%d]: %s", i, err.Error()), fmt.Sprintf("Elements[%d]", i))
		}
	}
	return nil
}

func (n *TupleLiteral) Validate() error {
	if len(n.Elements) == 0 {
		return coreerr.NewWithValues(coreerr.ExprTupleElemRequired, "TupleLiteral.Elements: at least one element is required", "Elements", "", "min=1")
	}
	for i, elem := range n.Elements {
		if elem == nil {
			return coreerr.New(coreerr.ExprTupleElemNil, fmt.Sprintf("TupleLiteral.Elements[%d]: is required", i), fmt.Sprintf("Elements[%d]", i))
		}
		if err := elem.Validate(); err != nil {
			return coreerr.New(coreerr.ExprTupleElemInvalid, fmt.Sprintf("TupleLiteral.Elements[%d]: %s", i, err.Error()), fmt.Sprintf("Elements[%d]", i))
		}
	}
	return nil
}

func (n *RecordLiteral) Validate() error {
	if len(n.Fields) == 0 {
		return coreerr.NewWithValues(coreerr.ExprRecordFieldRequired, "RecordLiteral.Fields: at least one field is required", "Fields", "", "min=1")
	}
	for i, field := range n.Fields {
		if field.Name == "" {
			return coreerr.New(coreerr.ExprRecordNameRequired, fmt.Sprintf("RecordLiteral.Fields[%d].Name: is required", i), fmt.Sprintf("Fields[%d].Name", i))
		}
		if field.Value == nil {
			return coreerr.New(coreerr.ExprRecordValueRequired, fmt.Sprintf("RecordLiteral.Fields[%d].Value: is required", i), fmt.Sprintf("Fields[%d].Value", i))
		}
		if err := field.Value.Validate(); err != nil {
			return coreerr.New(coreerr.ExprRecordValueInvalid, fmt.Sprintf("RecordLiteral.Fields[%d].Value: %s", i, err.Error()), fmt.Sprintf("Fields[%d].Value", i))
		}
	}
	return nil
}

func (n *SetConstant) Validate() error {
	if n.Kind == "" {
		return coreerr.NewWithValues(coreerr.ExprSetconstKindRequired, "SetConstant.Kind: is required", "Kind", "", "one of: nat, int, real, boolean")
	}
	if !validSetConstantKinds[n.Kind] {
		return coreerr.NewWithValues(coreerr.ExprSetconstKindInvalid, fmt.Sprintf("SetConstant.Kind: '%s' is not valid", string(n.Kind)), "Kind", string(n.Kind), "one of: nat, int, real, boolean")
	}
	return nil
}

// --- Reference validation ---

func (n *SelfRef) Validate() error { return nil }

func (n *AttributeRef) Validate() error {
	if err := n.AttributeKey.Validate(); err != nil {
		return coreerr.New(coreerr.ExprAttrkeyInvalid, fmt.Sprintf("AttributeRef.AttributeKey: %s", err.Error()), "AttributeKey")
	}
	return nil
}

func (n *LocalVar) Validate() error {
	if n.Name == "" {
		return coreerr.New(coreerr.ExprLocalvarNameRequired, "LocalVar.Name: is required", "Name")
	}
	return nil
}

func (n *PriorFieldValue) Validate() error {
	if n.Field == "" {
		return coreerr.New(coreerr.ExprPriorfieldRequired, "PriorFieldValue.Field: is required", "Field")
	}
	return nil
}

func (n *NextState) Validate() error {
	if n.Expr == nil {
		return coreerr.New(coreerr.ExprNextstateExprRequired, "NextState.Expr: is required", "Expr")
	}
	return n.Expr.Validate()
}

// --- Binary operator validation ---

func (n *BinaryArith) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "BinaryArith.Op: is required", "Op", "", "one of: add, sub, mul, div, mod, pow")
	}
	if !validArithOps[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("BinaryArith.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: add, sub, mul, div, mod, pow")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "BinaryArith.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("BinaryArith.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "BinaryArith.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("BinaryArith.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *BinaryLogic) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "BinaryLogic.Op: is required", "Op", "", "one of: and, or, implies, equiv")
	}
	if !validLogicOps[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("BinaryLogic.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: and, or, implies, equiv")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "BinaryLogic.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("BinaryLogic.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "BinaryLogic.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("BinaryLogic.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *Compare) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "Compare.Op: is required", "Op", "", "one of: lt, gt, lte, gte, eq, neq")
	}
	if !validCompareOps[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("Compare.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: lt, gt, lte, gte, eq, neq")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "Compare.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("Compare.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "Compare.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("Compare.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *SetOp) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "SetOp.Op: is required", "Op", "", "one of: union, intersect, difference")
	}
	if !validSetOpKinds[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("SetOp.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: union, intersect, difference")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "SetOp.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("SetOp.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "SetOp.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("SetOp.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *SetCompare) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "SetCompare.Op: is required", "Op", "", "one of: subset_eq, subset, superset_eq, superset")
	}
	if !validSetCompareOps[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("SetCompare.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: subset_eq, subset, superset_eq, superset")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "SetCompare.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("SetCompare.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "SetCompare.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("SetCompare.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *BagOp) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "BagOp.Op: is required", "Op", "", "one of: sum, difference")
	}
	if !validBagOpKinds[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("BagOp.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: sum, difference")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "BagOp.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("BagOp.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "BagOp.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("BagOp.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *BagCompare) Validate() error {
	if n.Op == "" {
		return coreerr.NewWithValues(coreerr.ExprOpRequired, "BagCompare.Op: is required", "Op", "", "one of: proper_sub_bag, sub_bag, proper_sup_bag, sup_bag")
	}
	if !validBagCompareOps[n.Op] {
		return coreerr.NewWithValues(coreerr.ExprOpInvalid, fmt.Sprintf("BagCompare.Op: '%s' is not valid", string(n.Op)), "Op", string(n.Op), "one of: proper_sub_bag, sub_bag, proper_sup_bag, sup_bag")
	}
	if n.Left == nil {
		return coreerr.New(coreerr.ExprLeftRequired, "BagCompare.Left: is required", "Left")
	}
	if err := n.Left.Validate(); err != nil {
		return coreerr.New(coreerr.ExprLeftInvalid, fmt.Sprintf("BagCompare.Left: %s", err.Error()), "Left")
	}
	if n.Right == nil {
		return coreerr.New(coreerr.ExprRightRequired, "BagCompare.Right: is required", "Right")
	}
	if err := n.Right.Validate(); err != nil {
		return coreerr.New(coreerr.ExprRightInvalid, fmt.Sprintf("BagCompare.Right: %s", err.Error()), "Right")
	}
	return nil
}

func (n *Membership) Validate() error {
	if n.Element == nil {
		return coreerr.New(coreerr.ExprElementRequired, "Membership.Element: is required", "Element")
	}
	if err := n.Element.Validate(); err != nil {
		return coreerr.New(coreerr.ExprElementInvalid, fmt.Sprintf("Membership.Element: %s", err.Error()), "Element")
	}
	if n.Set == nil {
		return coreerr.New(coreerr.ExprSetRequired, "Membership.Set: is required", "Set")
	}
	if err := n.Set.Validate(); err != nil {
		return coreerr.New(coreerr.ExprSetInvalid, fmt.Sprintf("Membership.Set: %s", err.Error()), "Set")
	}
	return nil
}

// --- Unary operator validation ---

func (n *Negate) Validate() error {
	if n.Expr == nil {
		return coreerr.New(coreerr.ExprExprRequired, "Negate.Expr: is required", "Expr")
	}
	return n.Expr.Validate()
}

func (n *Not) Validate() error {
	if n.Expr == nil {
		return coreerr.New(coreerr.ExprExprRequired, "Not.Expr: is required", "Expr")
	}
	return n.Expr.Validate()
}

// --- Collection validation ---

func (n *FieldAccess) Validate() error {
	if n.Field == "" {
		return coreerr.New(coreerr.ExprFieldRequired, "FieldAccess.Field: is required", "Field")
	}
	if n.Base == nil {
		return coreerr.New(coreerr.ExprBaseRequired, "FieldAccess.Base: is required", "Base")
	}
	if err := n.Base.Validate(); err != nil {
		return coreerr.New(coreerr.ExprBaseInvalid, fmt.Sprintf("FieldAccess.Base: %s", err.Error()), "Base")
	}
	return nil
}

func (n *TupleIndex) Validate() error {
	if n.Tuple == nil {
		return coreerr.New(coreerr.ExprTupleRequired, "TupleIndex.Tuple: is required", "Tuple")
	}
	if err := n.Tuple.Validate(); err != nil {
		return coreerr.New(coreerr.ExprTupleInvalid, fmt.Sprintf("TupleIndex.Tuple: %s", err.Error()), "Tuple")
	}
	if n.Index == nil {
		return coreerr.New(coreerr.ExprIndexRequired, "TupleIndex.Index: is required", "Index")
	}
	if err := n.Index.Validate(); err != nil {
		return coreerr.New(coreerr.ExprIndexInvalid, fmt.Sprintf("TupleIndex.Index: %s", err.Error()), "Index")
	}
	return nil
}

func (n *RecordUpdate) Validate() error {
	if n.Base == nil {
		return coreerr.New(coreerr.ExprBaseRequired, "RecordUpdate.Base: is required", "Base")
	}
	if err := n.Base.Validate(); err != nil {
		return coreerr.New(coreerr.ExprBaseInvalid, fmt.Sprintf("RecordUpdate.Base: %s", err.Error()), "Base")
	}
	if len(n.Alterations) == 0 {
		return coreerr.NewWithValues(coreerr.ExprAlterationsRequired, "RecordUpdate.Alterations: at least one alteration is required", "Alterations", "", "min=1")
	}
	for i, alt := range n.Alterations {
		if alt.Field == "" {
			return coreerr.New(coreerr.ExprAltFieldRequired, fmt.Sprintf("RecordUpdate.Alterations[%d].Field: is required", i), fmt.Sprintf("Alterations[%d].Field", i))
		}
		if alt.Value == nil {
			return coreerr.New(coreerr.ExprAltValueRequired, fmt.Sprintf("RecordUpdate.Alterations[%d].Value: is required", i), fmt.Sprintf("Alterations[%d].Value", i))
		}
		if err := alt.Value.Validate(); err != nil {
			return coreerr.New(coreerr.ExprAltValueInvalid, fmt.Sprintf("RecordUpdate.Alterations[%d].Value: %s", i, err.Error()), fmt.Sprintf("Alterations[%d].Value", i))
		}
	}
	return nil
}

func (n *StringIndex) Validate() error {
	if n.Str == nil {
		return coreerr.New(coreerr.ExprStrRequired, "StringIndex.Str: is required", "Str")
	}
	if err := n.Str.Validate(); err != nil {
		return coreerr.New(coreerr.ExprStrInvalid, fmt.Sprintf("StringIndex.Str: %s", err.Error()), "Str")
	}
	if n.Index == nil {
		return coreerr.New(coreerr.ExprIndexRequired, "StringIndex.Index: is required", "Index")
	}
	if err := n.Index.Validate(); err != nil {
		return coreerr.New(coreerr.ExprIndexInvalid, fmt.Sprintf("StringIndex.Index: %s", err.Error()), "Index")
	}
	return nil
}

func (n *StringConcat) Validate() error {
	if len(n.Operands) < 2 {
		return coreerr.NewWithValues(coreerr.ExprOperandsMinTwo, "StringConcat.Operands: at least two operands are required", "Operands", "", "min=2")
	}
	for i, op := range n.Operands {
		if op == nil {
			return coreerr.New(coreerr.ExprOperandRequired, fmt.Sprintf("StringConcat.Operands[%d]: is required", i), fmt.Sprintf("Operands[%d]", i))
		}
		if err := op.Validate(); err != nil {
			return coreerr.New(coreerr.ExprOperandInvalid, fmt.Sprintf("StringConcat.Operands[%d]: %s", i, err.Error()), fmt.Sprintf("Operands[%d]", i))
		}
	}
	return nil
}

func (n *TupleConcat) Validate() error {
	if len(n.Operands) < 2 {
		return coreerr.NewWithValues(coreerr.ExprOperandsMinTwo, "TupleConcat.Operands: at least two operands are required", "Operands", "", "min=2")
	}
	for i, op := range n.Operands {
		if op == nil {
			return coreerr.New(coreerr.ExprOperandRequired, fmt.Sprintf("TupleConcat.Operands[%d]: is required", i), fmt.Sprintf("Operands[%d]", i))
		}
		if err := op.Validate(); err != nil {
			return coreerr.New(coreerr.ExprOperandInvalid, fmt.Sprintf("TupleConcat.Operands[%d]: %s", i, err.Error()), fmt.Sprintf("Operands[%d]", i))
		}
	}
	return nil
}

// --- Control flow validation ---

func (n *IfThenElse) Validate() error {
	if n.Condition == nil {
		return coreerr.New(coreerr.ExprConditionRequired, "IfThenElse.Condition: is required", "Condition")
	}
	if err := n.Condition.Validate(); err != nil {
		return coreerr.New(coreerr.ExprConditionInvalid, fmt.Sprintf("IfThenElse.Condition: %s", err.Error()), "Condition")
	}
	if n.Then == nil {
		return coreerr.New(coreerr.ExprThenRequired, "IfThenElse.Then: is required", "Then")
	}
	if err := n.Then.Validate(); err != nil {
		return coreerr.New(coreerr.ExprThenInvalid, fmt.Sprintf("IfThenElse.Then: %s", err.Error()), "Then")
	}
	if n.Else == nil {
		return coreerr.New(coreerr.ExprElseRequired, "IfThenElse.Else: is required", "Else")
	}
	if err := n.Else.Validate(); err != nil {
		return coreerr.New(coreerr.ExprElseInvalid, fmt.Sprintf("IfThenElse.Else: %s", err.Error()), "Else")
	}
	return nil
}

func (n *Case) Validate() error {
	if len(n.Branches) == 0 {
		return coreerr.NewWithValues(coreerr.ExprBranchesRequired, "Case.Branches: at least one branch is required", "Branches", "", "min=1")
	}
	for i, branch := range n.Branches {
		if branch.Condition == nil {
			return coreerr.New(coreerr.ExprBranchCondRequired, fmt.Sprintf("Case.Branches[%d].Condition: is required", i), fmt.Sprintf("Branches[%d].Condition", i))
		}
		if err := branch.Condition.Validate(); err != nil {
			return coreerr.New(coreerr.ExprBranchCondInvalid, fmt.Sprintf("Case.Branches[%d].Condition: %s", i, err.Error()), fmt.Sprintf("Branches[%d].Condition", i))
		}
		if branch.Result == nil {
			return coreerr.New(coreerr.ExprBranchResultRequired, fmt.Sprintf("Case.Branches[%d].Result: is required", i), fmt.Sprintf("Branches[%d].Result", i))
		}
		if err := branch.Result.Validate(); err != nil {
			return coreerr.New(coreerr.ExprBranchResultInvalid, fmt.Sprintf("Case.Branches[%d].Result: %s", i, err.Error()), fmt.Sprintf("Branches[%d].Result", i))
		}
	}
	if n.Otherwise != nil {
		if err := n.Otherwise.Validate(); err != nil {
			return coreerr.New(coreerr.ExprOtherwiseInvalid, fmt.Sprintf("Case.Otherwise: %s", err.Error()), "Otherwise")
		}
	}
	return nil
}

// --- Quantifier validation ---

func (n *Quantifier) Validate() error {
	if n.Domain == nil {
		return coreerr.New(coreerr.ExprDomainRequired, "Quantifier.Domain: is required", "Domain")
	}
	if n.Predicate == nil {
		return coreerr.New(coreerr.ExprPredicateRequired, "Quantifier.Predicate: is required", "Predicate")
	}
	if n.Kind == "" {
		return coreerr.NewWithValues(coreerr.ExprQuantKindRequired, "Quantifier.Kind: is required", "Kind", "", "one of: forall, exists")
	}
	if !validQuantifierKinds[n.Kind] {
		return coreerr.NewWithValues(coreerr.ExprQuantKindInvalid, fmt.Sprintf("Quantifier.Kind: '%s' is not valid", string(n.Kind)), "Kind", string(n.Kind), "one of: forall, exists")
	}
	if n.Variable == "" {
		return coreerr.New(coreerr.ExprVariableRequired, "Quantifier.Variable: is required", "Variable")
	}
	if err := n.Domain.Validate(); err != nil {
		return coreerr.New(coreerr.ExprDomainInvalid, fmt.Sprintf("Quantifier.Domain: %s", err.Error()), "Domain")
	}
	if err := n.Predicate.Validate(); err != nil {
		return coreerr.New(coreerr.ExprPredicateInvalid, fmt.Sprintf("Quantifier.Predicate: %s", err.Error()), "Predicate")
	}
	return nil
}

func (n *SetFilter) Validate() error {
	if n.Set == nil {
		return coreerr.New(coreerr.ExprSetRequired, "SetFilter.Set: is required", "Set")
	}
	if n.Predicate == nil {
		return coreerr.New(coreerr.ExprPredicateRequired, "SetFilter.Predicate: is required", "Predicate")
	}
	if n.Variable == "" {
		return coreerr.New(coreerr.ExprVariableRequired, "SetFilter.Variable: is required", "Variable")
	}
	if err := n.Set.Validate(); err != nil {
		return coreerr.New(coreerr.ExprSetInvalid, fmt.Sprintf("SetFilter.Set: %s", err.Error()), "Set")
	}
	if err := n.Predicate.Validate(); err != nil {
		return coreerr.New(coreerr.ExprPredicateInvalid, fmt.Sprintf("SetFilter.Predicate: %s", err.Error()), "Predicate")
	}
	return nil
}

func (n *SetRange) Validate() error {
	if n.Start == nil {
		return coreerr.New(coreerr.ExprStartRequired, "SetRange.Start: is required", "Start")
	}
	if err := n.Start.Validate(); err != nil {
		return coreerr.New(coreerr.ExprStartInvalid, fmt.Sprintf("SetRange.Start: %s", err.Error()), "Start")
	}
	if n.End == nil {
		return coreerr.New(coreerr.ExprEndRequired, "SetRange.End: is required", "End")
	}
	if err := n.End.Validate(); err != nil {
		return coreerr.New(coreerr.ExprEndInvalid, fmt.Sprintf("SetRange.End: %s", err.Error()), "End")
	}
	return nil
}

// --- Call validation ---

func (n *ActionCall) Validate() error {
	if err := n.ActionKey.Validate(); err != nil {
		return coreerr.New(coreerr.ExprActionkeyInvalid, fmt.Sprintf("ActionCall.ActionKey: %s", err.Error()), "ActionKey")
	}
	for i, arg := range n.Args {
		if arg == nil {
			return coreerr.New(coreerr.ExprArgRequired, fmt.Sprintf("ActionCall.Args[%d]: is required", i), fmt.Sprintf("Args[%d]", i))
		}
		if err := arg.Validate(); err != nil {
			return coreerr.New(coreerr.ExprArgInvalid, fmt.Sprintf("ActionCall.Args[%d]: %s", i, err.Error()), fmt.Sprintf("Args[%d]", i))
		}
	}
	return nil
}

func (n *GlobalCall) Validate() error {
	if err := n.FunctionKey.Validate(); err != nil {
		return coreerr.New(coreerr.ExprFunctionkeyInvalid, fmt.Sprintf("GlobalCall.FunctionKey: %s", err.Error()), "FunctionKey")
	}
	for i, arg := range n.Args {
		if arg == nil {
			return coreerr.New(coreerr.ExprArgRequired, fmt.Sprintf("GlobalCall.Args[%d]: is required", i), fmt.Sprintf("Args[%d]", i))
		}
		if err := arg.Validate(); err != nil {
			return coreerr.New(coreerr.ExprArgInvalid, fmt.Sprintf("GlobalCall.Args[%d]: %s", i, err.Error()), fmt.Sprintf("Args[%d]", i))
		}
	}
	return nil
}

func (n *BuiltinCall) Validate() error {
	if n.Module == "" {
		return coreerr.New(coreerr.ExprModuleRequired, "BuiltinCall.Module: is required", "Module")
	}
	if n.Function == "" {
		return coreerr.New(coreerr.ExprFunctionRequired, "BuiltinCall.Function: is required", "Function")
	}
	for i, arg := range n.Args {
		if arg == nil {
			return coreerr.New(coreerr.ExprArgRequired, fmt.Sprintf("BuiltinCall.Args[%d]: is required", i), fmt.Sprintf("Args[%d]", i))
		}
		if err := arg.Validate(); err != nil {
			return coreerr.New(coreerr.ExprArgInvalid, fmt.Sprintf("BuiltinCall.Args[%d]: %s", i, err.Error()), fmt.Sprintf("Args[%d]", i))
		}
	}
	return nil
}

// --- Named set reference validation ---

func (n *NamedSetRef) Validate() error {
	if err := n.SetKey.Validate(); err != nil {
		return coreerr.New(coreerr.ExprSetkeyInvalid, fmt.Sprintf("NamedSetRef.SetKey: %s", err.Error()), "SetKey")
	}
	return nil
}
