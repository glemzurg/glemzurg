package logic_expression_type

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// ExpressionType is the interface implemented by all precise structural types.
// ExpressionTypes describe the shape and kind of values an expression can produce.
type ExpressionType interface {
	expressionType()
	TypeName() string
	Validate(ctx *coreerr.ValidationContext) error
}

// Type name constants.
const (
	TypeBoolean  = "boolean"
	TypeInteger  = "integer"
	TypeRational = "rational"
	TypeString   = "string"
	TypeEnum     = "enum"
	TypeSequence = "sequence"
	TypeTuple    = "tuple"
	TypeRecord   = "record"
	TypeFunction = "function"
	TypeObject   = "object"
)

// validateExpressionType validates an ExpressionType if it is non-nil.
func validateExpressionType(ctx *coreerr.ValidationContext, et ExpressionType) error {
	if et == nil {
		return nil
	}
	return et.Validate(ctx)
}
