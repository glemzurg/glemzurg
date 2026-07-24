package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
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

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, model_class.AttributeDetails{Name: "price", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	attrDoublePrice := helper.Must(model_class.NewAttribute(attrDoublePriceKey, model_class.AttributeDetails{Name: "doublePrice", Details: ""}, "", &derivationLogic, false, model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrPrice, attrDoublePrice})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(emptySchema())
	bindingsBuilder := state.NewBindingsBuilder(simState)
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(schema.New(model), bindingsBuilder, nil)
	s.Require().NoError(err)
	s.NotNil(dae)

	// Create an instance with price=10.
	attrs := object.NewRecord()
	attrs.Set("price", object.NewInteger(10))
	instance := simState.CreateInstance(classKey, attrs)

	derived, err := dae.ResolveDerived(instance)
	s.Require().NoError(err)
	s.NotNil(derived)

	// Map keys are attribute SubKeys so self.field and storage keys align.
	doublePriceVal, ok := derived["double_price"]
	s.True(ok, "derived map should contain SubKey 'double_price'")
	s.Equal("20", doublePriceVal.Inspect())
}

// TestDerivedAttributeEmptySpecification verifies that NewDerivedAttributeEvaluator
// silently skips an attribute with a DerivationPolicy that has an empty Specification.
// After LowerModel, empty specs remain with nil Expression and are skipped.
func (s *DerivedEvaluatorSuite) TestDerivedAttributeEmptySpecification() {
	classKey := mustKey("domain/d/subdomain/s/class/product")
	attrKey := mustKey("domain/d/subdomain/s/class/product/attribute/derived_field")

	derivationLogic := model_logic.NewLogic(mustKey("invariant/11"), model_logic.LogicTypeValue, "A derived field.", "", helper.Must(logic_spec.NewExpressionSpec("tla_plus", "", nil)), nil)

	attrDerived := helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "derivedField", Details: ""}, "", &derivationLogic, false, model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrDerived})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(emptySchema())
	bindingsBuilder := state.NewBindingsBuilder(simState)
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(schema.New(model), bindingsBuilder, nil)
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

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, model_class.AttributeDetails{Name: "price", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	attrDerived := helper.Must(model_class.NewAttribute(attrDerivedKey, model_class.AttributeDetails{Name: "derivedField", Details: ""}, "", &derivationLogic, false, model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrPrice, attrDerived})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(emptySchema())
	bindingsBuilder := state.NewBindingsBuilder(simState)
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(schema.New(model), bindingsBuilder, nil)
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

	attrPrice := helper.Must(model_class.NewAttribute(attrPriceKey, model_class.AttributeDetails{Name: "price", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	attrDoublePrice := helper.Must(model_class.NewAttribute(attrDoublePriceKey, model_class.AttributeDetails{Name: "doublePrice", Details: ""}, "", &derivationLogic, false, model_class.AttributeAnnotations{}))

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrPrice, attrDoublePrice})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(emptySchema())
	bindingsBuilder := state.NewBindingsBuilder(simState)
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(schema.New(model), bindingsBuilder, nil)
	s.Require().NoError(err)

	// Create an instance with price=5.
	attrs := object.NewRecord()
	attrs.Set("price", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	// Set the DerivedAttributeEvaluator as the resolver on a BindingsBuilder.
	bb := state.NewBindingsBuilder(simState)
	bb.SetDerivedResolver(dae)

	bindings := bb.BuildForInstance(instance)

	// The self record in bindings should include the derived field under its SubKey.
	selfRecord := bindings.Self()
	s.NotNil(selfRecord, "bindings should have self set")

	doublePriceVal := selfRecord.Get("double_price")
	s.NotNil(doublePriceVal, "self should contain derived attribute under SubKey 'double_price'")
	s.Equal("10", doublePriceVal.Inspect())
}

// TestDerivedAttributeSubKeyWhenDisplayNameDiffers ensures ResolveDerived keys by
// attribute SubKey even when the human-readable name differs (e.g. social_only).
func (s *DerivedEvaluatorSuite) TestDerivedAttributeSubKeyWhenDisplayNameDiffers() {
	classKey := mustKey("domain/d/subdomain/s/class/jurisdiction")
	socialKey := mustKey("domain/d/subdomain/s/class/jurisdiction/attribute/social_only")

	// Constant derivation: only the map key (SubKey vs display name) is under test.
	derivationLogic := model_logic.NewLogic(
		mustKey("invariant/14"),
		model_logic.LogicTypeValue,
		"Always social for this fixture.",
		"",
		productSpec(`TRUE`),
		nil,
	)

	attrSocial := helper.Must(model_class.NewAttribute(
		socialKey,
		model_class.AttributeDetails{Name: "Is Social Only", Details: ""},
		"",
		&derivationLogic,
		false,
		model_class.AttributeAnnotations{},
	))

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Jurisdiction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrSocial})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(emptySchema())
	bindingsBuilder := state.NewBindingsBuilder(simState)
	model := testModel(classEntry(class, classKey))

	dae, err := NewDerivedAttributeEvaluator(schema.New(model), bindingsBuilder, nil)
	s.Require().NoError(err)

	instance := simState.CreateInstance(classKey, object.NewRecord())

	derived, err := dae.ResolveDerived(instance)
	s.Require().NoError(err)
	s.Contains(derived, "social_only")
	s.NotContains(derived, "Is Social Only")
	boolVal, ok := derived["social_only"].(*object.Boolean)
	s.Require().True(ok)
	s.True(boolVal.Value())

	bb := state.NewBindingsBuilder(simState)
	bb.SetDerivedResolver(dae)
	selfRecord := bb.BuildForInstance(instance).Self()
	injected, ok := selfRecord.Get("social_only").(*object.Boolean)
	s.Require().True(ok)
	s.True(injected.Value())
}
