package ast

import (
	"bytes"
	"strconv"
	"strings"
)

// SetLiteralInt is a set of specific integer values.
// Pattern: {1, 2, 4, 6}
type SetLiteralInt struct {
	Values []int `validate:"required,min=1"` // The integer values
}

func (s *SetLiteralInt) expressionNode() {}

func (s *SetLiteralInt) String() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	strs := make([]string, len(s.Values))
	for i, v := range s.Values {
		strs[i] = strconv.Itoa(v)
	}
	out.WriteString(strings.Join(strs, ", "))
	out.WriteString("}")
	return out.String()
}

func (s *SetLiteralInt) Ascii() (value string) {
	return s.String()
}

func (s *SetLiteralInt) Validate() error {
	return _validate.Struct(s)
}
