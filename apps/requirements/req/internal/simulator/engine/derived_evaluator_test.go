package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

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

	derivationLogic := helper.Must(model_logic.NewLogic(mustKey("invariant/10"), model_logic.LogicTypeStateChange, "Double the price.", "price", model_logic.NotationTLAPlus, "self.price * 2"))

	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{
		attrPriceKey: {
			Key:  attrPriceKey,
			Name: "price",
		},
		attrDoublePriceKey: {
			Key:              attrDoublePriceKey,
			Name:             "doublePrice",
			DerivationPolicy: &derivationLogic,
		},
	}
	class.States = map[identity.Key]model_state.State{}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.NoError(err)
	s.NotNil(dae)

	// Create an instance with price=10.
	attrs := object.NewRecord()
	attrs.Set("price", object.NewInteger(10))
	instance := simState.CreateInstance(classKey, attrs)

	derived, err := dae.ResolveDerived(instance)
	s.NoError(err)
	s.NotNil(derived)

	doublePriceVal, ok := derived["doublePrice"]
	s.True(ok, "derived map should contain 'doublePrice'")
	s.Equal("20", doublePriceVal.Inspect())
}

// TestDerivedAttributeEmptySpecification verifies that NewDerivedAttributeEvaluator
// returns an error when an attribute has a DerivationPolicy with an empty Specification.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeEmptySpecification() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrKey := mustKey("domain/d/subdomain/s/class/product/attribute/derived_field")

	derivationLogic := helper.Must(model_logic.NewLogic(mustKey("invariant/11"), model_logic.LogicTypeStateChange, "A derived field.", "derived", model_logic.NotationTLAPlus, ""))

	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{
		attrKey: {
			Key:              attrKey,
			Name:             "derivedField",
			DerivationPolicy: &derivationLogic,
		},
	}
	class.States = map[identity.Key]model_state.State{}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Error(err)
	s.Nil(dae)
	s.Contains(err.Error(), "DerivationPolicy parse error")
}

// TestDerivedAttributeRejectsPrimedVars verifies that NewDerivedAttributeEvaluator
// returns an error when a DerivationPolicy specification contains primed variables.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeRejectsPrimedVars() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrPriceKey := mustKey("domain/d/subdomain/s/class/product/attribute/price")
	attrDerivedKey := mustKey("domain/d/subdomain/s/class/product/attribute/derived_field")

	derivationLogic := helper.Must(model_logic.NewLogic(mustKey("invariant/12"), model_logic.LogicTypeStateChange, "A derived field.", "price", model_logic.NotationTLAPlus, "self.price'"))

	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{
		attrPriceKey: {
			Key:  attrPriceKey,
			Name: "price",
		},
		attrDerivedKey: {
			Key:              attrDerivedKey,
			Name:             "derivedField",
			DerivationPolicy: &derivationLogic,
		},
	}
	class.States = map[identity.Key]model_state.State{}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.Error(err)
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

	derivationLogic := helper.Must(model_logic.NewLogic(mustKey("invariant/13"), model_logic.LogicTypeStateChange, "Double the price.", "price", model_logic.NotationTLAPlus, "self.price * 2"))

	class := helper.Must(model_class.NewClass(classKey, "Product", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{
		attrPriceKey: {
			Key:  attrPriceKey,
			Name: "price",
		},
		attrDoublePriceKey: {
			Key:              attrDoublePriceKey,
			Name:             "doublePrice",
			DerivationPolicy: &derivationLogic,
		},
	}
	class.States = map[identity.Key]model_state.State{}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

	simState := state.NewSimulationState()
	relationCtx := evaluator.NewRelationContext()
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(model, simState, relationCtx)
	s.NoError(err)

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
