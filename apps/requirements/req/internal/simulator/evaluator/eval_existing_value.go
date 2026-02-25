package evaluator

// evalExistingValue evaluates the @ reference (existing value in EXCEPT context).
func evalExistingValue(bindings *Bindings) *EvalResult {
	value := bindings.GetExistingValue()
	if value == nil {
		return NewEvalError("@ used outside of EXCEPT context")
	}

	return NewEvalResult(value)
}
