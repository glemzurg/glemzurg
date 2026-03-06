package ast

import (
	"bytes"
	"strings"
)

// SetLiteral is a general set literal containing expressions.
// Pattern: {expr1, expr2, expr3} or {} (empty set)
// This is more general than SetLiteralInt (integers only) or SetLiteralEnum (strings only).
type SetLiteral struct {
	Elements []Expression // The set elements (can be empty for {})
}

func (s *SetLiteral) expressionNode() {}

func (s *SetLiteral) String() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	strs := make([]string, len(s.Elements))
	for i, elem := range s.Elements {
		strs[i] = elem.String()
	}
	out.WriteString(strings.Join(strs, ", "))
	out.WriteString("}")
	return out.String()
}

func (s *SetLiteral) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	strs := make([]string, len(s.Elements))
	for i, elem := range s.Elements {
		strs[i] = elem.Ascii()
	}
	out.WriteString(strings.Join(strs, ", "))
	out.WriteString("}")
	return out.String()
}

func (s *SetLiteral) Validate() error {
	for _, elem := range s.Elements {
		if err := elem.Validate(); err != nil {
			return err
		}
	}
	return nil
}
