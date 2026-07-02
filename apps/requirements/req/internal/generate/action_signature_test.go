package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestActionDisplaySignaturePrependsSelfForNonCreationActions(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("d"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "s"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "behavior"))
	stateActiveKey := helper.Must(identity.NewStateKey(classKey, "active"))
	eventNewKey := helper.Must(identity.NewEventKey(classKey, "_new"))
	eventUpdateKey := helper.Must(identity.NewEventKey(classKey, "Update"))
	actionAddKey := helper.Must(identity.NewActionKey(classKey, "add"))
	actionUpdateKey := helper.Must(identity.NewActionKey(classKey, "update"))
	transAddKey := helper.Must(identity.NewTransitionKey(classKey, "", "_new", "", "add", "active"))
	transUpdateKey := helper.Must(identity.NewTransitionKey(classKey, "active", "update", "", "update", "active"))

	minBal := helper.Must(model_state.NewParameter(actionAddKey, "MinimumBalance", "Nat", false))
	topoff := helper.Must(model_state.NewParameter(actionUpdateKey, "TopoffBalance", "Nat", false))
	minBalUpdate := helper.Must(model_state.NewParameter(actionUpdateKey, "MinimumBalance", "Nat", false))
	addAction := model_state.NewAction(actionAddKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{minBal})
	updateAction := model_state.NewAction(actionUpdateKey, model_state.ActionDetails{Name: "Update", Details: ""}, nil, nil, nil, []model_state.Parameter{minBalUpdate, topoff})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Behavior", Details: ""})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: model_state.NewState(stateActiveKey, "Active", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventNewKey:    model_state.NewEvent(eventNewKey, "_new", "", nil),
		eventUpdateKey: model_state.NewEvent(eventUpdateKey, "Update", "", []string{"MinimumBalance", "TopoffBalance"}),
	})
	class.SetActions(map[identity.Key]model_state.Action{
		actionAddKey:    addAction,
		actionUpdateKey: updateAction,
	})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transAddKey: model_state.NewTransition(
			transAddKey, eventNewKey,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey},
			model_state.TransitionLogicKeys{ActionKey: &actionAddKey}, "",
		),
		transUpdateKey: model_state.NewTransition(
			transUpdateKey, eventUpdateKey,
			model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: &stateActiveKey},
			model_state.TransitionLogicKeys{ActionKey: &actionUpdateKey}, "",
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)

	require.Equal(t, "MinimumBalance", actionDisplaySignature(reqs, addAction))
	require.Equal(t, "self, MinimumBalance, TopoffBalance", actionDisplaySignature(reqs, updateAction))
}
