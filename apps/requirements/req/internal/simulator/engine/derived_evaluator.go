package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// derivedAttrInfo holds a pre-lowered DerivationPolicy expression for one attribute.
type derivedAttrInfo struct {
	attrKey identity.Key
	// attrSubKey is the storage / self.field key (attribute identity SubKey).
	attrSubKey string
	// attrName is the human-readable attribute name (errors and external reads).
	attrName   string
	expression me.Expression
}

// DerivedAttributeEvaluator computes derived attribute values on-demand.
// It loads all DerivationPolicy expressions at construction and evaluates
// them when an instance's derived attributes are requested.
type DerivedAttributeEvaluator struct {
	// byClass maps class key -> list of derived attribute info for that class.
	byClass map[identity.Key][]derivedAttrInfo

	// byAttrKey indexes the same infos for single-attribute evaluation.
	byAttrKey map[identity.Key]derivedAttrInfo

	// bindingsBuilder supplies association-aware bindings for each evaluation.
	bindingsBuilder *state.BindingsBuilder

	// evalCtx enables model global functions during derived evaluation.
	evalCtx *evaluator.EvalContext

	// catalog reports surface-unavailable derived attributes (out-of-scope deps).
	catalog *ClassCatalog
}

// NewDerivedAttributeEvaluator creates a new evaluator by scanning the model
// for attributes with DerivationPolicy. The model's ExpressionSpec.Expression
// fields must be populated (via parse functions passed to constructors).
// Returns an error if:
//   - any DerivationPolicy expression is not parsed (ParseOk() == false)
//   - any DerivationPolicy expression contains primed variables
func NewDerivedAttributeEvaluator(sch *schema.Schema, bindingsBuilder *state.BindingsBuilder, evalCtx *evaluator.EvalContext) (*DerivedAttributeEvaluator, error) {
	dae := &DerivedAttributeEvaluator{
		byClass:         make(map[identity.Key][]derivedAttrInfo),
		byAttrKey:       make(map[identity.Key]derivedAttrInfo),
		bindingsBuilder: bindingsBuilder,
		evalCtx:         evalCtx,
	}

	var buildErr error
	sch.ForEachClass(func(class model_class.Class) {
		if buildErr != nil {
			return
		}
		for _, attr := range class.Attributes {
			if attr.DerivationPolicy == nil {
				continue
			}

			expr := attr.DerivationPolicy.Spec.Expression
			if expr == nil {
				if attr.DerivationPolicy.Spec.Specification == "" {
					continue // Skip empty specs
				}
				buildErr = fmt.Errorf(
					"class %s attribute %s DerivationPolicy: expression not lowered",
					class.Name, attr.Name,
				)
				return
			}

			if model_bridge.ContainsAnyPrimedME(expr) {
				buildErr = fmt.Errorf(
					"class %s attribute %s DerivationPolicy must not contain primed variables",
					class.Name, attr.Name,
				)
				return
			}

			info := derivedAttrInfo{
				attrKey:    attr.Key,
				attrSubKey: attr.Key.SubKey,
				attrName:   attr.Name,
				expression: expr,
			}
			dae.byClass[class.Key] = append(dae.byClass[class.Key], info)
			dae.byAttrKey[attr.Key] = info
		}
	})
	if buildErr != nil {
		return nil, buildErr
	}

	return dae, nil
}

// SetCatalog wires surface unavailability checks for derived evaluation.
func (d *DerivedAttributeEvaluator) SetCatalog(catalog *ClassCatalog) {
	d.catalog = catalog
}

// ResolveDerived evaluates surface-available derived attributes for the given instance.
// Surface-unavailable attributes (out-of-scope association deps) are skipped so
// bindings inject only values that can be computed on this surface.
// Keys in the returned map are attribute SubKeys so they match stored fields and self.field access.
func (d *DerivedAttributeEvaluator) ResolveDerived(instance *instance.Instance) (map[string]object.Object, error) {
	infos := d.byClass[instance.ClassKey]
	if len(infos) == 0 {
		return make(map[string]object.Object), nil
	}

	bindings := d.bindingsBuilder.BuildForInstanceBase(instance)

	result := make(map[string]object.Object, len(infos))
	for _, info := range infos {
		if d.catalog != nil && d.catalog.IsSurfaceUnavailableDerived(info.attrKey) {
			continue
		}
		value, err := d.evalDerived(info, bindings)
		if err != nil {
			return nil, err
		}
		if value != nil {
			result[info.attrSubKey] = value
		}
	}

	return result, nil
}

// ResolveDerivedAttribute evaluates one derived attribute. When the attribute depends
// on out-of-scope classes, returns a surface-out-of-scope violation (not a hard error).
func (d *DerivedAttributeEvaluator) ResolveDerivedAttribute(
	instance *instance.Instance,
	attrKey identity.Key,
	attrName string,
) (object.Object, invariants.ViolationErrors, error) {
	if d.catalog != nil {
		if unavail, ok := d.catalog.SurfaceUnavailableDerived(attrKey); ok {
			return nil, invariants.ViolationErrors{
				invariants.NewSurfaceOutOfScopeViolation(
					instance.ClassKey, instance.ID, attrName, unavail.Reason(),
				),
			}, nil
		}
	}

	info, ok := d.byAttrKey[attrKey]
	if !ok {
		// Fall back to name match within the class (tests may use synthetic keys).
		for _, candidate := range d.byClass[instance.ClassKey] {
			if candidate.attrName == attrName {
				info = candidate
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, nil, fmt.Errorf("derived attribute %s not found on class", attrName)
	}

	bindings := d.bindingsBuilder.BuildForInstanceBase(instance)
	value, err := d.evalDerived(info, bindings)
	if err != nil {
		return nil, nil, err
	}
	return value, nil, nil
}

// SurfaceUnavailableDerivedReason returns the unavailability reason when known.
func (d *DerivedAttributeEvaluator) SurfaceUnavailableDerivedReason(attrKey identity.Key) (surface.UnavailableMember, bool) {
	if d.catalog == nil {
		return surface.UnavailableMember{}, false
	}
	return d.catalog.SurfaceUnavailableDerived(attrKey)
}

func (d *DerivedAttributeEvaluator) evalDerived(
	info derivedAttrInfo,
	bindings *evaluator.Bindings,
) (object.Object, error) {
	var evalResult *evaluator.EvalResult
	if d.evalCtx != nil {
		evalResult = evaluator.EvalWithContext(info.expression, bindings, d.evalCtx)
	} else {
		evalResult = evaluator.Eval(info.expression, bindings)
	}
	if evalResult.IsError() {
		return nil, fmt.Errorf(
			"derived attribute %s evaluation error: %s",
			info.attrName, evalResult.Error.Inspect(),
		)
	}
	return evalResult.Value, nil
}

// HasDerivedAttributes returns true if any class has derived attributes.
func (d *DerivedAttributeEvaluator) HasDerivedAttributes() bool {
	return len(d.byClass) > 0
}
