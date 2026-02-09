package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalTupleIndex evaluates tuple[index].
func evalTupleIndex(node *ast.ExpressionTupleIndex, bindings *Bindings) *EvalResult {
	tupleResult := Eval(node.Tuple, bindings)
	if tupleResult.IsError() {
		return tupleResult
	}

	indexResult := Eval(node.Index, bindings)
	if indexResult.IsError() {
		return indexResult
	}

	tuple, ok := tupleResult.Value.(*object.Tuple)
	if !ok {
		return NewEvalError("indexing requires Tuple, got %s", tupleResult.Value.Type())
	}

	indexNum, ok := indexResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("index must be Number, got %s", indexResult.Value.Type())
	}

	// TLA+ uses 1-based indexing
	index := int(indexNum.Rat().Num().Int64())
	value := tuple.At(index)
	if value == nil {
		return NewEvalError("index %d out of bounds", index)
	}

	return NewEvalResult(value)
}
