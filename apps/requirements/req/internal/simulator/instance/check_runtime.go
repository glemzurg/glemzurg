package instance

import (
	"fmt"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

const expressionReturnedNil = "expression returned nil"

// DeferredAssertion is a boolean expression checked against live state after mutation.
// Used for action/query postconditions and action safety rules.
type DeferredAssertion struct {
	Expression         me.Expression
	InstanceID         ID
	SourceKey          identity.Key
	SourceName         string
	Kind               AssertionKind
	Index              int
	OriginalExpression string
	// LetBindings are added to evaluation scope for safety rules (postconditions ignore).
	LetBindings map[string]object.Object
}

// AssertionKind selects how a failed deferred assertion is classified.
type AssertionKind int

const (
	// AssertionPostconditionAction is an action guarantee postcondition.
	AssertionPostconditionAction AssertionKind = iota
	// AssertionPostconditionQuery is a query guarantee postcondition.
	AssertionPostconditionQuery
	// AssertionSafetyRule is an action safety rule.
	AssertionSafetyRule
)

// AssessedFailure is one failed requires / parameter-invariant assessment
// (expression already evaluated by the caller or by instance helpers).
type AssessedFailure struct {
	Index   int
	Spec    string
	Message string
}

// CheckDeferredAssertions re-evaluates deferred boolean assertions on current instances.
// bindings builds self-scope for each target instance.
func (s *State) CheckDeferredAssertions(assertions []DeferredAssertion, bindings ExpressionBindings) ViolationErrors {
	if s == nil || bindings == nil || len(assertions) == 0 {
		return nil
	}
	var violations ViolationErrors
	for _, a := range assertions {
		target := s.GetInstance(a.InstanceID)
		if target == nil {
			continue
		}
		var evalBindings *evaluator.Bindings
		if len(a.LetBindings) > 0 {
			evalBindings = bindings.BuildForInstanceWithVariables(target, a.LetBindings)
		} else {
			evalBindings = bindings.BuildForInstance(target)
		}
		msg := evalBooleanFailure(a.Expression, evalBindings)
		if msg == "" {
			continue
		}
		violations = append(violations, deferredAssertionViolation(a, msg))
	}
	return violations
}

func deferredAssertionViolation(a DeferredAssertion, message string) *ViolationError {
	switch a.Kind {
	case AssertionPostconditionAction:
		return newActionGuaranteeViolation(a.SourceKey, a.SourceName, a.Index, a.OriginalExpression, a.InstanceID, message)
	case AssertionPostconditionQuery:
		return newQueryGuaranteeViolation(a.SourceKey, a.SourceName, a.Index, a.OriginalExpression, a.InstanceID, message)
	case AssertionSafetyRule:
		return newSafetyRuleViolation(a.SourceKey, a.SourceName, a.Index, a.OriginalExpression, a.InstanceID, message)
	}
	return newActionGuaranteeViolation(a.SourceKey, a.SourceName, a.Index, a.OriginalExpression, a.InstanceID, message)
}

func evalBooleanFailure(expr me.Expression, bindings *evaluator.Bindings) string {
	if expr == nil || bindings == nil {
		return ""
	}
	result := evaluator.Eval(expr, bindings)
	if result.IsError() {
		return fmt.Sprintf("evaluation error: %s", result.Error.Inspect())
	}
	if isTrueBooleanObject(result.Value) {
		return ""
	}
	if result.Value == nil {
		return expressionReturnedNil
	}
	return fmt.Sprintf("expression returned %s", result.Value.Inspect())
}

func isTrueBooleanObject(obj object.Object) bool {
	if obj == nil {
		return false
	}
	b, ok := obj.(*object.Boolean)
	return ok && b.Value()
}

// PeerEventUnavailableInput describes a peer event that could not be delivered.
// State assembles the message from live peer instance data unless Detail is set.
type PeerEventUnavailableInput struct {
	OwnerClassKey   identity.Key
	OwnerInstanceID ID
	AssociationName string
	PeerClassKey    identity.Key
	PeerClassName   string
	PeerInstanceID  ID
	EventKey        identity.Key
	EventName       string
	// Detail, when non-empty, is used as the full violation message (e.g. parameter binding failure).
	Detail string
}

// CheckPeerEventUnavailable builds a peer-event-unavailable violation using live peer state.
func (s *State) CheckPeerEventUnavailable(in PeerEventUnavailableInput) *ViolationError {
	msg := in.Detail
	if msg == "" {
		stateName := ""
		if s != nil && in.PeerInstanceID != 0 {
			if inst := s.GetInstance(in.PeerInstanceID); inst != nil {
				stateName = instanceStateName(inst)
			}
		}
		msg = fmt.Sprintf(
			"association %q sent event %s to class %s",
			in.AssociationName, in.EventName, in.PeerClassName,
		)
		if in.PeerInstanceID != 0 {
			if stateName != "" {
				msg = fmt.Sprintf(
					"%s but instance %d has no %s transition from state %s",
					msg, in.PeerInstanceID, in.EventName, stateName,
				)
			} else {
				msg = fmt.Sprintf("%s but instance %d is not available", msg, in.PeerInstanceID)
			}
		} else {
			msg = fmt.Sprintf("%s but the class has no %s creation transition", msg, in.EventName)
		}
	}
	return newPeerEventUnavailableViolation(PeerEventUnavailableParams{
		OwnerClassKey:   in.OwnerClassKey,
		OwnerInstanceID: in.OwnerInstanceID,
		AssociationName: in.AssociationName,
		PeerClassKey:    in.PeerClassKey,
		PeerInstanceID:  in.PeerInstanceID,
		EventKey:        in.EventKey,
		EventName:       in.EventName,
		Message:         msg,
	})
}

func instanceStateName(inst *Instance) string {
	if inst == nil {
		return ""
	}
	stateAttr := inst.GetAttribute("_state")
	if stateAttr == nil {
		return ""
	}
	if strObj, ok := stateAttr.(*object.String); ok {
		return strObj.Value()
	}
	return ""
}

// PackageActionRequireFailures converts assessed action requires into violations.
func PackageActionRequireFailures(actionKey identity.Key, actionName string, failures []AssessedFailure, instanceID ID) ViolationErrors {
	var violations ViolationErrors
	for _, f := range failures {
		violations = append(violations, newActionRequiresViolation(
			actionKey, actionName, f.Index, f.Spec, instanceID, f.Message,
		))
	}
	return violations
}

// PackageParameterInvariantFailures converts assessed parameter invariants into violations.
func PackageParameterInvariantFailures(ownerKey identity.Key, ownerName string, failures []AssessedFailure, instanceID ID) ViolationErrors {
	var violations ViolationErrors
	for _, f := range failures {
		violations = append(violations, newParameterInvariantViolation(
			ownerKey, ownerName, f.Index, f.Spec, instanceID, f.Message,
		))
	}
	return violations
}

// SurfaceMemberKind classifies a top-level surface member for availability checks.
type SurfaceMemberKind int

const (
	// SurfaceMemberDerived is an external derived attribute.
	SurfaceMemberDerived SurfaceMemberKind = iota
	// SurfaceMemberQuery is an external query.
	SurfaceMemberQuery
)

// CheckSurfaceMemberAccess reports a violation when a surface member depends on out-of-scope classes.
// When the member is available, returns nil.
func CheckSurfaceMemberAccess(
	sch *schema.Schema,
	kind SurfaceMemberKind,
	memberKey identity.Key,
	classKey identity.Key,
	instanceID ID,
	memberName string,
) *ViolationError {
	if sch == nil {
		return nil
	}
	var reason string
	switch kind {
	case SurfaceMemberDerived:
		unavail, ok := sch.SurfaceUnavailableDerived(memberKey)
		if !ok {
			return nil
		}
		reason = unavail.Reason()
	case SurfaceMemberQuery:
		unavail, ok := sch.SurfaceUnavailableQuery(memberKey)
		if !ok {
			return nil
		}
		reason = unavail.Reason()
	default:
		return nil
	}
	return newSurfaceOutOfScopeViolation(classKey, instanceID, memberName, reason)
}

// CheckStateMachineCompleteness reports in-scope classes whose SM omits system _new.
func CheckStateMachineCompleteness(sch *schema.Schema) ViolationErrors {
	if sch == nil {
		return nil
	}
	var classes []*schema.ClassSimInfo
	sch.EachInScopeClassSim(func(classInfo *schema.ClassSimInfo) {
		classes = append(classes, classInfo)
	})
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Class.Name < classes[j].Class.Name
	})

	var violations ViolationErrors
	for _, classInfo := range classes {
		if !classHasStateMachine(classInfo.Class) || classStateMachineHasNewEvent(classInfo.Class) {
			continue
		}
		violations = append(violations, newStateMachineIncompleteViolation(
			classInfo.ClassKey,
			classInfo.Class.Name,
		))
	}
	return violations
}

func classHasStateMachine(class model_class.Class) bool {
	return len(class.States) > 0 || len(class.Transitions) > 0
}

func classStateMachineHasNewEvent(class model_class.Class) bool {
	for _, event := range class.Events {
		if model_state.IsSystemCreationEvent(event.Name) {
			return true
		}
	}
	return false
}
