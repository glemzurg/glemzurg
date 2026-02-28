package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// derivedAttrInfo holds a parsed DerivationPolicy expression for one attribute.
type derivedAttrInfo struct {
	attrName   string
	expression ast.Expression
}

// DerivedAttributeEvaluator computes derived attribute values on-demand.
// It parses all DerivationPolicy expressions at construction and evaluates
// them when an instance's derived attributes are requested.
type DerivedAttributeEvaluator struct {
	// byClass maps class key -> list of derived attribute info for that class.
	byClass map[identity.Key][]derivedAttrInfo

	// bindingsBuilder is used to create bindings for evaluation.
	// Note: we build base bindings (without derived resolver) to avoid recursion.
	state *state.SimulationState

	// relationCtx for building bindings.
	relationCtx *evaluator.RelationContext
}

// NewDerivedAttributeEvaluator creates a new evaluator by scanning the model
// for attributes with DerivationPolicy. Returns an error if:
//   - any DerivationPolicy specification fails to parse
//   - any DerivationPolicy specification contains primed variables
func NewDerivedAttributeEvaluator(
	model *req_model.Model,
	simState *state.SimulationState,
	relationCtx *evaluator.RelationContext,
) (*DerivedAttributeEvaluator, error) {
	dae := &DerivedAttributeEvaluator{
		byClass:     make(map[identity.Key][]derivedAttrInfo),
		state:       simState,
		relationCtx: relationCtx,
	}

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, attr := range class.Attributes {
					if attr.DerivationPolicy == nil {
						continue
					}

					expr, err := parser.ParseExpression(attr.DerivationPolicy.Spec.Specification)
					if err != nil {
						return nil, fmt.Errorf(
							"class %s attribute %s DerivationPolicy parse error: %w",
							class.Name, attr.Name, err,
						)
					}

					if model_bridge.ContainsAnyPrimed(expr) {
						return nil, fmt.Errorf(
							"class %s attribute %s DerivationPolicy must not contain primed variables",
							class.Name, attr.Name,
						)
					}

					dae.byClass[class.Key] = append(dae.byClass[class.Key], derivedAttrInfo{
						attrName:   attr.Name,
						expression: expr,
					})
				}
			}
		}
	}

	return dae, nil
}

// ResolveDerived evaluates all derived attributes for the given instance.
// Returns a map of attribute name -> computed value.
func (d *DerivedAttributeEvaluator) ResolveDerived(instance *state.ClassInstance) (map[string]object.Object, error) {
	infos := d.byClass[instance.ClassKey]
	if len(infos) == 0 {
		return nil, nil
	}

	// Build bindings for this instance WITHOUT derived resolver to avoid recursion.
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(d.relationCtx)
	bindings = bindings.WithSelfAndClass(instance.Attributes, instance.ClassKey.String())

	result := make(map[string]object.Object, len(infos))
	for _, info := range infos {
		evalResult := evaluator.Eval(info.expression, bindings)
		if evalResult.IsError() {
			return nil, fmt.Errorf(
				"derived attribute %s evaluation error: %s",
				info.attrName, evalResult.Error.Inspect(),
			)
		}
		if evalResult.Value != nil {
			result[info.attrName] = evalResult.Value
		}
	}

	return result, nil
}

// HasDerivedAttributes returns true if any class has derived attributes.
func (d *DerivedAttributeEvaluator) HasDerivedAttributes() bool {
	return len(d.byClass) > 0
}
