// Package actions provides action and query execution for TLA+ simulation.
package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// DeferredPostCondition holds a post-condition guarantee to check after all
// primed assignments have been applied.
type DeferredPostCondition struct {
	// Expression is the parsed TLA+ post-condition expression.
	Expression ast.Expression

	// InstanceID identifies the instance this post-condition applies to.
	InstanceID state.InstanceID

	// SourceKey is the key of the action or query that produced this post-condition.
	SourceKey identity.Key

	// SourceName is the name of the action or query.
	SourceName string

	// SourceType is "action" or "query".
	SourceType string

	// Index is the index in the original TlaGuarantees array.
	Index int

	// OriginalExpression is the original TLA+ string (for error messages).
	OriginalExpression string
}

// DeferredSafetyRule holds a safety rule to check after all primed assignments
// have been applied. Safety rules are boolean assertions that must reference
// primed variables.
type DeferredSafetyRule struct {
	// Expression is the parsed TLA+ safety rule expression.
	Expression ast.Expression

	// InstanceID identifies the instance this safety rule applies to.
	InstanceID state.InstanceID

	// SourceKey is the key of the action that produced this safety rule.
	SourceKey identity.Key

	// SourceName is the name of the action.
	SourceName string

	// Index is the index in the original TlaSafetyRules array.
	Index int

	// OriginalExpression is the original TLA+ string (for error messages).
	OriginalExpression string
}

// ExecutionContext tracks the state of an action execution chain.
// It collects all primed assignments, post-conditions, and safety rules
// across the entire call chain, and enforces the re-entrancy constraint:
// an instance that has had primed values set by one action cannot have
// another action called on it.
type ExecutionContext struct {
	// collectedPrimed holds all primed assignments grouped by instance ID.
	// These are applied at the end of the top-level action.
	collectedPrimed map[state.InstanceID]map[string]object.Object

	// mutatedInstances tracks which instances have had actions set primed
	// values on them. These instances are "locked" for further action mutations.
	mutatedInstances map[state.InstanceID]bool

	// postConditions holds all post-conditions to check after primed
	// values are applied.
	postConditions []DeferredPostCondition

	// safetyRules holds all safety rules to check after primed values are applied.
	safetyRules []DeferredSafetyRule

	// depth tracks the current call chain depth (for debugging/limits).
	depth int
}

// NewExecutionContext creates a new top-level execution context.
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		collectedPrimed:  make(map[state.InstanceID]map[string]object.Object),
		mutatedInstances: make(map[state.InstanceID]bool),
		postConditions:   nil,
		depth:            0,
	}
}

// CanMutate checks if an action is allowed to set primed values on this instance.
// Returns false if the instance already has primed assignments from a different
// action in the chain.
func (ctx *ExecutionContext) CanMutate(instanceID state.InstanceID) bool {
	return !ctx.mutatedInstances[instanceID]
}

// RecordPrimedAssignment stores a primed assignment for deferred application.
// Returns an error if the field is "_state" (which can only be changed by transitions)
// or if the instance is locked by a different action.
func (ctx *ExecutionContext) RecordPrimedAssignment(
	instanceID state.InstanceID,
	fieldName string,
	value object.Object,
) error {
	if fieldName == "_state" {
		return fmt.Errorf("cannot set _state via action guarantee; state changes are driven by transitions")
	}

	if ctx.collectedPrimed[instanceID] == nil {
		ctx.collectedPrimed[instanceID] = make(map[string]object.Object)
	}
	ctx.collectedPrimed[instanceID][fieldName] = value
	ctx.mutatedInstances[instanceID] = true
	return nil
}

// AddPostCondition queues a post-condition for deferred checking.
func (ctx *ExecutionContext) AddPostCondition(pc DeferredPostCondition) {
	ctx.postConditions = append(ctx.postConditions, pc)
}

// GetAllPrimedAssignments returns all collected primed assignments grouped by instance.
func (ctx *ExecutionContext) GetAllPrimedAssignments() map[state.InstanceID]map[string]object.Object {
	return ctx.collectedPrimed
}

// GetAllPostConditions returns all queued post-conditions.
func (ctx *ExecutionContext) GetAllPostConditions() []DeferredPostCondition {
	return ctx.postConditions
}

// AddSafetyRule queues a safety rule for deferred checking.
func (ctx *ExecutionContext) AddSafetyRule(sr DeferredSafetyRule) {
	ctx.safetyRules = append(ctx.safetyRules, sr)
}

// GetAllSafetyRules returns all queued safety rules.
func (ctx *ExecutionContext) GetAllSafetyRules() []DeferredSafetyRule {
	return ctx.safetyRules
}

// MutatedInstanceIDs returns the set of instance IDs that have been mutated.
func (ctx *ExecutionContext) MutatedInstanceIDs() []state.InstanceID {
	ids := make([]state.InstanceID, 0, len(ctx.mutatedInstances))
	for id := range ctx.mutatedInstances {
		ids = append(ids, id)
	}
	return ids
}

// Depth returns the current call chain depth.
func (ctx *ExecutionContext) Depth() int {
	return ctx.depth
}

// IncrementDepth increments the call depth. Returns an error if the depth
// exceeds the maximum allowed (to prevent infinite recursion).
func (ctx *ExecutionContext) IncrementDepth() error {
	ctx.depth++
	if ctx.depth > 100 {
		return fmt.Errorf("action call chain depth exceeded maximum of 100")
	}
	return nil
}

// DecrementDepth decrements the call depth.
func (ctx *ExecutionContext) DecrementDepth() {
	ctx.depth--
}
