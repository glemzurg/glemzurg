package engine

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/require"
)

func TestEvenplayCurrencyAddSamplesISOFromNamedSet(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	action, ok := findCurrencyAddAction(model)
	require.True(t, ok)

	owner := actions.ParameterOwnerFromAction(action)
	logics, err := owner.SamplingLogicsFor(action.Parameters)
	require.NoError(t, err)
	constraints := actions.ExtractSamplingConstraintsForTest(logics)
	require.NotNil(t, constraints.NullableElseMembership)
	require.NotNil(t, constraints.NullableElseExclusionEquality)

	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	require.NoError(t, bb.RegisterNamedSets(model))

	binder := actions.NewParameterBinder()
	sampler := actions.NewParameterSampler(binder, bb.NamedSetValues())

	isoSet, ok := bb.NamedSetValues()["iso4217codes"].(*object.Set)
	require.True(t, ok)

	for seed := range 50 {
		result, err := sampler.SampleParameters(owner, action.Parameters, rand.New(rand.NewSource(int64(seed)))) //nolint:gosec // deterministic test seed
		require.NoError(t, err)
		if object.IsNull(result["ISO"]) {
			require.False(t, isoSet.Contains(result["Abbr"]))
			continue
		}
		iso := result["ISO"].(*object.String).Value()
		require.Len(t, iso, 3)
		require.Equal(t, iso, result["Abbr"].(*object.String).Value())
		require.True(t, isoSet.Contains(result["ISO"]))
	}
}

func findCurrencyAddAction(model *core.Model) (model_state.Action, bool) {
	for _, d := range model.Domains {
		for _, sd := range d.Subdomains {
			for _, class := range sd.Classes {
				if class.Name != "Currency" {
					continue
				}
				for _, act := range class.Actions {
					if act.Name == "Add" {
						return act, true
					}
				}
			}
		}
	}
	return model_state.Action{}, false
}

func loadEvenplayWalletModel(t *testing.T) *core.Model {
	t.Helper()
	modelPath := filepath.Join("../../../../../../data_sandbox/model/evenplay")
	parsed, failures, err := parser_human.Parse(modelPath)
	require.NoError(t, err)
	require.Empty(t, failures)
	active := &parsed

	subdomainKeys, err := surface.ResolveSubdomainKeysByPath(active, []string{"finance/wallet"})
	require.NoError(t, err)
	spec := &surface.SurfaceSpecification{IncludeSubdomains: subdomainKeys}
	resolved, err := surface.Resolve(spec, active)
	require.NoError(t, err)
	model, err := surface.BuildFilteredModel(active, resolved)
	require.NoError(t, err)
	require.NoError(t, convert.LowerModel(model))
	return model
}
