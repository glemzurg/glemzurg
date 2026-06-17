package ast

import "fmt"

// ChooseExpr picks one element from a set: CHOOSE var ∈ set : predicate.
type ChooseExpr struct {
	Membership Expression `validate:"required"`
	Predicate  Expression `validate:"required"`
}

func (e *ChooseExpr) expressionNode() {}

func (e *ChooseExpr) String() string {
	return "CHOOSE " + e.Membership.String() + " : " + e.Predicate.String()
}

func (e *ChooseExpr) ASCII() string {
	return "CHOOSE " + e.Membership.ASCII() + " : " + e.Predicate.ASCII()
}

func (e *ChooseExpr) Validate() error {
	if err := _validate.Struct(e); err != nil {
		return err
	}
	if err := e.Membership.Validate(); err != nil {
		return fmt.Errorf("membership: %w", err)
	}
	if err := e.Predicate.Validate(); err != nil {
		return fmt.Errorf("predicate: %w", err)
	}
	return nil
}
