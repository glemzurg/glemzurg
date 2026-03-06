package ast

import "bytes"

// Membership operators
const (
	MembershipOperatorIn    = "∈"
	MembershipOperatorNotIn = "∉"
)

var membershipAscii = map[string]string{
	MembershipOperatorIn:    `\in`,
	MembershipOperatorNotIn: `\notin`,
}

// Membership is a membership test between an element and a set.
type Membership struct {
	Operator string     `validate:"required,oneof=∈ ∉"` // The membership operator, e.g., ∈, ∉
	Left     Expression `validate:"required"`           // The element being tested
	Right    Expression `validate:"required"`           // The set being tested against (must be Set)
}

func (m *Membership) expressionNode() {}

func (m *Membership) String() (value string) {
	var out bytes.Buffer
	out.WriteString(m.Left.String())
	out.WriteString(" " + m.Operator + " ")
	out.WriteString(m.Right.String())
	return out.String()
}

func (m *Membership) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(m.Left.Ascii())
	out.WriteString(" " + membershipAscii[m.Operator] + " ")
	out.WriteString(m.Right.Ascii())
	return out.String()
}

func (m *Membership) Validate() error {
	if err := _validate.Struct(m); err != nil {
		return err
	}
	if err := m.Left.Validate(); err != nil {
		return err
	}
	if err := m.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicMembership is an alias for backwards compatibility.
// Deprecated: Use Membership instead.
type LogicMembership = Membership
