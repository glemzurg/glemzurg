package surface

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// MemberKind distinguishes surface members that pass association data.
type MemberKind string

const (
	// MemberDerived is a derived attribute (value from association or related data).
	MemberDerived MemberKind = "derived"
	// MemberQuery is a class query (parameterized read, same dependency rules as derived).
	MemberQuery MemberKind = "query"
)

// UnavailableMember records a derived attribute or query that cannot be simulated
// on the current surface because its expression depends on out-of-scope classes
// (typically via association navigation to single or multi peers).
type UnavailableMember struct {
	ClassKey       identity.Key
	ClassName      string
	Kind           MemberKind
	MemberKey      identity.Key
	MemberName     string
	MissingClasses []string // out-of-scope class display names, sorted
}

// Reason is a stable human-readable explanation for reports and violations.
func (u UnavailableMember) Reason() string {
	if len(u.MissingClasses) == 0 {
		return fmt.Sprintf("%s %q depends on classes outside the simulation surface", u.Kind, u.MemberName)
	}
	return fmt.Sprintf(
		"%s %q depends on out-of-scope class(es): %s",
		u.Kind, u.MemberName, strings.Join(u.MissingClasses, ", "),
	)
}

// CollectUnavailableMembers finds derived attributes and queries on surface classes
// whose expressions require classes not included in the surface.
func CollectUnavailableMembers(model *core.Model, resolved *ResolvedSurface) []UnavailableMember {
	if model == nil || resolved == nil {
		return nil
	}
	inScope := resolved.Classes
	nav := buildAssociationNavigationDeps(model)
	classByName := classLookupByNameAndTLA(model)
	classNames := classDisplayNames(model)

	var out []UnavailableMember
	for classKey, class := range resolved.Classes {
		for _, attr := range class.Attributes {
			if attr.DerivationPolicy == nil {
				continue
			}
			expr := attr.DerivationPolicy.Spec.Expression
			if expr == nil {
				continue
			}
			missing := missingClassesForExpression(expr, classKey, inScope, classByName, classNames, nav)
			if len(missing) == 0 {
				continue
			}
			out = append(out, UnavailableMember{
				ClassKey:       classKey,
				ClassName:      class.Name,
				Kind:           MemberDerived,
				MemberKey:      attr.Key,
				MemberName:     attr.Name,
				MissingClasses: missing,
			})
		}
		for _, query := range class.Queries {
			missing := missingClassesForLogics(query.Requires, classKey, inScope, classByName, classNames, nav)
			missing = mergeUniqueSorted(missing, missingClassesForLogics(query.Guarantees, classKey, inScope, classByName, classNames, nav))
			if len(missing) == 0 {
				continue
			}
			out = append(out, UnavailableMember{
				ClassKey:       classKey,
				ClassName:      class.Name,
				Kind:           MemberQuery,
				MemberKey:      query.Key,
				MemberName:     query.Name,
				MissingClasses: missing,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ClassKey.String() != out[j].ClassKey.String() {
			return out[i].ClassKey.String() < out[j].ClassKey.String()
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].MemberKey.String() < out[j].MemberKey.String()
	})
	return out
}

func missingClassesForLogics(
	logics []model_logic.Logic,
	ownerClassKey identity.Key,
	inScope map[identity.Key]model_class.Class,
	classByName map[string]identity.Key,
	classNames map[identity.Key]string,
	nav associationNavDeps,
) []string {
	var missing []string
	for _, logic := range logics {
		if logic.Spec.Expression != nil {
			missing = mergeUniqueSorted(missing, missingClassesForExpression(
				logic.Spec.Expression, ownerClassKey, inScope, classByName, classNames, nav,
			))
		}
		if logic.EndpointSelectorSpec.Expression != nil {
			missing = mergeUniqueSorted(missing, missingClassesForExpression(
				logic.EndpointSelectorSpec.Expression, ownerClassKey, inScope, classByName, classNames, nav,
			))
		}
	}
	return missing
}

func missingClassesForExpression(
	expr me.Expression,
	ownerClassKey identity.Key,
	inScope map[identity.Key]model_class.Class,
	classByName map[string]identity.Key,
	classNames map[identity.Key]string,
	nav associationNavDeps,
) []string {
	if expr == nil {
		return nil
	}
	required := make(map[identity.Key]bool)

	// Explicit class names / TLA class names and association-class member fields.
	for ident := range collectIdentifiersFromIR(expr) {
		if classKey, ok := classByName[ident]; ok {
			required[classKey] = true
		}
	}

	// Association navigations (forward, reverse, association-class members).
	collectAssociationClassKeys(expr, ownerClassKey, nav, required)

	var missing []string
	for classKey := range required {
		if _, ok := inScope[classKey]; ok {
			continue
		}
		if name := classNames[classKey]; name != "" {
			missing = append(missing, name)
		} else {
			missing = append(missing, classKey.String())
		}
	}
	sort.Strings(missing)
	return missing
}

// associationNavDeps maps owner class + field name to required peer/AC class keys.
type associationNavDeps map[identity.Key]map[string][]identity.Key

func buildAssociationNavigationDeps(model *core.Model) associationNavDeps {
	nav := make(associationNavDeps)
	add := func(owner identity.Key, field string, deps ...identity.Key) {
		if owner.String() == "" || field == "" {
			return
		}
		if nav[owner] == nil {
			nav[owner] = make(map[string][]identity.Key)
		}
		nav[owner][field] = appendUniqueKeys(nav[owner][field], deps...)
	}

	for _, assoc := range model.GetClassAssociations() {
		forward := model_class.AssociationTLAFieldName(assoc.Name)
		reverse := "_" + forward
		// From-side: forward field reaches to-endpoint (and AC row class when present).
		toDeps := []identity.Key{assoc.ToClassKey}
		fromDeps := []identity.Key{assoc.FromClassKey}
		if assoc.AssociationClassKey != nil {
			toDeps = append(toDeps, *assoc.AssociationClassKey)
			fromDeps = append(fromDeps, *assoc.AssociationClassKey)
			acName := ""
			if ac, ok := classByKey(model, *assoc.AssociationClassKey); ok {
				acName = model_class.ClassTLAName(ac.Name)
			}
			if acName != "" {
				// Host navigates AC member by class TLA name from either endpoint.
				add(assoc.FromClassKey, acName, *assoc.AssociationClassKey)
				add(assoc.ToClassKey, acName, *assoc.AssociationClassKey)
			}
		}
		add(assoc.FromClassKey, forward, toDeps...)
		add(assoc.ToClassKey, reverse, fromDeps...)
	}
	return nav
}

func collectAssociationClassKeys(
	expr me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) {
	if expr == nil {
		return
	}
	if collectAssocNavLeaf(expr, ownerClassKey, nav, required) {
		return
	}
	if collectAssocNavBinary(expr, ownerClassKey, nav, required) {
		return
	}
	if collectAssocNavCollection(expr, ownerClassKey, nav, required) {
		return
	}
	collectAssocNavControlFlow(expr, ownerClassKey, nav, required)
}

func collectAssocNavLeaf(
	expr me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) bool {
	switch e := expr.(type) {
	case *me.FieldAccess:
		if fields := nav[ownerClassKey]; fields != nil {
			for _, dep := range fields[e.Field] {
				required[dep] = true
			}
		}
		collectAssociationClassKeys(e.Base, ownerClassKey, nav, required)
	case *me.LocalVar, *me.SelfRef, *me.AttributeRef, *me.PriorFieldValue,
		*me.AssociationRef, *me.BoolLiteral, *me.IntLiteral, *me.RationalLiteral,
		*me.StringLiteral, *me.SetConstant, *me.NamedSetRef, *me.ClassRef:
		// Leaves with no further class-key deps.
	default:
		return false
	}
	return true
}

func collectAssocNavBinary(
	expr me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) bool {
	switch e := expr.(type) {
	case *me.BinaryArith:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.BinaryLogic:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.Compare:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.SetOp:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.SetCompare:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.BagOp:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.BagCompare:
		collectAssocNavPair(e.Left, e.Right, ownerClassKey, nav, required)
	case *me.Membership:
		collectAssocNavPair(e.Element, e.Set, ownerClassKey, nav, required)
	case *me.Negate:
		collectAssociationClassKeys(e.Expr, ownerClassKey, nav, required)
	case *me.Not:
		collectAssociationClassKeys(e.Expr, ownerClassKey, nav, required)
	case *me.NextState:
		collectAssociationClassKeys(e.Expr, ownerClassKey, nav, required)
	default:
		return false
	}
	return true
}

func collectAssocNavCollection(
	expr me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) bool {
	switch e := expr.(type) {
	case *me.SetLiteral:
		collectAssocNavSlice(e.Elements, ownerClassKey, nav, required)
	case *me.TupleLiteral:
		collectAssocNavSlice(e.Elements, ownerClassKey, nav, required)
	case *me.RecordLiteral:
		for _, f := range e.Fields {
			collectAssociationClassKeys(f.Value, ownerClassKey, nav, required)
		}
	case *me.TupleIndex:
		collectAssocNavPair(e.Tuple, e.Index, ownerClassKey, nav, required)
	case *me.StringIndex:
		collectAssocNavPair(e.Str, e.Index, ownerClassKey, nav, required)
	case *me.RecordUpdate:
		collectAssociationClassKeys(e.Base, ownerClassKey, nav, required)
		for _, alt := range e.Alterations {
			collectAssociationClassKeys(alt.Value, ownerClassKey, nav, required)
		}
	case *me.StringConcat:
		collectAssocNavSlice(e.Operands, ownerClassKey, nav, required)
	case *me.TupleConcat:
		collectAssocNavSlice(e.Operands, ownerClassKey, nav, required)
	default:
		return false
	}
	return true
}

func collectAssocNavControlFlow(
	expr me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) {
	switch e := expr.(type) {
	case *me.Quantifier:
		collectAssocNavPair(e.Domain, e.Predicate, ownerClassKey, nav, required)
	case *me.SetFilter:
		collectAssocNavPair(e.Set, e.Predicate, ownerClassKey, nav, required)
	case *me.SetMap:
		collectAssocNavPair(e.Set, e.Transform, ownerClassKey, nav, required)
	case *me.SetRange:
		collectAssocNavPair(e.Start, e.End, ownerClassKey, nav, required)
	case *me.IfThenElse:
		collectAssociationClassKeys(e.Condition, ownerClassKey, nav, required)
		collectAssociationClassKeys(e.Then, ownerClassKey, nav, required)
		collectAssociationClassKeys(e.Else, ownerClassKey, nav, required)
	case *me.LetExpr:
		collectAssocNavPair(e.Value, e.Body, ownerClassKey, nav, required)
	case *me.Choose:
		collectAssocNavPair(e.Set, e.Predicate, ownerClassKey, nav, required)
	case *me.Case:
		for _, branch := range e.Branches {
			collectAssocNavPair(branch.Condition, branch.Result, ownerClassKey, nav, required)
		}
		collectAssociationClassKeys(e.Otherwise, ownerClassKey, nav, required)
	case *me.ActionCall:
		collectAssocNavSlice(e.Args, ownerClassKey, nav, required)
	case *me.GlobalCall:
		collectAssocNavSlice(e.Args, ownerClassKey, nav, required)
	case *me.BuiltinCall:
		collectAssocNavSlice(e.Args, ownerClassKey, nav, required)
	case *me.EventCall:
		collectAssocNavSlice(e.Args, ownerClassKey, nav, required)
	}
}

func collectAssocNavPair(
	left, right me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) {
	collectAssociationClassKeys(left, ownerClassKey, nav, required)
	collectAssociationClassKeys(right, ownerClassKey, nav, required)
}

func collectAssocNavSlice(
	exprs []me.Expression,
	ownerClassKey identity.Key,
	nav associationNavDeps,
	required map[identity.Key]bool,
) {
	for _, el := range exprs {
		collectAssociationClassKeys(el, ownerClassKey, nav, required)
	}
}

func classLookupByNameAndTLA(model *core.Model) map[string]identity.Key {
	out := make(map[string]identity.Key)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				if class.Name != "" {
					out[class.Name] = classKey
					out[model_class.ClassTLAName(class.Name)] = classKey
				}
			}
		}
	}
	return out
}

func classDisplayNames(model *core.Model) map[identity.Key]string {
	out := make(map[identity.Key]string)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				out[classKey] = class.Name
			}
		}
	}
	return out
}

func classByKey(model *core.Model, key identity.Key) (model_class.Class, bool) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[key]; ok {
				return class, true
			}
		}
	}
	return model_class.Class{}, false
}

func appendUniqueKeys(dst []identity.Key, keys ...identity.Key) []identity.Key {
	seen := make(map[string]bool, len(dst))
	for _, k := range dst {
		seen[k.String()] = true
	}
	for _, k := range keys {
		if k.String() == "" || seen[k.String()] {
			continue
		}
		seen[k.String()] = true
		dst = append(dst, k)
	}
	return dst
}

func mergeUniqueSorted(a, b []string) []string {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	seen := make(map[string]bool, len(a)+len(b))
	var out []string
	for _, s := range a {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, s := range b {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
