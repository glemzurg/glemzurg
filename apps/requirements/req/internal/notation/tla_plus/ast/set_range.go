package ast

import (
	"bytes"
	"strconv"
)

// SetRange is a set of consecutive integers from Start to End (inclusive).
// Pattern: -5 .. 5 represents the set {-5, -4, -3, ..., 5}
type SetRange struct {
	Start int `validate:"ltefield=End"` // The starting integer (inclusive)
	End   int ``                        // The ending integer (inclusive)
}

func (s *SetRange) expressionNode() {}

func (s *SetRange) String() (value string) {
	var out bytes.Buffer
	out.WriteString(strconv.Itoa(s.Start))
	out.WriteString(" .. ")
	out.WriteString(strconv.Itoa(s.End))
	return out.String()
}

func (s *SetRange) Ascii() (value string) {
	return s.String()
}

func (s *SetRange) Validate() error {
	return _validate.Struct(s)
}
