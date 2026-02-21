package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/typechecker"
)

// EvalTyped evaluates a type-checked AST node.
// The type checker has already verified that all operations are type-safe,
// so this function can assume correct types without runtime checks.
//
// For most operations, we delegate to the regular evaluator since the runtime
// representation is the same. The key difference is that type errors should
// never occur - if they do, it indicates a bug in the type checker.
func EvalTyped(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	// The TypedNode wraps an ast.Node with type information.
	// We can use the Children field for sub-expressions that have already
	// been type-checked.
	return evalTypedNode(typed, bindings)
}

func evalTypedNode(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	switch n := typed.Node.(type) {

	// === Literals ===
	// Literals don't need type info - just evaluate directly

	case *ast.NumberLiteral:
		return evalNumberLiteral(n)

	case *ast.NumericPrefixExpression:
		return evalNumericPrefixExpression(n, bindings)

	case *ast.FractionExpr:
		return evalFractionExpr(n, bindings)

	case *ast.ParenExpr:
		return evalParenExpr(n, bindings)

	case *ast.StringLiteral:
		return evalStringLiteral(n)

	case *ast.BooleanLiteral:
		return evalBooleanLiteral(n)

	case *ast.TupleLiteral:
		return evalTypedTupleLiteral(typed, bindings)

	case *ast.SetLiteralInt:
		return evalSetLiteralInt(n)

	case *ast.SetLiteralEnum:
		return evalSetLiteralEnum(n)

	case *ast.SetRange:
		return evalSetRange(n, bindings)

	case *ast.SetConstant:
		return evalSetConstant(n)

	case *ast.RecordInstance:
		return evalTypedRecordInstance(typed, bindings)

	// === Identifiers ===

	case *ast.Identifier:
		return evalIdentifier(n, bindings)

	case *ast.FieldIdentifier:
		return evalTypedFieldIdentifier(typed, bindings)

	case *ast.ExistingValue:
		return evalExistingValue(bindings)

	// === Arithmetic ===
	// Type checker guarantees both operands are numeric

	case *ast.RealInfixExpression:
		return evalTypedRealInfix(typed, bindings)

	// === Logic ===
	// Type checker guarantees operands are boolean where required

	case *ast.LogicInfixExpression:
		return evalTypedLogicInfix(typed, bindings)

	case *ast.LogicPrefixExpression:
		return evalTypedLogicPrefix(typed, bindings)

	case *ast.LogicRealComparison:
		return evalTypedLogicRealComparison(typed, bindings)

	case *ast.LogicMembership:
		return evalTypedLogicMembership(typed, bindings)

	case *ast.LogicBoundQuantifier:
		return evalTypedLogicBoundQuantifier(typed, bindings)

	case *ast.LogicInfixSet:
		return evalTypedLogicInfixSet(typed, bindings)

	case *ast.LogicInfixBag:
		return evalTypedLogicInfixBag(typed, bindings)

	// === Sets ===

	case *ast.SetInfix:
		return evalTypedSetInfix(typed, bindings)

	case *ast.SetConditional:
		return evalTypedSetConditional(typed, bindings)

	// === Bags ===

	case *ast.BagInfix:
		return evalTypedBagInfix(typed, bindings)

	// === Tuples/Sequences ===

	case *ast.ExpressionTupleIndex:
		return evalTypedTupleIndex(typed, bindings)

	case *ast.TupleInfixExpression:
		return evalTypedTupleInfix(typed, bindings)

	// === Builtins ===

	case *ast.BuiltinCall:
		return evalTypedBuiltinCall(typed, bindings)

	// === Records ===

	case *ast.RecordAltered:
		return evalTypedRecordAltered(typed, bindings)

	// === Control Flow ===

	case *ast.ExpressionIfElse:
		return evalTypedIfElse(typed, bindings)

	case *ast.ExpressionCase:
		return evalTypedCase(typed, bindings)

	// === Calls ===

	case *ast.CallExpression:
		return evalTypedCallExpression(typed, bindings)

	// === Strings ===

	case *ast.StringIndex:
		return evalTypedStringIndex(typed, bindings)

	case *ast.StringInfixExpression:
		return evalTypedStringInfix(typed, bindings)

	// === Assignment ===

	case *ast.Assignment:
		return evalTypedAssignment(typed, bindings)

	default:
		return NewEvalError("unknown typed node type: %T", typed.Node)
	}
}

// === Typed Evaluation Helpers ===
// These use the type-checked children instead of re-evaluating sub-expressions

func evalTypedRealInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.RealInfixExpression)

	// Use typed children - type checker guarantees exactly 2 children
	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	// Type checker guarantees these are numbers
	leftNum := leftResult.Value.(*object.Number)
	rightNum := rightResult.Value.(*object.Number)

	var result *object.Number

	switch n.Operator {
	case "+":
		result = leftNum.Add(rightNum)
	case "-":
		result = leftNum.Sub(rightNum)
	case "*":
		result = leftNum.Mul(rightNum)
	case "÷", "/":
		if rightNum.IsZero() {
			return NewEvalError("division by zero")
		}
		result = leftNum.Div(rightNum)
	case "%":
		mod, err := leftNum.Mod(rightNum)
		if err != nil {
			return NewEvalError("modulo error: %v", err)
		}
		result = mod
	case "^":
		return NewEvalError("power operator not yet implemented")
	default:
		return NewEvalError("unknown arithmetic operator: %s", n.Operator)
	}

	return NewEvalResult(result)
}

func evalTypedLogicInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicInfixExpression)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	// Short-circuit evaluation for AND/OR
	leftBool := leftResult.Value.(*object.Boolean)

	switch n.Operator {
	case "∧", "/\\":
		if !leftBool.Value() {
			return NewEvalResult(FALSE)
		}
	case "∨", "\\/":
		if leftBool.Value() {
			return NewEvalResult(TRUE)
		}
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	rightBool := rightResult.Value.(*object.Boolean)

	var result bool
	switch n.Operator {
	case "∧", "/\\":
		result = leftBool.Value() && rightBool.Value()
	case "∨", "\\/":
		result = leftBool.Value() || rightBool.Value()
	case "⇒", "=>":
		result = !leftBool.Value() || rightBool.Value()
	case "≡", "<=>":
		result = leftBool.Value() == rightBool.Value()
	default:
		return NewEvalError("unknown logic operator: %s", n.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalTypedLogicPrefix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicPrefixExpression)

	operandResult := evalTypedNode(typed.Children[0], bindings)
	if operandResult.IsError() {
		return operandResult
	}

	operandBool := operandResult.Value.(*object.Boolean)

	switch n.Operator {
	case "¬", "~":
		return NewEvalResult(nativeBoolToBoolean(!operandBool.Value()))
	default:
		return NewEvalError("unknown logic prefix operator: %s", n.Operator)
	}
}

func evalTypedLogicRealComparison(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicRealComparison)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftNum := leftResult.Value.(*object.Number)
	rightNum := rightResult.Value.(*object.Number)

	cmp := leftNum.Cmp(rightNum)

	var result bool
	switch n.Operator {
	case "<":
		result = cmp < 0
	case ">":
		result = cmp > 0
	case "≤", "<=":
		result = cmp <= 0
	case "≥", ">=":
		result = cmp >= 0
	default:
		return NewEvalError("unknown comparison operator: %s", n.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalTypedLogicMembership(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicMembership)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	// Type checker guarantees right is a Set
	set := rightResult.Value.(*object.Set)

	var result bool
	switch n.Operator {
	case "∈", "\\in":
		result = set.Contains(leftResult.Value)
	case "∉", "\\notin":
		result = !set.Contains(leftResult.Value)
	default:
		return NewEvalError("unknown membership operator: %s", n.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalTypedLogicBoundQuantifier(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicBoundQuantifier)

	// Children: [0] = set type info, [1] = predicate type info
	// The set is from the membership
	membership := n.Membership.(*ast.LogicMembership)
	ident := membership.Left.(*ast.Identifier)

	setResult := evalTypedNode(typed.Children[0], bindings)
	if setResult.IsError() {
		return setResult
	}

	set := setResult.Value.(*object.Set)

	switch n.Quantifier {
	case "∀", "\\A":
		// ∀ - must be true for all elements
		for _, elem := range set.Elements() {
			innerBindings := NewEnclosedBindings(bindings)
			innerBindings.Set(ident.Value, elem, NamespaceLocal)

			predResult := evalTypedNode(typed.Children[1], innerBindings)
			if predResult.IsError() {
				return predResult
			}

			predBool := predResult.Value.(*object.Boolean)
			if !predBool.Value() {
				return NewEvalResult(FALSE)
			}
		}
		return NewEvalResult(TRUE)

	case "∃", "\\E":
		// ∃ - must be true for at least one element
		for _, elem := range set.Elements() {
			innerBindings := NewEnclosedBindings(bindings)
			innerBindings.Set(ident.Value, elem, NamespaceLocal)

			predResult := evalTypedNode(typed.Children[1], innerBindings)
			if predResult.IsError() {
				return predResult
			}

			predBool := predResult.Value.(*object.Boolean)
			if predBool.Value() {
				return NewEvalResult(TRUE)
			}
		}
		return NewEvalResult(FALSE)

	default:
		return NewEvalError("unknown quantifier: %s", n.Quantifier)
	}
}

func evalTypedLogicInfixSet(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicInfixSet)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftSet := leftResult.Value.(*object.Set)
	rightSet := rightResult.Value.(*object.Set)

	var result bool
	switch n.Operator {
	case "⊆", "\\subseteq":
		result = leftSet.IsSubsetOf(rightSet)
	case "⊂", "\\subset":
		result = leftSet.IsSubsetOf(rightSet) && !leftSet.Equals(rightSet)
	case "⊇", "\\supseteq":
		result = rightSet.IsSubsetOf(leftSet)
	case "⊃", "\\supset":
		result = rightSet.IsSubsetOf(leftSet) && !leftSet.Equals(rightSet)
	default:
		return NewEvalError("unknown set comparison operator: %s", n.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalTypedLogicInfixBag(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.LogicInfixBag)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftBag := leftResult.Value.(*object.Bag)
	rightBag := rightResult.Value.(*object.Bag)

	var result bool
	switch n.Operator {
	case "⊑", "\\sqsubseteq":
		result = leftBag.IsSubBagOf(rightBag)
	case "⊒", "\\sqsupseteq":
		result = rightBag.IsSubBagOf(leftBag)
	default:
		return NewEvalError("unknown bag comparison operator: %s", n.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalTypedSetInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.SetInfix)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftSet := leftResult.Value.(*object.Set)
	rightSet := rightResult.Value.(*object.Set)

	switch n.Operator {
	case "∪", "\\union":
		return NewEvalResult(leftSet.Union(rightSet))
	case "∩", "\\intersect":
		return NewEvalResult(leftSet.Intersection(rightSet))
	case "\\", "\\diff":
		return NewEvalResult(leftSet.Difference(rightSet))
	default:
		return NewEvalError("unknown set operator: %s", n.Operator)
	}
}

func evalTypedSetConditional(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.SetConditional)

	// Children: [0] = source set, [1] = predicate
	membership := n.Membership.(*ast.LogicMembership)
	ident := membership.Left.(*ast.Identifier)

	sourceResult := evalTypedNode(typed.Children[0], bindings)
	if sourceResult.IsError() {
		return sourceResult
	}

	sourceSet := sourceResult.Value.(*object.Set)
	result := object.NewSet()

	for _, elem := range sourceSet.Elements() {
		innerBindings := NewEnclosedBindings(bindings)
		innerBindings.Set(ident.Value, elem, NamespaceLocal)

		predResult := evalTypedNode(typed.Children[1], innerBindings)
		if predResult.IsError() {
			return predResult
		}

		predBool := predResult.Value.(*object.Boolean)
		if predBool.Value() {
			result.Add(elem)
		}
	}

	return NewEvalResult(result)
}

func evalTypedBagInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.BagInfix)

	leftResult := evalTypedNode(typed.Children[0], bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := evalTypedNode(typed.Children[1], bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftBag := leftResult.Value.(*object.Bag)
	rightBag := rightResult.Value.(*object.Bag)

	switch n.Operator {
	case "⊎", "(+)":
		return NewEvalResult(leftBag.Sum(rightBag))
	case "⊖", "(-)":
		return NewEvalResult(leftBag.Difference(rightBag))
	default:
		return NewEvalError("unknown bag operator: %s", n.Operator)
	}
}

func evalTypedTupleIndex(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.ExpressionTupleIndex)

	tupleResult := evalTypedNode(typed.Children[0], bindings)
	if tupleResult.IsError() {
		return tupleResult
	}

	indexResult := evalTypedNode(typed.Children[1], bindings)
	if indexResult.IsError() {
		return indexResult
	}

	tuple := tupleResult.Value.(*object.Tuple)
	index := indexResult.Value.(*object.Number)

	idx := int(index.Rat().Num().Int64())

	// TLA+ tuples are 1-indexed
	if idx < 1 || idx > tuple.Len() {
		return NewEvalError("tuple index %d out of bounds (length %d) for %s", idx, tuple.Len(), n.String())
	}

	return NewEvalResult(tuple.At(idx))
}

func evalTypedTupleInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.TupleInfixExpression)

	// All children are tuple operands
	var tuples []*object.Tuple
	for _, child := range typed.Children {
		result := evalTypedNode(child, bindings)
		if result.IsError() {
			return result
		}
		tuples = append(tuples, result.Value.(*object.Tuple))
	}

	switch n.Operator {
	case "∘", "\\o":
		result := object.NewTuple()
		for _, t := range tuples {
			for i := 1; i <= t.Len(); i++ {
				result.Append(t.At(i))
			}
		}
		return NewEvalResult(result)
	default:
		return NewEvalError("unknown tuple operator: %s", n.Operator)
	}
}

func evalTypedTupleLiteral(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	result := object.NewTuple()

	for _, child := range typed.Children {
		elemResult := evalTypedNode(child, bindings)
		if elemResult.IsError() {
			return elemResult
		}
		result = result.Append(elemResult.Value)
	}

	return NewEvalResult(result)
}

func evalTypedRecordInstance(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.RecordInstance)
	result := object.NewRecord()

	for i, binding := range n.Bindings {
		valueResult := evalTypedNode(typed.Children[i], bindings)
		if valueResult.IsError() {
			return valueResult
		}
		result.Set(binding.Field.Value, valueResult.Value)
	}

	return NewEvalResult(result)
}

func evalTypedRecordAltered(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.RecordAltered)

	// First child is the base record
	baseResult := evalTypedNode(typed.Children[0], bindings)
	if baseResult.IsError() {
		return baseResult
	}

	baseRecord := baseResult.Value.(*object.Record)
	result := baseRecord.Clone().(*object.Record)

	// Rest of children are alteration values
	for i, alter := range n.Alterations {
		valueResult := evalTypedNode(typed.Children[i+1], bindings)
		if valueResult.IsError() {
			return valueResult
		}
		result.Set(alter.Field.Member, valueResult.Value)
	}

	return NewEvalResult(result)
}

func evalTypedFieldIdentifier(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.FieldIdentifier)

	if n.Identifier == nil {
		// Field access on context record
		contextRecord := bindings.GetExistingValue()
		if contextRecord == nil {
			return NewEvalError("no context record for field access")
		}
		rec := contextRecord.(*object.Record)
		val := rec.Get(n.Member)
		if val == nil {
			return NewEvalError("record does not have field: %s", n.Member)
		}
		return NewEvalResult(val)
	}

	// First child is the record
	recordResult := evalTypedNode(typed.Children[0], bindings)
	if recordResult.IsError() {
		return recordResult
	}

	rec := recordResult.Value.(*object.Record)
	val := rec.Get(n.Member)
	if val == nil {
		return NewEvalError("record does not have field: %s", n.Member)
	}

	return NewEvalResult(val)
}

func evalTypedIfElse(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	// Children: [0] = condition, [1] = then, [2] = else
	condResult := evalTypedNode(typed.Children[0], bindings)
	if condResult.IsError() {
		return condResult
	}

	condBool := condResult.Value.(*object.Boolean)

	if condBool.Value() {
		return evalTypedNode(typed.Children[1], bindings)
	}
	return evalTypedNode(typed.Children[2], bindings)
}

func evalTypedCase(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.ExpressionCase)

	// Children are pairs of (condition, result), then optionally OTHER result
	childIdx := 0
	for range n.Branches {
		condResult := evalTypedNode(typed.Children[childIdx], bindings)
		if condResult.IsError() {
			return condResult
		}

		condBool := condResult.Value.(*object.Boolean)
		if condBool.Value() {
			return evalTypedNode(typed.Children[childIdx+1], bindings)
		}
		childIdx += 2
	}

	// Check OTHER
	if n.Other != nil {
		return evalTypedNode(typed.Children[childIdx], bindings)
	}

	return NewEvalError("no case branch matched and no OTHER clause")
}

func evalTypedBuiltinCall(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.BuiltinCall)

	// Evaluate all argument children
	args := make([]object.Object, len(typed.Children))
	for i, child := range typed.Children {
		result := evalTypedNode(child, bindings)
		if result.IsError() {
			return result
		}
		args[i] = result.Value
	}

	// Look up and call the builtin
	fn, ok := LookupBuiltin(n.Name)
	if !ok {
		return NewEvalError("unknown builtin: %s", n.Name)
	}

	return fn(args)
}

func evalTypedCallExpression(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	// Check if we have a registry context
	ctx := GetEvalContext()
	if ctx != nil && ctx.Registry != nil {
		return evalTypedCallExpressionWithRegistry(typed, bindings, ctx)
	}

	// Legacy behavior - return error
	return evalTypedCallExpressionLegacy(typed, bindings)
}

func evalTypedCallExpressionWithRegistry(typed *typechecker.TypedNode, bindings *Bindings, ctx *EvalContext) *EvalResult {
	n := typed.Node.(*ast.CallExpression)

	// Call the registry to resolve and evaluate
	result, err := ctx.Registry.ResolveAndEval(
		n,
		typed.Children, // These are the typed argument nodes
		bindings,
		ctx.ScopeLevel,
		ctx.Domain,
		ctx.Subdomain,
		ctx.Class,
	)
	if err != nil {
		return NewEvalError("call error: %s", err.Error())
	}

	return NewEvalResult(result)
}

func evalTypedCallExpressionLegacy(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.CallExpression)

	// First child is the parameter (record)
	if len(typed.Children) == 0 {
		return NewEvalError("call expression has no arguments")
	}

	paramResult := evalTypedNode(typed.Children[0], bindings)
	if paramResult.IsError() {
		return paramResult
	}

	// Build the function name
	var fnName string
	if n.ModelScope {
		fnName = "_" + n.FunctionName.Value
	} else {
		if n.Domain != nil {
			fnName = n.Domain.Value + "!"
		}
		if n.Subdomain != nil {
			fnName += n.Subdomain.Value + "!"
		}
		if n.Class != nil {
			fnName += n.Class.Value + "!"
		}
		fnName += n.FunctionName.Value
	}

	// For now, return an error since function definitions are not implemented
	return NewEvalError("function calls not yet implemented: %s", fnName)
}

func evalTypedStringIndex(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.StringIndex)

	strResult := evalTypedNode(typed.Children[0], bindings)
	if strResult.IsError() {
		return strResult
	}

	indexResult := evalTypedNode(typed.Children[1], bindings)
	if indexResult.IsError() {
		return indexResult
	}

	str := strResult.Value.(*object.String)
	index := indexResult.Value.(*object.Number)

	idx := int(index.Rat().Num().Int64())

	// TLA+ strings are 1-indexed
	runes := []rune(str.Value())
	if idx < 1 || idx > len(runes) {
		return NewEvalError("string index %d out of bounds (length %d) for %s", idx, len(runes), n.String())
	}

	return NewEvalResult(object.NewString(string(runes[idx-1])))
}

func evalTypedStringInfix(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.StringInfixExpression)

	var strs []string
	for _, child := range typed.Children {
		result := evalTypedNode(child, bindings)
		if result.IsError() {
			return result
		}
		strs = append(strs, result.Value.(*object.String).Value())
	}

	switch n.Operator {
	case "∘", "\\o":
		var result string
		for _, s := range strs {
			result += s
		}
		return NewEvalResult(object.NewString(result))
	default:
		return NewEvalError("unknown string operator: %s", n.Operator)
	}
}

func evalTypedAssignment(typed *typechecker.TypedNode, bindings *Bindings) *EvalResult {
	n := typed.Node.(*ast.Assignment)

	// The value to assign is in the first child (if we have typed children)
	var valueResult *EvalResult
	if len(typed.Children) > 0 {
		valueResult = evalTypedNode(typed.Children[0], bindings)
	} else {
		// Fallback for untyped value
		valueResult = Eval(n.Value, bindings)
	}

	if valueResult.IsError() {
		return valueResult
	}

	// Create primed binding
	primed := make(map[string]object.Object)
	primed[n.Target.Value] = valueResult.Value

	return NewEvalResultWithPrimed(EMPTY_SET, primed)
}
