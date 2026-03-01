package ast

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression_type"
)

// ConvertToExpressionType converts a TLA+ AST expression into a structural ExpressionType.
// This is the "type interpretation" pass — it treats a TLA+ expression as a type declaration
// rather than a value expression. Only a subset of AST nodes are valid type expressions.
func ConvertToExpressionType(expr Expression) (model_expression_type.ExpressionType, error) {
	if expr == nil {
		return nil, fmt.Errorf("type expression is nil")
	}

	switch e := expr.(type) {
	case *Identifier:
		return convertIdentifier(e)

	case *SetConstant:
		return convertSetConstant(e)

	case *SetLiteralEnum:
		return convertSetLiteralEnum(e)

	case *SetLiteral:
		return convertSetLiteral(e)

	case *FunctionCall:
		return convertFunctionCall(e)

	case *RecordTypeExpr:
		return convertRecordTypeExpr(e)

	case *CartesianProduct:
		return convertCartesianProduct(e)

	default:
		return nil, fmt.Errorf("not a valid type expression: %s", expr.String())
	}
}

// convertIdentifier handles identifiers that name built-in types.
// In the PEG grammar, Nat, Int, Real, BOOLEAN, STRING all parse as Identifiers.
func convertIdentifier(id *Identifier) (model_expression_type.ExpressionType, error) {
	switch id.Value {
	case "BOOLEAN":
		return &model_expression_type.BooleanType{}, nil
	case "Nat", "Int":
		return &model_expression_type.IntegerType{}, nil
	case "Real":
		return &model_expression_type.RationalType{}, nil
	case "STRING":
		return &model_expression_type.StringType{}, nil
	default:
		return nil, fmt.Errorf("unknown type identifier: %s", id.Value)
	}
}

// convertSetConstant handles the SetConstant AST node (used when the evaluator
// constructs AST nodes directly, not from the PEG parser).
func convertSetConstant(sc *SetConstant) (model_expression_type.ExpressionType, error) {
	switch sc.Value {
	case SetConstantBoolean:
		return &model_expression_type.BooleanType{}, nil
	case SetConstantNat, SetConstantInt:
		return &model_expression_type.IntegerType{}, nil
	case SetConstantReal:
		return &model_expression_type.RationalType{}, nil
	default:
		return nil, fmt.Errorf("unknown set constant for type: %s", sc.Value)
	}
}

// convertSetLiteralEnum handles {"val1", "val2", ...} → EnumType.
func convertSetLiteralEnum(sle *SetLiteralEnum) (model_expression_type.ExpressionType, error) {
	if len(sle.Values) == 0 {
		return nil, fmt.Errorf("enum type must have at least one value")
	}
	values := make([]string, len(sle.Values))
	copy(values, sle.Values)
	return &model_expression_type.EnumType{Values: values}, nil
}

// convertSetLiteral handles {expr1, expr2, ...} where all elements are string literals → EnumType.
func convertSetLiteral(sl *SetLiteral) (model_expression_type.ExpressionType, error) {
	if len(sl.Elements) == 0 {
		return nil, fmt.Errorf("set literal as type must have at least one element")
	}
	values := make([]string, 0, len(sl.Elements))
	for i, elem := range sl.Elements {
		strLit, ok := elem.(*StringLiteral)
		if !ok {
			return nil, fmt.Errorf("set literal element %d is not a string literal: %s", i, elem.String())
		}
		values = append(values, strLit.Value)
	}
	return &model_expression_type.EnumType{Values: values}, nil
}

// convertFunctionCall handles built-in module type constructors:
//   - _Seq!Seq(X)       → SequenceType{ElementType: X, Unique: false}
//   - _Seq!SeqUnique(X) → SequenceType{ElementType: X, Unique: true}
//   - _Set!_Set(X)      → SetType{ElementType: X}
//   - _Bags!_Bag(X)     → BagType{ElementType: X}
func convertFunctionCall(fc *FunctionCall) (model_expression_type.ExpressionType, error) {
	if len(fc.ScopePath) != 1 {
		return nil, fmt.Errorf("not a valid type expression: %s", fc.String())
	}

	module := fc.ScopePath[0].Value
	name := fc.Name.Value

	if len(fc.Args) != 1 {
		return nil, fmt.Errorf("type constructor %s!%s requires exactly 1 argument, got %d", module, name, len(fc.Args))
	}

	elemType, err := ConvertToExpressionType(fc.Args[0])
	if err != nil {
		return nil, fmt.Errorf("type constructor %s!%s argument: %w", module, name, err)
	}

	switch module {
	case "_Seq":
		switch name {
		case "Seq":
			return &model_expression_type.SequenceType{ElementType: elemType, Unique: false}, nil
		case "SeqUnique":
			return &model_expression_type.SequenceType{ElementType: elemType, Unique: true}, nil
		default:
			return nil, fmt.Errorf("unknown _Seq function for type: %s", name)
		}
	case "_Set":
		if name == "_Set" {
			return &model_expression_type.SetType{ElementType: elemType}, nil
		}
		return nil, fmt.Errorf("unknown _Set function for type: %s", name)
	case "_Bags":
		if name == "_Bag" {
			return &model_expression_type.BagType{ElementType: elemType}, nil
		}
		return nil, fmt.Errorf("unknown _Bags function for type: %s", name)
	default:
		return nil, fmt.Errorf("unknown module for type expression: %s", module)
	}
}

// convertRecordTypeExpr handles [name: STRING, age: Int] → RecordType.
func convertRecordTypeExpr(rt *RecordTypeExpr) (model_expression_type.ExpressionType, error) {
	fields := make([]model_expression_type.RecordFieldType, len(rt.Fields))
	for i, field := range rt.Fields {
		fieldType, err := ConvertToExpressionType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("record field %s: %w", field.Name.Value, err)
		}
		fields[i] = model_expression_type.RecordFieldType{
			Name: field.Name.Value,
			Type: fieldType,
		}
	}
	return &model_expression_type.RecordType{Fields: fields}, nil
}

// convertCartesianProduct handles S1 \X S2 → TupleType.
func convertCartesianProduct(cp *CartesianProduct) (model_expression_type.ExpressionType, error) {
	elemTypes := make([]model_expression_type.ExpressionType, len(cp.Operands))
	for i, operand := range cp.Operands {
		et, err := ConvertToExpressionType(operand)
		if err != nil {
			return nil, fmt.Errorf("cartesian product operand %d: %w", i, err)
		}
		elemTypes[i] = et
	}
	return &model_expression_type.TupleType{ElementTypes: elemTypes}, nil
}
