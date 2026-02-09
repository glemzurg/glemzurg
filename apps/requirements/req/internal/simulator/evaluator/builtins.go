package evaluator

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"

// BuiltinFn is the signature for all builtin functions.
// Args are pre-evaluated object.Object values.
type BuiltinFn func(args []object.Object) *EvalResult

// builtins maps function names to their implementations.
// Names follow _Module!Function syntax to avoid collision with user-defined names.
var builtins = map[string]BuiltinFn{
	// Sequences (_Seq module)
	"_Seq!Head":   builtinSeqHead,
	"_Seq!Tail":   builtinSeqTail,
	"_Seq!Append": builtinSeqAppend,
	"_Seq!Len":    builtinSeqLen,

	// Stack (LIFO) - custom module for data type support
	"_Stack!Push": builtinStackPush,
	"_Stack!Pop":  builtinStackPop,

	// Queue (FIFO) - custom module for data type support
	"_Queue!Enqueue": builtinQueueEnqueue,
	"_Queue!Dequeue": builtinQueueDequeue,

	// Bags
	"_Bags!SetToBag": builtinSetToBag,
	"_Bags!BagToSet": builtinBagToSet,
	"_Bags!CopiesIn": builtinCopiesIn,
	"_Bags!BagIn":    builtinBagIn,
}

// LookupBuiltin returns the builtin function for the given name.
func LookupBuiltin(name string) (BuiltinFn, bool) {
	fn, ok := builtins[name]
	return fn, ok
}

// === Sequence Builtins (_Seq module) ===

func builtinSeqHead(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Seq!Head requires 1 argument, got %d", len(args))
	}
	tuple, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Seq!Head requires Tuple, got %s", args[0].Type())
	}
	head := tuple.Head()
	if head == nil {
		return NewEvalError("_Seq!Head on empty tuple")
	}
	return NewEvalResult(head)
}

func builtinSeqTail(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Seq!Tail requires 1 argument, got %d", len(args))
	}
	tuple, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Seq!Tail requires Tuple, got %s", args[0].Type())
	}
	return NewEvalResult(tuple.Tail())
}

func builtinSeqAppend(args []object.Object) *EvalResult {
	if len(args) != 2 {
		return NewEvalError("_Seq!Append requires 2 arguments, got %d", len(args))
	}
	tuple, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Seq!Append requires Tuple as first arg, got %s", args[0].Type())
	}
	return NewEvalResult(tuple.Append(args[1]))
}

func builtinSeqLen(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Seq!Len requires 1 argument, got %d", len(args))
	}
	tuple, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Seq!Len requires Tuple, got %s", args[0].Type())
	}
	return NewEvalResult(object.NewNatural(int64(tuple.Len())))
}

// === Stack Builtins (LIFO) ===

func builtinStackPush(args []object.Object) *EvalResult {
	if len(args) != 2 {
		return NewEvalError("_Stack!Push requires 2 arguments, got %d", len(args))
	}
	stack, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Stack!Push requires Tuple as first arg, got %s", args[0].Type())
	}
	return NewEvalResult(stack.Prepend(args[1]))
}

func builtinStackPop(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Stack!Pop requires 1 argument, got %d", len(args))
	}
	stack, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Stack!Pop requires Tuple, got %s", args[0].Type())
	}
	head := stack.Head()
	if head == nil {
		return NewEvalError("_Stack!Pop on empty stack")
	}
	return NewEvalResult(head)
}

// === Queue Builtins (FIFO) ===

func builtinQueueEnqueue(args []object.Object) *EvalResult {
	if len(args) != 2 {
		return NewEvalError("_Queue!Enqueue requires 2 arguments, got %d", len(args))
	}
	queue, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Queue!Enqueue requires Tuple as first arg, got %s", args[0].Type())
	}
	return NewEvalResult(queue.Append(args[1]))
}

func builtinQueueDequeue(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Queue!Dequeue requires 1 argument, got %d", len(args))
	}
	queue, ok := args[0].(*object.Tuple)
	if !ok {
		return NewEvalError("_Queue!Dequeue requires Tuple, got %s", args[0].Type())
	}
	head := queue.Head()
	if head == nil {
		return NewEvalError("_Queue!Dequeue on empty queue")
	}
	return NewEvalResult(head)
}

// === Bag Builtins ===

func builtinSetToBag(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Bags!SetToBag requires 1 argument, got %d", len(args))
	}
	set, ok := args[0].(*object.Set)
	if !ok {
		return NewEvalError("_Bags!SetToBag requires Set, got %s", args[0].Type())
	}
	bag := object.NewBag()
	for _, elem := range set.Elements() {
		bag.Add(elem, 1)
	}
	return NewEvalResult(bag)
}

func builtinBagToSet(args []object.Object) *EvalResult {
	if len(args) != 1 {
		return NewEvalError("_Bags!BagToSet requires 1 argument, got %d", len(args))
	}
	bag, ok := args[0].(*object.Bag)
	if !ok {
		return NewEvalError("_Bags!BagToSet requires Bag, got %s", args[0].Type())
	}
	set := object.NewSet()
	for _, elem := range bag.Elements() {
		set.Add(elem)
	}
	return NewEvalResult(set)
}

func builtinCopiesIn(args []object.Object) *EvalResult {
	if len(args) != 2 {
		return NewEvalError("_Bags!CopiesIn requires 2 arguments, got %d", len(args))
	}
	bag, ok := args[1].(*object.Bag)
	if !ok {
		return NewEvalError("_Bags!CopiesIn requires Bag as second arg, got %s", args[1].Type())
	}
	count := bag.CopiesIn(args[0])
	return NewEvalResult(object.NewNatural(int64(count)))
}

func builtinBagIn(args []object.Object) *EvalResult {
	if len(args) != 2 {
		return NewEvalError("_Bags!BagIn requires 2 arguments, got %d", len(args))
	}
	bag, ok := args[1].(*object.Bag)
	if !ok {
		return NewEvalError("_Bags!BagIn requires Bag as second arg, got %s", args[1].Type())
	}
	contains := bag.Contains(args[0])
	return NewEvalResult(nativeBoolToBoolean(contains))
}
