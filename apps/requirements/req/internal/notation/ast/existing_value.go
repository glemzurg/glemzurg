package ast

// ExistingValue represents a reference to the existing value (output as @).
// This is used in contexts where the current value is being referenced or updated.
type ExistingValue struct{}

func (e *ExistingValue) expressionNode()        {}
func (e *ExistingValue) String() (value string) { return "@" }
func (e *ExistingValue) Ascii() (value string)  { return e.String() }
func (e *ExistingValue) Validate() error        { return nil }
