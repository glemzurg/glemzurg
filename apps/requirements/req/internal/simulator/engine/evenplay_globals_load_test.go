package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
	"github.com/stretchr/testify/require"
)

func TestLoadEvenplayWalletGlobals(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	result := model_bridge.NewLoader().LoadFromModel(model)
	t.Logf("success=%d errors=%d", result.SuccessCount(), result.ErrorCount())
	for _, e := range result.Errors {
		t.Logf("err: %v", e)
	}
	require.False(t, result.HasErrors(), "wallet surface model should load all expressions")
	_, params, ok := registry.NewRuntimeAdapter(result.Registry).LookupGlobal("amountsbag")
	require.True(t, ok, "_AmountsBag should be registered")
	require.Equal(t, []string{"adjustments"}, params)
}