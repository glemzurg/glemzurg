package ast

import "github.com/go-playground/validator/v10"

// _validate is the shared validator instance for all nodes.
var _validate = validator.New()

// Node is a node of the abstract syntax tree.
type Node interface {
	String() (value string)
	Ascii() (value string)
	Validate() error
}

// Statement is a node that changes state.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that produces a value.
// All value-producing nodes implement this interface.
// Semantic types (Boolean, Number, Set, etc.) are determined by the type checker,
// not by Go interface types.
type Expression interface {
	Node
	expressionNode()
}
