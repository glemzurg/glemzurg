package ast

import "bytes"

// SetFilter is a set comprehension that filters elements from a set.
// Pattern: {x ∈ S : p} - the set of all x in S where p is true.
type SetFilter struct {
	Membership Expression `validate:"required"` // The membership test (must be Membership)
	Predicate  Expression `validate:"required"` // The filtering predicate (must be Boolean)
}

func (sc *SetFilter) expressionNode() {}

func (sc *SetFilter) String() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Membership.String())
	out.WriteString(" : ")
	out.WriteString(sc.Predicate.String())
	out.WriteString("}")
	return out.String()
}

func (sc *SetFilter) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Membership.Ascii())
	out.WriteString(" : ")
	out.WriteString(sc.Predicate.Ascii())
	out.WriteString("}")
	return out.String()
}

func (sc *SetFilter) Validate() error {
	if err := _validate.Struct(sc); err != nil {
		return err
	}
	if err := sc.Membership.Validate(); err != nil {
		return err
	}
	if err := sc.Predicate.Validate(); err != nil {
		return err
	}
	return nil
}

// SetConditional is an alias for backwards compatibility.
// Deprecated: Use SetFilter instead.
type SetConditional = SetFilter
