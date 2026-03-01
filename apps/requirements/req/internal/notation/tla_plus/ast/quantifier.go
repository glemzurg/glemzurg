package ast

import "bytes"

// Quantifier operators
const (
	QuantifierForAll = "∀" // Universal quantifier
	QuantifierExists = "∃" // Existential quantifier
)

var quantifierAscii = map[string]string{
	QuantifierForAll: `\A`,
	QuantifierExists: `\E`,
}

// Backwards compatibility constants.
// Deprecated: Use Quantifier* constants instead.
const (
	LogicQuantifierForAll = QuantifierForAll
	LogicQuantifierExists = QuantifierExists
)

// Quantifier is a quantified logic expression over a set.
// Examples: ∀x ∈ S : p, ∃x ∈ S : p
type Quantifier struct {
	Quantifier string     `validate:"required,oneof=∀ ∃"` // The quantifier, e.g., ∀, ∃
	Membership Expression `validate:"required"`           // The membership test (must be Membership)
	Predicate  Expression `validate:"required"`           // The predicate expression (must be Boolean)
}

func (q *Quantifier) expressionNode() {}

func (q *Quantifier) String() (value string) {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(q.Quantifier)
	out.WriteString(q.Membership.String())
	out.WriteString(" : ")
	out.WriteString(q.Predicate.String())
	out.WriteString(")")
	return out.String()
}

func (q *Quantifier) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(quantifierAscii[q.Quantifier])
	out.WriteString(" ")
	out.WriteString(q.Membership.Ascii())
	out.WriteString(" : ")
	out.WriteString(q.Predicate.Ascii())
	out.WriteString(")")
	return out.String()
}

func (q *Quantifier) Validate() error {
	if err := _validate.Struct(q); err != nil {
		return err
	}
	if err := q.Membership.Validate(); err != nil {
		return err
	}
	if err := q.Predicate.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicBoundQuantifier is an alias for backwards compatibility.
// Deprecated: Use Quantifier instead.
type LogicBoundQuantifier = Quantifier
