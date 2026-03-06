package ast

import (
	"strconv"
	"strings"
)

// BooleanLiteral is a boolean value.
type BooleanLiteral struct {
	Value bool
}

func (b *BooleanLiteral) expressionNode()        {}
func (b *BooleanLiteral) String() (value string) { return strings.ToUpper(strconv.FormatBool(b.Value)) }
func (b *BooleanLiteral) Ascii() (value string)  { return b.String() }
func (b *BooleanLiteral) Validate() error        { return _validate.Struct(b) }
