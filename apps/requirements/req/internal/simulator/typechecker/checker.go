package typechecker

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
)

// RegistryInterface abstracts the registry for type checking purposes.
// This allows the type checker to resolve custom function calls without
// a direct dependency on the registry package.
type RegistryInterface interface {
	// ResolveCallExpression resolves a call expression to a function definition.
	// Returns the definition key, parameter types, and return type.
	ResolveCallExpression(call *ast.CallExpression, scopeLevel int, domain, subdomain, class string) (
		key string,
		paramTypes []types.Type,
		returnType types.Type,
		err error,
	)
}

// DependencyTracker records dependencies discovered during type checking.
type DependencyTracker interface {
	// RecordDependency records that the current definition depends on another.
	RecordDependency(fromKey, toKey string)
}

// TypeChecker performs Hindley-Milner type inference on AST nodes.
type TypeChecker struct {
	env   *TypeEnv
	subst types.Substitution

	// Optional registry for custom function resolution
	registry    RegistryInterface
	scopeLevel  int    // Current scope level (0=global, 1=domain, 2=subdomain, 3=class)
	scopeDomain string // Current domain (if scopeLevel >= 1)
	scopeSub    string // Current subdomain (if scopeLevel >= 2)
	scopeClass  string // Current class (if scopeLevel >= 3)

	// Optional dependency tracker
	depTracker DependencyTracker
	currentKey string // Key of the definition being type-checked

	// Relation types for association fields
	// Maps: classKey -> fieldName -> type
	// Used to type-check .Name (forward) and ._Name (reverse) relation access
	relationTypes map[string]map[string]types.Type
}

// NewTypeChecker creates a new type checker with an initial environment.
func NewTypeChecker() *TypeChecker {
	tc := &TypeChecker{
		env:   NewEnv(),
		subst: make(types.Substitution),
	}
	// Add builtin type signatures
	tc.addBuiltins()
	return tc
}

// SetRegistry sets the registry interface for custom function resolution.
func (tc *TypeChecker) SetRegistry(registry RegistryInterface) {
	tc.registry = registry
}

// SetScope sets the current scope for type checking.
// This determines how relative function calls are resolved.
//
// Scope levels:
//   - 0 (global): Can call Domain!Subdomain!Class!Func
//   - 1 (domain): Can call Subdomain!Class!Func
//   - 2 (subdomain): Can call Class!Func
//   - 3 (class): Can call Func
func (tc *TypeChecker) SetScope(level int, domain, subdomain, class string) {
	tc.scopeLevel = level
	tc.scopeDomain = domain
	tc.scopeSub = subdomain
	tc.scopeClass = class
}

// SetDependencyTracker sets the dependency tracker for recording call dependencies.
func (tc *TypeChecker) SetDependencyTracker(tracker DependencyTracker, currentKey string) {
	tc.depTracker = tracker
	tc.currentKey = currentKey
}

// ClearDependencyTracker removes the dependency tracker.
func (tc *TypeChecker) ClearDependencyTracker() {
	tc.depTracker = nil
	tc.currentKey = ""
}

// SetRelationTypes sets the relation types for a class.
// The fieldTypes map contains field names (including "_" prefix for reverse) to their types.
// Forward relations (e.g., "Lines") and reverse relations (e.g., "_Lines") should both be included.
func (tc *TypeChecker) SetRelationTypes(classKey string, fieldTypes map[string]types.Type) {
	if tc.relationTypes == nil {
		tc.relationTypes = make(map[string]map[string]types.Type)
	}
	tc.relationTypes[classKey] = fieldTypes
}

// GetRelationType returns the type for a relation field on a class.
// Returns nil if no relation type is registered.
func (tc *TypeChecker) GetRelationType(classKey, fieldName string) types.Type {
	if tc.relationTypes == nil {
		return nil
	}
	classMap, ok := tc.relationTypes[classKey]
	if !ok {
		return nil
	}
	return classMap[fieldName]
}

// ClearRelationTypes removes all relation type registrations.
func (tc *TypeChecker) ClearRelationTypes() {
	tc.relationTypes = nil
}

// Env returns the type environment.
func (tc *TypeChecker) Env() *TypeEnv {
	return tc.env
}

// TypedNode wraps an AST node with its inferred type.
type TypedNode struct {
	Node     ast.Node
	Type     types.Type
	Children []*TypedNode // For compound expressions
}

// TypeError represents a type checking error with location info.
type TypeError struct {
	Node    ast.Node
	Message string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("type error: %s in %s", e.Message, e.Node.String())
}

// Check performs type checking on an AST node.
// Returns a TypedNode with type annotations, or an error.
func (tc *TypeChecker) Check(node ast.Node) (*TypedNode, error) {
	return tc.infer(node, tc.env)
}

// DeclareVariable adds a variable with a known type to the type environment.
func (tc *TypeChecker) DeclareVariable(name string, typ types.Type) {
	tc.env.BindMono(name, typ)
}

// freshTypeVar creates a new type variable.
func (tc *TypeChecker) freshTypeVar(name string) types.TypeVar {
	return types.NewTypeVar(name)
}

// instantiate creates a fresh instance of a type scheme.
// Replaces bound type variables with fresh ones.
func (tc *TypeChecker) instantiate(scheme types.Scheme) types.Type {
	if len(scheme.TypeVars) == 0 {
		return scheme.Type
	}

	// Create fresh type variables for each bound variable
	freshVars := make(types.Substitution)
	for _, id := range scheme.TypeVars {
		freshVars[id] = tc.freshTypeVar("")
	}

	// Substitute bound vars with fresh vars
	return freshVars.Apply(scheme.Type)
}

// generalize creates a type scheme by quantifying over free type variables
// that are not free in the environment.
func (tc *TypeChecker) generalize(t types.Type, env *TypeEnv) types.Scheme {
	// Apply current substitution first
	t = tc.subst.Apply(t)

	freeInType := t.FreeTypeVars()
	freeInEnv := env.FreeTypeVars()

	// Quantify variables free in type but not in environment
	var toQuantify []int
	for id := range freeInType {
		if _, inEnv := freeInEnv[id]; !inEnv {
			toQuantify = append(toQuantify, id)
		}
	}

	return types.Scheme{TypeVars: toQuantify, Type: t}
}

// unify unifies two types and updates the substitution.
func (tc *TypeChecker) unify(t1, t2 types.Type) error {
	newSubst, err := UnifyWithSubst(t1, t2, tc.subst)
	if err != nil {
		return err
	}
	tc.subst = newSubst
	return nil
}

// apply applies the current substitution to a type.
func (tc *TypeChecker) apply(t types.Type) types.Type {
	return tc.subst.Apply(t)
}

// infer implements Algorithm W for type inference.
func (tc *TypeChecker) infer(node ast.Node, env *TypeEnv) (*TypedNode, error) {
	switch n := node.(type) {
	// === Literals ===

	case *ast.BooleanLiteral:
		return &TypedNode{Node: n, Type: types.Boolean{}}, nil

	case *ast.NumberLiteral:
		return &TypedNode{Node: n, Type: types.Number{}}, nil

	case *ast.NumericPrefixExpression:
		return tc.inferNumericPrefix(n, env)

	case *ast.FractionExpr:
		return tc.inferFractionExpr(n, env)

	case *ast.ParenExpr:
		return tc.infer(n.Inner, env)

	case *ast.StringLiteral:
		return &TypedNode{Node: n, Type: types.String{}}, nil

	// === Identifiers ===

	case *ast.Identifier:
		scheme, ok := env.Lookup(n.Value)
		if !ok {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("unbound variable: %s", n.Value)}
		}
		t := tc.instantiate(scheme)
		return &TypedNode{Node: n, Type: t}, nil

	// === Arithmetic ===

	case *ast.RealInfixExpression:
		return tc.inferRealInfix(n, env)

	// === Logic ===

	case *ast.LogicInfixExpression:
		return tc.inferLogicInfix(n, env)

	case *ast.LogicPrefixExpression:
		return tc.inferLogicPrefix(n, env)

	case *ast.LogicRealComparison:
		return tc.inferRealComparison(n, env)

	case *ast.LogicMembership:
		return tc.inferMembership(n, env)

	case *ast.LogicBoundQuantifier:
		return tc.inferQuantifier(n, env)

	case *ast.LogicInfixSet:
		return tc.inferLogicInfixSet(n, env)

	case *ast.LogicInfixBag:
		return tc.inferLogicInfixBag(n, env)

	// === Sets ===

	case *ast.SetLiteralInt:
		return &TypedNode{Node: n, Type: types.Set{Element: types.Number{}}}, nil

	case *ast.SetLiteralEnum:
		return tc.inferSetLiteralEnum(n, env)

	case *ast.SetRange:
		return tc.inferSetRange(n, env)

	case *ast.SetInfix:
		return tc.inferSetInfix(n, env)

	case *ast.SetConditional:
		return tc.inferSetConditional(n, env)

	case *ast.SetConstant:
		return tc.inferSetConstant(n, env)

	// === Tuples ===

	case *ast.TupleLiteral:
		return tc.inferTupleLiteral(n, env)

	case *ast.ExpressionTupleIndex:
		return tc.inferTupleIndex(n, env)

	case *ast.TupleInfixExpression:
		return tc.inferTupleInfix(n, env)

	// === Records ===

	case *ast.RecordInstance:
		return tc.inferRecordInstance(n, env)

	case *ast.RecordAltered:
		return tc.inferRecordAltered(n, env)

	case *ast.FieldIdentifier:
		return tc.inferFieldAccess(n, env)

	// === Bags ===

	case *ast.BagInfix:
		return tc.inferBagInfix(n, env)

	// === Strings ===

	case *ast.StringInfixExpression:
		return tc.inferStringInfix(n, env)

	case *ast.StringIndex:
		return tc.inferStringIndex(n, env)

	// === Control Flow ===

	case *ast.ExpressionIfElse:
		return tc.inferIfElse(n, env)

	case *ast.ExpressionCase:
		return tc.inferCase(n, env)

	// === Calls ===

	case *ast.BuiltinCall:
		return tc.inferBuiltinCall(n, env)

	case *ast.CallExpression:
		return tc.inferCallExpression(n, env)

	// === Special ===

	case *ast.ExistingValue:
		return tc.inferExistingValue(n, env)

	default:
		return nil, &TypeError{Node: node, Message: fmt.Sprintf("unsupported node type: %T", node)}
	}
}

// === Inference Helpers ===

func (tc *TypeChecker) inferRealInfix(n *ast.RealInfixExpression, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(left.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Number, got %s", n.Operator, tc.apply(left.Type))}
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(right.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Number, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Number{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferLogicInfix(n *ast.LogicInfixExpression, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(left.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Boolean, got %s", n.Operator, tc.apply(left.Type))}
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(right.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Boolean, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferLogicPrefix(n *ast.LogicPrefixExpression, env *TypeEnv) (*TypedNode, error) {
	operand, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(operand.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("operand of %s: expected Boolean, got %s", n.Operator, tc.apply(operand.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{operand},
	}, nil
}

func (tc *TypeChecker) inferRealComparison(n *ast.LogicRealComparison, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(left.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Number, got %s", n.Operator, tc.apply(left.Type))}
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(right.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Number, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferMembership(n *ast.LogicMembership, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}

	// Right must be a Set[T] where T matches left
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(right.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Set, got %s", n.Operator, tc.apply(right.Type))}
	}

	if err := tc.unify(left.Type, elemType); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("element type mismatch in %s", n.Operator)}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferQuantifier(n *ast.LogicBoundQuantifier, env *TypeEnv) (*TypedNode, error) {
	// Process membership to get bound variable
	membership, ok := n.Membership.(*ast.LogicMembership)
	if !ok {
		return nil, &TypeError{Node: n, Message: "quantifier requires LogicMembership"}
	}

	ident, ok := membership.Left.(*ast.Identifier)
	if !ok {
		return nil, &TypeError{Node: n, Message: "quantifier binding must be an identifier"}
	}

	// Infer set type
	setTyped, err := tc.infer(membership.Right, env)
	if err != nil {
		return nil, err
	}

	// Get element type from set
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(setTyped.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("quantifier domain: expected Set, got %s", tc.apply(setTyped.Type))}
	}

	// Extend environment with bound variable
	innerEnv := env.Extend()
	innerEnv.BindMono(ident.Value, tc.apply(elemType))

	// Type check predicate in extended environment
	pred, err := tc.infer(n.Predicate, innerEnv)
	if err != nil {
		return nil, err
	}

	if err := tc.unify(pred.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("quantifier predicate: expected Boolean, got %s", tc.apply(pred.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{setTyped, pred},
	}, nil
}

func (tc *TypeChecker) inferLogicInfixSet(n *ast.LogicInfixSet, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}

	// Both must be sets with compatible element types
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(left.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Set, got %s", n.Operator, tc.apply(left.Type))}
	}
	if err := tc.unify(right.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Set, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferLogicInfixBag(n *ast.LogicInfixBag, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}

	// Both must be bags
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(left.Type, types.Bag{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand: expected Bag, got %s", tc.apply(left.Type))}
	}
	if err := tc.unify(right.Type, types.Bag{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand: expected Bag, got %s", tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Boolean{},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferSetLiteralEnum(n *ast.SetLiteralEnum, env *TypeEnv) (*TypedNode, error) {
	// SetLiteralEnum has []string Values, so it's always a Set[String]
	if len(n.Values) == 0 {
		// Empty set has polymorphic element type
		elemType := tc.freshTypeVar("elem")
		return &TypedNode{Node: n, Type: types.Set{Element: elemType}}, nil
	}

	// Enum sets contain strings
	return &TypedNode{
		Node: n,
		Type: types.Set{Element: types.String{}},
	}, nil
}

func (tc *TypeChecker) inferSetRange(n *ast.SetRange, env *TypeEnv) (*TypedNode, error) {
	// SetRange has Start and End as int fields, so it's always Set[Number]
	return &TypedNode{
		Node: n,
		Type: types.Set{Element: types.Number{}},
	}, nil
}

func (tc *TypeChecker) inferSetInfix(n *ast.SetInfix, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}

	// Both must be sets with compatible element types
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(left.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Set, got %s", n.Operator, tc.apply(left.Type))}
	}
	if err := tc.unify(right.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Set, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Set{Element: tc.apply(elemType)},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferSetConditional(n *ast.SetConditional, env *TypeEnv) (*TypedNode, error) {
	// {x ∈ S : P(x)} - uses Membership and Predicate fields
	// Get the membership which contains the binding
	membership, ok := n.Membership.(*ast.LogicMembership)
	if !ok {
		return nil, &TypeError{Node: n, Message: "set comprehension requires LogicMembership"}
	}

	ident, ok := membership.Left.(*ast.Identifier)
	if !ok {
		return nil, &TypeError{Node: n, Message: "set comprehension binding must be an identifier"}
	}

	// Infer source set type
	source, err := tc.infer(membership.Right, env)
	if err != nil {
		return nil, err
	}

	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(source.Type, types.Set{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("set comprehension source: expected Set, got %s", tc.apply(source.Type))}
	}

	// Extend environment
	innerEnv := env.Extend()
	innerEnv.BindMono(ident.Value, tc.apply(elemType))

	// Check predicate
	pred, err := tc.infer(n.Predicate, innerEnv)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(pred.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("set comprehension predicate: expected Boolean, got %s", tc.apply(pred.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Set{Element: tc.apply(elemType)},
		Children: []*TypedNode{source, pred},
	}, nil
}

func (tc *TypeChecker) inferSetConstant(n *ast.SetConstant, env *TypeEnv) (*TypedNode, error) {
	// Constants like BOOLEAN, Nat, Int, Real
	switch n.Value {
	case "BOOLEAN":
		return &TypedNode{Node: n, Type: types.Set{Element: types.Boolean{}}}, nil
	case "Nat", "Int", "Real":
		return &TypedNode{Node: n, Type: types.Set{Element: types.Number{}}}, nil
	default:
		// Unknown constant - use polymorphic type
		elemType := tc.freshTypeVar("elem")
		return &TypedNode{Node: n, Type: types.Set{Element: elemType}}, nil
	}
}

func (tc *TypeChecker) inferTupleLiteral(n *ast.TupleLiteral, env *TypeEnv) (*TypedNode, error) {
	if len(n.Elements) == 0 {
		elemType := tc.freshTypeVar("elem")
		return &TypedNode{Node: n, Type: types.Tuple{Element: elemType}}, nil
	}

	// TLA+ tuples are homogeneous sequences
	elemType := tc.freshTypeVar("elem")
	var children []*TypedNode

	for _, elem := range n.Elements {
		typed, err := tc.infer(elem, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(typed.Type, elemType); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("tuple element type mismatch: %s", tc.apply(typed.Type))}
		}
		children = append(children, typed)
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Tuple{Element: tc.apply(elemType)},
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferTupleIndex(n *ast.ExpressionTupleIndex, env *TypeEnv) (*TypedNode, error) {
	tuple, err := tc.infer(n.Tuple, env)
	if err != nil {
		return nil, err
	}

	index, err := tc.infer(n.Index, env)
	if err != nil {
		return nil, err
	}

	// Index must be a number
	if err := tc.unify(index.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("tuple index: expected Number, got %s", tc.apply(index.Type))}
	}

	// Tuple type
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(tuple.Type, types.Tuple{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("tuple indexing: expected Tuple, got %s", tc.apply(tuple.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(elemType),
		Children: []*TypedNode{tuple, index},
	}, nil
}

func (tc *TypeChecker) inferTupleInfix(n *ast.TupleInfixExpression, env *TypeEnv) (*TypedNode, error) {
	if len(n.Operands) < 2 {
		return nil, &TypeError{Node: n, Message: "tuple concatenation requires at least 2 operands"}
	}

	elemType := tc.freshTypeVar("elem")
	var children []*TypedNode

	for i, operand := range n.Operands {
		typed, err := tc.infer(operand, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(typed.Type, types.Tuple{Element: elemType}); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("operand %d of %s: expected Tuple, got %s", i+1, n.Operator, tc.apply(typed.Type))}
		}
		children = append(children, typed)
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Tuple{Element: tc.apply(elemType)},
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferRecordInstance(n *ast.RecordInstance, env *TypeEnv) (*TypedNode, error) {
	fields := make(map[string]types.Type)
	var children []*TypedNode

	for _, binding := range n.Bindings {
		// binding.Field is already *ast.Identifier
		typed, err := tc.infer(binding.Expression, env)
		if err != nil {
			return nil, err
		}

		fields[binding.Field.Value] = tc.apply(typed.Type)
		children = append(children, typed)
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Record{Fields: fields},
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferRecordAltered(n *ast.RecordAltered, env *TypeEnv) (*TypedNode, error) {
	// RecordAltered has Identifier (the base record) and Alterations
	record, err := tc.infer(n.Identifier, env)
	if err != nil {
		return nil, err
	}

	// Base must be a record
	baseRecord, ok := tc.apply(record.Type).(types.Record)
	if !ok {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("record update: expected Record, got %s", tc.apply(record.Type))}
	}

	// Create new fields map with updates
	newFields := make(map[string]types.Type)
	for name, ft := range baseRecord.Fields {
		newFields[name] = ft
	}

	var children []*TypedNode
	children = append(children, record)

	for _, alter := range n.Alterations {
		// alter.Field is *ast.FieldIdentifier which has Member field
		typed, err := tc.infer(alter.Expression, env)
		if err != nil {
			return nil, err
		}

		newFields[alter.Field.Member] = tc.apply(typed.Type)
		children = append(children, typed)
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Record{Fields: newFields},
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferFieldAccess(n *ast.FieldIdentifier, env *TypeEnv) (*TypedNode, error) {
	if n.Identifier == nil {
		// Field access without explicit identifier (uses context)
		return nil, &TypeError{Node: n, Message: "field access without identifier requires context"}
	}

	record, err := tc.infer(n.Identifier, env)
	if err != nil {
		return nil, err
	}

	// Record must be a record type
	recType, ok := tc.apply(record.Type).(types.Record)
	if !ok {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("field access: expected Record, got %s", tc.apply(record.Type))}
	}

	// First, check if this is a regular record field
	fieldType, exists := recType.Fields[n.Member]
	if exists {
		return &TypedNode{
			Node:     n,
			Type:     fieldType,
			Children: []*TypedNode{record},
		}, nil
	}

	// If not found, check if it's a relation field (when in class scope)
	if tc.scopeLevel == 3 && tc.scopeClass != "" {
		// Build class key from scope (this is a simplified key construction)
		// In practice, the full identity.Key would be used
		classKey := tc.scopeDomain + "/" + tc.scopeSub + "/" + tc.scopeClass
		if relationType := tc.GetRelationType(classKey, n.Member); relationType != nil {
			return &TypedNode{
				Node:     n,
				Type:     relationType,
				Children: []*TypedNode{record},
			}, nil
		}
	}

	return nil, &TypeError{Node: n, Message: fmt.Sprintf("record does not have field: %s", n.Member)}
}

func (tc *TypeChecker) inferBagInfix(n *ast.BagInfix, env *TypeEnv) (*TypedNode, error) {
	left, err := tc.infer(n.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}

	// Both must be bags with compatible element types
	elemType := tc.freshTypeVar("elem")
	if err := tc.unify(left.Type, types.Bag{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("left operand of %s: expected Bag, got %s", n.Operator, tc.apply(left.Type))}
	}
	if err := tc.unify(right.Type, types.Bag{Element: elemType}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("right operand of %s: expected Bag, got %s", n.Operator, tc.apply(right.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Bag{Element: tc.apply(elemType)},
		Children: []*TypedNode{left, right},
	}, nil
}

func (tc *TypeChecker) inferStringInfix(n *ast.StringInfixExpression, env *TypeEnv) (*TypedNode, error) {
	if len(n.Operands) < 2 {
		return nil, &TypeError{Node: n, Message: "string concatenation requires at least 2 operands"}
	}

	var children []*TypedNode

	for i, operand := range n.Operands {
		typed, err := tc.infer(operand, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(typed.Type, types.String{}); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("operand %d of %s: expected String, got %s", i+1, n.Operator, tc.apply(typed.Type))}
		}
		children = append(children, typed)
	}

	return &TypedNode{
		Node:     n,
		Type:     types.String{},
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferStringIndex(n *ast.StringIndex, env *TypeEnv) (*TypedNode, error) {
	str, err := tc.infer(n.Str, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(str.Type, types.String{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("string indexing: expected String, got %s", tc.apply(str.Type))}
	}

	index, err := tc.infer(n.Index, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(index.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("string index: expected Number, got %s", tc.apply(index.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.String{},
		Children: []*TypedNode{str, index},
	}, nil
}

func (tc *TypeChecker) inferIfElse(n *ast.ExpressionIfElse, env *TypeEnv) (*TypedNode, error) {
	cond, err := tc.infer(n.Condition, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(cond.Type, types.Boolean{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("if condition: expected Boolean, got %s", tc.apply(cond.Type))}
	}

	thenBranch, err := tc.infer(n.Then, env)
	if err != nil {
		return nil, err
	}

	elseBranch, err := tc.infer(n.Else, env)
	if err != nil {
		return nil, err
	}

	// Both branches must have same type
	if err := tc.unify(thenBranch.Type, elseBranch.Type); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("if branches must have same type: %s vs %s", tc.apply(thenBranch.Type), tc.apply(elseBranch.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(thenBranch.Type),
		Children: []*TypedNode{cond, thenBranch, elseBranch},
	}, nil
}

func (tc *TypeChecker) inferCase(n *ast.ExpressionCase, env *TypeEnv) (*TypedNode, error) {
	if len(n.Branches) == 0 {
		return nil, &TypeError{Node: n, Message: "case expression must have at least one branch"}
	}

	resultType := tc.freshTypeVar("result")
	var children []*TypedNode

	for _, branch := range n.Branches {
		condTyped, err := tc.infer(branch.Condition, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(condTyped.Type, types.Boolean{}); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("case condition: expected Boolean, got %s", tc.apply(condTyped.Type))}
		}

		resultTyped, err := tc.infer(branch.Result, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(resultTyped.Type, resultType); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("case branches must have same type")}
		}

		children = append(children, condTyped, resultTyped)
	}

	// Check OTHER branch if present
	if n.Other != nil {
		otherTyped, err := tc.infer(n.Other, env)
		if err != nil {
			return nil, err
		}
		if err := tc.unify(otherTyped.Type, resultType); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("case OTHER branch must match other branches")}
		}
		children = append(children, otherTyped)
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(resultType),
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferBuiltinCall(n *ast.BuiltinCall, env *TypeEnv) (*TypedNode, error) {
	scheme, ok := env.Lookup(n.Name)
	if !ok {
		// Check builtins registry
		scheme, ok = tc.env.Lookup(n.Name)
		if !ok {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("unknown builtin: %s", n.Name)}
		}
	}

	// Instantiate the polymorphic type
	fnType := tc.instantiate(scheme)

	// Must be a function type
	fn, ok := fnType.(types.Function)
	if !ok {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("builtin %s is not a function: %s", n.Name, fnType)}
	}

	// Check argument count
	if len(n.Args) != len(fn.Params) {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("builtin %s expects %d arguments, got %d", n.Name, len(fn.Params), len(n.Args))}
	}

	// Type check arguments and unify with parameters
	var children []*TypedNode
	for i, arg := range n.Args {
		argTyped, err := tc.infer(arg, env)
		if err != nil {
			return nil, err
		}

		if err := tc.unify(argTyped.Type, fn.Params[i]); err != nil {
			return nil, &TypeError{Node: n, Message: fmt.Sprintf("argument %d of %s: expected %s, got %s", i+1, n.Name, tc.apply(fn.Params[i]), tc.apply(argTyped.Type))}
		}

		children = append(children, argTyped)
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(fn.Return),
		Children: children,
	}, nil
}

func (tc *TypeChecker) inferCallExpression(n *ast.CallExpression, env *TypeEnv) (*TypedNode, error) {
	// Use registry-based resolution if available
	if tc.registry != nil {
		return tc.inferCallExpressionWithRegistry(n, env)
	}

	// Legacy behavior: look up in type environment directly
	return tc.inferCallExpressionLegacy(n, env)
}

// inferCallExpressionWithRegistry uses the registry for function resolution.
func (tc *TypeChecker) inferCallExpressionWithRegistry(n *ast.CallExpression, env *TypeEnv) (*TypedNode, error) {
	// Resolve using registry with current scope context
	key, paramTypes, returnType, err := tc.registry.ResolveCallExpression(
		n,
		tc.scopeLevel,
		tc.scopeDomain,
		tc.scopeSub,
		tc.scopeClass,
	)
	if err != nil {
		return nil, &TypeError{Node: n, Message: err.Error()}
	}

	// Record dependency if tracking
	if tc.depTracker != nil && tc.currentKey != "" {
		tc.depTracker.RecordDependency(tc.currentKey, key)
	}

	// Type check arguments
	var children []*TypedNode

	// Get arguments from the call - they could be in Parameter field as a tuple/record
	// or could be multiple positional arguments
	args, err := tc.extractCallArguments(n, env)
	if err != nil {
		return nil, err
	}

	// Check argument count
	if len(args) != len(paramTypes) {
		return nil, &TypeError{
			Node:    n,
			Message: fmt.Sprintf("function %s expects %d arguments, got %d", key, len(paramTypes), len(args)),
		}
	}

	// Type check and unify each argument
	for i, arg := range args {
		argTyped, err := tc.infer(arg, env)
		if err != nil {
			return nil, err
		}

		if err := tc.unify(argTyped.Type, paramTypes[i]); err != nil {
			return nil, &TypeError{
				Node:    n,
				Message: fmt.Sprintf("argument %d of %s: expected %s, got %s", i+1, key, tc.apply(paramTypes[i]), tc.apply(argTyped.Type)),
			}
		}

		children = append(children, argTyped)
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(returnType),
		Children: children,
	}, nil
}

// extractCallArguments extracts the argument list from a call expression.
// Arguments can be in Parameter field as individual expressions.
func (tc *TypeChecker) extractCallArguments(n *ast.CallExpression, env *TypeEnv) ([]ast.Expression, error) {
	if n.Parameter == nil {
		return nil, nil // No arguments
	}

	// Check if Parameter is a tuple literal (multiple args)
	if tuple, ok := n.Parameter.(*ast.TupleLiteral); ok {
		return tuple.Elements, nil
	}

	// Single argument
	return []ast.Expression{n.Parameter}, nil
}

// inferCallExpressionLegacy is the original implementation for backward compatibility.
func (tc *TypeChecker) inferCallExpressionLegacy(n *ast.CallExpression, env *TypeEnv) (*TypedNode, error) {
	// Build the fully qualified function name
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

	// Look up the function definition
	scheme, ok := env.Lookup(fnName)
	if !ok {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("undefined function: %s", fnName)}
	}

	fnType := tc.instantiate(scheme)

	fn, ok := fnType.(types.Function)
	if !ok {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("function %s is not callable", fnName)}
	}

	// Type check the parameter (which is a record)
	paramTyped, err := tc.infer(n.Parameter, env)
	if err != nil {
		return nil, err
	}

	// Functions take a single record parameter
	if len(fn.Params) != 1 {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("function %s has invalid signature", fnName)}
	}

	if err := tc.unify(paramTyped.Type, fn.Params[0]); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("parameter of %s: type mismatch", fnName)}
	}

	return &TypedNode{
		Node:     n,
		Type:     tc.apply(fn.Return),
		Children: []*TypedNode{paramTyped},
	}, nil
}

func (tc *TypeChecker) inferExistingValue(n *ast.ExistingValue, env *TypeEnv) (*TypedNode, error) {
	// ExistingValue (@) references the current value in an update context
	// Its type depends on the context - we use a fresh type variable
	// that will be unified based on usage
	t := tc.freshTypeVar("existing")
	return &TypedNode{Node: n, Type: t}, nil
}

func (tc *TypeChecker) inferNumericPrefix(n *ast.NumericPrefixExpression, env *TypeEnv) (*TypedNode, error) {
	operand, err := tc.infer(n.Right, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(operand.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("operand of %s: expected Number, got %s", n.Operator, tc.apply(operand.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Number{},
		Children: []*TypedNode{operand},
	}, nil
}

func (tc *TypeChecker) inferFractionExpr(n *ast.FractionExpr, env *TypeEnv) (*TypedNode, error) {
	numerator, err := tc.infer(n.Numerator, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(numerator.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("fraction numerator: expected Number, got %s", tc.apply(numerator.Type))}
	}

	denominator, err := tc.infer(n.Denominator, env)
	if err != nil {
		return nil, err
	}
	if err := tc.unify(denominator.Type, types.Number{}); err != nil {
		return nil, &TypeError{Node: n, Message: fmt.Sprintf("fraction denominator: expected Number, got %s", tc.apply(denominator.Type))}
	}

	return &TypedNode{
		Node:     n,
		Type:     types.Number{},
		Children: []*TypedNode{numerator, denominator},
	}, nil
}
