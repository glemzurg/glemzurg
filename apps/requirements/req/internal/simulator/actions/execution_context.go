// Package actions provides action and query execution for TLA+ simulation.
package actions

import (
	"fmt"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// DeferredPostCondition holds a post-condition guarantee to check after all
// primed assignments have been applied.
type DeferredPostCondition struct {
	// Expression is the lowered post-condition expression.
	Expression me.Expression

	// InstanceID identifies the instance this post-condition applies to.
	InstanceID state.InstanceID

	// SourceKey is the key of the action or query that produced this post-condition.
	SourceKey identity.Key

	// SourceName is the name of the action or query.
	SourceName string

	// SourceType is "action" or "query".
	SourceType string

	// Index is the index in the original Guarantees array.
	Index int

	// OriginalExpression is the original TLA+ string (for error messages).
	OriginalExpression string
}

// DeferredPeerCreation creates a peer and/or association-class row for set-add / bulk-create.
type DeferredPeerCreation struct {
	FromInstanceID state.InstanceID
	AssocKey       identity.Key
	ToClassKey     identity.Key
	// ToInstanceID, when set, links an existing to-endpoint (no new to-class create).
	ToInstanceID *state.InstanceID
	// Params are creation-event parameters for the to-class (plain set-add) or the
	// association class when ToInstanceID is set / AC materialization carries them.
	Params map[string]object.Object
}

// DeferredPeerUpdate fires a peer-class event on an existing association link target.
type DeferredPeerUpdate struct {
	OwnerInstanceID state.InstanceID
	AssocKey        identity.Key
	PeerInstanceID  state.InstanceID
	ToClassKey      identity.Key
	EventKey        identity.Key
	EventName       string
	Params          map[string]object.Object
	RemovesLink     bool
}

// DeferredSafetyRule holds a safety rule to check after all primed assignments
// have been applied. Safety rules are boolean assertions that must reference
// primed variables.
type DeferredSafetyRule struct {
	// Expression is the lowered safety rule expression.
	Expression me.Expression

	// InstanceID identifies the instance this safety rule applies to.
	InstanceID state.InstanceID

	// SourceKey is the key of the action that produced this safety rule.
	SourceKey identity.Key

	// SourceName is the name of the action.
	SourceName string

	// Index is the index in the original SafetyRules array.
	Index int

	// OriginalExpression is the original TLA+ string (for error messages).
	OriginalExpression string

	// LetBindings contains let variable values computed before this safety rule
	// in the same list. These are added to the evaluation bindings when the
	// safety rule is evaluated after primed assignments are applied.
	LetBindings map[string]object.Object
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

	// mutatedInstances tracks which instances have had primed values recorded.
	mutatedInstances map[state.InstanceID]bool

	// instanceActionOwner maps instance ID to the action key that first claimed
	// mutation rights in this chain (blocks a different chained action).
	instanceActionOwner map[state.InstanceID]identity.Key

	// postConditions holds all post-conditions to check after primed
	// values are applied.
	postConditions []DeferredPostCondition

	// safetyRules holds all safety rules to check after primed values are applied.
	safetyRules []DeferredSafetyRule

	// peerCreations materialize association set-add guarantees: create the to-class
	// via its creation transition and link back to the owning instance.
	peerCreations []DeferredPeerCreation

	// peerUpdates fire peer-class events for association set-map guarantees.
	peerUpdates []DeferredPeerUpdate

	// peerTransitions records peer-class transitions for trace output.
	peerTransitions []PeerTransitionRecord

	// peerViolations records association peer events the target class could not accept.
	peerViolations invariants.ViolationErrors

	// depth tracks the current call chain depth (for debugging/limits).
	depth int

	// requiresViolations holds precondition failures that block guarantee application.
	requiresViolations invariants.ViolationErrors

	// associationRemovedPeers records peers dropped by association state_change guarantees.
	associationRemovedPeers map[associationRemovalKey][]state.InstanceID

	// associationDestroyCandidates tracks removed peers targeted for peer _destroy. While targeted,
	// association links stay put; unavailable _destroy records PeerEventUnavailable and leaves links.
	associationDestroyCandidates map[associationRemovalKey]map[state.InstanceID]bool
}

type associationRemovalKey struct {
	OwnerInstanceID state.InstanceID
	AssocKey        identity.Key
}

// NewExecutionContext creates a new top-level execution context.
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		collectedPrimed:     make(map[state.InstanceID]map[string]object.Object),
		mutatedInstances:    make(map[state.InstanceID]bool),
		instanceActionOwner: make(map[state.InstanceID]identity.Key),
		postConditions:      nil,
		depth:               0,
	}
}

// ClaimInstanceForAction records that actionKey may mutate instanceID in this chain.
// Returns false if another action already claimed the instance.
func (ctx *ExecutionContext) ClaimInstanceForAction(instanceID state.InstanceID, actionKey identity.Key) bool {
	if owner, ok := ctx.instanceActionOwner[instanceID]; ok && owner != actionKey {
		return false
	}
	ctx.instanceActionOwner[instanceID] = actionKey
	return true
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

// AddPeerCreation queues a peer instance to create and link for an association set-add guarantee.
func (ctx *ExecutionContext) AddPeerCreation(pc DeferredPeerCreation) {
	ctx.peerCreations = append(ctx.peerCreations, pc)
}

// GetPeerCreations returns queued peer creations.
func (ctx *ExecutionContext) GetPeerCreations() []DeferredPeerCreation {
	return ctx.peerCreations
}

// AddPeerUpdate queues a peer-class event for an association set-map guarantee.
func (ctx *ExecutionContext) AddPeerUpdate(pu DeferredPeerUpdate) {
	ctx.peerUpdates = append(ctx.peerUpdates, pu)
}

// GetPeerUpdates returns queued peer updates.
func (ctx *ExecutionContext) GetPeerUpdates() []DeferredPeerUpdate {
	return ctx.peerUpdates
}

// AddPeerTransition records a peer-class transition for cascaded trace output.
func (ctx *ExecutionContext) AddPeerTransition(rec PeerTransitionRecord) {
	ctx.peerTransitions = append(ctx.peerTransitions, rec)
}

// GetPeerTransitions returns recorded peer-class transitions.
func (ctx *ExecutionContext) GetPeerTransitions() []PeerTransitionRecord {
	return ctx.peerTransitions
}

// AddPeerViolation records an association peer event the target class could not accept.
func (ctx *ExecutionContext) AddPeerViolation(v *invariants.ViolationError) {
	if v == nil {
		return
	}
	ctx.peerViolations = append(ctx.peerViolations, v)
}

// GetPeerViolations returns association peer event violations.
func (ctx *ExecutionContext) GetPeerViolations() invariants.ViolationErrors {
	return ctx.peerViolations
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

// SetRequiresViolations records precondition failures for the current action chain.
func (ctx *ExecutionContext) SetRequiresViolations(violations invariants.ViolationErrors) {
	ctx.requiresViolations = violations
}

// RequiresViolations returns precondition failures recorded during execution.
func (ctx *ExecutionContext) RequiresViolations() invariants.ViolationErrors {
	return ctx.requiresViolations
}

// SetAssociationRemovedPeers records peers removed from an association by state_change.
func (ctx *ExecutionContext) SetAssociationRemovedPeers(
	ownerInstanceID state.InstanceID,
	assocKey identity.Key,
	peerIDs []state.InstanceID,
) {
	if len(peerIDs) == 0 {
		return
	}
	key := associationRemovalKey{OwnerInstanceID: ownerInstanceID, AssocKey: assocKey}
	if ctx.associationRemovedPeers == nil {
		ctx.associationRemovedPeers = make(map[associationRemovalKey][]state.InstanceID)
	}
	ctx.associationRemovedPeers[key] = append(ctx.associationRemovedPeers[key], peerIDs...)
}

// AssociationRemovedPeers returns peers recorded as removed for one association.
func (ctx *ExecutionContext) AssociationRemovedPeers(
	ownerInstanceID state.InstanceID,
	assocKey identity.Key,
) []state.InstanceID {
	key := associationRemovalKey{OwnerInstanceID: ownerInstanceID, AssocKey: assocKey}
	return ctx.associationRemovedPeers[key]
}

// MarkAssociationDestroyCandidate records a removed peer targeted by a destroy guarantee.
func (ctx *ExecutionContext) MarkAssociationDestroyCandidate(
	ownerInstanceID state.InstanceID,
	assocKey identity.Key,
	peerID state.InstanceID,
) {
	key := associationRemovalKey{OwnerInstanceID: ownerInstanceID, AssocKey: assocKey}
	if ctx.associationDestroyCandidates == nil {
		ctx.associationDestroyCandidates = make(map[associationRemovalKey]map[state.InstanceID]bool)
	}
	if ctx.associationDestroyCandidates[key] == nil {
		ctx.associationDestroyCandidates[key] = make(map[state.InstanceID]bool)
	}
	ctx.associationDestroyCandidates[key][peerID] = true
}

// AssociationDestroyCandidate reports whether a removed peer was selected for peer _destroy.
func (ctx *ExecutionContext) AssociationDestroyCandidate(key associationRemovalKey, peerID state.InstanceID) bool {
	if ctx.associationDestroyCandidates == nil {
		return false
	}
	return ctx.associationDestroyCandidates[key][peerID]
}

// associationRemovedPeerSets returns all association removal batches from state_change.
func (ctx *ExecutionContext) associationRemovedPeerSets() map[associationRemovalKey][]state.InstanceID {
	if len(ctx.associationRemovedPeers) == 0 {
		return nil
	}
	return ctx.associationRemovedPeers
}
