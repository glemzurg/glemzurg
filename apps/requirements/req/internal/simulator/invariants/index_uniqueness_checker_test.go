package invariants

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// --- Helpers for building test models with indexes ---

func indexTestModel(attrs []model_class.Attribute) (*core.Model, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/plane")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Plane", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(attrs)

	subdomainKey := mustKey("domain/d/subdomain/s")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: class,
	}

	domainKey := mustKey("domain/d")
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	return &model, classKey
}

func spanAttr(name string, indexNums []uint) model_class.Attribute {
	return helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/plane/attribute/"+name), model_class.AttributeDetails{Name: name, Details: ""}, "[0, 10000]", nil, false, model_class.AttributeAnnotations{IndexNums: indexNums}))
}

func enumAttr(name string, values []string, indexNums []uint) model_class.Attribute {
	dataTypeRules := "{" + strings.Join(values, ", ") + "}"
	return helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/plane/attribute/"+name), model_class.AttributeDetails{Name: name, Details: ""}, dataTypeRules, nil, false, model_class.AttributeAnnotations{IndexNums: indexNums}))
}

// --- Tests ---

func (s *InvariantsSuite) TestIndexCheckerNoIndexes() {
	attr := helper.Must(model_class.NewAttribute(mustKey("domain/d/subdomain/s/class/plane/attribute/name"), model_class.AttributeDetails{Name: "name", Details: ""}, "string", nil, false, model_class.AttributeAnnotations{}))
	model, _ := indexTestModel([]model_class.Attribute{attr})

	checker := NewIndexUniquenessChecker(model)
	s.False(checker.HasIndexes())

	simState := state.NewSimulationState()
	violations := checker.CheckState(simState)
	s.False(violations.HasViolations())
}

func (s *InvariantsSuite) TestIndexCheckerSingleAttrNoConflict() {
	attr := spanAttr("tail_number", []uint{1})
	model, classKey := indexTestModel([]model_class.Attribute{attr})

	checker := NewIndexUniquenessChecker(model)
	s.True(checker.HasIndexes())

	simState := state.NewSimulationState()

	attrs1 := object.NewRecord()
	attrs1.Set("tail_number", object.NewInteger(100))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("tail_number", object.NewInteger(200))
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.False(violations.HasViolations())
}

func (s *InvariantsSuite) TestIndexCheckerSingleAttrConflict() {
	attr := spanAttr("tail_number", []uint{1})
	model, classKey := indexTestModel([]model_class.Attribute{attr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	attrs1 := object.NewRecord()
	attrs1.Set("tail_number", object.NewInteger(100))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("tail_number", object.NewInteger(100)) // duplicate!
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeIndexUniqueness, violations[0].Type)
	s.Contains(violations[0].Message, "tail_number")
	s.Contains(violations[0].Message, "index 1")
}

func (s *InvariantsSuite) TestIndexCheckerCompositeNoConflict() {
	emailAttr := enumAttr("email", []string{"a@b.com", "c@d.com"}, []uint{1})
	tenantAttr := enumAttr("tenant", []string{"acme", "globex"}, []uint{1})

	model, classKey := indexTestModel([]model_class.Attribute{emailAttr, tenantAttr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	// Same email, different tenant — no conflict
	attrs1 := object.NewRecord()
	attrs1.Set("email", object.NewString("a@b.com"))
	attrs1.Set("tenant", object.NewString("acme"))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("email", object.NewString("a@b.com"))
	attrs2.Set("tenant", object.NewString("globex"))
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.False(violations.HasViolations())
}

func (s *InvariantsSuite) TestIndexCheckerCompositeConflict() {
	emailAttr := enumAttr("email", []string{"a@b.com", "c@d.com"}, []uint{1})
	tenantAttr := enumAttr("tenant", []string{"acme", "globex"}, []uint{1})

	model, classKey := indexTestModel([]model_class.Attribute{emailAttr, tenantAttr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	// Same (email, tenant) tuple — conflict!
	attrs1 := object.NewRecord()
	attrs1.Set("email", object.NewString("a@b.com"))
	attrs1.Set("tenant", object.NewString("acme"))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("email", object.NewString("a@b.com"))
	attrs2.Set("tenant", object.NewString("acme"))
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Equal(ViolationTypeIndexUniqueness, violations[0].Type)
}

func (s *InvariantsSuite) TestIndexCheckerMultipleIndexes() {
	// Index 1: tail_number, Index 2: callsign
	tailAttr := spanAttr("tail_number", []uint{1})
	callAttr := enumAttr("callsign", []string{"AA1", "AA2", "BB1"}, []uint{2})

	model, classKey := indexTestModel([]model_class.Attribute{tailAttr, callAttr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	// Index 1 OK (different tail_numbers), Index 2 violated (same callsign)
	attrs1 := object.NewRecord()
	attrs1.Set("tail_number", object.NewInteger(100))
	attrs1.Set("callsign", object.NewString("AA1"))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("tail_number", object.NewInteger(200))
	attrs2.Set("callsign", object.NewString("AA1")) // duplicate callsign
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
	s.Contains(violations[0].Message, "callsign")
	s.Contains(violations[0].Message, "index 2")
}

func (s *InvariantsSuite) TestIndexCheckerNilValuesDuplicate() {
	attr := spanAttr("tail_number", []uint{1})
	model, classKey := indexTestModel([]model_class.Attribute{attr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	// Both instances have nil tail_number — treated as duplicate
	attrs1 := object.NewRecord()
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.True(violations.HasViolations())
	s.Len(violations, 1)
}

func (s *InvariantsSuite) TestIndexCheckerUsesAttributeFieldKeyNotDisplayName() {
	abbrAttr := helper.Must(model_class.NewAttribute(
		mustKey("domain/d/subdomain/s/class/currency/attribute/abbr"),
		model_class.AttributeDetails{Name: "Abbr", Details: ""},
		"unconstrained",
		nil,
		false,
		model_class.AttributeAnnotations{IndexNums: []uint{0}},
	))
	model, classKey := indexTestModel([]model_class.Attribute{abbrAttr})

	checker := NewIndexUniquenessChecker(model)
	info := checker.GetClassIndexInfo(classKey)
	s.Require().NotNil(info)
	s.Require().Len(info.Indexes, 1)
	s.Equal([]string{"abbr"}, info.Indexes[0].AttrNames)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("abbr", object.NewString("USD"))
	simState.CreateInstance(classKey, attrs)

	attrs2 := object.NewRecord()
	attrs2.Set("abbr", object.NewString("USD"))
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.True(violations.HasViolations())
}

func (s *InvariantsSuite) TestIndexCheckerMixedTypesNotEqual() {
	// Number 42 should not equal String "42"
	attr := spanAttr("id", []uint{1})
	model, classKey := indexTestModel([]model_class.Attribute{attr})

	checker := NewIndexUniquenessChecker(model)

	simState := state.NewSimulationState()

	attrs1 := object.NewRecord()
	attrs1.Set("id", object.NewInteger(42))
	simState.CreateInstance(classKey, attrs1)

	attrs2 := object.NewRecord()
	attrs2.Set("id", object.NewString("42"))
	simState.CreateInstance(classKey, attrs2)

	violations := checker.CheckState(simState)
	s.False(violations.HasViolations(), "Number 42 and String '42' should not be treated as equal")
}
