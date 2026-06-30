package ast

import "bytes"

// SetMap is a set comprehension that maps an expression over each element of a set.
// Pattern: {f(x) : x ∈ S} — the image of S under f.
type SetMap struct {
	Transform  Expression `validate:"required"` // Expression applied to each element
	Membership Expression `validate:"required"` // The membership binding (must be Membership)
}

func (sc *SetMap) expressionNode() {}

func (sc *SetMap) String() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Transform.String())
	out.WriteString(" : ")
	out.WriteString(sc.Membership.String())
	out.WriteString("}")
	return out.String()
}

func (sc *SetMap) ASCII() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Transform.ASCII())
	out.WriteString(" : ")
	out.WriteString(sc.Membership.ASCII())
	out.WriteString("}")
	return out.String()
}

func (sc *SetMap) Validate() error {
	if err := _validate.Struct(sc); err != nil {
		return err
	}
	if err := sc.Transform.Validate(); err != nil {
		return err
	}
	if err := sc.Membership.Validate(); err != nil {
		return err
	}
	return nil
}
