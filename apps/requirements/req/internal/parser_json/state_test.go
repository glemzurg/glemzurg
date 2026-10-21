package parser_json

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestWriteReadStateMachine(t *testing.T) {
	model := stateMachineModel(t)
	require.NoError(t, model.Validate())

	first := writeModel(t, model)
	second := writeModel(t, model)
	assert.Equal(t, first, second)

	text := string(first)
	assert.NotContains(t, text, `\u003c`)
	assert.NotContains(t, text, `"data_type":`)
	assert.NotContains(t, text, `"expression"`)
	assert.Contains(t, text, `a < b`)
	assert.Contains(t, text, `self.name' = \"placed\"`)
	assert.Less(t, strings.Index(text, `/state/closed`), strings.Index(text, `/state/open`))
	assert.Less(t, strings.Index(text, `/event/_destroy`), strings.Index(text, `/event/_new`))
	assert.Less(t, strings.Index(text, `/event/_new`), strings.Index(text, `/event/place`))
	assert.Less(t, strings.Index(text, `"when":"entry"`), strings.Index(text, `"when":"do"`))
	assert.Less(t, strings.Index(text, `"when":"do"`), strings.Index(text, `"when":"exit"`))
	assert.Less(t, strings.Index(text, `"details":"small"`), strings.Index(text, `"details":"large"`))

	got := readModel(t, first)
	assert.Equal(t, model, got)
}

func stateMachineModel(t *testing.T) core.Model {
	t.Helper()

	domainKey := mustKey(t, func() (identity.Key, error) { return identity.NewDomainKey("commerce") })
	subKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(domainKey, "default") })
	classKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(subKey, "order") })
	nameKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeKey(classKey, "name") })
	openKey := mustKey(t, func() (identity.Key, error) { return identity.NewStateKey(classKey, "open") })
	closedKey := mustKey(t, func() (identity.Key, error) { return identity.NewStateKey(classKey, "closed") })
	newEventKey := mustKey(t, func() (identity.Key, error) { return identity.NewEventKey(classKey, "_new") })
	placeEventKey := mustKey(t, func() (identity.Key, error) { return identity.NewEventKey(classKey, "place") })
	destroyEventKey := mustKey(t, func() (identity.Key, error) { return identity.NewEventKey(classKey, "_destroy") })
	guardKey := mustKey(t, func() (identity.Key, error) { return identity.NewGuardKey(classKey, "can_place") })
	actionKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionKey(classKey, "place") })
	queryKey := mustKey(t, func() (identity.Key, error) { return identity.NewQueryKey(classKey, "lookup") })

	nameAttr, err := model_class.NewAttribute(nameKey, model_class.AttributeDetails{Name: "Name"}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	require.NoError(t, err)

	entryKey := mustKey(t, func() (identity.Key, error) { return identity.NewStateActionKey(openKey, "entry", "place") })
	doKey := mustKey(t, func() (identity.Key, error) { return identity.NewStateActionKey(openKey, "do", "place") })
	exitKey := mustKey(t, func() (identity.Key, error) { return identity.NewStateActionKey(openKey, "exit", "place") })
	open := model_state.NewState(openKey, "Open", "Accepting orders.", "open for business")
	open.SetActions([]model_state.StateAction{
		model_state.NewStateAction(exitKey, actionKey, "exit"),
		model_state.NewStateAction(doKey, actionKey, "do"),
		model_state.NewStateAction(entryKey, actionKey, "entry"),
	})
	closed := model_state.NewState(closedKey, "Closed", "", "")

	qty := mustParameter(t, actionKey, "Qty", "unconstrained", false)
	qtyInvKey := mustKey(t, func() (identity.Key, error) { return identity.NewParameterInvariantKey(qty.Key, "1") })
	qty.SetInvariants([]model_logic.Logic{
		mustLogic(t, qtyInvKey, model_logic.LogicTypeAssessment, "positive", "", "qty > 0"),
	})
	qty.SetSimulation(qtySimulation(t, qty.Key))
	note := mustParameter(t, actionKey, "Note", "unconstrained", true)

	letKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionRequireKey(actionKey, "1") })
	requireKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionRequireKey(actionKey, "2") })
	changeKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionGuaranteeKey(actionKey, "1") })
	eventsKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionGuaranteeKey(actionKey, "2") })
	safetyKey := mustKey(t, func() (identity.Key, error) { return identity.NewActionSafetyKey(actionKey, "1") })
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Place", Details: "Place the order."},
		[]model_logic.Logic{
			mustLogic(t, letKey, model_logic.LogicTypeLet, "", "bound", "qty"),
			mustLogic(t, requireKey, model_logic.LogicTypeAssessment, "ready", "", "bound > 0"),
		},
		[]model_logic.Logic{
			mustLogic(t, changeKey, model_logic.LogicTypeStateChange, "rename", "name", `self.name' = "placed"`),
			mustLogic(t, eventsKey, model_logic.LogicTypeEvents, "broadcast", "", "{ Notify(self) : TRUE }"),
		},
		[]model_logic.Logic{
			mustLogic(t, safetyKey, model_logic.LogicTypeSafetyRule, "named", "", `self.name' # ""`),
		},
		[]model_state.Parameter{qty, note},
	)

	filter := mustParameter(t, queryKey, "Filter", "unconstrained", false)
	queryRequireKey := mustKey(t, func() (identity.Key, error) { return identity.NewQueryRequireKey(queryKey, "1") })
	queryGuaranteeKey := mustKey(t, func() (identity.Key, error) { return identity.NewQueryGuaranteeKey(queryKey, "1") })
	query := model_state.NewQuery(queryKey, "Lookup", "Find an order.",
		[]model_logic.Logic{mustLogic(t, queryRequireKey, model_logic.LogicTypeAssessment, "ready", "", "TRUE")},
		[]model_logic.Logic{mustLogic(t, queryGuaranteeKey, model_logic.LogicTypeQuery, "match", "found", "TRUE")},
		[]model_state.Parameter{filter},
	)

	createKey := mustKey(t, func() (identity.Key, error) { return identity.NewTransitionKey(classKey, "", "_new", "", "", "open") })
	placeKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewTransitionKey(classKey, "open", "place", "can_place", "place", "closed")
	})
	finalKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewTransitionKey(classKey, "closed", "_destroy", "", "", "")
	})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: "A customer order."})
	class.SetAttributes([]model_class.Attribute{nameAttr})
	class.SetStates(map[identity.Key]model_state.State{openKey: open, closedKey: closed})
	class.SetEvents(map[identity.Key]model_state.Event{
		newEventKey:     model_state.NewEvent(newEventKey, model_state.EventNameNew, "", nil),
		placeEventKey:   model_state.NewEvent(placeEventKey, "place", "Customer places the order.", []string{"Qty", "Note"}),
		destroyEventKey: model_state.NewEvent(destroyEventKey, model_state.EventNameDestroy, "", nil),
	})
	class.SetGuards(map[identity.Key]model_state.Guard{
		guardKey: model_state.NewGuard(guardKey, "Can Place", mustLogic(t, guardKey, model_logic.LogicTypeAssessment, "affordable", "", "a < b")),
	})
	class.SetActions(map[identity.Key]model_state.Action{actionKey: action})
	class.SetQueries(map[identity.Key]model_state.Query{queryKey: query})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		createKey: model_state.NewTransition(createKey, newEventKey, model_state.TransitionStateKeys{ToStateKey: &openKey}, model_state.TransitionLogicKeys{}, ""),
		placeKey: model_state.NewTransition(placeKey, placeEventKey, model_state.TransitionStateKeys{FromStateKey: &openKey, ToStateKey: &closedKey}, model_state.TransitionLogicKeys{
			GuardKey:  &guardKey,
			ActionKey: &actionKey,
		}, "places the order"),
		finalKey: model_state.NewTransition(finalKey, destroyEventKey, model_state.TransitionStateKeys{FromStateKey: &closedKey}, model_state.TransitionLogicKeys{}, ""),
	})

	subdomain := model_domain.NewSubdomain(subKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Commerce", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subKey: subdomain}
	model := core.NewModel("sample", core.ModelDetails{Name: "Sample"}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	return model
}

func qtySimulation(t *testing.T, paramKey identity.Key) *model_state.ParameterSimulation {
	t.Helper()
	req0 := mustKey(t, func() (identity.Key, error) { return identity.NewParameterSimulationRequireKey(paramKey, "0") })
	spec0 := mustKey(t, func() (identity.Key, error) { return identity.NewParameterSimulationSpecKey(paramKey, "0") })
	req1 := mustKey(t, func() (identity.Key, error) { return identity.NewParameterSimulationRequireKey(paramKey, "1") })
	spec1 := mustKey(t, func() (identity.Key, error) { return identity.NewParameterSimulationSpecKey(paramKey, "1") })
	small := mustLogic(t, spec0, model_logic.LogicTypeValue, "small value", "", "1")
	large := mustLogic(t, spec1, model_logic.LogicTypeValue, "large value", "", "2")
	return &model_state.ParameterSimulation{
		Details: "sample qty",
		Rules: []model_state.ParameterSimulationRule{
			{
				Details: "small",
				Requires: []model_logic.Logic{
					mustLogic(t, req0, model_logic.LogicTypeAssessment, "positive", "", "TRUE"),
				},
				Specification: &small,
			},
			{
				Details: "large",
				Requires: []model_logic.Logic{
					mustLogic(t, req1, model_logic.LogicTypeAssessment, "room", "", "TRUE"),
				},
				Specification: &large,
			},
		},
	}
}

func mustParameter(t *testing.T, parent identity.Key, name, rules string, nullable bool) model_state.Parameter {
	t.Helper()
	param, err := model_state.NewParameter(parent, name, rules, nullable)
	require.NoError(t, err)
	return param
}

func mustLogic(t *testing.T, key identity.Key, kind, description, target, specification string) model_logic.Logic {
	t.Helper()
	return model_logic.NewLogic(key, kind, description, target, mustSpec(t, specification), nil)
}
