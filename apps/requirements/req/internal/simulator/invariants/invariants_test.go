package invariants

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_data_type"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_domain"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_state"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type InvariantsSuite struct {
	suite.Suite
}

func TestInvariantsSuite(t *testing.T) {
	suite.Run(t, new(InvariantsSuite))
}

// Helper to create an identity.Key from a string
func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

// Helper to create a basic model with a class
func createTestModel() *req_model.Model {
	classKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order")

	// Create a data type for status (enumeration)
	statusEnums := []model_data_type.AtomicEnum{
		{Value: "pending", SortOrder: 0},
		{Value: "active", SortOrder: 1},
		{Value: "completed", SortOrder: 2},
	}
	ordered := true
	statusDataType := &model_data_type.DataType{
		CollectionType: model_data_type.CollectionTypeAtomic,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.ConstraintTypeEnumeration,
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
		CollectionType: model_data_type.CollectionTypeAtomic,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.ConstraintTypeSpan,
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
		CollectionType: model_data_type.CollectionTypeAtomic,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.ConstraintTypeUnconstrained,
		},
	}

	// Create attributes
	statusAttr := model_class.Attribute{
		Key:           mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/status"),
		Name:          "status",
		DataTypeRules: "{pending, active, completed}",
		Nullable:      false,
		DataType:      statusDataType,
	}

	amountAttr := model_class.Attribute{
		Key:           mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/amount"),
		Name:          "amount",
		DataTypeRules: "[0, 1000000] cents",
		Nullable:      false,
		DataType:      amountDataType,
	}

	nameAttr := model_class.Attribute{
		Key:           mustKey("domain/test_domain/subdomain/test_subdomain/class/order/attribute/name"),
		Name:          "name",
		DataTypeRules: "string",
		Nullable:      true,
		DataType:      nameDataType,
	}

	// Create an action with a post-condition guarantee
	actionKey := mustKey("domain/test_domain/subdomain/test_subdomain/class/order/action/complete")
	completeAction := model_state.Action{
		Key:  actionKey,
		Name: "complete",
		TlaRequires: []string{
			"self.status = \"active\"",
		},
		TlaGuarantees: []string{
			"self.status' = \"completed\"", // This is a primed assignment, not a post-condition
			"self.status' # self.status",   // This is a post-condition invariant
		},
	}

	// Create the class
	class := model_class.Class{
		Key:  classKey,
		Name: "Order",
		Attributes: map[identity.Key]model_class.Attribute{
			statusAttr.Key: statusAttr,
			amountAttr.Key: amountAttr,
			nameAttr.Key:   nameAttr,
		},
		Actions: map[identity.Key]model_state.Action{
			actionKey: completeAction,
		},
	}

	// Create the subdomain
	subdomain := model_domain.Subdomain{
		Key:  mustKey("domain/test_domain/subdomain/test_subdomain"),
		Name: "TestSubdomain",
		Classes: map[identity.Key]model_class.Class{
			classKey: class,
		},
	}

	// Create the domain
	domain := model_domain.Domain{
		Key:  mustKey("domain/test_domain"),
		Name: "TestDomain",
		Subdomains: map[identity.Key]model_domain.Subdomain{
			subdomain.Key: subdomain,
		},
	}

	// Create the model
	return &req_model.Model{
		Key:  "test_model",
		Name: "TestModel",
		TlaInvariants: []string{
			"TRUE", // Always passes
		},
		Domains: map[identity.Key]model_domain.Domain{
			domain.Key: domain,
		},
	}
}

// Test: DataTypeChecker detects unparsed data types
func (s *InvariantsSuite) TestDataTypeCheckerDetectsUnparsedDataType() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	// Create an attribute without a parsed DataType
	attr := model_class.Attribute{
		Key:           mustKey("domain/d/subdomain/s/class/c/attribute/bad"),
		Name:          "bad",
		DataTypeRules: "invalid data type rules",
		Nullable:      false,
		DataType:      nil, // Not parsed!
	}

	class := model_class.Class{
		Key:  classKey,
		Name: "BadClass",
		Attributes: map[identity.Key]model_class.Attribute{
			attr.Key: attr,
		},
	}

	subdomain := model_domain.Subdomain{
		Key:     mustKey("domain/d/subdomain/s"),
		Name:    "S",
		Classes: map[identity.Key]model_class.Class{classKey: class},
	}

	domain := model_domain.Domain{
		Key:        mustKey("domain/d"),
		Name:       "D",
		Subdomains: map[identity.Key]model_domain.Subdomain{subdomain.Key: subdomain},
	}

	model := &req_model.Model{
		Key:     "test",
		Name:    "Test",
		Domains: map[identity.Key]model_domain.Domain{domain.Key: domain},
	}

	checker, violations := NewDataTypeChecker(model)
	s.NotNil(checker)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeUnparsedDataType, violations[0].Type)
	s.Equal("bad", violations[0].AttributeName)
}

// Test: DataTypeChecker validates required attributes
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

// Test: DataTypeChecker validates span constraints
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

// Test: DataTypeChecker validates enumeration constraints
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

// Test: DataTypeChecker passes for valid instance
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

// Test: DataTypeChecker handles nullable attributes correctly
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

// Test: Violation types and messages
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

// Test: ViolationList filtering
func (s *InvariantsSuite) TestViolationListFiltering() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	violations := ViolationList{
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

// Test: InvariantChecker creation
func (s *InvariantsSuite) TestInvariantCheckerCreation() {
	model := createTestModel()
	checker, err := NewInvariantChecker(model)
	s.NoError(err)
	s.NotNil(checker)
	s.Equal(1, checker.GetModelInvariantCount())
}

// Test: InvariantChecker model invariant that passes
func (s *InvariantsSuite) TestInvariantCheckerModelInvariantPasses() {
	model := createTestModel()
	checker, err := NewInvariantChecker(model)
	s.NoError(err)

	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	violations := checker.CheckModelInvariants(simState, bindingsBuilder)
	s.False(violations.HasViolations())
}

// Test: InvariantChecker model invariant that fails
func (s *InvariantsSuite) TestInvariantCheckerModelInvariantFails() {
	model := &req_model.Model{
		Key:  "test",
		Name: "Test",
		TlaInvariants: []string{
			"FALSE", // Always fails
		},
		Domains: map[identity.Key]model_domain.Domain{},
	}

	checker, err := NewInvariantChecker(model)
	s.NoError(err)

	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	violations := checker.CheckModelInvariants(simState, bindingsBuilder)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeModelInvariant, violations[0].Type)
}

// Test: InvariantChecker with invalid TLA+ expression
func (s *InvariantsSuite) TestInvariantCheckerInvalidExpression() {
	model := &req_model.Model{
		Key:  "test",
		Name: "Test",
		TlaInvariants: []string{
			"this is not valid TLA+",
		},
		Domains: map[identity.Key]model_domain.Domain{},
	}

	checker, err := NewInvariantChecker(model)
	s.Error(err)
	s.Nil(checker)
}

// Test: CheckAllInvariants combines data type and TLA+ checks
func (s *InvariantsSuite) TestCheckAllInvariants() {
	model := createTestModel()

	// Update model invariant to check something real
	model.TlaInvariants = []string{"TRUE"}

	invChecker, err := NewInvariantChecker(model)
	s.NoError(err)

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

// Test: Span with open bounds
func (s *InvariantsSuite) TestDataTypeCheckerSpanOpenBounds() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	// Create data type with open bounds: (0, 100)
	lowerValue := 0
	higherValue := 100
	lowerDenom := 1
	higherDenom := 1
	dataType := &model_data_type.DataType{
		CollectionType: model_data_type.CollectionTypeAtomic,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.ConstraintTypeSpan,
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

	attr := model_class.Attribute{
		Key:           mustKey("domain/d/subdomain/s/class/c/attribute/value"),
		Name:          "value",
		DataTypeRules: "(0, 100) items",
		Nullable:      false,
		DataType:      dataType,
	}

	class := model_class.Class{
		Key:        classKey,
		Name:       "Test",
		Attributes: map[identity.Key]model_class.Attribute{attr.Key: attr},
	}

	subdomain := model_domain.Subdomain{
		Key:     mustKey("domain/d/subdomain/s"),
		Name:    "S",
		Classes: map[identity.Key]model_class.Class{classKey: class},
	}

	domain := model_domain.Domain{
		Key:        mustKey("domain/d"),
		Name:       "D",
		Subdomains: map[identity.Key]model_domain.Subdomain{subdomain.Key: subdomain},
	}

	model := &req_model.Model{
		Key:     "test",
		Name:    "Test",
		Domains: map[identity.Key]model_domain.Domain{domain.Key: domain},
	}

	checker, violations := NewDataTypeChecker(model)
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

// Test: InvariantChecker rejects model invariants containing primed variables
func (s *InvariantsSuite) TestInvariantCheckerRejectsPrimedInvariants() {
	model := &req_model.Model{
		Key:  "test",
		Name: "Test",
		TlaInvariants: []string{
			"x' > 0",
		},
		Domains: map[identity.Key]model_domain.Domain{},
	}

	checker, err := NewInvariantChecker(model)
	s.Error(err)
	s.Nil(checker)
	s.Contains(err.Error(), "must not contain primed variables")
}
