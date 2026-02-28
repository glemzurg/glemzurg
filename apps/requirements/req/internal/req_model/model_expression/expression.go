package model_expression

import "github.com/go-playground/validator/v10"

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// Expression is the interface implemented by all model expression nodes.
// Model expressions are notation-independent representations of formal logic.
type Expression interface {
	expressionNode()
	NodeType() string
	Validate() error
}

// Node type constants. These match the SQL expression_node_type enum values.
const (
	// Literals.
	NodeBoolLiteral     = "bool_literal"
	NodeIntLiteral      = "int_literal"
	NodeRationalLiteral = "rational_literal"
	NodeStringLiteral   = "string_literal"
	NodeSetLiteral      = "set_literal"
	NodeTupleLiteral    = "tuple_literal"
	NodeRecordLiteral   = "record_literal"
	NodeSetConstant     = "set_constant"

	// References.
	NodeSelfRef         = "self_ref"
	NodeAttributeRef    = "attribute_ref"
	NodeLocalVar        = "local_var"
	NodePriorFieldValue = "prior_field_value"
	NodeNextState       = "next_state"

	// Binary operators.
	NodeBinaryArith  = "binary_arith"
	NodeBinaryLogic  = "binary_logic"
	NodeCompare      = "compare"
	NodeSetOp        = "set_op"
	NodeSetCompare   = "set_compare"
	NodeBagOp        = "bag_op"
	NodeBagCompare   = "bag_compare"
	NodeMembership   = "membership"

	// Unary operators.
	NodeNegate = "negate"
	NodeNot    = "not"

	// Collections.
	NodeFieldAccess   = "field_access"
	NodeTupleIndex    = "tuple_index"
	NodeRecordUpdate  = "record_update"
	NodeStringIndex   = "string_index"
	NodeStringConcat  = "string_concat"
	NodeTupleConcat   = "tuple_concat"

	// Control flow.
	NodeIfThenElse = "if_then_else"
	NodeCase       = "case"

	// Quantifiers.
	NodeQuantifier = "quantifier"
	NodeSetFilter  = "set_filter"
	NodeSetRange   = "set_range"

	// Calls.
	NodeActionCall   = "action_call"
	NodeGlobalCall   = "global_call"
	NodeBuiltinCall  = "builtin_call"
	NodeNamedSetRef  = "named_set_ref"
)

// ValidateExpression validates an Expression if it is non-nil.
func ValidateExpression(expr Expression) error {
	if expr == nil {
		return nil
	}
	return expr.Validate()
}
