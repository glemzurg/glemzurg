package engine

import (
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/require"
)

func TestEvenplayCurrencyAttributeInvariantViolatesInvalidISO(t *testing.T) {
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

	checker, err := invariants.NewInvariantChecker(model)
	require.NoError(t, err)

	var currencyClassKey string
	for _, d := range model.Domains {
		for _, sd := range d.Subdomains {
			for k, c := range sd.Classes {
				if c.Name == "Currency" {
					currencyClassKey = k.String()
					for _, attr := range c.Attributes {
						if attr.Name == "ISO Code" || attr.Key.SubKey == "iso" {
							t.Logf("attr %q nullable=%v invariants=%d specs=%v", attr.Name, attr.Nullable, len(attr.Invariants), attr.Invariants)
						}
					}
				}
			}
		}
	}
	require.NotEmpty(t, currencyClassKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("abbr", object.NewString("NS"))
	attrs.Set("iso", object.NewString("zQu9MxNm"))
	attrs.Set("name", object.NewString("T"))
	attrs.Set("precision", object.NewInteger(2))
	attrs.Set("type", object.NewString("SOCIAL"))
	simState.CreateInstance(mustKey(currencyClassKey), attrs)

	bindingsBuilder := state.NewBindingsBuilder(simState)
	require.NoError(t, bindingsBuilder.RegisterNamedSets(model))

	violations := checker.CheckAttributeInvariants(simState, bindingsBuilder)
	require.True(t, violations.HasViolations(), "expected attribute invariant violation for invalid ISO, got none")
	t.Logf("violations: %v", violations)
}
