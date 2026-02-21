package ast

const (
	SetConstantBoolean = "BOOLEAN" // The set of {TRUE, FALSE}
	SetConstantNat     = "Nat"     // The set of natural numbers (0, 1, 2, ...)
	SetConstantInt     = "Int"     // The set of integers (..., -2, -1, 0, 1, 2, ...)
	SetConstantReal    = "Real"    // The set of real numbers
)

// SetConstant is a built-in set constant.
type SetConstant struct {
	Value string `validate:"required,oneof=BOOLEAN Nat Int Real"` // The built-in set name, e.g., BOOLEAN, Nat, Int, Real
}

func (s *SetConstant) expressionNode()        {}
func (s *SetConstant) String() (value string) { return s.Value }
func (s *SetConstant) Ascii() (value string)  { return s.String() }
func (s *SetConstant) Validate() error        { return _validate.Struct(s) }
