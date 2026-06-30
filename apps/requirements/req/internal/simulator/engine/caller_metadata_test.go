package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateCallerDataFromModel_UseCaseEventSender(t *testing.T) {
	subdomainKey := mustKey("domain/finance/subdomain/wallet")
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "maintains_partners"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "main"))
	fromObjectKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "administrator"))
	toObjectKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "partner"))
	stepKey := helper.Must(identity.NewScenarioStepKey(scenarioKey, "1"))
	eventKey := mustKey("domain/finance/subdomain/wallet/class/partner/event/add")
	adminClassKey := mustKey("domain/finance/subdomain/wallet/class/administrator")
	partnerClassKey := mustKey("domain/finance/subdomain/wallet/class/partner")

	leafType := model_scenario.LEAF_TYPE_EVENT
	step := model_scenario.Step{
		Key:           stepKey,
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      &leafType,
		FromObjectKey: &fromObjectKey,
		ToObjectKey:   &toObjectKey,
		EventKey:      &eventKey,
	}

	scenario := model_scenario.NewScenario(scenarioKey, "Main", "")
	scenario.SetObjects(map[identity.Key]model_scenario.Object{
		fromObjectKey: model_scenario.NewObject(fromObjectKey, 1, model_scenario.ObjectDiagramName{Name: "Admin", NameStyle: "name"}, adminClassKey, false, ""),
		toObjectKey:   model_scenario.NewObject(toObjectKey, 2, model_scenario.ObjectDiagramName{Name: "Acme", NameStyle: "name"}, partnerClassKey, false, ""),
	})
	scenario.Steps = &step

	useCase := model_use_case.NewUseCase(useCaseKey, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSea}, model_use_case.GeneralizationRefs{}, model_use_case.UseCaseDetails{Name: "Maintain Partners"})
	useCase.SetScenarios(map[identity.Key]model_scenario.Scenario{
		scenarioKey: scenario,
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		partnerClassKey: buildPartnerClassForCallerTest(partnerClassKey, eventKey),
	}
	subdomain.UseCases = map[identity.Key]model_use_case.UseCase{
		useCaseKey: useCase,
	}

	domainKey := mustKey("domain/finance")
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := core.NewModel("evenplay", core.ModelDetails{Name: "Evenplay", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}

	catalog := NewClassCatalog(&model)
	PopulateCallerDataFromModel(&model, catalog)

	cd := catalog.CallerData()
	require.Contains(t, cd.EventSentBy, eventKey)
	assert.Contains(t, cd.EventSentBy[eventKey], adminClassKey)
}

func TestExternalCreationEvents_FiltersSimulatableSender(t *testing.T) {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	catalog := NewClassCatalog(model)
	createEventKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	catalog.SetEventSentBy(createEventKey, []identity.Key{orderKey})

	ext := catalog.ExternalCreationEvents(itemKey)
	assert.Empty(t, ext, "creation event sent by simulatable in-scope class is internal")
}

func TestExternalCreationEvents_ExcludesAssociationSetAddPeer(t *testing.T) {
	subdomainKey := mustKey("domain/finance/subdomain/wallet")
	fromKey := mustKey("domain/finance/subdomain/wallet/class/currency_wallet_definition")
	toKey := mustKey("domain/finance/subdomain/wallet/class/social_currency_behavior")
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "applies_social"))
	actionKey := helper.Must(identity.NewActionKey(fromKey, "set_social"))
	eventNewTo := helper.Must(identity.NewEventKey(toKey, "_new"))

	fromStateKey := helper.Must(identity.NewStateKey(fromKey, "active"))
	fromClass := model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency Wallet Definition"})
	fromClass.SetStates(map[identity.Key]model_state.State{
		fromStateKey: model_state.NewState(fromStateKey, "Active", "", ""),
	})
	fromClass.SetActions(map[identity.Key]model_state.Action{
		actionKey: model_state.NewAction(
			actionKey,
			model_state.ActionDetails{Name: "SetSocialBehavior", Details: ""},
			nil,
			[]model_logic.Logic{
				model_logic.NewLogic(
					helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
					model_logic.LogicTypeStateChange,
					"",
					"AppliesSocialCurrencyLogic",
					logic_spec.ExpressionSpec{
						Notation:      model_logic.NotationTLAPlus,
						Specification: `AppliesSocialCurrencyLogic \union {_new(MinimumBalance, TopoffBalance)}`,
					},
					nil,
				),
			},
			nil,
			nil,
		),
	})

	toClass, _ := testItemClass()
	toClassKey := toClass.Key
	toClass = model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Social Currency Behavior"})
	stateActiveKey := helper.Must(identity.NewStateKey(toKey, "active"))
	transCreateKey := helper.Must(identity.NewTransitionKey(toKey, "", "_new", "", "", "active"))
	toClass.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: model_state.NewState(stateActiveKey, "Active", "", ""),
	})
	toClass.SetEvents(map[identity.Key]model_state.Event{
		eventNewTo: model_state.NewEvent(eventNewTo, "_new", "", []string{"MinimumBalance", "TopoffBalance"}),
	})
	toClass.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: model_state.NewTransition(
			transCreateKey,
			eventNewTo,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil},
			"",
		),
	})
	_ = toClassKey

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Applies Social Currency Logic", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(model_class.NewMultiplicity("0..1"))},
		model_class.Multiplicity{},
		model_class.AssociationOptions{},
	)

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{fromKey: fromClass, toKey: toClass}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	domainKey := mustKey("domain/finance")
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}

	model := core.NewModel("evenplay", core.ModelDetails{Name: "Evenplay", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	catalog := NewClassCatalog(&model)
	PopulateCallerDataFromModel(&model, catalog)

	cd := catalog.CallerData()
	require.Contains(t, cd.EventSentBy, eventNewTo)
	assert.Contains(t, cd.EventSentBy[eventNewTo], fromKey)
	assert.Empty(t, catalog.ExternalCreationEvents(toKey))
}

func buildPartnerClassForCallerTest(classKey, eventKey identity.Key) model_class.Class {
	stateKey := helper.Must(identity.NewStateKey(classKey, "active"))
	transKey := helper.Must(identity.NewTransitionKey(classKey, "", "add", "", "", "active"))

	class := model_class.NewClass(
		classKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Partner", Details: "", UnfinishedNotes: "", UmlComment: ""},
	)
	class.States = map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "Active", "", ""),
	}
	class.Events = map[identity.Key]model_state.Event{
		eventKey: model_state.NewEvent(eventKey, "Add", "", nil),
	}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: model_state.NewTransition(transKey, eventKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, ""),
	}
	return class
}
