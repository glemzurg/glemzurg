package ast

import "bytes"

// Set comparison operators
const (
	SetComparisonEqual      = "="
	SetComparisonNotEqual   = "≠"
	SetComparisonSubsetEq   = "⊆"
	SetComparisonSubset     = "⊂"
	SetComparisonSupersetEq = "⊇"
	SetComparisonSuperset   = "⊃"
)

var setComparisonAscii = map[string]string{
	SetComparisonEqual:      `=`,
	SetComparisonNotEqual:   `/=`,
	SetComparisonSubsetEq:   `\subseteq`,
	SetComparisonSubset:     `\subset`,
	SetComparisonSupersetEq: `\supseteq`,
	SetComparisonSuperset:   `\supset`,
}

// Backwards compatibility constants.
// Deprecated: Use SetComparison* constants instead.
const (
	LogicSetOperatorEqual      = SetComparisonEqual
	LogicSetOperatorNotEqual   = SetComparisonNotEqual
	LogicSetOperatorSubsetEq   = SetComparisonSubsetEq
	LogicSetOperatorSubset     = SetComparisonSubset
	LogicSetOperatorSupersetEq = SetComparisonSupersetEq
	LogicSetOperatorSuperset   = SetComparisonSuperset
)

// BinarySetComparison is a binary comparison operation between two sets.
type BinarySetComparison struct {
	Operator string     `validate:"required,oneof== ≠ ⊆ ⊂ ⊇ ⊃"` // The comparison operator, e.g., =, ≠, ⊆, ⊂, ⊇, ⊃
	Left     Expression `validate:"required"`                     // Must be Set
	Right    Expression `validate:"required"`                     // Must be Set
}

func (is *BinarySetComparison) expressionNode() {}

func (is *BinarySetComparison) String() (value string) {
	var out bytes.Buffer
	out.WriteString(is.Left.String())
	out.WriteString(" " + is.Operator + " ")
	out.WriteString(is.Right.String())
	return out.String()
}

func (is *BinarySetComparison) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(is.Left.Ascii())
	out.WriteString(" " + setComparisonAscii[is.Operator] + " ")
	out.WriteString(is.Right.Ascii())
	return out.String()
}

func (is *BinarySetComparison) Validate() error {
	if err := _validate.Struct(is); err != nil {
		return err
	}
	if err := is.Left.Validate(); err != nil {
		return err
	}
	if err := is.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicInfixSet is an alias for backwards compatibility.
// Deprecated: Use BinarySetComparison instead.
type LogicInfixSet = BinarySetComparison
