package model_bridge

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
)

// GuaranteeKind classifies what kind of guarantee expression this is.
type GuaranteeKind int

const (
	// GuaranteeUnknown means the guarantee kind hasn't been determined yet.
	GuaranteeUnknown GuaranteeKind = iota
	// GuaranteePrimedAssignment is a primed assignment like `self.count' = value` or `result' = value`.
	// These define state changes or outputs.
	GuaranteePrimedAssignment
	// GuaranteePostCondition is a boolean invariant that must be TRUE after state changes.
	// Example: `record'.count > record.count`
	GuaranteePostCondition
)

func (k GuaranteeKind) String() string {
	switch k {
	case GuaranteePrimedAssignment:
		return "primed_assignment"
	case GuaranteePostCondition:
		return "post_condition"
	default:
		return "unknown"
	}
}

// DefinitionBuilder builds registry definitions from extracted expressions.
type DefinitionBuilder struct{}

// NewDefinitionBuilder creates a new DefinitionBuilder.
func NewDefinitionBuilder() *DefinitionBuilder {
	return &DefinitionBuilder{}
}

// BuildResult contains the result of building a definition.
type BuildResult struct {
	// Definition is the built definition, nil if building failed.
	Definition *registry.Definition

	// Error is the error that occurred during building, nil if successful.
	Error error

	// Source is the original extracted expression.
	Source ExtractedExpression

	// GuaranteeKind classifies guarantee expressions.
	// Only set for SourceActionGuarantees and SourceQueryGuarantees.
	GuaranteeKind GuaranteeKind
}

// IsSuccess returns true if the definition was built successfully.
func (r *BuildResult) IsSuccess() bool {
	return r.Error == nil && r.Definition != nil
}

// Build parses an extracted expression and registers it in the registry.
// Returns the built definition or an error.
func (b *DefinitionBuilder) Build(expr ExtractedExpression, reg *registry.Registry) *BuildResult {
	result := &BuildResult{Source: expr}

	// Parse the TLA+ expression
	parsedExpr, err := parser.ParseExpression(expr.Expression)
	if err != nil {
		result.Error = fmt.Errorf("parse error for %s: %w", expr.Source, err)
		return result
	}

	// Build the definition based on source type
	switch expr.Source {
	case SourceModelInvariant:
		def, err := b.buildModelInvariant(expr, parsedExpr, reg)
		result.Definition = def
		result.Error = err

	case SourceTlaDefinition:
		def, err := b.buildTlaDefinition(expr, parsedExpr, reg)
		result.Definition = def
		result.Error = err

	case SourceActionRequires:
		def, err := b.buildActionRequires(expr, parsedExpr, reg)
		result.Definition = def
		result.Error = err

	case SourceActionGuarantees:
		def, kind, err := b.buildActionGuarantees(expr, parsedExpr, reg)
		result.Definition = def
		result.GuaranteeKind = kind
		result.Error = err

	case SourceQueryRequires:
		def, err := b.buildQueryRequires(expr, parsedExpr, reg)
		result.Definition = def
		result.Error = err

	case SourceQueryGuarantees:
		def, kind, err := b.buildQueryGuarantees(expr, parsedExpr, reg)
		result.Definition = def
		result.GuaranteeKind = kind
		result.Error = err

	case SourceGuardCondition:
		def, err := b.buildGuardExpression(expr, parsedExpr, reg)
		result.Definition = def
		result.Error = err

	default:
		result.Error = fmt.Errorf("unsupported expression source: %s", expr.Source)
	}

	return result
}

// buildModelInvariant builds a model invariant definition.
// Model invariants are registered at global scope with a generated name.
func (b *DefinitionBuilder) buildModelInvariant(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, error) {
	// Generate a unique name for the invariant
	name := fmt.Sprintf("Invariant%d", expr.Index)

	def, err := reg.RegisterGlobalFunction(name, parsedExpr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register model invariant %d: %w", expr.Index, err)
	}

	return def, nil
}

// buildTlaDefinition builds a global TLA+ definition.
// TLA definitions are registered at global scope with their declared name.
func (b *DefinitionBuilder) buildTlaDefinition(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, error) {
	// Convert parameter names to registry.Parameter (types will be inferred later)
	params := make([]registry.Parameter, len(expr.Parameters))
	for i, paramName := range expr.Parameters {
		params[i] = registry.Parameter{Name: paramName}
	}

	def, err := reg.RegisterGlobalFunction(expr.Name, parsedExpr, params)
	if err != nil {
		return nil, fmt.Errorf("failed to register TLA definition %s: %w", expr.Name, err)
	}

	return def, nil
}

// buildActionRequires builds an action precondition expression.
// Each requires expression must evaluate to TRUE/FALSE and they are ANDed together.
// Actions are registered at class scope.
func (b *DefinitionBuilder) buildActionRequires(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, error) {
	if expr.ScopeKey == nil {
		return nil, fmt.Errorf("action expression requires a scope key")
	}

	domain, subdomain, class, err := extractClassScope(expr.ScopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to extract scope for action %s: %w", expr.Name, err)
	}

	name := fmt.Sprintf("%s_Requires%d", expr.Name, expr.Index)

	def, err := reg.RegisterClassFunction(domain, subdomain, class, name, parsedExpr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register action requires %s: %w", name, err)
	}

	return def, nil
}

// buildActionGuarantees builds an action guarantee expression.
// Guarantees can be either:
//   - Primed assignments: `self.field' = value` or `result' = value` (state changes/outputs)
//   - Post-conditions: boolean expressions that must be TRUE after state changes
//
// Actions are registered at class scope.
func (b *DefinitionBuilder) buildActionGuarantees(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, GuaranteeKind, error) {
	if expr.ScopeKey == nil {
		return nil, GuaranteeUnknown, fmt.Errorf("action expression requires a scope key")
	}

	domain, subdomain, class, err := extractClassScope(expr.ScopeKey)
	if err != nil {
		return nil, GuaranteeUnknown, fmt.Errorf("failed to extract scope for action %s: %w", expr.Name, err)
	}

	// Classify the guarantee
	kind := ClassifyGuarantee(parsedExpr)

	name := fmt.Sprintf("%s_Guarantees%d", expr.Name, expr.Index)

	def, err := reg.RegisterClassFunction(domain, subdomain, class, name, parsedExpr, nil)
	if err != nil {
		return nil, GuaranteeUnknown, fmt.Errorf("failed to register action guarantees %s: %w", name, err)
	}

	return def, kind, nil
}

// buildQueryRequires builds a query precondition expression.
// Each requires expression must evaluate to TRUE/FALSE and they are ANDed together.
// Queries are registered at class scope.
func (b *DefinitionBuilder) buildQueryRequires(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, error) {
	if expr.ScopeKey == nil {
		return nil, fmt.Errorf("query expression requires a scope key")
	}

	domain, subdomain, class, err := extractClassScope(expr.ScopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to extract scope for query %s: %w", expr.Name, err)
	}

	name := fmt.Sprintf("%s_Requires%d", expr.Name, expr.Index)

	def, err := reg.RegisterClassFunction(domain, subdomain, class, name, parsedExpr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register query requires %s: %w", name, err)
	}

	return def, nil
}

// buildQueryGuarantees builds a query guarantee expression.
// Guarantees can be either:
//   - Primed assignments: `result' = value` (query outputs)
//   - Post-conditions: boolean expressions that must be TRUE (filtering criteria)
//
// Queries are registered at class scope.
func (b *DefinitionBuilder) buildQueryGuarantees(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, GuaranteeKind, error) {
	if expr.ScopeKey == nil {
		return nil, GuaranteeUnknown, fmt.Errorf("query expression requires a scope key")
	}

	domain, subdomain, class, err := extractClassScope(expr.ScopeKey)
	if err != nil {
		return nil, GuaranteeUnknown, fmt.Errorf("failed to extract scope for query %s: %w", expr.Name, err)
	}

	// Classify the guarantee
	kind := ClassifyGuarantee(parsedExpr)

	name := fmt.Sprintf("%s_Guarantees%d", expr.Name, expr.Index)

	def, err := reg.RegisterClassFunction(domain, subdomain, class, name, parsedExpr, nil)
	if err != nil {
		return nil, GuaranteeUnknown, fmt.Errorf("failed to register query guarantees %s: %w", name, err)
	}

	return def, kind, nil
}

// ClassifyGuarantee determines if a guarantee expression is a primed assignment
// or a post-condition invariant.
//
// A primed assignment has the form: primed_expr = value
// where primed_expr contains a Primed node (e.g., self.count', result')
//
// A post-condition is any other boolean expression.
func ClassifyGuarantee(expr ast.Expression) GuaranteeKind {
	// Check if this is an equality with a primed LHS
	if eq, ok := expr.(*ast.BinaryEquality); ok {
		if eq.Operator == "=" && containsPrimed(eq.Left) {
			return GuaranteePrimedAssignment
		}
	}

	// Otherwise it's a post-condition
	return GuaranteePostCondition
}

// containsPrimed checks if a LHS expression contains a Primed node.
// This is used specifically for classifying guarantee LHS (shallow walk).
func containsPrimed(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Primed:
		return true
	case *ast.FieldAccess:
		if e.Base != nil {
			return containsPrimed(e.Base)
		}
		if e.Identifier != nil {
			return containsPrimed(e.Identifier)
		}
		return false
	case *ast.TupleIndex:
		return containsPrimed(e.Tuple)
	default:
		return false
	}
}

// ContainsAnyPrimed recursively walks the entire AST to detect any Primed node.
// Unlike containsPrimed (which only checks LHS patterns), this walks all
// branches of the expression tree.
func ContainsAnyPrimed(expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *ast.Primed:
		return true

	// Leaf nodes.
	case *ast.Identifier, *ast.NumberLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.SetConstant, *ast.SetLiteralEnum,
		*ast.SetLiteralInt, *ast.SetRange, *ast.ExistingValue:
		return false

	// Binary operators.
	case *ast.BinaryArithmetic:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinaryLogic:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinaryComparison:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinaryEquality:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinarySetComparison:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinarySetOperation:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinaryBagComparison:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.BinaryBagOperation:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.Membership:
		return ContainsAnyPrimed(e.Left) || ContainsAnyPrimed(e.Right)
	case *ast.Fraction:
		return ContainsAnyPrimed(e.Numerator) || ContainsAnyPrimed(e.Denominator)

	// Unary operators.
	case *ast.UnaryLogic:
		return ContainsAnyPrimed(e.Right)
	case *ast.UnaryNegation:
		return ContainsAnyPrimed(e.Right)
	case *ast.Parenthesized:
		return ContainsAnyPrimed(e.Inner)

	// Access and indexing.
	case *ast.FieldAccess:
		return ContainsAnyPrimed(e.Base)
	case *ast.TupleIndex:
		return ContainsAnyPrimed(e.Tuple) || ContainsAnyPrimed(e.Index)
	case *ast.StringIndex:
		return ContainsAnyPrimed(e.Str) || ContainsAnyPrimed(e.Index)

	// Concatenation.
	case *ast.StringConcat:
		for _, op := range e.Operands {
			if ContainsAnyPrimed(op) {
				return true
			}
		}
		return false
	case *ast.TupleConcat:
		for _, op := range e.Operands {
			if ContainsAnyPrimed(op) {
				return true
			}
		}
		return false

	// Conditionals.
	case *ast.IfThenElse:
		return ContainsAnyPrimed(e.Condition) || ContainsAnyPrimed(e.Then) || ContainsAnyPrimed(e.Else)
	case *ast.CaseExpr:
		for _, branch := range e.Branches {
			if ContainsAnyPrimed(branch.Condition) || ContainsAnyPrimed(branch.Result) {
				return true
			}
		}
		return ContainsAnyPrimed(e.Other)

	// Quantification and filtering.
	case *ast.Quantifier:
		return ContainsAnyPrimed(e.Membership) || ContainsAnyPrimed(e.Predicate)
	case *ast.SetFilter:
		return ContainsAnyPrimed(e.Membership) || ContainsAnyPrimed(e.Predicate)

	// Literals with expression children.
	case *ast.SetLiteral:
		for _, elem := range e.Elements {
			if ContainsAnyPrimed(elem) {
				return true
			}
		}
		return false
	case *ast.SetRangeExpr:
		return ContainsAnyPrimed(e.Start) || ContainsAnyPrimed(e.End)
	case *ast.TupleLiteral:
		for _, elem := range e.Elements {
			if ContainsAnyPrimed(elem) {
				return true
			}
		}
		return false
	case *ast.RecordInstance:
		for _, binding := range e.Bindings {
			if ContainsAnyPrimed(binding.Expression) {
				return true
			}
		}
		return false
	case *ast.RecordAltered:
		for _, alt := range e.Alterations {
			if ContainsAnyPrimed(alt.Expression) {
				return true
			}
		}
		return false

	// Function calls.
	case *ast.FunctionCall:
		for _, arg := range e.Args {
			if ContainsAnyPrimed(arg) {
				return true
			}
		}
		return false
	case *ast.BuiltinCall:
		for _, arg := range e.Args {
			if ContainsAnyPrimed(arg) {
				return true
			}
		}
		return false
	case *ast.ScopedCall:
		return ContainsAnyPrimed(e.Parameter)

	default:
		return false
	}
}

// buildGuardExpression builds a guard condition expression.
// Guards are registered at class scope.
func (b *DefinitionBuilder) buildGuardExpression(
	expr ExtractedExpression,
	parsedExpr ast.Expression,
	reg *registry.Registry,
) (*registry.Definition, error) {
	if expr.ScopeKey == nil {
		return nil, fmt.Errorf("guard expression requires a scope key")
	}

	domain, subdomain, class, err := extractClassScope(expr.ScopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to extract scope for guard %s: %w", expr.Name, err)
	}

	// Generate a unique name combining guard name and index
	name := fmt.Sprintf("%s_Guard%d", expr.Name, expr.Index)

	def, err := reg.RegisterClassFunction(domain, subdomain, class, name, parsedExpr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register guard expression %s: %w", name, err)
	}

	return def, nil
}

// extractClassScope extracts domain, subdomain, and class from an action/query/guard key.
// The key format is: domain/DOMAIN_NAME/subdomain/SUBDOMAIN_NAME/class/CLASS_NAME/[action|query|guard]/NAME
func extractClassScope(key *identity.Key) (domain, subdomain, class string, err error) {
	// Get the parent key string which should be the class key
	parentKeyStr := key.ParentKey()
	if parentKeyStr == "" {
		return "", "", "", fmt.Errorf("key has no parent: %s", key.String())
	}

	// Parse the parent key to extract class components
	// Format: domain/DOMAIN_NAME/subdomain/SUBDOMAIN_NAME/class/CLASS_NAME
	parts := strings.Split(parentKeyStr, "/")
	if len(parts) < 6 {
		return "", "", "", fmt.Errorf("invalid parent key format: %s", parentKeyStr)
	}

	// Find the domain, subdomain, and class components
	for i := 0; i < len(parts)-1; i += 2 {
		keyType := parts[i]
		value := parts[i+1]

		switch keyType {
		case "domain":
			domain = value
		case "subdomain":
			subdomain = value
		case "class":
			class = value
		}
	}

	if domain == "" || subdomain == "" || class == "" {
		return "", "", "", fmt.Errorf("could not extract all scope components from: %s", parentKeyStr)
	}

	return domain, subdomain, class, nil
}
