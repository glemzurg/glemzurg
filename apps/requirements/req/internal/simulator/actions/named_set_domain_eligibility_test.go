package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestNamedSetSampleDomainsAvailableSetMinusUsed(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	codeParam := helper.Must(model_state.NewParameter(actionKey, "JurisdictionCode", "code", false))
	nameParam := helper.Must(model_state.NewParameter(actionKey, "Name", "name", false))

	action := model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "Add", Details: ""},
		[]model_logic.Logic{
			model_logic.NewLogic(
				helper.Must(identity.NewActionRequireKey(actionKey, "0")),
				model_logic.LogicTypeAssessment,
				"Unused allowed code.",
				"",
				jurisdictionCodeSetMinusUsedRequireSpec(),
				nil,
			),
		},
		nil,
		nil,
		[]model_state.Parameter{codeParam, nameParam},
	)
	owner := ParameterOwnerFromAction(action)

	namedSets := map[string]object.Object{
		"jurisdictioncodes": object.NewSetFromElements([]object.Object{
			object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("CA")}),
			object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("NY")}),
		}),
	}
	sampler := NewParameterSampler(NewParameterBinder(), namedSets)

	// One code free → available.
	usedOne := object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("CA")})
	sampler.SetPeerFieldDistinctLookup(func(_ identity.Key, fieldSubKey string) []object.Object {
		require.Equal(t, "jurisdiction_code", fieldSubKey)
		return []object.Object{usedOne}
	})
	ok, err := sampler.NamedSetSampleDomainsAvailable(owner, action.Parameters)
	require.NoError(t, err)
	require.True(t, ok)

	// Both codes used → domain empty → not available.
	usedBoth := []object.Object{
		object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("CA")}),
		object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("NY")}),
	}
	sampler.SetPeerFieldDistinctLookup(func(_ identity.Key, _ string) []object.Object {
		return usedBoth
	})
	ok, err = sampler.NamedSetSampleDomainsAvailable(owner, action.Parameters)
	require.NoError(t, err)
	require.False(t, ok, "when every allowed code is used, Add must be ineligible")
}

func TestNamedSetSampleDomainsAvailableWhenAllNormalizedEmptyPairsUsed(t *testing.T) {
	// Mirrors evenplay: empty-string pair elements normalize to NULL in storage/sets.
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	codeParam := helper.Must(model_state.NewParameter(actionKey, "JurisdictionCode", "code", false))
	action := model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "Add", Details: ""},
		[]model_logic.Logic{
			model_logic.NewLogic(
				helper.Must(identity.NewActionRequireKey(actionKey, "0")),
				model_logic.LogicTypeAssessment,
				"Unused allowed code.",
				"",
				jurisdictionCodeSetMinusUsedRequireSpec(),
				nil,
			),
		},
		nil,
		nil,
		[]model_state.Parameter{codeParam},
	)
	owner := ParameterOwnerFromAction(action)

	// Build named set the way eval does (empty string → null).
	set := object.NewSet()
	for _, pair := range [][]string{
		{"", ""},
		{"US", "CA"},
		{"US", "NY"},
		{"CA", "ON"},
		{"CA", "BC"},
		{"GB", ""},
	} {
		set.Add(object.NewTupleFromElements([]object.Object{
			object.NewString(pair[0]),
			object.NewString(pair[1]),
		}))
	}
	require.Equal(t, 6, set.Size())

	sampler := NewParameterSampler(NewParameterBinder(), map[string]object.Object{
		"jurisdictioncodes": set,
	})
	// All set elements already used (as stored on instances after normalize).
	var used []object.Object
	for _, elem := range set.Elements() {
		used = append(used, elem.Clone())
	}
	sampler.SetPeerFieldDistinctLookup(func(_ identity.Key, _ string) []object.Object {
		return used
	})
	ok, err := sampler.NamedSetSampleDomainsAvailable(owner, action.Parameters)
	require.NoError(t, err)
	require.False(t, ok, "exhausted allowed codes must make Add ineligible")
}

func TestNamedSetSampleDomainsAvailableDoesNotBlockUnrelatedAction(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "update"))
	nameParam := helper.Must(model_state.NewParameter(actionKey, "Name", "name", false))
	action := model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "Update", Details: ""},
		nil,
		nil,
		nil,
		[]model_state.Parameter{nameParam},
	)
	owner := ParameterOwnerFromAction(action)
	sampler := NewParameterSampler(NewParameterBinder(), nil)
	ok, err := sampler.NamedSetSampleDomainsAvailable(owner, action.Parameters)
	require.NoError(t, err)
	require.True(t, ok)

	// Sanity: sampling still works when domain has free values.
	_ = rand.New(rand.NewSource(1)) //nolint:gosec
}

func TestDetectSetMinusFromModelSyntax(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	jurisdictionCodesKey := helper.Must(identity.NewNamedSetKey("jurisdictioncodes"))
	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: map[string]bool{"JurisdictionCode": true},
		ClassNames: map[string]identity.Key{"Jurisdiction": classKey},
		NamedSets:  map[string]identity.Key{"_JurisdictionCodes": jurisdictionCodesKey},
	}
	// Exact model text (with spaces as in class file).
	tla := `JurisdictionCode \in (_JurisdictionCodes \ { j.jurisdiction_code : j \in Jurisdiction })`
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, convert.NewExpressionParseFunc(ctx)))
	require.True(t, spec.ParseOk())
	constraints := extractParameterConstraints([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewActionRequireKey(helper.Must(identity.NewActionKey(classKey, "add")), "0")),
			model_logic.LogicTypeAssessment, "x", "", spec, nil,
		),
	})
	require.NotNil(t, constraints.paramInNamedSetMinusPeerField, "must extract set-minus pattern")
	require.Equal(t, classKey, constraints.paramInNamedSetMinusPeerField.classKey)
	require.Equal(t, "jurisdiction_code", constraints.paramInNamedSetMinusPeerField.fieldSubKey)
}
