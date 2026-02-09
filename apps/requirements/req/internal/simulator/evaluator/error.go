package evaluator

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// newError creates a new Error object.
func newError(format string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

// isError checks if an object is an Error.
func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.TypeError
}

// isResultError checks if an EvalResult contains an error.
func isResultError(result *EvalResult) bool {
	return result != nil && result.IsError()
}

// wrapError wraps an existing error into an EvalResult.
func wrapError(err *object.Error) *EvalResult {
	return &EvalResult{Error: err}
}
