package ast

import (
	"bytes"
	"fmt"
)

// CaseBranch represents a single branch in a CASE expression.
// Pattern: condition → expression
type CaseBranch struct {
	Condition Expression `validate:"required"` // Must be Boolean
	Result    Expression `validate:"required"`
}

// CaseExpr represents a CASE expression with multiple branches.
// Pattern: CASE n ≥ 0 → e □ n ≤ 0 → e □ OTHER → e
// The Other branch is optional.
type CaseExpr struct {
	Branches []*CaseBranch `validate:"required,min=1"` // At least one branch
	Other    Expression    // Optional OTHER branch
}

func (e *CaseExpr) expressionNode() {}

func (e *CaseExpr) String() (value string) {
	var out bytes.Buffer
	out.WriteString("CASE ")
	for i, branch := range e.Branches {
		if i > 0 {
			out.WriteString(" □ ")
		}
		out.WriteString(branch.Condition.String())
		out.WriteString(" → ")
		out.WriteString(branch.Result.String())
	}
	if e.Other != nil {
		out.WriteString(" □ OTHER → ")
		out.WriteString(e.Other.String())
	}
	return out.String()
}

func (e *CaseExpr) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("CASE ")
	for i, branch := range e.Branches {
		if i > 0 {
			out.WriteString(" [] ")
		}
		out.WriteString(branch.Condition.Ascii())
		out.WriteString(" -> ")
		out.WriteString(branch.Result.Ascii())
	}
	if e.Other != nil {
		out.WriteString(" [] OTHER -> ")
		out.WriteString(e.Other.Ascii())
	}
	return out.String()
}

func (e *CaseExpr) Validate() error {
	if err := _validate.Struct(e); err != nil {
		return err
	}
	for i, branch := range e.Branches {
		if branch == nil {
			return fmt.Errorf("Branches[%d]: is nil", i)
		}
		if err := branch.Condition.Validate(); err != nil {
			return fmt.Errorf("Branches[%d].Condition: %w", i, err)
		}
		if err := branch.Result.Validate(); err != nil {
			return fmt.Errorf("Branches[%d].Result: %w", i, err)
		}
	}
	if e.Other != nil {
		if err := e.Other.Validate(); err != nil {
			return fmt.Errorf("Other: %w", err)
		}
	}
	return nil
}

// ExpressionCase is an alias for backwards compatibility.
// Deprecated: Use CaseExpr instead.
type ExpressionCase = CaseExpr
