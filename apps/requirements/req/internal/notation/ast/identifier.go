package ast

// Identifier is a variable name.
type Identifier struct {
	Value string `validate:"required"` // The identifier name
}

func (i *Identifier) expressionNode()        {}
func (i *Identifier) String() (value string) { return i.Value }
func (i *Identifier) Ascii() (value string)  { return i.String() }
func (i *Identifier) Validate() error        { return _validate.Struct(i) }
