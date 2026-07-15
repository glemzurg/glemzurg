package convert

import (
	"fmt"
	"maps"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

// LowerModel walks the entire model tree, parsing and lowering every ExpressionSpec
// that has a TLA+ specification string into a logic_expression.Expression.
// It returns the first error encountered, leaving the model partially populated.
func LowerModel(model *core.Model) error {
	// Build model-level lookup maps for global functions and named sets.
	globalFunctions := BuildGlobalFunctionMap(model)
	namedSets := BuildNamedSetMap(model)

	// Build cross-class action lookup (AllActions) across the entire model.
	allActions := BuildAllActionsMap(model)

	// 1. Lower model-level invariants (no class context).
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	for i := range model.Invariants {
		if err := lowerLogicSpec(&model.Invariants[i].Spec, modelCtx); err != nil {
			return fmt.Errorf("model invariant %d: %w", i, err)
		}
	}

	// 2. Lower global functions.
	for gfKey, gf := range model.GlobalFunctions {
		params := make(map[string]bool, len(gf.Parameters))
		for _, p := range gf.Parameters {
			params[p] = true
		}
		gfCtx := &LowerContext{
			GlobalFunctions: globalFunctions,
			NamedSets:       namedSets,
			AllActions:      allActions,
			Parameters:      params,
		}
		if err := lowerLogicSpec(&gf.Logic.Spec, gfCtx); err != nil {
			return fmt.Errorf("global function %q: %w", gfKey.String(), err)
		}
		model.GlobalFunctions[gfKey] = gf
	}

	// 3. Lower named sets.
	for nsKey, ns := range model.NamedSets {
		if err := lowerLogicSpec(&ns.Spec, modelCtx); err != nil {
			return fmt.Errorf("named set %q: %w", nsKey.String(), err)
		}
		model.NamedSets[nsKey] = ns
	}

	allAssociations := model.GetClassAssociations()

	// 4. Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				if err := lowerClass(&class, globalFunctions, namedSets, allActions, allAssociations, subdomain.Classes); err != nil {
					return fmt.Errorf("class %q: %w", cKey.String(), err)
				}
				subdomain.Classes[cKey] = class
			}
			domain.Subdomains[sKey] = subdomain
		}
		model.Domains[dKey] = domain
	}

	return nil
}

// lowerClass populates all ExpressionSpec.Expression fields within a class.
func lowerClass(
	class *model_class.Class,
	globalFunctions map[string]identity.Key,
	namedSets map[string]identity.Key,
	allActions map[string]identity.Key,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) error {
	classCtx := NewClassLowerContext(class, globalFunctions, namedSets, allActions, associations, classes)

	// Class invariants.
	for i := range class.Invariants {
		if err := lowerLogicSpec(&class.Invariants[i].Spec, classCtx); err != nil {
			return fmt.Errorf("class invariant %d: %w", i, err)
		}
	}

	// Attributes: derivation policy and invariants.
	for i := range class.Attributes {
		if err := lowerAttribute(&class.Attributes[i], classCtx); err != nil {
			return fmt.Errorf("attribute %q: %w", class.Attributes[i].Key.String(), err)
		}
	}

	// Guards.
	for gKey, guard := range class.Guards {
		if err := lowerLogicSpec(&guard.Logic.Spec, classCtx); err != nil {
			return fmt.Errorf("guard %q: %w", gKey.String(), err)
		}
		class.Guards[gKey] = guard
	}

	// Actions.
	for actKey, action := range class.Actions {
		if err := lowerAction(&action, classCtx); err != nil {
			return fmt.Errorf("action %q: %w", actKey.String(), err)
		}
		class.Actions[actKey] = action
	}

	// Queries.
	for qKey, query := range class.Queries {
		if err := lowerQuery(&query, classCtx); err != nil {
			return fmt.Errorf("query %q: %w", qKey.String(), err)
		}
		class.Queries[qKey] = query
	}

	return nil
}

func lowerAttribute(attr *model_class.Attribute, baseCtx *LowerContext) error {
	if attr.DerivationPolicy != nil {
		if err := lowerLogicSpec(&attr.DerivationPolicy.Spec, baseCtx); err != nil {
			return fmt.Errorf("derivation policy: %w", err)
		}
	}
	for i := range attr.Invariants {
		if err := lowerLogicSpec(&attr.Invariants[i].Spec, baseCtx); err != nil {
			return fmt.Errorf("invariant %d: %w", i, err)
		}
	}
	return nil
}

func lowerAction(action *model_state.Action, baseCtx *LowerContext) error {
	// Actions have parameters that become local vars.
	ctx := ContextWithParameters(baseCtx, action.Parameters)

	for i := range action.Requires {
		if err := lowerLogicSpec(&action.Requires[i].Spec, ctx); err != nil {
			return fmt.Errorf("require %d: %w", i, err)
		}
	}
	for i := range action.Guarantees {
		guar := &action.Guarantees[i]
		guarCtx := lowerContextWithPriorLetGuarantees(ctx, action.Guarantees[:i])
		if model_logic.IsAssociationClassReify(*guar) {
			if err := lowerAssociationClassReifyGuarantee(guar, guarCtx); err != nil {
				return fmt.Errorf("guarantee %d: %w", i, err)
			}
			continue
		}
		if err := lowerLogicSpec(&guar.Spec, guarCtx); err != nil {
			return fmt.Errorf("guarantee %d: %w", i, err)
		}
		if guar.Type == model_logic.LogicTypeDestroy {
			if err := lowerDestroyGuaranteeEvent(guar, guarCtx); err != nil {
				return fmt.Errorf("guarantee %d destroy_event: %w", i, err)
			}
		}
	}
	for i := range action.SafetyRules {
		if err := lowerLogicSpec(&action.SafetyRules[i].Spec, ctx); err != nil {
			return fmt.Errorf("safety rule %d: %w", i, err)
		}
	}
	for i := range action.Parameters {
		for j := range action.Parameters[i].Invariants {
			if err := lowerLogicSpec(&action.Parameters[i].Invariants[j].Spec, ctx); err != nil {
				return fmt.Errorf("parameter %q invariant %d: %w", action.Parameters[i].Name, j, err)
			}
		}
		if err := lowerParameterSimulation(&action.Parameters[i], ctx); err != nil {
			return err
		}
	}
	return nil
}

// lowerParameterSimulation lowers each rule's requires and specification.
func lowerParameterSimulation(param *model_state.Parameter, ctx *LowerContext) error {
	if param.Simulation == nil {
		return nil
	}
	for r := range param.Simulation.Rules {
		rule := &param.Simulation.Rules[r]
		for j := range rule.Requires {
			if err := lowerLogicSpec(&rule.Requires[j].Spec, ctx); err != nil {
				return fmt.Errorf("parameter %q simulation rule %d require %d: %w", param.Name, r, j, err)
			}
		}
		if rule.Specification != nil {
			if err := lowerLogicSpec(&rule.Specification.Spec, ctx); err != nil {
				return fmt.Errorf("parameter %q simulation rule %d specification: %w", param.Name, r, err)
			}
		}
	}
	return nil
}

func lowerQuery(query *model_state.Query, baseCtx *LowerContext) error {
	ctx := ContextWithParameters(baseCtx, query.Parameters)

	for i := range query.Requires {
		if err := lowerLogicSpec(&query.Requires[i].Spec, ctx); err != nil {
			return fmt.Errorf("require %d: %w", i, err)
		}
	}
	for i := range query.Guarantees {
		if err := lowerLogicSpec(&query.Guarantees[i].Spec, ctx); err != nil {
			return fmt.Errorf("guarantee %d: %w", i, err)
		}
	}
	for i := range query.Parameters {
		for j := range query.Parameters[i].Invariants {
			if err := lowerLogicSpec(&query.Parameters[i].Invariants[j].Spec, ctx); err != nil {
				return fmt.Errorf("parameter %q invariant %d: %w", query.Parameters[i].Name, j, err)
			}
		}
	}
	return nil
}

// contextWithParameters creates a child context with action/query parameters added.
func ContextWithParameters(base *LowerContext, params []model_state.Parameter) *LowerContext {
	if len(params) == 0 {
		return base
	}
	child := *base
	child.Parameters = make(map[string]bool)
	if base.Parameters != nil {
		maps.Copy(child.Parameters, base.Parameters)
	}
	for _, p := range params {
		child.Parameters[p.Name] = true
	}
	return &child
}

func lowerContextWithPriorLetGuarantees(base *LowerContext, prior []model_logic.Logic) *LowerContext {
	ctx := base
	for i := range prior {
		if prior[i].Type == model_logic.LogicTypeLet && prior[i].Target != "" {
			ctx = withLocalVar(ctx, prior[i].Target)
		}
	}
	return ctx
}

func lowerDestroyGuaranteeEvent(guar *model_logic.Logic, ctx *LowerContext) error {
	deleteCtx := ctx
	if sf, ok := guar.Spec.Expression.(*me.SetFilter); ok {
		deleteCtx = withLocalVar(ctx, sf.Variable)
	}
	if boundVar := destroyEventBoundVariable(guar.DestroyEventSpec.Specification); boundVar != "" {
		deleteCtx = withLocalVar(deleteCtx, boundVar)
	}
	return lowerLogicSpec(&guar.DestroyEventSpec, deleteCtx)
}

// lowerAssociationClassReifyGuarantee lowers endpoint_selector then Spec.
// When endpoint_selector is a set-map (LET-like domain), its binder is in scope for Spec.
func lowerAssociationClassReifyGuarantee(guar *model_logic.Logic, ctx *LowerContext) error {
	if err := lowerLogicSpec(&guar.EndpointSelectorSpec, ctx); err != nil {
		return fmt.Errorf("endpoint_selector: %w", err)
	}
	reifyCtx := ctx
	if setMap, ok := guar.EndpointSelectorSpec.Expression.(*me.SetMap); ok && setMap.Variable != "" {
		reifyCtx = withLocalVar(ctx, setMap.Variable)
	}
	if err := lowerLogicSpec(&guar.Spec, reifyCtx); err != nil {
		return fmt.Errorf("creation specification: %w", err)
	}
	return nil
}

// destroyEventBoundVariable returns the first destroy_event call argument name.
// That identifier is a bound variable for lowering only; the simulator skips it at runtime.
func destroyEventBoundVariable(specification string) string {
	if specification == "" {
		return ""
	}
	astExpr, err := parser.ParseExpression(specification)
	if err != nil {
		return ""
	}
	call, ok := astExpr.(*ast.FunctionCall)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	id, ok := call.Args[0].(*ast.Identifier)
	if !ok {
		return ""
	}
	return id.Value
}

// lowerLogicSpec parses and lowers a single ExpressionSpec if it has a TLA+ specification
// and hasn't been lowered yet.
func lowerLogicSpec(spec *logic_spec.ExpressionSpec, ctx *LowerContext) error {
	// Skip if not TLA+, no specification text, or already lowered.
	if spec.Notation != "tla_plus" || spec.Specification == "" || spec.Expression != nil {
		return nil
	}

	// Parse the TLA+ specification string to AST.
	astExpr, err := parser.ParseExpression(spec.Specification)
	if err != nil {
		return fmt.Errorf("parse %q: %w", spec.Specification, err)
	}

	// Lower AST to logic_expression.
	meExpr, err := Lower(astExpr, ctx)
	if err != nil {
		return fmt.Errorf("lower %q: %w", spec.Specification, err)
	}

	spec.Expression = meExpr
	return nil
}

// --- Map builders ---

func BuildGlobalFunctionMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key, len(model.GlobalFunctions))
	for _, gf := range model.GlobalFunctions {
		m[gf.Name] = gf.Key
	}
	return m
}

// SubdomainClassMaps holds per-subdomain class and association indexes for lowering.
type SubdomainClassMaps struct {
	Associations map[identity.Key]model_class.Association
	Classes      map[identity.Key]model_class.Class
}

func BuildNamedSetMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key, len(model.NamedSets))
	for _, ns := range model.NamedSets {
		m[ns.Name] = ns.Key
	}
	return m
}

// BuildClassNameMap maps class display names to identity keys across the whole model.
func BuildClassNameMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				m[class.Name] = class.Key
			}
		}
	}
	return m
}

// BuildClassKeyToNameMap maps class identity keys to display names across the whole model.
func BuildClassKeyToNameMap(model *core.Model) map[identity.Key]string {
	m := make(map[identity.Key]string)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				m[class.Key] = class.Name
			}
		}
	}
	return m
}

func BuildAllActionsMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key)
	for _, domain := range model.Domains {
		dName := domain.Name
		for _, subdomain := range domain.Subdomains {
			sName := subdomain.Name
			for _, class := range subdomain.Classes {
				cName := class.Name
				for _, action := range class.Actions {
					// Register with multiple scoping levels for lookup flexibility.
					// Full: Domain!Subdomain!Class!Action
					m[dName+"!"+sName+"!"+cName+"!"+action.Name] = action.Key
					// Partial: Subdomain!Class!Action
					m[sName+"!"+cName+"!"+action.Name] = action.Key
					// Partial: Class!Action
					m[cName+"!"+action.Name] = action.Key
				}
				for _, query := range class.Queries {
					m[dName+"!"+sName+"!"+cName+"!"+query.Name] = query.Key
					m[sName+"!"+cName+"!"+query.Name] = query.Key
					m[cName+"!"+query.Name] = query.Key
				}
			}
		}
	}
	return m
}

func BuildAttributeNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Attributes))
	for _, attr := range class.Attributes {
		m[attr.Name] = attr.Key
	}
	return m
}

func BuildActionNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Actions))
	for _, action := range class.Actions {
		m[action.Name] = action.Key
	}
	return m
}

func BuildQueryNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Queries))
	for _, query := range class.Queries {
		m[query.Name] = query.Key
	}
	return m
}

// NewClassLowerContext builds a LowerContext for expressions scoped to one class.
func NewClassLowerContext(
	class *model_class.Class,
	globalFunctions map[string]identity.Key,
	namedSets map[string]identity.Key,
	allActions map[string]identity.Key,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) *LowerContext {
	return &LowerContext{
		ClassKey:         class.Key,
		AttributeNames:   BuildAttributeNameMap(class),
		ActionNames:      BuildActionNameMap(class),
		QueryNames:       BuildQueryNameMap(class),
		AssociationNames: BuildOutgoingAssociationFieldNameMap(class.Key, associations),
		SystemEventNames: BuildSystemEventNameMap(class),
		PeerEventNames:   BuildPeerEventNameMap(class.Key, associations, classes),
		GlobalFunctions:  globalFunctions,
		NamedSets:        namedSets,
		ClassNames:       BuildClassNamesForLower(classes),
		AllActions:       allActions,
	}
}

// BuildClassNamesForLower maps TLA-friendly class names to keys for quantifier domains.
// Both the display name and the space-stripped form resolve (e.g. "Account Definition" and AccountDefinition).
func BuildClassNamesForLower(classes map[identity.Key]model_class.Class) map[string]identity.Key {
	if len(classes) == 0 {
		return nil
	}
	m := make(map[string]identity.Key, len(classes)*2)
	for key, class := range classes {
		if class.Name != "" {
			m[class.Name] = key
			compact := strings.ReplaceAll(class.Name, " ", "")
			if compact != class.Name {
				m[compact] = key
			}
		}
		if key.SubKey != "" {
			// PascalCase of subkey is not automatic; subkey alone is rarely used in TLA.
			_ = key
		}
	}
	return m
}

// BuildOutgoingAssociationFieldNameMap maps TLA field names to association keys for from-class links.
func BuildOutgoingAssociationFieldNameMap(classKey identity.Key, associations map[identity.Key]model_class.Association) map[string]identity.Key {
	if len(associations) == 0 {
		return nil
	}
	m := make(map[string]identity.Key)
	for _, assoc := range associations {
		if assoc.FromClassKey != classKey {
			continue
		}
		m[model_class.AssociationTLAFieldName(assoc.Name)] = assoc.Key
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// BuildPeerEventNameMap maps peer-class event names reachable via outgoing associations
// and via object-of action parameters (for peer-domain event set-maps).
func BuildPeerEventNameMap(
	fromClassKey identity.Key,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) map[string]identity.Key {
	if len(classes) == 0 {
		return nil
	}
	m := make(map[string]identity.Key)
	addPeerClassEvents := func(peerClass model_class.Class) {
		for _, event := range peerClass.Events {
			m[event.Name] = event.Key
			if model_state.IsSystemCreationEvent(event.Name) || model_state.IsSystemFinalEvent(event.Name) {
				m[model_state.SystemEventTLAName(event.Name)] = event.Key
			}
		}
	}
	for _, assoc := range associations {
		if assoc.FromClassKey != fromClassKey {
			continue
		}
		if peerClass, ok := classes[assoc.ToClassKey]; ok {
			addPeerClassEvents(peerClass)
		}
	}
	// Peer-domain events: actions may fire events on instances from object-of parameters.
	if fromClass, ok := classes[fromClassKey]; ok {
		for _, action := range fromClass.Actions {
			for _, param := range action.Parameters {
				for _, classKey := range objectOfClassKeysInDataType(param.DataType, classes) {
					if peerClass, ok := classes[classKey]; ok {
						addPeerClassEvents(peerClass)
					}
				}
			}
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// objectOfClassKeysInDataType collects class keys referenced by object-of atomics in a data type tree.
func objectOfClassKeysInDataType(dt *model_data_type.DataType, classes map[identity.Key]model_class.Class) []identity.Key {
	if dt == nil {
		return nil
	}
	var keys []identity.Key
	if dt.Atomic != nil && dt.Atomic.ObjectClassKey != nil {
		want := *dt.Atomic.ObjectClassKey
		for ck, c := range classes {
			if ck.SubKey == want || ck.String() == want || c.Name == want || identity.NormalizeSubKey(c.Name) == want {
				keys = append(keys, ck)
			}
		}
	}
	if dt.ElementDataType != nil {
		keys = append(keys, objectOfClassKeysInDataType(dt.ElementDataType, classes)...)
	}
	for i := range dt.RecordFields {
		keys = append(keys, objectOfClassKeysInDataType(dt.RecordFields[i].FieldDataType, classes)...)
	}
	return keys
}

// BuildPeerEventRaiseNameMap maps peer-class event keys to their declared names.
func BuildPeerEventRaiseNameMap(
	fromClassKey identity.Key,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) map[identity.Key]string {
	nameMap := BuildPeerEventNameMap(fromClassKey, associations, classes)
	if len(nameMap) == 0 {
		return nil
	}
	out := make(map[identity.Key]string, len(nameMap))
	for name, key := range nameMap {
		out[key] = name
	}
	return out
}

// BuildSystemEventNameMap maps system event spellings to event keys declared on the class.
// Both ASCII (_new) and canonical TLA («new») forms resolve to the same key.
func BuildSystemEventNameMap(class *model_class.Class) map[string]identity.Key {
	if len(class.Events) == 0 {
		return nil
	}
	m := make(map[string]identity.Key)
	for _, event := range class.Events {
		if model_state.IsSystemCreationEvent(event.Name) || model_state.IsSystemFinalEvent(event.Name) {
			m[event.Name] = event.Key
			m[model_state.SystemEventTLAName(event.Name)] = event.Key
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// BuildSystemEventRaiseNameMap maps event keys to canonical TLA spellings («new», «destroy»).
func BuildSystemEventRaiseNameMap(class *model_class.Class) map[identity.Key]string {
	if len(class.Events) == 0 {
		return nil
	}
	m := make(map[identity.Key]string)
	for _, event := range class.Events {
		if model_state.IsSystemCreationEvent(event.Name) || model_state.IsSystemFinalEvent(event.Name) {
			m[event.Key] = model_state.SystemEventTLAName(event.Name)
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
