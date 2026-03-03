package convert

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	met "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression_type"
)

// RaiseType converts a model ExpressionType into a TLA+ string representation.
// It constructs AST nodes internally and prints them via ast.Print.
func RaiseType(et met.ExpressionType, ctx *RaiseContext) (string, error) {
	expr, err := raiseTypeToAST(et, ctx)
	if err != nil {
		return "", err
	}
	return ast.Print(expr), nil
}

// raiseTypeToAST converts an ExpressionType to an AST expression suitable
// for printing as a TLA+ type expression.
func raiseTypeToAST(et met.ExpressionType, ctx *RaiseContext) (ast.Expression, error) {
	if et == nil {
		return nil, fmt.Errorf("cannot raise nil ExpressionType")
	}

	switch t := et.(type) {
	case *met.BooleanType:
		return &ast.Identifier{Value: "BOOLEAN"}, nil

	case *met.IntegerType:
		return &ast.Identifier{Value: "Int"}, nil

	case *met.RationalType:
		return &ast.Identifier{Value: "Real"}, nil

	case *met.StringType:
		return &ast.Identifier{Value: "STRING"}, nil

	case *met.EnumType:
		return raiseEnumType(t)

	case *met.SetType:
		return raiseSetType(t, ctx)

	case *met.SequenceType:
		return raiseSequenceType(t, ctx)

	case *met.BagType:
		return raiseBagType(t, ctx)

	case *met.TupleType:
		return raiseTupleType(t, ctx)

	case *met.RecordType:
		return raiseRecordType(t, ctx)

	case *met.FunctionType:
		return nil, fmt.Errorf("FunctionType cannot be represented as a TLA+ type expression")

	case *met.ObjectType:
		return raiseObjectType(t, ctx)

	default:
		return nil, fmt.Errorf("unsupported ExpressionType: %T", et)
	}
}

func raiseEnumType(t *met.EnumType) (ast.Expression, error) {
	return &ast.SetLiteralEnum{Values: t.Values}, nil
}

func raiseSetType(t *met.SetType, ctx *RaiseContext) (ast.Expression, error) {
	elemAST, err := raiseTypeToAST(t.ElementType, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetType.ElementType: %w", err)
	}
	return &ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: "_Set"}},
		Name:      &ast.Identifier{Value: "_Set"},
		Args:      []ast.Expression{elemAST},
	}, nil
}

func raiseSequenceType(t *met.SequenceType, ctx *RaiseContext) (ast.Expression, error) {
	elemAST, err := raiseTypeToAST(t.ElementType, ctx)
	if err != nil {
		return nil, fmt.Errorf("SequenceType.ElementType: %w", err)
	}
	funcName := "Seq"
	if t.Unique {
		funcName = "SeqUnique"
	}
	return &ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: "_Seq"}},
		Name:      &ast.Identifier{Value: funcName},
		Args:      []ast.Expression{elemAST},
	}, nil
}

func raiseBagType(t *met.BagType, ctx *RaiseContext) (ast.Expression, error) {
	elemAST, err := raiseTypeToAST(t.ElementType, ctx)
	if err != nil {
		return nil, fmt.Errorf("BagType.ElementType: %w", err)
	}
	return &ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: "_Bags"}},
		Name:      &ast.Identifier{Value: "_Bag"},
		Args:      []ast.Expression{elemAST},
	}, nil
}

func raiseTupleType(t *met.TupleType, ctx *RaiseContext) (ast.Expression, error) {
	operands := make([]ast.Expression, len(t.ElementTypes))
	for i, elemType := range t.ElementTypes {
		elemAST, err := raiseTypeToAST(elemType, ctx)
		if err != nil {
			return nil, fmt.Errorf("TupleType.ElementTypes[%d]: %w", i, err)
		}
		operands[i] = elemAST
	}
	return &ast.CartesianProduct{Operands: operands}, nil
}

func raiseRecordType(t *met.RecordType, ctx *RaiseContext) (ast.Expression, error) {
	fields := make([]*ast.RecordTypeField, len(t.Fields))
	for i, f := range t.Fields {
		typeAST, err := raiseTypeToAST(f.Type, ctx)
		if err != nil {
			return nil, fmt.Errorf("RecordType.Fields[%d]: %w", i, err)
		}
		fields[i] = &ast.RecordTypeField{
			Name: &ast.Identifier{Value: f.Name},
			Type: typeAST,
		}
	}
	return &ast.RecordTypeExpr{Fields: fields}, nil
}

func raiseObjectType(t *met.ObjectType, ctx *RaiseContext) (ast.Expression, error) {
	// ObjectType references a class by key. Use the key's SubKey (class name).
	return &ast.Identifier{Value: t.ClassKey.SubKey}, nil
}
