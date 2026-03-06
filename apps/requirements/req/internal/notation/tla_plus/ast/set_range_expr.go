package ast

import "bytes"

// SetRangeExpr is a dynamic set range expression: start .. end.
// Unlike SetRange (which has static int bounds), this allows any expression
// as start and end values, enabling ranges like x..y or (n-1)..(n+1).
type SetRangeExpr struct {
	Start Expression `validate:"required"` // The starting value (inclusive)
	End   Expression `validate:"required"` // The ending value (inclusive)
}

func (s *SetRangeExpr) expressionNode() {}

func (s *SetRangeExpr) String() (value string) {
	var out bytes.Buffer
	out.WriteString(s.Start.String())
	out.WriteString(" .. ")
	out.WriteString(s.End.String())
	return out.String()
}

func (s *SetRangeExpr) Ascii() (value string) {
	return s.String()
}

func (s *SetRangeExpr) Validate() error {
	if err := _validate.Struct(s); err != nil {
		return err
	}
	if err := s.Start.Validate(); err != nil {
		return err
	}
	if err := s.End.Validate(); err != nil {
		return err
	}
	return nil
}
