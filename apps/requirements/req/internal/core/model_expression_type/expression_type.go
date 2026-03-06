package model_expression_type

import "github.com/go-playground/validator/v10"

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// ExpressionType is the interface implemented by all precise structural types.
// ExpressionTypes describe the shape and kind of values an expression can produce.
type ExpressionType interface {
	expressionType()
	TypeName() string
	Validate() error
}

// Type name constants.
const (
	TypeBoolean     = "boolean"
	TypeInteger     = "integer"
	TypeRational    = "rational"
	TypeString      = "string"
	TypeEnum        = "enum"
	TypeSet         = "set"
	TypeSequence    = "sequence"
	TypeBag         = "bag"
	TypeTuple       = "tuple"
	TypeRecord      = "record"
	TypeFunction    = "function"
	TypeObject = "object"
)

// ValidateExpressionType validates an ExpressionType if it is non-nil.
func ValidateExpressionType(et ExpressionType) error {
	if et == nil {
		return nil
	}
	return et.Validate()
}
