package ast

import (
	"bytes"
	"strings"
)

// SetLiteralEnum is a set of string values (enumeration).
// Pattern: {"value1", "value2", "value3"}
type SetLiteralEnum struct {
	Values []string `validate:"required,min=1"` // The enumeration values
}

func (s *SetLiteralEnum) expressionNode() {}

func (s *SetLiteralEnum) String() (value string) {
	var out bytes.Buffer
	out.WriteString("{")
	quoted := make([]string, len(s.Values))
	for i, v := range s.Values {
		quoted[i] = `"` + v + `"`
	}
	out.WriteString(strings.Join(quoted, ", "))
	out.WriteString("}")
	return out.String()
}

func (s *SetLiteralEnum) Ascii() (value string) {
	return s.String()
}

func (s *SetLiteralEnum) Validate() error {
	return _validate.Struct(s)
}
