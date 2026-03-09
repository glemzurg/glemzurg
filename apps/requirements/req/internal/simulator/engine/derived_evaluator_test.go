package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

// productSpec parses a TLA+ expression in the context of a Product class
// with attributes: price, double_price, derived_field.
func productSpec(tla string) logic_spec.ExpressionSpec {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"price":         helper.Must(identity.NewAttributeKey(classKey, "price")),
			"double_price":  helper.Must(identity.NewAttributeKey(classKey, "double_price")),
			"derived_field": helper.Must(identity.NewAttributeKey(classKey, "derived_field")),
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

type DerivedEvaluatorSuite struct {
	suite.Suite
}

func TestDerivedEvaluatorSuite(t *testing.T) {
	suite.Run(t, new(DerivedEvaluatorSuite))
}

// ========================================================================
// Tests
// ========================================================================

// TestDerivedAttributeEvaluation verifies that a derived attribute with a
// DerivationPolicy expression is correctly evaluated. A "doublePrice"
// attribute defined as "self.price * 2" should resolve to 20 when price=10.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeEvaluation() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrPriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/price")
	attrDoublePriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/double_price")

	derivationLogic := model_logic.NewLogic(mustKey("invariant/10"), model_logic.LogicTypeValue, "Double the price.", "", productSpec("self.price * 2"), nil)

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, "price", "", "", nil, false,
		model_class.AttributeAnnotations{}))
	attrDoublePrice := helper.Must(model_class.NewAttribute(attrDoublePriceKey, "doublePrice", "", "", &derivationLogic, false,
		model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, "Product", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrPriceKey:       attrPrice,
		attrDoublePriceKey: attrDoublePrice,
	})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Require().NoError(err)
	s.NotNil(dae)

	// Create an instance with price=10.
	attrs := object.NewRecord()
	attrs.Set("price", object.NewInteger(10))
	instance := simState.CreateInstance(classKey, attrs)

	derived, err := dae.ResolveDerived(instance)
	s.Require().NoError(err)
	s.NotNil(derived)

	doublePriceVal, ok := derived["doublePrice"]
	s.True(ok, "derived map should contain 'doublePrice'")
	s.Equal("20", doublePriceVal.Inspect())
}

// TestDerivedAttributeEmptySpecification verifies that NewDerivedAttributeEvaluator
// silently skips an attribute with a DerivationPolicy that has an empty Specification.
// After LowerModel, empty specs remain with nil Expression and are skipped.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeEmptySpecification() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrKey := mustKey("domain/d/subdomain/s/class/product/attribute/derived_field")

	derivationLogic := model_logic.NewLogic(mustKey("invariant/11"), model_logic.LogicTypeValue, "A derived field.", "", helper.Must(logic_spec.NewExpressionSpec("tla_plus", "", nil)), nil)

	attrDerived := helper.Must(model_class.NewAttribute(attrKey, "derivedField", "", "", &derivationLogic, false,
		model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, "Product", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrKey: attrDerived,
	})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Require().NoError(err)
	s.NotNil(dae)
	// Empty specification is silently skipped — no derived attributes.
	s.False(dae.HasDerivedAttributes())
}

// TestDerivedAttributeRejectsPrimedVars verifies that NewDerivedAttributeEvaluator
// returns an error when a DerivationPolicy specification contains primed variables.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeRejectsPrimedVars() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrPriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/price")
	attrDerivedKey := mustKey("domain/d/subdomain/s/class/product/attribute/derived_field")

	derivationLogic := model_logic.NewLogic(mustKey("invariant/12"), model_logic.LogicTypeValue, "A derived field.", "", productSpec("self.price'"), nil)

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, "price", "", "", nil, false,
		model_class.AttributeAnnotations{}))
	attrDerived := helper.Must(model_class.NewAttribute(attrDerivedKey, "derivedField", "", "", &derivationLogic, false,
		model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, "Product", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrPriceKey:   attrPrice,
		attrDerivedKey: attrDerived,
	})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Require().Error(err)
	s.Nil(dae)
	s.Contains(err.Error(), "must not contain primed variables")
}

// TestDerivedAttributeInBindings verifies that when a DerivedAttributeEvaluator
// is set as the DerivedResolver on a BindingsBuilder, building bindings for an
// instance includes the derived attribute value in self.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeInBindings() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrPriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/price")
	attrDoublePriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/double_price")

	derivationLogic := model_logic.NewLogic(mustKey("invariant/13"), model_logic.LogicTypeValue, "Double the price.", "", productSpec("self.price * 2"), nil)

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, "price", "", "", nil, false,
		model_class.AttributeAnnotations{}))
	attrDoublePrice := helper.Must(model_class.NewAttribute(attrDoublePriceKey, "doublePrice", "", "", &derivationLogic, false,
		model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, "Product", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrPriceKey:       attrPrice,
		attrDoublePriceKey: attrDoublePrice,
	})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Require().NoError(err)

	// Create an instance with price=5.
	attrs := object.NewRecord()
	attrs.Set("price", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	// Set the DerivedAttributeEvaluator as the resolver on a BindingsBuilder.
	bb := state.NewBindingsBuilder(simState)
	bb.SetDerivedResolver(dae)

	bindings := bb.BuildForInstance(instance)

	// The self record in bindings should include the derived "doublePrice" field.
	selfRecord := bindings.Self()
	s.NotNil(selfRecord, "bindings should have self set")

	doublePriceVal := selfRecord.Get("doublePrice")
	s.NotNil(doublePriceVal, "self should contain derived attribute 'doublePrice'")
	s.Equal("10", doublePriceVal.Inspect())
}
