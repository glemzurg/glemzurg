package model_expression

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
		return &coreerr.ValidationError{
			Code:    coreerr.ExprIntValueRequired,
			Message: "IntLiteral.Value: is required",
			Field:   "Value",
		}
	}
	return nil
}

func (n *StringLiteral) Validate() error { return nil }

func (n *RationalLiteral) Validate() error {
	if n.Value == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRatValueRequired,
			Message: "RationalLiteral.Value: is required",
			Field:   "Value",
		}
	}
	return nil
}

func (n *SetLiteral) Validate() error {
	for i, elem := range n.Elements {
		if elem == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprSetElemRequired,
				Message: fmt.Sprintf("SetLiteral.Elements[%d]: is required", i),
				Field:   fmt.Sprintf("Elements[%d]", i),
			}
		}
		if err := elem.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprSetElemInvalid,
				Message: fmt.Sprintf("SetLiteral.Elements[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Elements[%d]", i),
			}
		}
	}
	return nil
}

func (n *TupleLiteral) Validate() error {
	if len(n.Elements) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprTupleElemRequired,
			Message: "TupleLiteral.Elements: at least one element is required",
			Field:   "Elements",
			Want:    "min=1",
		}
	}
	for i, elem := range n.Elements {
		if elem == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprTupleElemNil,
				Message: fmt.Sprintf("TupleLiteral.Elements[%d]: is required", i),
				Field:   fmt.Sprintf("Elements[%d]", i),
			}
		}
		if err := elem.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprTupleElemInvalid,
				Message: fmt.Sprintf("TupleLiteral.Elements[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Elements[%d]", i),
			}
		}
	}
	return nil
}

func (n *RecordLiteral) Validate() error {
	if len(n.Fields) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRecordFieldRequired,
			Message: "RecordLiteral.Fields: at least one field is required",
			Field:   "Fields",
			Want:    "min=1",
		}
	}
	for i, field := range n.Fields {
		if field.Name == "" {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprRecordNameRequired,
				Message: fmt.Sprintf("RecordLiteral.Fields[%d].Name: is required", i),
				Field:   fmt.Sprintf("Fields[%d].Name", i),
			}
		}
		if field.Value == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprRecordValueRequired,
				Message: fmt.Sprintf("RecordLiteral.Fields[%d].Value: is required", i),
				Field:   fmt.Sprintf("Fields[%d].Value", i),
			}
		}
		if err := field.Value.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprRecordValueInvalid,
				Message: fmt.Sprintf("RecordLiteral.Fields[%d].Value: %s", i, err.Error()),
				Field:   fmt.Sprintf("Fields[%d].Value", i),
			}
		}
	}
	return nil
}

func (n *SetConstant) Validate() error {
	if n.Kind == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetconstKindRequired,
			Message: "SetConstant.Kind: is required",
			Field:   "Kind",
			Want:    "one of: nat, int, real, boolean",
		}
	}
	if !validSetConstantKinds[n.Kind] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetconstKindInvalid,
			Message: fmt.Sprintf("SetConstant.Kind: '%s' is not valid", string(n.Kind)),
			Field:   "Kind",
			Got:     string(n.Kind),
			Want:    "one of: nat, int, real, boolean",
		}
	}
	return nil
}

// --- Reference validation ---

func (n *SelfRef) Validate() error { return nil }

func (n *AttributeRef) Validate() error {
	if err := n.AttributeKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprAttrkeyInvalid,
			Message: fmt.Sprintf("AttributeRef.AttributeKey: %s", err.Error()),
			Field:   "AttributeKey",
		}
	}
	return nil
}

func (n *LocalVar) Validate() error {
	if n.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLocalvarNameRequired,
			Message: "LocalVar.Name: is required",
			Field:   "Name",
		}
	}
	return nil
}

func (n *PriorFieldValue) Validate() error {
	if n.Field == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprPriorfieldRequired,
			Message: "PriorFieldValue.Field: is required",
			Field:   "Field",
		}
	}
	return nil
}

func (n *NextState) Validate() error {
	if n.Expr == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprNextstateExprRequired,
			Message: "NextState.Expr: is required",
			Field:   "Expr",
		}
	}
	return n.Expr.Validate()
}

// --- Binary operator validation ---

func (n *BinaryArith) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "BinaryArith.Op: is required",
			Field:   "Op",
			Want:    "one of: add, sub, mul, div, mod, pow",
		}
	}
	if !validArithOps[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("BinaryArith.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: add, sub, mul, div, mod, pow",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "BinaryArith.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("BinaryArith.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "BinaryArith.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("BinaryArith.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *BinaryLogic) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "BinaryLogic.Op: is required",
			Field:   "Op",
			Want:    "one of: and, or, implies, equiv",
		}
	}
	if !validLogicOps[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("BinaryLogic.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: and, or, implies, equiv",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "BinaryLogic.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("BinaryLogic.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "BinaryLogic.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("BinaryLogic.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *Compare) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "Compare.Op: is required",
			Field:   "Op",
			Want:    "one of: lt, gt, lte, gte, eq, neq",
		}
	}
	if !validCompareOps[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("Compare.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: lt, gt, lte, gte, eq, neq",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "Compare.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("Compare.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "Compare.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("Compare.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *SetOp) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "SetOp.Op: is required",
			Field:   "Op",
			Want:    "one of: union, intersect, difference",
		}
	}
	if !validSetOpKinds[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("SetOp.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: union, intersect, difference",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "SetOp.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("SetOp.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "SetOp.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("SetOp.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *SetCompare) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "SetCompare.Op: is required",
			Field:   "Op",
			Want:    "one of: subset_eq, subset, superset_eq, superset",
		}
	}
	if !validSetCompareOps[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("SetCompare.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: subset_eq, subset, superset_eq, superset",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "SetCompare.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("SetCompare.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "SetCompare.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("SetCompare.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *BagOp) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "BagOp.Op: is required",
			Field:   "Op",
			Want:    "one of: sum, difference",
		}
	}
	if !validBagOpKinds[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("BagOp.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: sum, difference",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "BagOp.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("BagOp.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "BagOp.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("BagOp.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *BagCompare) Validate() error {
	if n.Op == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpRequired,
			Message: "BagCompare.Op: is required",
			Field:   "Op",
			Want:    "one of: proper_sub_bag, sub_bag, proper_sup_bag, sup_bag",
		}
	}
	if !validBagCompareOps[n.Op] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOpInvalid,
			Message: fmt.Sprintf("BagCompare.Op: '%s' is not valid", string(n.Op)),
			Field:   "Op",
			Got:     string(n.Op),
			Want:    "one of: proper_sub_bag, sub_bag, proper_sup_bag, sup_bag",
		}
	}
	if n.Left == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftRequired,
			Message: "BagCompare.Left: is required",
			Field:   "Left",
		}
	}
	if err := n.Left.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprLeftInvalid,
			Message: fmt.Sprintf("BagCompare.Left: %s", err.Error()),
			Field:   "Left",
		}
	}
	if n.Right == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightRequired,
			Message: "BagCompare.Right: is required",
			Field:   "Right",
		}
	}
	if err := n.Right.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprRightInvalid,
			Message: fmt.Sprintf("BagCompare.Right: %s", err.Error()),
			Field:   "Right",
		}
	}
	return nil
}

func (n *Membership) Validate() error {
	if n.Element == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprElementRequired,
			Message: "Membership.Element: is required",
			Field:   "Element",
		}
	}
	if err := n.Element.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprElementInvalid,
			Message: fmt.Sprintf("Membership.Element: %s", err.Error()),
			Field:   "Element",
		}
	}
	if n.Set == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetRequired,
			Message: "Membership.Set: is required",
			Field:   "Set",
		}
	}
	if err := n.Set.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetInvalid,
			Message: fmt.Sprintf("Membership.Set: %s", err.Error()),
			Field:   "Set",
		}
	}
	return nil
}

// --- Unary operator validation ---

func (n *Negate) Validate() error {
	if n.Expr == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprExprRequired,
			Message: "Negate.Expr: is required",
			Field:   "Expr",
		}
	}
	return n.Expr.Validate()
}

func (n *Not) Validate() error {
	if n.Expr == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprExprRequired,
			Message: "Not.Expr: is required",
			Field:   "Expr",
		}
	}
	return n.Expr.Validate()
}

// --- Collection validation ---

func (n *FieldAccess) Validate() error {
	if n.Field == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprFieldRequired,
			Message: "FieldAccess.Field: is required",
			Field:   "Field",
		}
	}
	if n.Base == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprBaseRequired,
			Message: "FieldAccess.Base: is required",
			Field:   "Base",
		}
	}
	if err := n.Base.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprBaseInvalid,
			Message: fmt.Sprintf("FieldAccess.Base: %s", err.Error()),
			Field:   "Base",
		}
	}
	return nil
}

func (n *TupleIndex) Validate() error {
	if n.Tuple == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprTupleRequired,
			Message: "TupleIndex.Tuple: is required",
			Field:   "Tuple",
		}
	}
	if err := n.Tuple.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprTupleInvalid,
			Message: fmt.Sprintf("TupleIndex.Tuple: %s", err.Error()),
			Field:   "Tuple",
		}
	}
	if n.Index == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprIndexRequired,
			Message: "TupleIndex.Index: is required",
			Field:   "Index",
		}
	}
	if err := n.Index.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprIndexInvalid,
			Message: fmt.Sprintf("TupleIndex.Index: %s", err.Error()),
			Field:   "Index",
		}
	}
	return nil
}

func (n *RecordUpdate) Validate() error {
	if n.Base == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprBaseRequired,
			Message: "RecordUpdate.Base: is required",
			Field:   "Base",
		}
	}
	if err := n.Base.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprBaseInvalid,
			Message: fmt.Sprintf("RecordUpdate.Base: %s", err.Error()),
			Field:   "Base",
		}
	}
	if len(n.Alterations) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprAlterationsRequired,
			Message: "RecordUpdate.Alterations: at least one alteration is required",
			Field:   "Alterations",
			Want:    "min=1",
		}
	}
	for i, alt := range n.Alterations {
		if alt.Field == "" {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprAltFieldRequired,
				Message: fmt.Sprintf("RecordUpdate.Alterations[%d].Field: is required", i),
				Field:   fmt.Sprintf("Alterations[%d].Field", i),
			}
		}
		if alt.Value == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprAltValueRequired,
				Message: fmt.Sprintf("RecordUpdate.Alterations[%d].Value: is required", i),
				Field:   fmt.Sprintf("Alterations[%d].Value", i),
			}
		}
		if err := alt.Value.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprAltValueInvalid,
				Message: fmt.Sprintf("RecordUpdate.Alterations[%d].Value: %s", i, err.Error()),
				Field:   fmt.Sprintf("Alterations[%d].Value", i),
			}
		}
	}
	return nil
}

func (n *StringIndex) Validate() error {
	if n.Str == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprStrRequired,
			Message: "StringIndex.Str: is required",
			Field:   "Str",
		}
	}
	if err := n.Str.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprStrInvalid,
			Message: fmt.Sprintf("StringIndex.Str: %s", err.Error()),
			Field:   "Str",
		}
	}
	if n.Index == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprIndexRequired,
			Message: "StringIndex.Index: is required",
			Field:   "Index",
		}
	}
	if err := n.Index.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprIndexInvalid,
			Message: fmt.Sprintf("StringIndex.Index: %s", err.Error()),
			Field:   "Index",
		}
	}
	return nil
}

func (n *StringConcat) Validate() error {
	if len(n.Operands) < 2 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOperandsMinTwo,
			Message: "StringConcat.Operands: at least two operands are required",
			Field:   "Operands",
			Want:    "min=2",
		}
	}
	for i, op := range n.Operands {
		if op == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprOperandRequired,
				Message: fmt.Sprintf("StringConcat.Operands[%d]: is required", i),
				Field:   fmt.Sprintf("Operands[%d]", i),
			}
		}
		if err := op.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprOperandInvalid,
				Message: fmt.Sprintf("StringConcat.Operands[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Operands[%d]", i),
			}
		}
	}
	return nil
}

func (n *TupleConcat) Validate() error {
	if len(n.Operands) < 2 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprOperandsMinTwo,
			Message: "TupleConcat.Operands: at least two operands are required",
			Field:   "Operands",
			Want:    "min=2",
		}
	}
	for i, op := range n.Operands {
		if op == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprOperandRequired,
				Message: fmt.Sprintf("TupleConcat.Operands[%d]: is required", i),
				Field:   fmt.Sprintf("Operands[%d]", i),
			}
		}
		if err := op.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprOperandInvalid,
				Message: fmt.Sprintf("TupleConcat.Operands[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Operands[%d]", i),
			}
		}
	}
	return nil
}

// --- Control flow validation ---

func (n *IfThenElse) Validate() error {
	if n.Condition == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprConditionRequired,
			Message: "IfThenElse.Condition: is required",
			Field:   "Condition",
		}
	}
	if err := n.Condition.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprConditionInvalid,
			Message: fmt.Sprintf("IfThenElse.Condition: %s", err.Error()),
			Field:   "Condition",
		}
	}
	if n.Then == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprThenRequired,
			Message: "IfThenElse.Then: is required",
			Field:   "Then",
		}
	}
	if err := n.Then.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprThenInvalid,
			Message: fmt.Sprintf("IfThenElse.Then: %s", err.Error()),
			Field:   "Then",
		}
	}
	if n.Else == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprElseRequired,
			Message: "IfThenElse.Else: is required",
			Field:   "Else",
		}
	}
	if err := n.Else.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprElseInvalid,
			Message: fmt.Sprintf("IfThenElse.Else: %s", err.Error()),
			Field:   "Else",
		}
	}
	return nil
}

func (n *Case) Validate() error {
	if len(n.Branches) == 0 {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprBranchesRequired,
			Message: "Case.Branches: at least one branch is required",
			Field:   "Branches",
			Want:    "min=1",
		}
	}
	for i, branch := range n.Branches {
		if branch.Condition == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprBranchCondRequired,
				Message: fmt.Sprintf("Case.Branches[%d].Condition: is required", i),
				Field:   fmt.Sprintf("Branches[%d].Condition", i),
			}
		}
		if err := branch.Condition.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprBranchCondInvalid,
				Message: fmt.Sprintf("Case.Branches[%d].Condition: %s", i, err.Error()),
				Field:   fmt.Sprintf("Branches[%d].Condition", i),
			}
		}
		if branch.Result == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprBranchResultRequired,
				Message: fmt.Sprintf("Case.Branches[%d].Result: is required", i),
				Field:   fmt.Sprintf("Branches[%d].Result", i),
			}
		}
		if err := branch.Result.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprBranchResultInvalid,
				Message: fmt.Sprintf("Case.Branches[%d].Result: %s", i, err.Error()),
				Field:   fmt.Sprintf("Branches[%d].Result", i),
			}
		}
	}
	if n.Otherwise != nil {
		if err := n.Otherwise.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprOtherwiseInvalid,
				Message: fmt.Sprintf("Case.Otherwise: %s", err.Error()),
				Field:   "Otherwise",
			}
		}
	}
	return nil
}

// --- Quantifier validation ---

func (n *Quantifier) Validate() error {
	if n.Domain == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprDomainRequired,
			Message: "Quantifier.Domain: is required",
			Field:   "Domain",
		}
	}
	if n.Predicate == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprPredicateRequired,
			Message: "Quantifier.Predicate: is required",
			Field:   "Predicate",
		}
	}
	if n.Kind == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprQuantKindRequired,
			Message: "Quantifier.Kind: is required",
			Field:   "Kind",
			Want:    "one of: forall, exists",
		}
	}
	if !validQuantifierKinds[n.Kind] {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprQuantKindInvalid,
			Message: fmt.Sprintf("Quantifier.Kind: '%s' is not valid", string(n.Kind)),
			Field:   "Kind",
			Got:     string(n.Kind),
			Want:    "one of: forall, exists",
		}
	}
	if n.Variable == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprVariableRequired,
			Message: "Quantifier.Variable: is required",
			Field:   "Variable",
		}
	}
	if err := n.Domain.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprDomainInvalid,
			Message: fmt.Sprintf("Quantifier.Domain: %s", err.Error()),
			Field:   "Domain",
		}
	}
	if err := n.Predicate.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprPredicateInvalid,
			Message: fmt.Sprintf("Quantifier.Predicate: %s", err.Error()),
			Field:   "Predicate",
		}
	}
	return nil
}

func (n *SetFilter) Validate() error {
	if n.Set == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetRequired,
			Message: "SetFilter.Set: is required",
			Field:   "Set",
		}
	}
	if n.Predicate == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprPredicateRequired,
			Message: "SetFilter.Predicate: is required",
			Field:   "Predicate",
		}
	}
	if n.Variable == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprVariableRequired,
			Message: "SetFilter.Variable: is required",
			Field:   "Variable",
		}
	}
	if err := n.Set.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetInvalid,
			Message: fmt.Sprintf("SetFilter.Set: %s", err.Error()),
			Field:   "Set",
		}
	}
	if err := n.Predicate.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprPredicateInvalid,
			Message: fmt.Sprintf("SetFilter.Predicate: %s", err.Error()),
			Field:   "Predicate",
		}
	}
	return nil
}

func (n *SetRange) Validate() error {
	if n.Start == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprStartRequired,
			Message: "SetRange.Start: is required",
			Field:   "Start",
		}
	}
	if err := n.Start.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprStartInvalid,
			Message: fmt.Sprintf("SetRange.Start: %s", err.Error()),
			Field:   "Start",
		}
	}
	if n.End == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprEndRequired,
			Message: "SetRange.End: is required",
			Field:   "End",
		}
	}
	if err := n.End.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprEndInvalid,
			Message: fmt.Sprintf("SetRange.End: %s", err.Error()),
			Field:   "End",
		}
	}
	return nil
}

// --- Call validation ---

func (n *ActionCall) Validate() error {
	if err := n.ActionKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprActionkeyInvalid,
			Message: fmt.Sprintf("ActionCall.ActionKey: %s", err.Error()),
			Field:   "ActionKey",
		}
	}
	for i, arg := range n.Args {
		if arg == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgRequired,
				Message: fmt.Sprintf("ActionCall.Args[%d]: is required", i),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
		if err := arg.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgInvalid,
				Message: fmt.Sprintf("ActionCall.Args[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
	}
	return nil
}

func (n *GlobalCall) Validate() error {
	if err := n.FunctionKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprFunctionkeyInvalid,
			Message: fmt.Sprintf("GlobalCall.FunctionKey: %s", err.Error()),
			Field:   "FunctionKey",
		}
	}
	for i, arg := range n.Args {
		if arg == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgRequired,
				Message: fmt.Sprintf("GlobalCall.Args[%d]: is required", i),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
		if err := arg.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgInvalid,
				Message: fmt.Sprintf("GlobalCall.Args[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
	}
	return nil
}

func (n *BuiltinCall) Validate() error {
	if n.Module == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprModuleRequired,
			Message: "BuiltinCall.Module: is required",
			Field:   "Module",
		}
	}
	if n.Function == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprFunctionRequired,
			Message: "BuiltinCall.Function: is required",
			Field:   "Function",
		}
	}
	for i, arg := range n.Args {
		if arg == nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgRequired,
				Message: fmt.Sprintf("BuiltinCall.Args[%d]: is required", i),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
		if err := arg.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ExprArgInvalid,
				Message: fmt.Sprintf("BuiltinCall.Args[%d]: %s", i, err.Error()),
				Field:   fmt.Sprintf("Args[%d]", i),
			}
		}
	}
	return nil
}

// --- Named set reference validation ---

func (n *NamedSetRef) Validate() error {
	if err := n.SetKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ExprSetkeyInvalid,
			Message: fmt.Sprintf("NamedSetRef.SetKey: %s", err.Error()),
			Field:   "SetKey",
		}
	}
	return nil
}
