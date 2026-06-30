package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/stretchr/testify/require"
)

func TestEvenplayRemoveSocialBehaviorDeleteGuaranteeForm(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	action, ok := findCurrencyWalletDefinitionAction(model, "RemoveSocialBehavior")
	require.True(t, ok)
	require.Len(t, action.Guarantees, 1)

	guar := action.Guarantees[0]
	_, selection, eventCall, ok := model_class.MatchAssociationDeleteGuarantee(guar)
	require.True(t, ok)
	require.True(t, model_class.DeleteGuaranteeHasInlineStateChange(guar))
	require.Equal(t, "AppliesSocialCurrencyLogic", guar.Target)
	require.Equal(t, "b", selection.Variable)
	require.Equal(t, "_delete", eventCall.EventKey.SubKey)
}

func TestEvenplayRemoveSocialBehaviorTraceShowsNestedDelete(t *testing.T) {
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
		peerDeletes := countPeerDeleteTransitions(step)
		nestedDeletes := countNestedDeleteCascades(step.CascadedSteps)
		if peerDeletes == 0 {
			continue
		}
		require.Equal(t, peerDeletes, nestedDeletes,
			"step %d wallet#%d: peer delete transitions must appear as nested trace cascades",
			step.StepNumber, step.InstanceID)
	}

	step60 := result.Steps[59]
	require.Equal(t, "RemoveSocialBehavior", step60.EventName)
	require.Equal(t, state.InstanceID(33), step60.InstanceID)
	require.Empty(t, step60.CascadedSteps, "wallet #33 has no linked behaviors before step 60 in seed 42")
	require.Equal(t, 0, countPeerDeleteTransitions(step60))

	var step71 *SimulationStep
	for _, step := range result.Steps {
		if step.StepNumber == 71 && step.EventName == "RemoveSocialBehavior" && step.InstanceID == 24 {
			step71 = step
			break
		}
	}
	require.NotNil(t, step71, "seed 42 step 71 should remove social behavior from wallet #24")
	require.Len(t, step71.CascadedSteps, 1)
	require.Equal(t, StepKindDeletion, step71.CascadedSteps[0].Kind)
	require.Equal(t, model_state.EventNameDelete, step71.CascadedSteps[0].EventName)
}

func countPeerDeleteTransitions(step *SimulationStep) int {
	if step.TransitionResult == nil || step.TransitionResult.ActionResult == nil {
		return 0
	}
	count := 0
	for _, peer := range step.TransitionResult.ActionResult.PeerTransitions {
		if peer.Result != nil && peer.Result.WasDeletion {
			count++
		}
	}
	return count
}

func countNestedDeleteCascades(steps []*SimulationStep) int {
	count := 0
	for _, step := range steps {
		if step.Kind == StepKindDeletion {
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
