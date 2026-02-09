package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalStringIndex evaluates string[index].
func evalStringIndex(node *ast.StringIndex, bindings *Bindings) *EvalResult {
	strResult := Eval(node.Str, bindings)
	if strResult.IsError() {
		return strResult
	}

	indexResult := Eval(node.Index, bindings)
	if indexResult.IsError() {
		return indexResult
	}

	str, ok := strResult.Value.(*object.String)
	if !ok {
		return NewEvalError("string indexing requires String, got %s", strResult.Value.Type())
	}

	indexNum, ok := indexResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("string index must be Number, got %s", indexResult.Value.Type())
	}

	// TLA+ uses 1-based indexing
	index := int(indexNum.Rat().Num().Int64())
	strVal := str.Value()

	if index < 1 || index > len(strVal) {
		return NewEvalError("string index %d out of bounds (length %d)", index, len(strVal))
	}

	// Return single character as string
	return NewEvalResult(object.NewString(string(strVal[index-1])))
}
