package actions

import (
	"math/rand"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_data_type"
	"github.com/glemzurg/go-tlaplus/internal/simulator/invariants"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
)

// --- Helpers for index value generator tests ---

func spanAttrDef(name string, lower, upper int) *model_class.Attribute {
	return &model_class.Attribute{
		Key:       mustKey("domain/d/subdomain/s/class/c/attribute/" + name),
		Name:      name,
		IndexNums: []uint{1},
		DataType: &model_data_type.DataType{
			CollectionType: model_data_type.CollectionTypeAtomic,
			Atomic: &model_data_type.Atomic{
				ConstraintType: model_data_type.ConstraintTypeSpan,
				Span: &model_data_type.AtomicSpan{
					LowerType:   "closed",
					LowerValue:  &lower,
					HigherType:  "closed",
					HigherValue: &upper,
				},
			},
		},
	}
}

func enumAttrDef(name string, values []string) *model_class.Attribute {
	enums := make([]model_data_type.AtomicEnum, len(values))
	for i, v := range values {
		enums[i] = model_data_type.AtomicEnum{Value: v, SortOrder: i}
	}
	return &model_class.Attribute{
		Key:       mustKey("domain/d/subdomain/s/class/c/attribute/" + name),
		Name:      name,
		IndexNums: []uint{1},
		DataType: &model_data_type.DataType{
			CollectionType: model_data_type.CollectionTypeAtomic,
			Atomic: &model_data_type.Atomic{
				ConstraintType: model_data_type.ConstraintTypeEnumeration,
				Enums:          enums,
			},
		},
	}
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
	class := model_class.Class{Key: classKey, Name: "C"}

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

	class := model_class.Class{Key: classKey, Name: "C"}
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

	class := model_class.Class{Key: classKey, Name: "C"}
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

	class := model_class.Class{Key: classKey, Name: "C"}
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

	class := model_class.Class{Key: classKey, Name: "C"}
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

	class := model_class.Class{Key: classKey, Name: "C"}

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
