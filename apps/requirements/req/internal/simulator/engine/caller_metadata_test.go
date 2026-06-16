package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
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
		fromObjectKey: model_scenario.NewObject(fromObjectKey, 1, "Admin", "name", adminClassKey, false, ""),
		toObjectKey:   model_scenario.NewObject(toObjectKey, 2, "Acme", "name", partnerClassKey, false, ""),
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

	model := core.NewModel("evenplay", "Evenplay", "", "", nil, nil, nil)
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
		transKey: model_state.NewTransition(transKey, nil, eventKey, nil, nil, &stateKey, ""),
	}
	return class
}
