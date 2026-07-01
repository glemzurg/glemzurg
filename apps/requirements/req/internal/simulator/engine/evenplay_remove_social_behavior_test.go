package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/require"
)

func TestEvenplayRemoveSocialBehaviorDestroyGuaranteeForm(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	action, ok := findCurrencyWalletDefinitionAction(model, "RemoveSocialBehavior")
	require.True(t, ok)
	require.Len(t, action.Guarantees, 1)

	guar := action.Guarantees[0]
	_, selection, eventCall, ok := model_class.MatchAssociationDestroyGuarantee(guar)
	require.True(t, ok)
	require.True(t, model_class.DestroyGuaranteeHasInlineStateChange(guar))
	require.Equal(t, "AppliesSocialCurrencyLogic", guar.Target)
	require.Equal(t, "b", selection.Variable)
	require.Equal(t, "_destroy", eventCall.EventKey.SubKey)
}

func TestEvenplayRemoveSocialBehaviorTraceShowsNestedDestroy(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	subdomainKeys, err := surface.ResolveSubdomainKeysByPath(model, []string{"finance/wallet"})
	require.NoError(t, err)
	spec := &surface.SurfaceSpecification{IncludeSubdomains: subdomainKeys}

	eng, err := NewSimulationEngine(model, SimulationConfig{
		MaxSteps:        100,
		RandomSeed:      42,
		StopOnViolation: false,
		Surface:         spec,
	})
	require.NoError(t, err)
	result, err := eng.Run()
	require.NoError(t, err)

	for _, step := range result.Steps {
		if step.EventName != "RemoveSocialBehavior" {
			continue
		}
		peerDestroys := countPeerDestroyTransitions(step)
		nestedDestroys := countNestedDeleteCascades(step.CascadedSteps)
		if peerDestroys == 0 {
			continue
		}
		require.Equal(t, peerDestroys, nestedDestroys,
			"step %d wallet#%d: peer destroy transitions must appear as nested trace cascades",
			step.StepNumber, step.InstanceID)
	}

	var stepWithoutPeerDestroy *SimulationStep
	for _, step := range result.Steps {
		if step.EventName != "RemoveSocialBehavior" {
			continue
		}
		if countPeerDestroyTransitions(step) != 0 {
			continue
		}
		stepWithoutPeerDestroy = step
		break
	}
	require.NotNil(t, stepWithoutPeerDestroy, "seed 42 should include RemoveSocialBehavior without peer destroys")
	require.Empty(t, stepWithoutPeerDestroy.CascadedSteps,
		"wallet #%d has no linked behaviors for no-op RemoveSocialBehavior in seed 42", stepWithoutPeerDestroy.InstanceID)

	var stepWithDestroy *SimulationStep
	for _, step := range result.Steps {
		if step.EventName != "RemoveSocialBehavior" {
			continue
		}
		if countPeerDestroyTransitions(step) == 0 {
			continue
		}
		stepWithDestroy = step
		break
	}
	require.NotNil(t, stepWithDestroy, "seed 42 should include a RemoveSocialBehavior step with peer destroys")
	require.NotEmpty(t, stepWithDestroy.CascadedSteps)
	require.Equal(t, StepKindDestroy, stepWithDestroy.CascadedSteps[0].Kind)
	require.Equal(t, model_state.EventNameDestroy, stepWithDestroy.CascadedSteps[0].EventName)
}

func countPeerDestroyTransitions(step *SimulationStep) int {
	if step.TransitionResult == nil || step.TransitionResult.ActionResult == nil {
		return 0
	}
	count := 0
	for _, peer := range step.TransitionResult.ActionResult.PeerTransitions {
		if peer.Result != nil && peer.Result.WasDestroy {
			count++
		}
	}
	return count
}

func countNestedDeleteCascades(steps []*SimulationStep) int {
	count := 0
	for _, step := range steps {
		if step.Kind == StepKindDestroy {
			count++
		}
		count += countNestedDeleteCascades(step.CascadedSteps)
	}
	return count
}

func findCurrencyWalletDefinitionAction(model *core.Model, name string) (model_state.Action, bool) {
	for _, d := range model.Domains {
		for _, sd := range d.Subdomains {
			for _, class := range sd.Classes {
				if class.Name != "Currency Wallet Definition" {
					continue
				}
				for _, act := range class.Actions {
					if act.Name == name {
						return act, true
					}
				}
			}
		}
	}
	return model_state.Action{}, false
}
