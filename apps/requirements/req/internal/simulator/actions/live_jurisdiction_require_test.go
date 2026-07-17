package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestLiveEvenplayJurisdictionAddDomainExhaustedIsIneligible(t *testing.T) {
	modelPath := filepath.Clean(filepath.Join("/workspaces/glemzurg", "data_sandbox", "model", "evenplay"))
	if _, err := os.Stat(modelPath); err != nil {
		t.Skip("evenplay model not present")
	}
	model, failures, err := parser_human.Parse(modelPath)
	require.NoError(t, err)
	require.Empty(t, failures)
	require.NoError(t, convert.LowerModel(&model))

	var actionOwner ParameterOwner
	var classKey identity.Key
	found := false
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if class.Key.SubKey != "jurisdiction" {
					continue
				}
				classKey = class.Key
				for _, action := range class.Actions {
					if action.Name != "Add" {
						continue
					}
					actionOwner = ParameterOwnerFromAction(action)
					found = true
					break
				}
			}
		}
	}
	require.True(t, found, "jurisdiction Add action")

	logics, err := actionOwner.SamplingLogicsFor(actionOwner.Parameters)
	require.NoError(t, err)
	constraints := extractParameterConstraints(logics)
	require.NotNil(t, constraints.paramInNamedSetMinusPeerField,
		"live Add require must extract set-minus-used pattern; logics=%d requires=%d",
		len(logics), len(actionOwner.Requires))
	require.Equal(t, "jurisdiction_code", constraints.paramInNamedSetMinusPeerField.fieldSubKey)
	require.Equal(t, classKey, constraints.paramInNamedSetMinusPeerField.classKey)

	setSubKey := constraints.paramInNamedSetMinusPeerField.setSubKey
	namedSets := map[string]object.Object{}
	for _, ns := range model.NamedSets {
		require.NotNil(t, ns.Spec.Expression, ns.Name)
		result := evaluator.Eval(ns.Spec.Expression, evaluator.NewBindings())
		require.False(t, result.IsError(), ns.Name)
		namedSets[ns.Key.SubKey] = result.Value
	}
	liveSet, ok := namedSets[setSubKey].(*object.Set)
	require.True(t, ok)
	require.Positive(t, liveSet.Size())

	sampler := NewParameterSampler(NewParameterBinder(), namedSets)
	var used []object.Object
	for _, elem := range liveSet.Elements() {
		used = append(used, elem.Clone())
	}
	sampler.SetPeerFieldDistinctLookup(func(ck identity.Key, field string) []object.Object {
		require.Equal(t, classKey, ck)
		require.Equal(t, "jurisdiction_code", field)
		return used
	})
	okAvail, err := sampler.NamedSetSampleDomainsAvailable(actionOwner, actionOwner.Parameters)
	require.NoError(t, err)
	require.False(t, okAvail, "when all jurisdiction codes are used, Add must be ineligible")
}
