package model_expression

import "fmt"

// --- Literal validation ---

func (n *BoolLiteral) Validate() error     { return nil }
func (n *IntLiteral) Validate() error      { return nil }
func (n *StringLiteral) Validate() error   { return nil }

func (n *RationalLiteral) Validate() error {
	if n.Denominator == 0 {
		return fmt.Errorf("RationalLiteral: denominator cannot be zero")
	}
	return nil
}

func (n *SetLiteral) Validate() error {
	for i, elem := range n.Elements {
		if elem == nil {
			return fmt.Errorf("SetLiteral.Elements[%d]: is required", i)
		}
		if err := elem.Validate(); err != nil {
			return fmt.Errorf("SetLiteral.Elements[%d]: %w", i, err)
		}
	}
	return nil
}

func (n *TupleLiteral) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	for i, elem := range n.Elements {
		if elem == nil {
			return fmt.Errorf("TupleLiteral.Elements[%d]: is required", i)
		}
		if err := elem.Validate(); err != nil {
			return fmt.Errorf("TupleLiteral.Elements[%d]: %w", i, err)
		}
	}
	return nil
}

func (n *RecordLiteral) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	for i, field := range n.Fields {
		if field.Name == "" {
			return fmt.Errorf("RecordLiteral.Fields[%d].Name: is required", i)
		}
		if field.Value == nil {
			return fmt.Errorf("RecordLiteral.Fields[%d].Value: is required", i)
		}
		if err := field.Value.Validate(); err != nil {
			return fmt.Errorf("RecordLiteral.Fields[%d].Value: %w", i, err)
		}
	}
	return nil
}

func (n *SetConstant) Validate() error {
	return _validate.Struct(n)
}

// --- Reference validation ---

func (n *SelfRef) Validate() error { return nil }

func (n *AttributeRef) Validate() error {
	if err := n.AttributeKey.Validate(); err != nil {
		return fmt.Errorf("AttributeRef.AttributeKey: %w", err)
	}
	return nil
}

func (n *LocalVar) Validate() error {
	return _validate.Struct(n)
}

func (n *PriorFieldValue) Validate() error {
	return _validate.Struct(n)
}

func (n *NextState) Validate() error {
	if n.Expr == nil {
		return fmt.Errorf("NextState.Expr: is required")
	}
	return n.Expr.Validate()
}

// --- Binary operator validation ---

func (n *BinaryArith) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("BinaryArith.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("BinaryArith.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("BinaryArith.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("BinaryArith.Right: %w", err)
	}
	return nil
}

func (n *BinaryLogic) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("BinaryLogic.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("BinaryLogic.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("BinaryLogic.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("BinaryLogic.Right: %w", err)
	}
	return nil
}

func (n *Compare) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("Compare.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("Compare.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("Compare.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("Compare.Right: %w", err)
	}
	return nil
}

func (n *SetOp) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("SetOp.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("SetOp.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("SetOp.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("SetOp.Right: %w", err)
	}
	return nil
}

func (n *SetCompare) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("SetCompare.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("SetCompare.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("SetCompare.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("SetCompare.Right: %w", err)
	}
	return nil
}

func (n *BagOp) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("BagOp.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("BagOp.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("BagOp.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("BagOp.Right: %w", err)
	}
	return nil
}

func (n *BagCompare) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Left == nil {
		return fmt.Errorf("BagCompare.Left: is required")
	}
	if err := n.Left.Validate(); err != nil {
		return fmt.Errorf("BagCompare.Left: %w", err)
	}
	if n.Right == nil {
		return fmt.Errorf("BagCompare.Right: is required")
	}
	if err := n.Right.Validate(); err != nil {
		return fmt.Errorf("BagCompare.Right: %w", err)
	}
	return nil
}

func (n *Membership) Validate() error {
	if n.Element == nil {
		return fmt.Errorf("Membership.Element: is required")
	}
	if err := n.Element.Validate(); err != nil {
		return fmt.Errorf("Membership.Element: %w", err)
	}
	if n.Set == nil {
		return fmt.Errorf("Membership.Set: is required")
	}
	if err := n.Set.Validate(); err != nil {
		return fmt.Errorf("Membership.Set: %w", err)
	}
	return nil
}

// --- Unary operator validation ---

func (n *Negate) Validate() error {
	if n.Expr == nil {
		return fmt.Errorf("Negate.Expr: is required")
	}
	return n.Expr.Validate()
}

func (n *Not) Validate() error {
	if n.Expr == nil {
		return fmt.Errorf("Not.Expr: is required")
	}
	return n.Expr.Validate()
}

// --- Collection validation ---

func (n *FieldAccess) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if n.Base == nil {
		return fmt.Errorf("FieldAccess.Base: is required")
	}
	return n.Base.Validate()
}

func (n *TupleIndex) Validate() error {
	if n.Tuple == nil {
		return fmt.Errorf("TupleIndex.Tuple: is required")
	}
	if err := n.Tuple.Validate(); err != nil {
		return fmt.Errorf("TupleIndex.Tuple: %w", err)
	}
	if n.Index == nil {
		return fmt.Errorf("TupleIndex.Index: is required")
	}
	if err := n.Index.Validate(); err != nil {
		return fmt.Errorf("TupleIndex.Index: %w", err)
	}
	return nil
}

func (n *RecordUpdate) Validate() error {
	if n.Base == nil {
		return fmt.Errorf("RecordUpdate.Base: is required")
	}
	if err := n.Base.Validate(); err != nil {
		return fmt.Errorf("RecordUpdate.Base: %w", err)
	}
	if len(n.Alterations) == 0 {
		return fmt.Errorf("RecordUpdate.Alterations: at least one alteration is required")
	}
	for i, alt := range n.Alterations {
		if alt.Field == "" {
			return fmt.Errorf("RecordUpdate.Alterations[%d].Field: is required", i)
		}
		if alt.Value == nil {
			return fmt.Errorf("RecordUpdate.Alterations[%d].Value: is required", i)
		}
		if err := alt.Value.Validate(); err != nil {
			return fmt.Errorf("RecordUpdate.Alterations[%d].Value: %w", i, err)
		}
	}
	return nil
}

func (n *StringIndex) Validate() error {
	if n.Str == nil {
		return fmt.Errorf("StringIndex.Str: is required")
	}
	if err := n.Str.Validate(); err != nil {
		return fmt.Errorf("StringIndex.Str: %w", err)
	}
	if n.Index == nil {
		return fmt.Errorf("StringIndex.Index: is required")
	}
	if err := n.Index.Validate(); err != nil {
		return fmt.Errorf("StringIndex.Index: %w", err)
	}
	return nil
}

func (n *StringConcat) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	for i, op := range n.Operands {
		if op == nil {
			return fmt.Errorf("StringConcat.Operands[%d]: is required", i)
		}
		if err := op.Validate(); err != nil {
			return fmt.Errorf("StringConcat.Operands[%d]: %w", i, err)
		}
	}
	return nil
}

func (n *TupleConcat) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	for i, op := range n.Operands {
		if op == nil {
			return fmt.Errorf("TupleConcat.Operands[%d]: is required", i)
		}
		if err := op.Validate(); err != nil {
			return fmt.Errorf("TupleConcat.Operands[%d]: %w", i, err)
		}
	}
	return nil
}

// --- Control flow validation ---

func (n *IfThenElse) Validate() error {
	if n.Condition == nil {
		return fmt.Errorf("IfThenElse.Condition: is required")
	}
	if err := n.Condition.Validate(); err != nil {
		return fmt.Errorf("IfThenElse.Condition: %w", err)
	}
	if n.Then == nil {
		return fmt.Errorf("IfThenElse.Then: is required")
	}
	if err := n.Then.Validate(); err != nil {
		return fmt.Errorf("IfThenElse.Then: %w", err)
	}
	if n.Else == nil {
		return fmt.Errorf("IfThenElse.Else: is required")
	}
	if err := n.Else.Validate(); err != nil {
		return fmt.Errorf("IfThenElse.Else: %w", err)
	}
	return nil
}

func (n *Case) Validate() error {
	if len(n.Branches) == 0 {
		return fmt.Errorf("Case.Branches: at least one branch is required")
	}
	for i, branch := range n.Branches {
		if branch.Condition == nil {
			return fmt.Errorf("Case.Branches[%d].Condition: is required", i)
		}
		if err := branch.Condition.Validate(); err != nil {
			return fmt.Errorf("Case.Branches[%d].Condition: %w", i, err)
		}
		if branch.Result == nil {
			return fmt.Errorf("Case.Branches[%d].Result: is required", i)
		}
		if err := branch.Result.Validate(); err != nil {
			return fmt.Errorf("Case.Branches[%d].Result: %w", i, err)
		}
	}
	if n.Otherwise != nil {
		if err := n.Otherwise.Validate(); err != nil {
			return fmt.Errorf("Case.Otherwise: %w", err)
		}
	}
	return nil
}

// --- Quantifier validation ---

func (n *Quantifier) Validate() error {
	if n.Domain == nil {
		return fmt.Errorf("Quantifier.Domain: is required")
	}
	if n.Predicate == nil {
		return fmt.Errorf("Quantifier.Predicate: is required")
	}
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if err := n.Domain.Validate(); err != nil {
		return fmt.Errorf("Quantifier.Domain: %w", err)
	}
	if err := n.Predicate.Validate(); err != nil {
		return fmt.Errorf("Quantifier.Predicate: %w", err)
	}
	return nil
}

func (n *SetFilter) Validate() error {
	if n.Set == nil {
		return fmt.Errorf("SetFilter.Set: is required")
	}
	if n.Predicate == nil {
		return fmt.Errorf("SetFilter.Predicate: is required")
	}
	if err := _validate.Struct(n); err != nil {
		return err
	}
	if err := n.Set.Validate(); err != nil {
		return fmt.Errorf("SetFilter.Set: %w", err)
	}
	if err := n.Predicate.Validate(); err != nil {
		return fmt.Errorf("SetFilter.Predicate: %w", err)
	}
	return nil
}

func (n *SetRange) Validate() error {
	if n.Start == nil {
		return fmt.Errorf("SetRange.Start: is required")
	}
	if err := n.Start.Validate(); err != nil {
		return fmt.Errorf("SetRange.Start: %w", err)
	}
	if n.End == nil {
		return fmt.Errorf("SetRange.End: is required")
	}
	if err := n.End.Validate(); err != nil {
		return fmt.Errorf("SetRange.End: %w", err)
	}
	return nil
}

// --- Call validation ---

func (n *ActionCall) Validate() error {
	if err := n.ActionKey.Validate(); err != nil {
		return fmt.Errorf("ActionCall.ActionKey: %w", err)
	}
	for i, arg := range n.Args {
		if arg == nil {
			return fmt.Errorf("ActionCall.Args[%d]: is required", i)
		}
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("ActionCall.Args[%d]: %w", i, err)
		}
	}
	return nil
}

func (n *GlobalCall) Validate() error {
	if err := n.FunctionKey.Validate(); err != nil {
		return fmt.Errorf("GlobalCall.FunctionKey: %w", err)
	}
	for i, arg := range n.Args {
		if arg == nil {
			return fmt.Errorf("GlobalCall.Args[%d]: is required", i)
		}
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("GlobalCall.Args[%d]: %w", i, err)
		}
	}
	return nil
}

func (n *BuiltinCall) Validate() error {
	if err := _validate.Struct(n); err != nil {
		return err
	}
	for i, arg := range n.Args {
		if arg == nil {
			return fmt.Errorf("BuiltinCall.Args[%d]: is required", i)
		}
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("BuiltinCall.Args[%d]: %w", i, err)
		}
	}
	return nil
}

// --- Named set reference validation ---

func (n *NamedSetRef) Validate() error {
	if err := n.SetKey.Validate(); err != nil {
		return fmt.Errorf("NamedSetRef.SetKey: %w", err)
	}
	return nil
}
