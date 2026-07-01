package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type InvariantsSuite struct {
	suite.Suite
}

func TestInvariantsSuite(t *testing.T) {
	suite.Run(t, new(InvariantsSuite))
}

// Helper to create an identity.Key from a string.
func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

// parsedSpec creates a TLA+ ExpressionSpec with the expression parsed via the convert pipeline.
func parsedSpec(tla string) logic_spec.ExpressionSpec {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// orderSpec parses a TLA+ expression in the context of the Order class
// used by createTestModel, with attributes: status, amount, name.
func orderSpec(tla string) logic_spec.ExpressionSpec {
	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"status": helper.Must(identity.NewAttributeKey(classKey, "status")),
			"amount": helper.Must(identity.NewAttributeKey(classKey, "amount")),
			"name":   helper.Must(identity.NewAttributeKey(classKey, "name")),
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// Helper to create a basic model with a class.
func createTestModel() *core.Model {
	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")

	// Create a data type for status (enumeration)
	statusEnums := []model_data_type.AtomicEnum{
		{Value: "pending", SortOrder: 0},
		{Value: "active", SortOrder: 1},
		{Value: "completed", SortOrder: 2},
	}
	ordered := true
	statusDataType := &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "enumeration",
			Enums:          statusEnums,
			EnumOrdered:    &ordered,
		},
	}

	// Create a data type for amount (span)
	lowerValue := 0
	higherValue := 1000000
	lowerDenom := 1
	higherDenom := 1
	amountDataType := &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "span",
			Span: &model_data_type.AtomicSpan{
				LowerType:         "closed",
				LowerValue:        &lowerValue,
				LowerDenominator:  &lowerDenom,
				HigherType:        "closed",
				HigherValue:       &higherValue,
				HigherDenominator: &higherDenom,
				Units:             "cents",
				Precision:         1,
			},
		},
	}

	// Create a data type for name (unconstrained)
	nameDataType := &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "unconstrained",
		},
	}

	stringTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil))
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))

	// Create attributes
	statusAttr := helper.Must(model_class.NewAttribute(
		mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/status"),
		model_class.AttributeDetails{Name: "status", Details: ""},
		"enum of {pending, active, completed}",
		nil, false, model_class.AttributeAnnotations{},
	))
	statusAttr.DataType = statusDataType
	statusAttr.DataType.TypeSpec = &stringTypeSpec

	amountAttr := helper.Must(model_class.NewAttribute(mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/amount"), model_class.AttributeDetails{Name: "amount", Details: ""}, "[0, 1000000] cents", nil, false, model_class.AttributeAnnotations{}))
	amountAttr.DataType = amountDataType
	amountAttr.DataType.TypeSpec = &natTypeSpec

	nameAttr := helper.Must(model_class.NewAttribute(mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/name"), model_class.AttributeDetails{Name: "name", Details: ""}, "string", nil, true, model_class.AttributeAnnotations{}))
	nameAttr.DataType = nameDataType
	nameAttr.DataType.TypeSpec = &stringTypeSpec

	// Create an action with a post-condition guarantee
	actionKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order/action/complete")
	requires := []model_logic.Logic{
		model_logic.NewLogic(helper.Must(identity.NewActionRequireKey(actionKey, "0")), model_logic.LogicTypeAssessment, "Precondition.", "", orderSpec("self.status = \"active\""), nil),
	}
	guarantees := []model_logic.Logic{
		model_logic.NewLogic(helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")), model_logic.LogicTypeStateChange, "Postcondition.", "status", parsedSpec("\"completed\""), nil),  // This is a primed assignment, not a post-condition
		model_logic.NewLogic(helper.Must(identity.NewActionGuaranteeKey(actionKey, "1")), model_logic.LogicTypeStateChange, "Postcondition.", "amount", orderSpec("self.amount + 1"), nil), // A second guarantee on a different attribute
	}
	completeAction := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "complete", Details: ""}, requires, guarantees, nil, nil)

	// Create the class
	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{statusAttr, amountAttr, nameAttr})
	class.SetActions(map[identity.Key]model_state.Action{
		actionKey: completeAction,
	})

	// Create the subdomain
	subdomainKey := mustKey("domain/test_domain/subdomain/test_subdomain")
	subdomain := model_domain.NewSubdomain(subdomainKey, "TestSubdomain", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: class,
	}

	// Create the domain
	domainKey := mustKey("domain/test_domain")
	domain := model_domain.NewDomain(domainKey, "TestDomain", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	// Create the model
	invariants := []model_logic.Logic{
		model_logic.NewLogic(helper.Must(identity.NewInvariantKey("0")), model_logic.LogicTypeAssessment, "Always true.", "", parsedSpec("TRUE"), nil),
	}
	model := core.NewModel("test_model", core.ModelDetails{Name: "TestModel", Details: ""}, "", invariants, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	return &model
}

// Test: DataTypeChecker detects unparsed data types.
func (s *InvariantsSuite) TestDataTypeCheckerDetectsUnparsedDataType() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	// Create an attribute without a parsed DataType
	attr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/c/attribute/bad"), model_class.AttributeDetails{Name: "bad", Details: ""}, "invalid data type rules", nil, false, model_class.AttributeAnnotations{}))
	attr.DataType = nil // Not parsed!

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "BadClass", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.Attributes = []model_class.Attribute{attr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, violations := NewDataTypeChecker(&model)
	s.NotNil(checker)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeUnparsedDataType, violations[0].Type)
	s.Equal("bad", violations[0].AttributeName)
}

// Test: DataTypeChecker validates required attributes.
func (s *InvariantsSuite) TestDataTypeCheckerRequiredAttribute() {
	model := createTestModel()
	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create an instance without status (which is required)
	// Not setting status means Get("status") returns nil
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(100))
	// status is deliberately not set - required but missing

	instance := simState.CreateInstance(classKey, attrs)

	// Check should find the required attribute violation
	instanceViolations := checker.CheckInstance(instance)
	s.True(instanceViolations.HasViolations())

	var foundRequiredViolation bool
	for _, v := range instanceViolations {
		if v.Type == ViolationTypeRequiredAttribute && v.AttributeName == "status" {
			foundRequiredViolation = true
		}
	}
	s.True(foundRequiredViolation, "Should detect missing required attribute 'status'")
}

// Test: DataTypeChecker validates span constraints.
func (s *InvariantsSuite) TestDataTypeCheckerSpanConstraint() {
	model := createTestModel()
	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create an instance with amount outside the allowed range
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("active"))
	attrs.Set("amount", object.NewInteger(2000000)) // Exceeds max of 1000000

	instance := simState.CreateInstance(classKey, attrs)

	// Check should find the span violation
	instanceViolations := checker.CheckInstance(instance)
	s.True(instanceViolations.HasViolations())

	var foundSpanViolation bool
	for _, v := range instanceViolations {
		if v.Type == ViolationTypeSpanConstraint && v.AttributeName == "amount" {
			foundSpanViolation = true
		}
	}
	s.True(foundSpanViolation, "Should detect amount exceeding span constraint")
}

// Test: DataTypeChecker validates enumeration constraints.
func (s *InvariantsSuite) TestDataTypeCheckerEnumConstraint() {
	model := createTestModel()
	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create an instance with status not in the enumeration
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("invalid_status")) // Not in {pending, active, completed}
	attrs.Set("amount", object.NewInteger(100))

	instance := simState.CreateInstance(classKey, attrs)

	// Check should find the enum violation
	instanceViolations := checker.CheckInstance(instance)
	s.True(instanceViolations.HasViolations())

	var foundEnumViolation bool
	for _, v := range instanceViolations {
		if v.Type == ViolationTypeEnumConstraint && v.AttributeName == "status" {
			foundEnumViolation = true
		}
	}
	s.True(foundEnumViolation, "Should detect status not in enumeration")
}

// Test: DataTypeChecker passes for valid instance.
func (s *InvariantsSuite) TestDataTypeCheckerValidInstance() {
	model := createTestModel()
	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create a valid instance
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("active"))
	attrs.Set("amount", object.NewInteger(500))
	attrs.Set("name", object.NewString("Test Order"))

	instance := simState.CreateInstance(classKey, attrs)

	// Check should pass
	instanceViolations := checker.CheckInstance(instance)
	s.False(instanceViolations.HasViolations())
}

// Test: DataTypeChecker flags values on attributes that declare no TLA+ type_spec.
func (s *InvariantsSuite) TestDataTypeCheckerMissingAttributeTypeSpec() {
	classKey := mustKey("domain/d/subdomain/s/class/currency")

	abbrAttr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/currency/attribute/abbr"), model_class.AttributeDetails{Name: "Abbr", Details: ""}, "unconstrained", nil, false, model_class.AttributeAnnotations{}))
	abbrAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
	}

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency"})
	class.Attributes = []model_class.Attribute{abbrAttr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, violations := NewDataTypeChecker(&model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("abbr", object.NewString("USD"))
	instance := simState.CreateInstance(classKey, attrs)

	instanceViolations := checker.CheckInstance(instance)
	s.Require().Len(instanceViolations, 1)
	s.Equal(ViolationTypeMissingAttributeTypeSpec, instanceViolations[0].Type)
	s.Equal("Abbr", instanceViolations[0].AttributeName)
}

func (s *InvariantsSuite) TestDataTypeCheckerUnparsedAttributeOnInstance() {
	classKey := mustKey("domain/d/subdomain/s/class/currency")
	badAttr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/currency/attribute/bad"), model_class.AttributeDetails{Name: "Bad", Details: ""}, "not a valid data type rule", nil, false, model_class.AttributeAnnotations{}))
	badAttr.DataType = nil

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency"})
	class.Attributes = []model_class.Attribute{badAttr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, setupViolations := NewDataTypeChecker(&model)
	s.Require().Len(setupViolations, 1)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("bad", object.NewString("value"))
	instance := simState.CreateInstance(classKey, attrs)

	instanceViolations := checker.CheckInstance(instance)
	s.Require().Len(instanceViolations, 1)
	s.Equal(ViolationTypeUnparsedDataType, instanceViolations[0].Type)
	s.Equal(instance.ID, instanceViolations[0].InstanceID)
}

func (s *InvariantsSuite) TestCheckParameterTypeSpecsUnparsedRules() {
	actionKey := mustKey("domain/d/subdomain/s/class/currency/action/add")
	classKey := mustKey("domain/d/subdomain/s/class/currency")
	param := helper.Must(model_state.NewParameter(actionKey, "Name", "not a valid rule", false))
	param.DataType = nil

	violations := CheckParameterTypeSpecs(
		[]model_state.Parameter{param}, actionKey, "Add", "action", 1, classKey,
	)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeUnparsedDataType, violations[0].Type)
	s.Equal("Name", violations[0].AttributeName)
}

func (s *InvariantsSuite) TestCheckParameterTypeSpecs() {
	actionKey := mustKey("domain/d/subdomain/s/class/currency/action/add")
	classKey := mustKey("domain/d/subdomain/s/class/currency")
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "Name", "unconstrained", false)),
	}
	stringTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil))
	typed := helper.Must(model_state.NewParameter(actionKey, "Type", "enum of SOCIAL, REAL", false))
	typed.DataType.TypeSpec = &stringTypeSpec

	violations := CheckParameterTypeSpecs(params, actionKey, "Add", "action", 1, classKey)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeMissingParameterTypeSpec, violations[0].Type)
	s.Equal("Name", violations[0].AttributeName)

	violations = CheckParameterTypeSpecs(
		[]model_state.Parameter{typed}, actionKey, "Add", "action", 1, classKey,
	)
	s.Empty(violations)
}

func (s *InvariantsSuite) TestCheckParameterTypeSpecsDateTime() {
	actionKey := mustKey("domain/d/subdomain/s/class/event/action/record")
	classKey := mustKey("domain/d/subdomain/s/class/event")
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	stringTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil))

	missing := helper.Must(model_state.NewParameter(actionKey, "When", "datetime", false))
	wrongSpec := helper.Must(model_state.NewParameter(actionKey, "When", "datetime", false))
	wrongSpec.DataType.TypeSpec = &stringTypeSpec
	correct := helper.Must(model_state.NewParameter(actionKey, "When", "datetime", false))
	correct.DataType.TypeSpec = &natTypeSpec

	violations := CheckParameterTypeSpecs(
		[]model_state.Parameter{missing}, actionKey, "Record", "action", 1, classKey,
	)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeMissingParameterTypeSpec, violations[0].Type)

	violations = CheckParameterTypeSpecs(
		[]model_state.Parameter{wrongSpec}, actionKey, "Record", "action", 1, classKey,
	)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeDateTimeTypeSpecMismatch, violations[0].Type)
	s.Equal("STRING", violations[0].ActualValue)

	violations = CheckParameterTypeSpecs(
		[]model_state.Parameter{correct}, actionKey, "Record", "action", 1, classKey,
	)
	s.Empty(violations)
}

func (s *InvariantsSuite) TestDataTypeCheckerDateTimeTypeSpecMismatch() {
	classKey := mustKey("domain/d/subdomain/s/class/event")
	attrKey := mustKey("domain/d/subdomain/s/class/event/attribute/when")
	attr := helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "When", Details: ""}, "datetime", nil, false, model_class.AttributeAnnotations{}))
	stringTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil))
	attr.DataType.TypeSpec = &stringTypeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Event"})
	class.Attributes = []model_class.Attribute{attr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, setupViolations := NewDataTypeChecker(&model)
	s.Empty(setupViolations)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("when", object.NewNatural(42))
	instance := simState.CreateInstance(classKey, attrs)

	instanceViolations := checker.CheckInstance(instance)
	s.Require().Len(instanceViolations, 1)
	s.Equal(ViolationTypeDateTimeTypeSpecMismatch, instanceViolations[0].Type)
}

func (s *InvariantsSuite) TestDataTypeCheckerDateTimeConstraint() {
	classKey := mustKey("domain/d/subdomain/s/class/event")
	attrKey := mustKey("domain/d/subdomain/s/class/event/attribute/when")
	attr := helper.Must(model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "When", Details: ""}, "datetime", nil, false, model_class.AttributeAnnotations{}))
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	attr.DataType.TypeSpec = &natTypeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Event"})
	class.Attributes = []model_class.Attribute{attr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, setupViolations := NewDataTypeChecker(&model)
	s.Empty(setupViolations)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("when", object.NewNatural(42))
	instance := simState.CreateInstance(classKey, attrs)
	s.Empty(checker.CheckInstance(instance))

	attrs.Set("when", object.NewNatural(0))
	instance = simState.CreateInstance(classKey, attrs)
	violations := checker.CheckInstance(instance)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeDateTimeConstraint, violations[0].Type)

	attrs.Set("when", object.NewNatural(model_data_type.DateTimeValueMax+1))
	instance = simState.CreateInstance(classKey, attrs)
	violations = checker.CheckInstance(instance)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeDateTimeConstraint, violations[0].Type)
}

// Test: DataTypeChecker handles nullable attributes correctly.
func (s *InvariantsSuite) TestDataTypeCheckerNullableAttribute() {
	model := createTestModel()
	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create an instance with nil name (which is nullable - OK)
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("pending"))
	attrs.Set("amount", object.NewInteger(100))
	// name is not set (nil) - should be fine since nullable

	instance := simState.CreateInstance(classKey, attrs)

	// Check should pass - name is nullable
	instanceViolations := checker.CheckInstance(instance)
	s.False(instanceViolations.HasViolations())
}

// Test: Violation types and messages.
func (s *InvariantsSuite) TestViolationTypes() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	// Test required attribute violation
	v1 := NewRequiredAttributeViolation(1, classKey, "field")
	s.Equal(ViolationTypeRequiredAttribute, v1.Type)
	s.Contains(v1.Message, "field")
	s.Contains(v1.Message, "required")

	// Test span constraint violation
	v2 := NewSpanConstraintViolation(1, classKey, "amount", "500", "[0, 100]")
	s.Equal(ViolationTypeSpanConstraint, v2.Type)
	s.Contains(v2.Message, "amount")
	s.Contains(v2.Message, "500")
	s.Contains(v2.Message, "[0, 100]")

	// Test enum constraint violation
	v3 := NewEnumConstraintViolation(1, classKey, "status", "bad", []string{"good", "ok"})
	s.Equal(ViolationTypeEnumConstraint, v3.Type)
	s.Contains(v3.Message, "status")
	s.Contains(v3.Message, "bad")

	// Test collection size violation
	minSize := 1
	maxSize := 5
	v4 := NewCollectionSizeViolation(1, classKey, "items", 10, &minSize, &maxSize)
	s.Equal(ViolationTypeCollectionSize, v4.Type)
	s.Contains(v4.Message, "items")
	s.Contains(v4.Message, "10")

	// Test model invariant violation
	v5 := NewModelInvariantViolation(0, "x > 5", "expression returned FALSE")
	s.Equal(ViolationTypeModelInvariant, v5.Type)
	s.Contains(v5.Message, "x > 5")

	// Test unparsed data type violation
	v6 := NewUnparsedDataTypeViolation(classKey, "bad_field", "invalid syntax")
	s.Equal(ViolationTypeUnparsedDataType, v6.Type)
	s.Contains(v6.Message, "bad_field")
	s.Contains(v6.Message, "unparsed")
}

// Test: ViolationErrors filtering.
func (s *InvariantsSuite) TestViolationErrorsFiltering() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	violations := ViolationErrors{
		NewRequiredAttributeViolation(1, classKey, "a"),
		NewSpanConstraintViolation(1, classKey, "b", "1", "[0,0]"),
		NewModelInvariantViolation(0, "TRUE", "failed"),
		NewEnumConstraintViolation(1, classKey, "c", "x", []string{"y"}),
	}

	// Filter by type
	required := violations.ByType(ViolationTypeRequiredAttribute)
	s.Len(required, 1)
	s.Equal("a", required[0].AttributeName)

	// TLA violations
	tla := violations.TLAViolations()
	s.Len(tla, 1)
	s.Equal(ViolationTypeModelInvariant, tla[0].Type)

	// Data type violations
	dataType := violations.DataTypeViolations()
	s.Len(dataType, 3)
}

// Test: InvariantChecker creation.
func (s *InvariantsSuite) TestInvariantCheckerCreation() {
	model := createTestModel()
	checker, err := NewInvariantChecker(model)
	s.Require().NoError(err)
	s.NotNil(checker)
	s.Equal(1, checker.GetModelInvariantCount())
}

// Test: InvariantChecker model invariant that passes.
func (s *InvariantsSuite) TestInvariantCheckerModelInvariantPasses() {
	model := createTestModel()
	checker, err := NewInvariantChecker(model)
	s.Require().NoError(err)

	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	violations := checker.CheckModelInvariants(simState, bindingsBuilder)
	s.False(violations.HasViolations())
}

// Test: InvariantChecker model invariant that fails.
func (s *InvariantsSuite) TestInvariantCheckerModelInvariantFails() {
	invariants := []model_logic.Logic{
		model_logic.NewLogic(helper.Must(identity.NewInvariantKey("0")), model_logic.LogicTypeAssessment, "Always false.", "", parsedSpec("FALSE"), nil),
	}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", invariants, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{}

	checker, err := NewInvariantChecker(&model)
	s.Require().NoError(err)

	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	violations := checker.CheckModelInvariants(simState, bindingsBuilder)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeModelInvariant, violations[0].Type)
}

// Test: Invalid TLA+ expression results in ParseOk() == false (no expression parsed).
func (s *InvariantsSuite) TestInvariantCheckerInvalidExpression() {
	spec := parsedSpec("this is not valid TLA+")
	// Invalid TLA+ silently produces nil Expression.
	s.False(spec.ParseOk())

	invariants := []model_logic.Logic{
		model_logic.NewLogic(helper.Must(identity.NewInvariantKey("0")), model_logic.LogicTypeAssessment, "Invalid expression.", "", spec, nil),
	}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", invariants, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{}

	// The checker should handle nil Expression (skip unparsed invariants).
	checker, err := NewInvariantChecker(&model)
	s.Require().NoError(err)
	s.Equal(0, checker.GetModelInvariantCount()) // Unparsed invariants are not counted.
}

// Test: CheckAllInvariants combines data type and TLA+ checks.
func (s *InvariantsSuite) TestCheckAllInvariants() {
	model := createTestModel()

	// Update model invariant to check something real (already parsed via parsedSpec in createTestModel).
	invChecker, err := NewInvariantChecker(model)
	s.Require().NoError(err)

	dtChecker, dtViolations := NewDataTypeChecker(model)
	s.False(dtViolations.HasViolations())

	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	simState := state.NewSimulationState()

	// Create an invalid instance
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("bad_status")) // Invalid enum
	attrs.Set("amount", object.NewInteger(100))

	simState.CreateInstance(classKey, attrs)

	bindingsBuilder := state.NewBindingsBuilder(simState)

	violations := invChecker.CheckAllInvariants(simState, bindingsBuilder, dtChecker, nil)
	s.True(violations.HasViolations())

	// Should have data type violation
	dtv := violations.DataTypeViolations()
	s.True(dtv.HasViolations())
}

// Test: Span with open bounds.
func (s *InvariantsSuite) TestDataTypeCheckerSpanOpenBounds() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	// Create data type with open bounds: (0, 100)
	lowerValue := 0
	higherValue := 100
	lowerDenom := 1
	higherDenom := 1
	dataType := &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "span",
			Span: &model_data_type.AtomicSpan{
				LowerType:         "open", // Exclude 0
				LowerValue:        &lowerValue,
				LowerDenominator:  &lowerDenom,
				HigherType:        "open", // Exclude 100
				HigherValue:       &higherValue,
				HigherDenominator: &higherDenom,
				Units:             "items",
				Precision:         1,
			},
		},
	}

	attr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/c/attribute/value"), model_class.AttributeDetails{Name: "value", Details: ""}, "(0, 100) items", nil, false, model_class.AttributeAnnotations{}))
	attr.DataType = dataType
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	attr.DataType.TypeSpec = &natTypeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Test", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.Attributes = []model_class.Attribute{attr}

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}

	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, violations := NewDataTypeChecker(&model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	simState := state.NewSimulationState()

	// Test value at lower bound (should fail with open)
	attrs1 := object.NewRecord()
	attrs1.Set("value", object.NewInteger(0))
	instance1 := simState.CreateInstance(classKey, attrs1)
	v1 := checker.CheckInstance(instance1)
	s.True(v1.HasViolations(), "Value 0 should violate open lower bound")

	// Test value at upper bound (should fail with open)
	attrs2 := object.NewRecord()
	attrs2.Set("value", object.NewInteger(100))
	instance2 := simState.CreateInstance(classKey, attrs2)
	v2 := checker.CheckInstance(instance2)
	s.True(v2.HasViolations(), "Value 100 should violate open upper bound")

	// Test value inside bounds (should pass)
	attrs3 := object.NewRecord()
	attrs3.Set("value", object.NewInteger(50))
	instance3 := simState.CreateInstance(classKey, attrs3)
	v3 := checker.CheckInstance(instance3)
	s.False(v3.HasViolations(), "Value 50 should pass")
}

func (s *InvariantsSuite) TestDataTypeCheckerUsesAttributeFieldKey() {
	classKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction")
	stringTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil))
	boolTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "BOOLEAN", nil))

	nameAttrKey := helper.Must(identity.NewAttributeKey(classKey, "name"))
	nameAttr := helper.Must(model_class.NewAttribute(nameAttrKey, model_class.AttributeDetails{Name: "Display Name", Details: ""}, "unconstrained", nil, false, model_class.AttributeAnnotations{}))
	nameAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
	}
	nameAttr.DataType.TypeSpec = &stringTypeSpec

	socialAttrKey := helper.Must(identity.NewAttributeKey(classKey, "social_only"))
	socialAttr := helper.Must(model_class.NewAttribute(socialAttrKey, model_class.AttributeDetails{Name: "Is Social Only", Details: ""}, "enum of TRUE, FALSE", nil, false, model_class.AttributeAnnotations{}))
	socialAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "enumeration",
			Enums: []model_data_type.AtomicEnum{
				{Value: "TRUE", SortOrder: 0},
				{Value: "FALSE", SortOrder: 1},
			},
		},
	}
	socialAttr.DataType.TypeSpec = &boolTypeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{nameAttr, socialAttr})

	subdomainKey := mustKey("domain/finance/subdomain/wallet")
	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/finance")
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, violations := NewDataTypeChecker(&model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("name", object.NewString("UK"))
	attrs.Set("social_only", object.NewBoolean(false))
	instance := simState.CreateInstance(classKey, attrs)

	instanceViolations := checker.CheckInstance(instance)
	s.False(instanceViolations.HasViolations())
}

func (s *InvariantsSuite) TestDataTypeCheckerNormalizesEmptyStringToNull() {
	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	nameAttrKey := helper.Must(identity.NewAttributeKey(classKey, "name"))
	nameAttr := helper.Must(model_class.NewAttribute(nameAttrKey, model_class.AttributeDetails{Name: "name", Details: ""}, "unconstrained", nil, false, model_class.AttributeAnnotations{}))
	nameAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "unconstrained",
		},
	}
	typeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil)
	s.Require().NoError(err)
	nameAttr.DataType.TypeSpec = &typeSpec

	countryAttrKey := helper.Must(identity.NewAttributeKey(classKey, "country_code"))
	countryAttr := helper.Must(model_class.NewAttribute(countryAttrKey, model_class.AttributeDetails{Name: "country_code", Details: ""}, "unconstrained", nil, true, model_class.AttributeAnnotations{}))
	countryAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "unconstrained",
		},
		TypeSpec: &typeSpec,
	}

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{nameAttr, countryAttr})

	subdomainKey := mustKey("domain/test_domain/subdomain/test_subdomain")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/test_domain")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, violations := NewDataTypeChecker(&model)
	s.NotNil(checker)
	s.False(violations.HasViolations())

	simState := state.NewSimulationState()

	emptyNameAttrs := object.NewRecord()
	emptyNameAttrs.Set("name", object.NewString(""))
	instanceEmpty := simState.CreateInstance(classKey, emptyNameAttrs)
	s.True(object.IsNull(instanceEmpty.GetAttribute("name")))
	emptyNameViolations := checker.CheckInstance(instanceEmpty)
	s.True(emptyNameViolations.HasViolations())
	s.Equal(ViolationTypeRequiredAttribute, emptyNameViolations[0].Type)

	nullCountryAttrs := object.NewRecord()
	nullCountryAttrs.Set("name", object.NewString("valid"))
	nullCountryAttrs.Set("country_code", object.NewSet())
	instanceNull := simState.CreateInstance(classKey, nullCountryAttrs)
	nullCountryViolations := checker.CheckInstance(instanceNull)
	s.False(nullCountryViolations.HasViolations())

	emptyCountryAttrs := object.NewRecord()
	emptyCountryAttrs.Set("name", object.NewString("valid"))
	emptyCountryAttrs.Set("country_code", object.NewString(""))
	instanceEmptyCountry := simState.CreateInstance(classKey, emptyCountryAttrs)
	s.True(object.IsNull(instanceEmptyCountry.GetAttribute("country_code")))
	emptyCountryViolations := checker.CheckInstance(instanceEmptyCountry)
	s.False(emptyCountryViolations.HasViolations())
}

func (s *InvariantsSuite) TestCheckAttributeInvariantSkipsWhenNullableAndUnset() {
	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")
	scoreAttrKey := helper.Must(identity.NewAttributeKey(classKey, "score"))
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	scoreAttr := helper.Must(model_class.NewAttribute(scoreAttrKey, model_class.AttributeDetails{Name: "score", Details: ""}, "positive when set", nil, true, model_class.AttributeAnnotations{}))
	scoreAttr.DataType = &model_data_type.DataType{
		CollectionType: "atomic",
		Atomic:         &model_data_type.Atomic{ConstraintType: "unconstrained"},
	}
	scoreAttr.DataType.TypeSpec = &natTypeSpec
	scoreCtx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"score": scoreAttrKey,
		},
	}
	scorePF := convert.NewExpressionParseFunc(scoreCtx)
	scoreInvariantSpec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", "self.score > 0", scorePF))
	scoreAttr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewAttributeInvariantKey(scoreAttrKey, "0")),
			model_logic.LogicTypeAssessment,
			"Score must be positive when provided.",
			"",
			scoreInvariantSpec,
			nil,
		),
	})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{scoreAttr})

	subdomainKey := mustKey("domain/test_domain/subdomain/test_subdomain")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domainKey := mustKey("domain/test_domain")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	checker, err := NewInvariantChecker(&model)
	s.Require().NoError(err)

	simState := state.NewSimulationState()
	unsetAttrs := object.NewRecord()
	unsetAttrs.Set("status", object.NewString("active"))
	unsetAttrs.Set("amount", object.NewInteger(1))
	simState.CreateInstance(classKey, unsetAttrs)

	bindingsBuilder := state.NewBindingsBuilder(simState)
	violations := checker.CheckAttributeInvariants(simState, bindingsBuilder)
	s.False(violations.HasViolations())

	violatingAttrs := object.NewRecord()
	violatingAttrs.Set("status", object.NewString("active"))
	violatingAttrs.Set("amount", object.NewInteger(1))
	violatingAttrs.Set("score", object.NewInteger(0))
	simState2 := state.NewSimulationState()
	simState2.CreateInstance(classKey, violatingAttrs)

	violations = checker.CheckAttributeInvariants(simState2, state.NewBindingsBuilder(simState2))
	s.True(violations.HasViolations())
	s.Equal(ViolationTypeAttributeInvariant, violations[0].Type)
}

// Test: Primed variables in model-level invariants fail to parse without class context.
func (s *InvariantsSuite) TestInvariantCheckerRejectsPrimedInvariants() {
	// Primed unresolved identifier fails to parse (no class context).
	spec := parsedSpec("x' > 0")
	s.False(spec.ParseOk()) // Cannot parse primed variable without class context.
}
