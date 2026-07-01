package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/require"
)

func TestEvenplayWalletSurfaceListsAccountBalanceDerived(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	catalog := NewClassCatalog(model)
	PopulateDerivedAttributeCallersFromModel(model, catalog)

	var accountKey string
	for _, classInfo := range catalog.AllScopedClasses() {
		if classInfo.Class.Name == "Account" {
			accountKey = classInfo.ClassKey.String()
			break
		}
	}
	require.NotEmpty(t, accountKey)

	derived := catalog.ExternalDerivedAttributes(mustKey(accountKey))
	require.NotEmpty(t, derived)
	require.Equal(t, "Balance", derived[0].Name)

	text := BuildSurfaceReport(catalog).FormatText()
	require.Contains(t, text, "derived: Balance")
}

func TestEvenplaySimulationReadsAccountBalanceDerived(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	subdomainKeys, err := surface.ResolveSubdomainKeysByPath(model, []string{"finance/wallet"})
	require.NoError(t, err)

	eng, err := NewSimulationEngine(model, SimulationConfig{
		MaxSteps:   500,
		RandomSeed: 7,
		Surface:    &surface.SurfaceSpecification{IncludeSubdomains: subdomainKeys},
	})
	require.NoError(t, err)

	result, err := eng.Run()
	require.NoError(t, err)

	var found bool
	for _, step := range result.Steps {
		if step.DerivedAttributeName == "Balance" {
			found = true
			require.NotNil(t, step.DerivedReadValue)
			break
		}
	}
	require.True(t, found, "simulation should include at least one Balance derived read step")
}