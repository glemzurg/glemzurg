package evaluator

import (
	"strings"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ============================================================
// Literals
// ============================================================

func evalIntLiteral(n *me.IntLiteral) *EvalResult {
	return NewEvalResult(object.NewInteger(n.Value.Int64()))
}

func evalRationalLiteral(n *me.RationalLiteral) *EvalResult {
	num := n.Value.Num().Int64()
	denom := n.Value.Denom().Int64()
	return NewEvalResult(object.NewRational(num, denom))
}

func evalBoolLiteral(n *me.BoolLiteral) *EvalResult {
	return NewEvalResult(nativeBoolToBoolean(n.Value))
}

func evalMEStringLiteral(n *me.StringLiteral) *EvalResult {
	return NewEvalResult(object.NewString(n.Value))
}

func evalMETupleLiteral(n *me.TupleLiteral, bindings *Bindings) *EvalResult {
	elements := make([]object.Object, 0, len(n.Elements))
	for _, elem := range n.Elements {
		result := Eval(elem, bindings)
		if result.IsError() {
			return result
		}
		elements = append(elements, result.Value)
	}
	return NewEvalResult(object.NewTupleFromElements(elements))
}

func evalMESetLiteral(n *me.SetLiteral, bindings *Bindings) *EvalResult {
	elements := make([]object.Object, 0, len(n.Elements))
	for _, elem := range n.Elements {
		result := Eval(elem, bindings)
		if result.IsError() {
			return result
		}
		elements = append(elements, result.Value)
	}
	return NewEvalResult(object.NewSetFromElements(elements))
}

func evalMESetConstant(n *me.SetConstant) *EvalResult {
	switch n.Kind {
	case me.SetConstantBoolean:
		elements := []object.Object{
			object.NewBoolean(true),
			object.NewBoolean(false),
		}
		return NewEvalResult(object.NewSetFromElements(elements))
	case me.SetConstantNat:
		return NewEvalError("cannot enumerate infinite set: Nat")
	case me.SetConstantInt:
		return NewEvalError("cannot enumerate infinite set: Int")
	case me.SetConstantReal:
		return NewEvalError("cannot enumerate infinite set: Real")
	default:
		return NewEvalError("unknown set constant: %s", n.Kind)
	}
}

func evalMESetRange(n *me.SetRange, bindings *Bindings) *EvalResult {
	startResult := Eval(n.Start, bindings)
	if startResult.IsError() {
		return startResult
	}
	startNum, ok := startResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("set range start must be a number, got %s", startResult.Value.Type())
	}
	if !startNum.Rat().IsInt() {
		return NewEvalError("set range start must be an integer, got %s", startNum.Inspect())
	}

	endResult := Eval(n.End, bindings)
	if endResult.IsError() {
		return endResult
	}
	endNum, ok := endResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("set range end must be a number, got %s", endResult.Value.Type())
	}
	if !endNum.Rat().IsInt() {
		return NewEvalError("set range end must be an integer, got %s", endNum.Inspect())
	}

	start := startNum.Rat().Num().Int64()
	end := endNum.Rat().Num().Int64()

	elements := make([]object.Object, 0)
	for i := start; i <= end; i++ {
		elements = append(elements, object.NewInteger(i))
	}
	return NewEvalResult(object.NewSetFromElements(elements))
}

func evalMERecordLiteral(n *me.RecordLiteral, bindings *Bindings) *EvalResult {
	fields := make(map[string]object.Object, len(n.Fields))
	for _, field := range n.Fields {
		result := Eval(field.Value, bindings)
		if result.IsError() {
			return result
		}
		fields[field.Name] = result.Value
	}
	return NewEvalResult(object.NewRecordFromFields(fields))
}

// ============================================================
// References
// ============================================================

func evalSelfRef(bindings *Bindings) *EvalResult {
	self := bindings.Self()
	if self == nil {
		return NewEvalError("self is not defined in this scope")
	}
	return NewEvalResult(self)
}

func evalAttributeRef(n *me.AttributeRef, bindings *Bindings) *EvalResult {
	// AttributeRef resolves to a field on the self record.
	// The attribute name is the last segment of the key.
	self := bindings.Self()
	if self == nil {
		return NewEvalError("attribute reference requires self scope")
	}
	attrName := n.AttributeKey.SubKey
	value := self.Get(attrName)
	if value == nil {
		return NewEvalError("attribute not found: %s", attrName)
	}
	return NewEvalResult(value)
}

func evalLocalVar(n *me.LocalVar, bindings *Bindings) *EvalResult {
	value, found := bindings.GetValue(n.Name)
	if !found {
		return NewEvalError("identifier not found: %s", n.Name)
	}
	return NewEvalResult(value)
}

func evalPriorFieldValue(bindings *Bindings) *EvalResult {
	value := bindings.GetExistingValue()
	if value == nil {
		return NewEvalError("@ used outside of EXCEPT context")
	}
	return NewEvalResult(value)
}

func evalNextState(n *me.NextState, bindings *Bindings) *EvalResult {
	// NextState wraps an expression whose identifiers should resolve to primed values.
	// For simple identifiers and field accesses, check primed values first.
	switch inner := n.Expr.(type) {
	case *me.LocalVar:
		if val, found := bindings.GetPrimedValue(inner.Name); found {
			return NewEvalResult(val)
		}
		val, found := bindings.GetValue(inner.Name)
		if !found {
			return NewEvalError("identifier not found: %s", inner.Name)
		}
		return NewEvalResult(val)

	case *me.AttributeRef:
		attrName := inner.AttributeKey.SubKey
		if val, found := bindings.GetPrimedValue(attrName); found {
			return NewEvalResult(val)
		}
		self := bindings.Self()
		if self == nil {
			return NewEvalError("attribute reference requires self scope")
		}
		value := self.Get(attrName)
		if value == nil {
			return NewEvalError("attribute not found: %s", attrName)
		}
		return NewEvalResult(value)

	case *me.FieldAccess:
		return evalNextStateFieldAccess(inner, bindings)

	case *me.SelfRef:
		// self' - look for primed self
		self := bindings.Self()
		if self == nil {
			return NewEvalError("self is not defined in this scope")
		}
		return NewEvalResult(self)

	default:
		return NewEvalError("primed expression requires an identifier or field access, got %T", n.Expr)
	}
}

// evalNextStateFieldAccess handles primed field access chains like record.field'.
func evalNextStateFieldAccess(fa *me.FieldAccess, bindings *Bindings) *EvalResult {
	// Collect the field access chain and find the root.
	var fields []string
	var current me.Expression = fa

	for {
		access, ok := current.(*me.FieldAccess)
		if !ok {
			break
		}
		fields = append([]string{access.Field}, fields...)
		current = access.Base
	}

	// Get the root value (check primed first).
	var rootValue object.Object

	switch root := current.(type) {
	case *me.LocalVar:
		if val, found := bindings.GetPrimedValue(root.Name); found {
			rootValue = val
		} else {
			val, found := bindings.GetValue(root.Name)
			if !found {
				return NewEvalError("identifier not found: %s", root.Name)
			}
			rootValue = val
		}
	case *me.AttributeRef:
		attrName := root.AttributeKey.SubKey
		if val, found := bindings.GetPrimedValue(attrName); found {
			rootValue = val
		} else {
			self := bindings.Self()
			if self == nil {
				return NewEvalError("attribute reference requires self scope")
			}
			value := self.Get(attrName)
			if value == nil {
				return NewEvalError("attribute not found: %s", attrName)
			}
			rootValue = value
		}
	case *me.SelfRef:
		self := bindings.Self()
		if self == nil {
			return NewEvalError("self is not defined in this scope")
		}
		rootValue = self
	default:
		// For other expression types, evaluate normally.
		result := Eval(current, bindings)
		if result.IsError() {
			return result
		}
		rootValue = result.Value
	}

	// Apply the field access chain.
	currentValue := rootValue
	for _, field := range fields {
		record, ok := currentValue.(*object.Record)
		if !ok {
			return NewEvalError("field access requires Record, got %s", currentValue.Type())
		}
		fieldValue := record.Get(field)
		if fieldValue == nil {
			return NewEvalError("field not found: %s", field)
		}
		currentValue = fieldValue
	}

	return NewEvalResult(currentValue)
}

func evalNamedSetRef(n *me.NamedSetRef, bindings *Bindings) *EvalResult {
	// Named sets are resolved at eval time by looking up their name in bindings.
	name := n.SetKey.SubKey
	value, found := bindings.GetValue(name)
	if !found {
		return NewEvalError("named set not found: %s", name)
	}
	return NewEvalResult(value)
}

// ============================================================
// Binary operators
// ============================================================

func evalMEBinaryArith(n *me.BinaryArith, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftNum, ok := leftResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("left operand must be Number, got %s", leftResult.Value.Type())
	}
	rightNum, ok := rightResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("right operand must be Number, got %s", rightResult.Value.Type())
	}

	var result *object.Number
	switch n.Op {
	case me.ArithAdd:
		result = leftNum.Add(rightNum)
	case me.ArithSub:
		result = leftNum.Sub(rightNum)
	case me.ArithMul:
		result = leftNum.Mul(rightNum)
	case me.ArithDiv:
		if rightNum.IsZero() {
			return NewEvalError("division by zero")
		}
		result = leftNum.Div(rightNum)
	case me.ArithMod:
		mod, err := leftNum.Mod(rightNum)
		if err != nil {
			return NewEvalError("modulo error: %v", err)
		}
		result = mod
	case me.ArithPow:
		pow, err := leftNum.Pow(rightNum)
		if err != nil {
			return NewEvalError("power error: %v", err)
		}
		result = pow
	default:
		return NewEvalError("unknown arithmetic operator: %s", n.Op)
	}
	return NewEvalResult(result)
}

func evalMEBinaryLogic(n *me.BinaryLogic, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	leftBool, ok := leftResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("left operand must be Boolean, got %s", leftResult.Value.Type())
	}

	// Short-circuit evaluation
	switch n.Op {
	case me.LogicAnd:
		if !leftBool.Value() {
			return NewEvalResult(FALSE)
		}
	case me.LogicOr:
		if leftBool.Value() {
			return NewEvalResult(TRUE)
		}
	case me.LogicImplies:
		if !leftBool.Value() {
			return NewEvalResult(TRUE)
		}
	case me.LogicEquiv:
		// No short-circuit possible for equivalence.
	default:
		return NewEvalError("unknown logic operator: %s", n.Op)
	}

	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}
	rightBool, ok := rightResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("right operand must be Boolean, got %s", rightResult.Value.Type())
	}

	var result bool
	switch n.Op {
	case me.LogicAnd:
		result = leftBool.Value() && rightBool.Value()
	case me.LogicOr:
		result = leftBool.Value() || rightBool.Value()
	case me.LogicImplies:
		result = !leftBool.Value() || rightBool.Value()
	case me.LogicEquiv:
		result = leftBool.Value() == rightBool.Value()
	default:
		return NewEvalError("unknown logic operator: %s", n.Op)
	}
	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalMECompare(n *me.Compare, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	// Handle equality/inequality for all types
	if n.Op == me.CompareEq || n.Op == me.CompareNeq {
		equals := objectsEqual(leftResult.Value, rightResult.Value)
		if n.Op == me.CompareNeq {
			equals = !equals
		}
		return NewEvalResult(nativeBoolToBoolean(equals))
	}

	// Numeric comparisons
	leftNum, ok := leftResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("left operand must be Number, got %s", leftResult.Value.Type())
	}
	rightNum, ok := rightResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("right operand must be Number, got %s", rightResult.Value.Type())
	}

	cmp := leftNum.Cmp(rightNum)
	var result bool
	switch n.Op {
	case me.CompareLt:
		result = cmp < 0
	case me.CompareGt:
		result = cmp > 0
	case me.CompareLte:
		result = cmp <= 0
	case me.CompareGte:
		result = cmp >= 0
	case me.CompareEq, me.CompareNeq:
		// Already handled above; should not reach here.
		return NewEvalError("unexpected equality operator in numeric comparison: %s", n.Op)
	default:
		return NewEvalError("unknown comparison operator: %s", n.Op)
	}
	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalMESetOp(n *me.SetOp, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftSet, ok := leftResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("left operand must be Set, got %s", leftResult.Value.Type())
	}
	rightSet, ok := rightResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("right operand must be Set, got %s", rightResult.Value.Type())
	}

	var result *object.Set
	switch n.Op {
	case me.SetUnion:
		result = leftSet.Union(rightSet)
	case me.SetIntersect:
		result = leftSet.Intersection(rightSet)
	case me.SetDifference:
		result = leftSet.Difference(rightSet)
	default:
		return NewEvalError("unknown set operator: %s", n.Op)
	}
	return NewEvalResult(result)
}

func evalMESetCompare(n *me.SetCompare, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftSet, ok := leftResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("left operand must be Set, got %s", leftResult.Value.Type())
	}
	rightSet, ok := rightResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("right operand must be Set, got %s", rightResult.Value.Type())
	}

	var result bool
	switch n.Op {
	case me.SetCompareSubsetEq:
		result = leftSet.IsSubsetOf(rightSet)
	case me.SetCompareSubset:
		result = leftSet.IsSubsetOf(rightSet) && !leftSet.Equals(rightSet)
	case me.SetCompareSupersetEq:
		result = rightSet.IsSubsetOf(leftSet)
	case me.SetCompareSuperset:
		result = rightSet.IsSubsetOf(leftSet) && !leftSet.Equals(rightSet)
	default:
		return NewEvalError("unknown set comparison operator: %s", n.Op)
	}
	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalMEBagOp(n *me.BagOp, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftBag, ok := leftResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("left operand must be Bag, got %s", leftResult.Value.Type())
	}
	rightBag, ok := rightResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("right operand must be Bag, got %s", rightResult.Value.Type())
	}

	var result *object.Bag
	switch n.Op {
	case me.BagSum:
		result = leftBag.Sum(rightBag)
	case me.BagDifference:
		result = leftBag.Difference(rightBag)
	default:
		return NewEvalError("unknown bag operator: %s", n.Op)
	}
	return NewEvalResult(result)
}

func evalMEBagCompare(n *me.BagCompare, bindings *Bindings) *EvalResult {
	leftResult := Eval(n.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}
	rightResult := Eval(n.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftBag, ok := leftResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("left operand must be Bag, got %s", leftResult.Value.Type())
	}
	rightBag, ok := rightResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("right operand must be Bag, got %s", rightResult.Value.Type())
	}

	var result bool
	switch n.Op {
	case me.BagCompareProperSubBag:
		result = leftBag.IsProperSubBagOf(rightBag)
	case me.BagCompareSubBag:
		result = leftBag.IsSubBagOf(rightBag)
	case me.BagCompareProperSupBag:
		result = leftBag.IsProperSuperBagOf(rightBag)
	case me.BagCompareSupBag:
		result = leftBag.IsSuperBagOf(rightBag)
	default:
		return NewEvalError("unknown bag comparison operator: %s", n.Op)
	}
	return NewEvalResult(nativeBoolToBoolean(result))
}

func evalMEMembership(n *me.Membership, bindings *Bindings) *EvalResult {
	elemResult := Eval(n.Element, bindings)
	if elemResult.IsError() {
		return elemResult
	}
	setResult := Eval(n.Set, bindings)
	if setResult.IsError() {
		return setResult
	}

	set, ok := setResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("membership test requires Set, got %s", setResult.Value.Type())
	}

	contains := set.Contains(elemResult.Value)
	if n.Negated {
		contains = !contains
	}
	return NewEvalResult(nativeBoolToBoolean(contains))
}

// ============================================================
// Unary operators
// ============================================================

func evalMENegate(n *me.Negate, bindings *Bindings) *EvalResult {
	result := Eval(n.Expr, bindings)
	if result.IsError() {
		return result
	}
	num, ok := result.Value.(*object.Number)
	if !ok {
		return NewEvalError("cannot negate non-numeric value: %T", result.Value)
	}
	return NewEvalResult(num.Neg())
}

func evalMENot(n *me.Not, bindings *Bindings) *EvalResult {
	result := Eval(n.Expr, bindings)
	if result.IsError() {
		return result
	}
	boolVal, ok := result.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("operand must be Boolean, got %s", result.Value.Type())
	}
	return NewEvalResult(nativeBoolToBoolean(!boolVal.Value()))
}

// ============================================================
// Collections
// ============================================================

func evalMEFieldAccess(n *me.FieldAccess, bindings *Bindings) *EvalResult {
	baseResult := Eval(n.Base, bindings)
	if baseResult.IsError() {
		return baseResult
	}

	record, ok := baseResult.Value.(*object.Record)
	if !ok {
		return NewEvalError("field access requires Record, got %s", baseResult.Value.Type())
	}

	// Check for relation traversal.
	classKey := bindings.SelfClassKey()
	relCtx := bindings.RelationContext()
	if classKey != "" && relCtx != nil {
		if relInfo := lookupRelation(classKey, n.Field, relCtx); relInfo != nil {
			return evalRelationTraversal(record, relInfo, relCtx)
		}
	}

	value := record.Get(n.Field)
	if value == nil {
		return NewEvalError("field not found: %s", n.Field)
	}
	return NewEvalResult(value)
}

func evalMETupleIndex(n *me.TupleIndex, bindings *Bindings) *EvalResult {
	tupleResult := Eval(n.Tuple, bindings)
	if tupleResult.IsError() {
		return tupleResult
	}
	indexResult := Eval(n.Index, bindings)
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

	index := int(indexNum.Rat().Num().Int64())
	value := tuple.At(index)
	if value == nil {
		return NewEvalError("index %d out of bounds", index)
	}
	return NewEvalResult(value)
}

func evalMERecordUpdate(n *me.RecordUpdate, bindings *Bindings) *EvalResult {
	baseResult := Eval(n.Base, bindings)
	if baseResult.IsError() {
		return baseResult
	}

	baseRecord, ok := baseResult.Value.(*object.Record)
	if !ok {
		return NewEvalError("EXCEPT requires Record, got %s", baseResult.Value.Type())
	}

	result := baseRecord.Clone().(*object.Record)

	for _, alt := range n.Alterations {
		currentValue := result.Get(alt.Field)

		childBindings := NewEnclosedBindings(bindings)
		if currentValue != nil {
			childBindings.SetExistingValue(currentValue)
		}

		newValueResult := Eval(alt.Value, childBindings)
		if newValueResult.IsError() {
			return newValueResult
		}

		result.Set(alt.Field, newValueResult.Value)
	}

	return NewEvalResult(result)
}

func evalMEStringIndex(n *me.StringIndex, bindings *Bindings) *EvalResult {
	strResult := Eval(n.Str, bindings)
	if strResult.IsError() {
		return strResult
	}
	indexResult := Eval(n.Index, bindings)
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

	index := int(indexNum.Rat().Num().Int64())
	strVal := str.Value()
	if index < 1 || index > len(strVal) {
		return NewEvalError("string index %d out of bounds (length %d)", index, len(strVal))
	}
	return NewEvalResult(object.NewString(string(strVal[index-1])))
}

func evalMEStringConcat(n *me.StringConcat, bindings *Bindings) *EvalResult {
	var builder strings.Builder
	for i, operand := range n.Operands {
		opResult := Eval(operand, bindings)
		if opResult.IsError() {
			return opResult
		}
		str, ok := opResult.Value.(*object.String)
		if !ok {
			return NewEvalError("operand %d must be String, got %s", i+1, opResult.Value.Type())
		}
		builder.WriteString(str.Value())
	}
	return NewEvalResult(object.NewString(builder.String()))
}

func evalMETupleConcat(n *me.TupleConcat, bindings *Bindings) *EvalResult {
	firstResult := Eval(n.Operands[0], bindings)
	if firstResult.IsError() {
		return firstResult
	}
	result, ok := firstResult.Value.(*object.Tuple)
	if !ok {
		return NewEvalError("operand 1 must be Tuple, got %s", firstResult.Value.Type())
	}

	for i := 1; i < len(n.Operands); i++ {
		opResult := Eval(n.Operands[i], bindings)
		if opResult.IsError() {
			return opResult
		}
		tuple, ok := opResult.Value.(*object.Tuple)
		if !ok {
			return NewEvalError("operand %d must be Tuple, got %s", i+1, opResult.Value.Type())
		}
		result = result.Concat(tuple)
	}
	return NewEvalResult(result)
}

// ============================================================
// Control flow
// ============================================================

func evalMEIfThenElse(n *me.IfThenElse, bindings *Bindings) *EvalResult {
	condResult := Eval(n.Condition, bindings)
	if condResult.IsError() {
		return condResult
	}
	condBool, ok := condResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("IF condition must be Boolean, got %s", condResult.Value.Type())
	}
	if condBool.Value() {
		return Eval(n.Then, bindings)
	}
	return Eval(n.Else, bindings)
}

func evalMECase(n *me.Case, bindings *Bindings) *EvalResult {
	for _, branch := range n.Branches {
		condResult := Eval(branch.Condition, bindings)
		if condResult.IsError() {
			return condResult
		}
		condBool, ok := condResult.Value.(*object.Boolean)
		if !ok {
			return NewEvalError("CASE branch condition must be Boolean, got %s", condResult.Value.Type())
		}
		if condBool.Value() {
			return Eval(branch.Result, bindings)
		}
	}
	if n.Otherwise != nil {
		return Eval(n.Otherwise, bindings)
	}
	return NewEvalError("CASE: no branch matched and no OTHER clause")
}

// ============================================================
// Quantifiers
// ============================================================

func evalMEQuantifier(n *me.Quantifier, bindings *Bindings) *EvalResult {
	setResult := Eval(n.Domain, bindings)
	if setResult.IsError() {
		return setResult
	}
	set, ok := setResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("quantifier requires Set, got %s", setResult.Value.Type())
	}

	elements := set.Elements()

	switch n.Kind {
	case me.QuantifierForall:
		for _, elem := range elements {
			childBindings := NewEnclosedBindings(bindings)
			childBindings.Set(n.Variable, elem, NamespaceLocal)
			predResult := Eval(n.Predicate, childBindings)
			if predResult.IsError() {
				return predResult
			}
			predBool, ok := predResult.Value.(*object.Boolean)
			if !ok {
				return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
			}
			if !predBool.Value() {
				return NewEvalResult(FALSE)
			}
		}
		return NewEvalResult(TRUE)

	case me.QuantifierExists:
		for _, elem := range elements {
			childBindings := NewEnclosedBindings(bindings)
			childBindings.Set(n.Variable, elem, NamespaceLocal)
			predResult := Eval(n.Predicate, childBindings)
			if predResult.IsError() {
				return predResult
			}
			predBool, ok := predResult.Value.(*object.Boolean)
			if !ok {
				return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
			}
			if predBool.Value() {
				return NewEvalResult(TRUE)
			}
		}
		return NewEvalResult(FALSE)

	default:
		return NewEvalError("unknown quantifier: %s", n.Kind)
	}
}

func evalMESetFilter(n *me.SetFilter, bindings *Bindings) *EvalResult {
	setResult := Eval(n.Set, bindings)
	if setResult.IsError() {
		return setResult
	}
	sourceSet, ok := setResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("set filter requires Set, got %s", setResult.Value.Type())
	}

	resultElements := make([]object.Object, 0)
	for _, elem := range sourceSet.Elements() {
		childBindings := NewEnclosedBindings(bindings)
		childBindings.Set(n.Variable, elem, NamespaceLocal)
		predResult := Eval(n.Predicate, childBindings)
		if predResult.IsError() {
			return predResult
		}
		predBool, ok := predResult.Value.(*object.Boolean)
		if !ok {
			return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
		}
		if predBool.Value() {
			resultElements = append(resultElements, elem)
		}
	}
	return NewEvalResult(object.NewSetFromElements(resultElements))
}

// ============================================================
// Calls
// ============================================================

func evalMEBuiltinCall(n *me.BuiltinCall, bindings *Bindings) *EvalResult {
	args := make([]object.Object, len(n.Args))
	for i, argExpr := range n.Args {
		result := Eval(argExpr, bindings)
		if result.IsError() {
			return result
		}
		args[i] = result.Value
	}

	// Build the full builtin name: Module!Function or just Function.
	var fullName string
	if n.Module != "" {
		fullName = n.Module + "!" + n.Function
	} else {
		fullName = n.Function
	}

	fn, ok := LookupBuiltin(fullName)
	if !ok {
		return NewEvalError("unknown builtin: %s", fullName)
	}
	return fn(args)
}

func evalMEGlobalCall(n *me.GlobalCall, bindings *Bindings) *EvalResult {
	funcName := n.FunctionKey.SubKey

	// Try builtins first.
	args := make([]object.Object, len(n.Args))
	for i, argExpr := range n.Args {
		result := Eval(argExpr, bindings)
		if result.IsError() {
			return result
		}
		args[i] = result.Value
	}

	fn, ok := LookupBuiltin(funcName)
	if ok {
		return fn(args)
	}

	// Fall back to registry for user-defined global functions.
	ctx := GetEvalContext()
	if ctx != nil && ctx.IRRegistry != nil {
		body, params, found := ctx.IRRegistry.LookupGlobal(funcName)
		if found {
			return evalRegistryCall(body, params, args, bindings)
		}
	}

	return NewEvalError("unknown global function: %s", funcName)
}

func evalMEActionCall(n *me.ActionCall, bindings *Bindings) *EvalResult {
	// Action calls represent cross-class action/query invocations.
	// These are resolved at the simulator level, not the expression evaluator level.
	return NewEvalError("action calls not yet supported in expression evaluator: %s", n.ActionKey.String())
}

// evalRegistryCall evaluates a registry function body with parameter bindings.
func evalRegistryCall(body me.Expression, params []string, args []object.Object, bindings *Bindings) *EvalResult {
	if len(params) != len(args) {
		return NewEvalError("function expects %d arguments, got %d", len(params), len(args))
	}

	childBindings := NewEnclosedBindings(bindings)
	for i, paramName := range params {
		childBindings.Set(paramName, args[i], NamespaceLocal)
	}

	return Eval(body, childBindings)
}
