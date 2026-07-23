package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestPeerEffectCascade_RecursivePeerTransitions(t *testing.T) {
	fromKey := mustKey("domain/finance/subdomain/wallet/class/currency_wallet_definition")
	toKey := mustKey("domain/finance/subdomain/wallet/class/social_currency_behavior")
	grandKey := mustKey("domain/finance/subdomain/wallet/class/grandchild")
	updateEventKey := mustKey("domain/finance/subdomain/wallet/class/social_currency_behavior/event/update")
	updateGrandEventKey := mustKey("domain/finance/subdomain/wallet/class/grandchild/event/update")

	parentResult := &actions.TransitionResult{
		InstanceID: 28,
		FromState:  "Active",
		ToState:    "Active",
		ActionResult: &actions.ActionResult{
			PeerTransitions: []actions.PeerTransitionRecord{
				{
					ClassKey:  toKey,
					ClassName: "Social Currency Behavior",
					EventKey:  updateEventKey,
					EventName: "Update",
					Parameters: map[string]object.Object{
						"MinimumBalance": object.NewInteger(81),
						"TopoffBalance":  object.NewInteger(38),
					},
					Result: &actions.TransitionResult{
						InstanceID: 12,
						FromState:  "Active",
						ToState:    "Active",
						ActionResult: &actions.ActionResult{
							PeerTransitions: []actions.PeerTransitionRecord{
								{
									ClassKey:  grandKey,
									ClassName: "Grandchild",
									EventKey:  updateGrandEventKey,
									EventName: "Update",
									Result: &actions.TransitionResult{
										InstanceID: 99,
										FromState:  "Active",
										ToState:    "Active",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	parentStep := &SimulationStep{
		StepNumber:       95,
		Kind:             StepKindNormal,
		ClassKey:         fromKey,
		ClassName:        "Currency Wallet Definition",
		EventName:        "SetSocialBehavior",
		InstanceID:       28,
		FromState:        "Active",
		ToState:          "Active",
		TransitionResult: parentResult,
	}

	catalog := NewClassCatalog(testModel())
	chainHandler := NewCreationChainHandler(catalog, nil, nil, nil, nil)
	exec := NewStepExecutor(StepExecutorDeps{ChainHandler: chainHandler, Catalog: catalog})

	require.NoError(t, exec.appendAssociationPeerCascades(parentStep, parentResult, instance.NewState(nil)))

	require.Len(t, parentStep.CascadedSteps, 1)
	updateStep := parentStep.CascadedSteps[0]
	require.Equal(t, "Update", updateStep.EventName)
	require.Equal(t, instance.ID(12), updateStep.InstanceID)

	require.Len(t, updateStep.CascadedSteps, 1)
	grandStep := updateStep.CascadedSteps[0]
	require.Equal(t, "Grandchild", grandStep.ClassName)
	require.Equal(t, instance.ID(99), grandStep.InstanceID)
}
