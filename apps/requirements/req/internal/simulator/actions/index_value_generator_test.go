package actions

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// --- Helpers for index value generator tests ---

func spanAttrDef(name string, lower, upper int) *model_class.Attribute {
	dataTypeRules := fmt.Sprintf("[%d, %d]", lower, upper)
	attr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/c/attribute/"+name), name, "", dataTypeRules, nil, false, "", []uint{1}))
	attr.DataType = &model_data_type.DataType{
		Key:            attr.Key.String(),
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "span",
			Span: &model_data_type.AtomicSpan{
				LowerType:   "closed",
				LowerValue:  &lower,
				HigherType:  "closed",
				HigherValue: &upper,
			},
		},
	}
	return &attr
}

func enumAttrDef(name string, values []string) *model_class.Attribute {
	dataTypeRules := "{" + strings.Join(values, ", ") + "}"
	attr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/c/attribute/"+name), name, "", dataTypeRules, nil, false, "", []uint{1}))
	enums := make([]model_data_type.AtomicEnum, len(values))
	for i, v := range values {
		enums[i] = model_data_type.AtomicEnum{Value: v, SortOrder: i}
	}
	attr.DataType = &model_data_type.DataType{
		Key:            attr.Key.String(),
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "enumeration",
			Enums:          enums,
		},
	}
	return &attr
}

func makeIndexInfo(classKey identity.Key, indexes []invariants.IndexDefinition) *invariants.ClassIndexInfo {
	return &invariants.ClassIndexInfo{
		ClassKey: classKey,
		Indexes:  indexes,
	}
}

// --- Tests ---

func (s *ActionsSuite) TestGenerateIndexSafeValuesNoIndexes() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	indexInfo := makeIndexInfo(classKey, nil) // no indexes
	attrs := object.NewRecord()
	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))

	err := generateIndexSafeValues(attrs, indexInfo, nil, class, rng)
	s.NoError(err)
}

func (s *ActionsSuite) TestGenerateIndexSafeValuesSpanUnique() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	idAttr := spanAttrDef("id", 1, 10000)
	indexInfo := makeIndexInfo(classKey, []invariants.IndexDefinition{
		{
			IndexNum:  1,
			AttrNames: []string{"id"},
			AttrDefs:  []*model_class.Attribute{idAttr},
		},
	})

	// Create an existing instance with id=42
	simState := state.NewSimulationState()
	existAttrs := object.NewRecord()
	existAttrs.Set("id", object.NewInteger(42))
	simState.CreateInstance(classKey, existAttrs)

	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))
	newAttrs := object.NewRecord()

	err := generateIndexSafeValues(newAttrs, indexInfo, simState.InstancesByClass(classKey), class, rng)
	s.NoError(err)

	// The generated id should not be 42
	generatedID := newAttrs.Get("id")
	s.NotNil(generatedID)
	s.NotEqual("42", generatedID.Inspect(), "Generated ID should differ from existing")
}

func (s *ActionsSuite) TestGenerateIndexSafeValuesEnumUnique() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	colorAttr := enumAttrDef("color", []string{"red", "green", "blue"})
	indexInfo := makeIndexInfo(classKey, []invariants.IndexDefinition{
		{
			IndexNum:  1,
			AttrNames: []string{"color"},
			AttrDefs:  []*model_class.Attribute{colorAttr},
		},
	})

	// Existing instances have "red" and "green"
	simState := state.NewSimulationState()
	a1 := object.NewRecord()
	a1.Set("color", object.NewString("red"))
	simState.CreateInstance(classKey, a1)

	a2 := object.NewRecord()
	a2.Set("color", object.NewString("green"))
	simState.CreateInstance(classKey, a2)

	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))
	newAttrs := object.NewRecord()

	err := generateIndexSafeValues(newAttrs, indexInfo, simState.InstancesByClass(classKey), class, rng)
	s.NoError(err)

	// The generated color should be "blue" (the only unused value)
	generated := newAttrs.Get("color")
	s.NotNil(generated)
	strVal := generated.(*object.String).Value()
	s.NotEqual("red", strVal)
	s.NotEqual("green", strVal)
}

func (s *ActionsSuite) TestGenerateIndexSafeValuesEnumExhausted() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	// Only 2 possible enum values
	colorAttr := enumAttrDef("color", []string{"red", "green"})
	indexInfo := makeIndexInfo(classKey, []invariants.IndexDefinition{
		{
			IndexNum:  1,
			AttrNames: []string{"color"},
			AttrDefs:  []*model_class.Attribute{colorAttr},
		},
	})

	// Both values already taken
	simState := state.NewSimulationState()
	a1 := object.NewRecord()
	a1.Set("color", object.NewString("red"))
	simState.CreateInstance(classKey, a1)

	a2 := object.NewRecord()
	a2.Set("color", object.NewString("green"))
	simState.CreateInstance(classKey, a2)

	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))
	newAttrs := object.NewRecord()

	err := generateIndexSafeValues(newAttrs, indexInfo, simState.InstancesByClass(classKey), class, rng)
	s.Error(err, "Should fail when all enum values are exhausted")
	s.Contains(err.Error(), "unable to generate unique")
}

func (s *ActionsSuite) TestGenerateIndexSafeValuesComposite() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	emailAttr := enumAttrDef("email", []string{"a@b.com", "c@d.com", "e@f.com"})
	tenantAttr := enumAttrDef("tenant", []string{"acme", "globex"})
	// Both in index 1
	emailAttr.IndexNums = []uint{1}
	tenantAttr.IndexNums = []uint{1}

	indexInfo := makeIndexInfo(classKey, []invariants.IndexDefinition{
		{
			IndexNum:  1,
			AttrNames: []string{"email", "tenant"},
			AttrDefs:  []*model_class.Attribute{emailAttr, tenantAttr},
		},
	})

	// One existing tuple: (a@b.com, acme)
	simState := state.NewSimulationState()
	a1 := object.NewRecord()
	a1.Set("email", object.NewString("a@b.com"))
	a1.Set("tenant", object.NewString("acme"))
	simState.CreateInstance(classKey, a1)

	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))
	newAttrs := object.NewRecord()

	err := generateIndexSafeValues(newAttrs, indexInfo, simState.InstancesByClass(classKey), class, rng)
	s.NoError(err)

	// The generated tuple should not be (a@b.com, acme)
	email := newAttrs.Get("email").(*object.String).Value()
	tenant := newAttrs.Get("tenant").(*object.String).Value()
	s.False(email == "a@b.com" && tenant == "acme", "Generated tuple should not match existing")
}

func (s *ActionsSuite) TestGenerateIndexSafeValuesPresetAttribute() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	rng := rand.New(rand.NewSource(42))

	emailAttr := enumAttrDef("email", []string{"a@b.com", "c@d.com"})
	tenantAttr := enumAttrDef("tenant", []string{"acme", "globex"})
	emailAttr.IndexNums = []uint{1}
	tenantAttr.IndexNums = []uint{1}

	indexInfo := makeIndexInfo(classKey, []invariants.IndexDefinition{
		{
			IndexNum:  1,
			AttrNames: []string{"email", "tenant"},
			AttrDefs:  []*model_class.Attribute{emailAttr, tenantAttr},
		},
	})

	// Existing: (a@b.com, acme)
	simState := state.NewSimulationState()
	a1 := object.NewRecord()
	a1.Set("email", object.NewString("a@b.com"))
	a1.Set("tenant", object.NewString("acme"))
	simState.CreateInstance(classKey, a1)

	class := helper.Must(model_class.NewClass(classKey, "C", "", nil, nil, nil, ""))

	// Pre-set email to "a@b.com" — generator should pick a different tenant
	newAttrs := object.NewRecord()
	newAttrs.Set("email", object.NewString("a@b.com"))

	err := generateIndexSafeValues(newAttrs, indexInfo, simState.InstancesByClass(classKey), class, rng)
	s.NoError(err)

	// Email should still be what we pre-set
	s.Equal("a@b.com", newAttrs.Get("email").(*object.String).Value())
	// Tenant should not be "acme" (to avoid duplicate tuple)
	tenant := newAttrs.Get("tenant").(*object.String).Value()
	s.NotEqual("acme", tenant)
}
