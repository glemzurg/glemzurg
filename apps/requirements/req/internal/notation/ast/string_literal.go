package ast

// StringLiteral is a string value.
type StringLiteral struct {
	Value string // The string value (without quotes)
}

func (s *StringLiteral) expressionNode()        {}
func (s *StringLiteral) String() (value string) { return `"` + s.Value + `"` }
func (s *StringLiteral) Ascii() (value string)  { return s.String() }
func (s *StringLiteral) Validate() error        { return _validate.Struct(s) }
