package parser_json

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestWriteReadUseCases(t *testing.T) {
	model := useCaseModel(t)
	require.NoError(t, model.Validate())

	first := writeModel(t, model)
	second := writeModel(t, model)
	assert.Equal(t, first, second)

	text := string(first)
	assert.NotContains(t, text, `\u003c`)
	assert.Contains(t, text, `a < b`)
	assert.Less(t, strings.Index(text, `/class/clerk`), strings.Index(text, `/class/customer`))
	assert.Less(t, strings.Index(text, `/usecase/place_order`), strings.Index(text, `/usecase/rush_order`))
	assert.Less(t, strings.Index(text, `/usecase/rush_order`), strings.Index(text, `/usecase/sign_in`))
	assert.Less(t, strings.Index(text, `"mud_use_case_key":"domain/commerce/subdomain/default/usecase/rush_order"`), strings.Index(text, `"mud_use_case_key":"domain/commerce/subdomain/default/usecase/sign_in"`))
	assert.Less(t, strings.Index(text, `/scenario/alt`), strings.Index(text, `/scenario/happy`))
	assert.Less(t, strings.Index(text, `/sobject/buyer`), strings.Index(text, `/sobject/order`))
	assert.Contains(t, text, `"leaf_type":"destroy"`)
	assert.Contains(t, text, `"step_type":"switch"`)
	assert.Contains(t, text, `"share_type":"include"`)
	assert.Contains(t, text, `"share_type":"extend"`)

	got := readModel(t, first)
	assert.Equal(t, model, got)
}

func useCaseModel(t *testing.T) core.Model {
	t.Helper()

	domainKey := mustKey(t, func() (identity.Key, error) { return identity.NewDomainKey("commerce") })
	subKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(domainKey, "default") })
	clerkActorKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("clerk") })
	customerActorKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("customer") })
	clerkKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(subKey, "clerk") })
	customerKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(subKey, "customer") })
	orderKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(subKey, "order") })
	eventKey := mustKey(t, func() (identity.Key, error) { return identity.NewEventKey(orderKey, "place") })
	queryKey := mustKey(t, func() (identity.Key, error) { return identity.NewQueryKey(orderKey, "lookup") })
	genKey := mustKey(t, func() (identity.Key, error) { return identity.NewUseCaseGeneralizationKey(subKey, "checkout") })
	placeKey := mustKey(t, func() (identity.Key, error) { return identity.NewUseCaseKey(subKey, "place_order") })
	rushKey := mustKey(t, func() (identity.Key, error) { return identity.NewUseCaseKey(subKey, "rush_order") })
	signInKey := mustKey(t, func() (identity.Key, error) { return identity.NewUseCaseKey(subKey, "sign_in") })
	altKey := mustKey(t, func() (identity.Key, error) { return identity.NewScenarioKey(placeKey, "alt") })
	happyKey := mustKey(t, func() (identity.Key, error) { return identity.NewScenarioKey(placeKey, "happy") })
	buyerKey := mustKey(t, func() (identity.Key, error) { return identity.NewScenarioObjectKey(happyKey, "buyer") })
	orderObjKey := mustKey(t, func() (identity.Key, error) { return identity.NewScenarioObjectKey(happyKey, "order") })

	order := model_class.NewClass(orderKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order"})
	order.SetEvents(map[identity.Key]model_state.Event{
		eventKey: model_state.NewEvent(eventKey, "place", "", nil),
	})
	order.SetQueries(map[identity.Key]model_state.Query{
		queryKey: model_state.NewQuery(queryKey, "Lookup", "", nil, nil, nil),
	})
	clerk := model_class.NewClass(clerkKey, model_class.ClassLinks{ActorKey: &clerkActorKey}, model_class.ClassDetails{Name: "Clerk"})
	customer := model_class.NewClass(customerKey, model_class.ClassLinks{ActorKey: &customerActorKey}, model_class.ClassDetails{Name: "Customer"})

	happy := model_scenario.NewScenario(happyKey, "Happy Path", "Main flow.")
	happy.SetObjects(map[identity.Key]model_scenario.Object{
		buyerKey:    model_scenario.NewObject(buyerKey, 2, model_scenario.ObjectDiagramName{Name: "Buyer", NameStyle: "name"}, customerKey, false, ""),
		orderObjKey: model_scenario.NewObject(orderObjKey, 1, model_scenario.ObjectDiagramName{Name: "42", NameStyle: "id"}, orderKey, true, "the order"),
	})
	happy.Steps = &model_scenario.Step{
		Key:      scenarioStepKey(t, happyKey, "root"),
		StepType: model_scenario.STEP_TYPE_SEQUENCE,
		Statements: []model_scenario.Step{
			{
				Key:       scenarioStepKey(t, happyKey, "loop"),
				StepType:  model_scenario.STEP_TYPE_LOOP,
				Condition: "a < b",
				Statements: []model_scenario.Step{
					leafStep(t, happyKey, "place", model_scenario.LEAF_TYPE_EVENT, "places", &buyerKey, &orderObjKey, &eventKey, nil, nil),
				},
			},
			{
				Key:      scenarioStepKey(t, happyKey, "switch"),
				StepType: model_scenario.STEP_TYPE_SWITCH,
				Statements: []model_scenario.Step{
					{
						Key:       scenarioStepKey(t, happyKey, "case_discount"),
						StepType:  model_scenario.STEP_TYPE_CASE,
						Condition: "discount",
						Statements: []model_scenario.Step{
							leafStep(t, happyKey, "lookup", model_scenario.LEAF_TYPE_QUERY, "looks up", &buyerKey, &orderObjKey, nil, &queryKey, nil),
						},
					},
				},
			},
			leafStep(t, happyKey, "alt", model_scenario.LEAF_TYPE_SCENARIO, "", &buyerKey, &orderObjKey, nil, nil, &altKey),
			leafStep(t, happyKey, "destroy", model_scenario.LEAF_TYPE_DESTROY, "gone", &orderObjKey, nil, nil, nil, nil),
		},
	}

	place := model_use_case.NewUseCase(placeKey, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSea}, model_use_case.GeneralizationRefs{SuperclassOfKey: &genKey}, model_use_case.UseCaseDetails{
		Name:       "Place Order",
		Details:    "Customer places an order.",
		UmlComment: "primary",
	})
	place.SetActors(map[identity.Key]model_use_case.Actor{
		clerkKey:    model_use_case.NewActor(""),
		customerKey: model_use_case.NewActor("buyer"),
	})
	place.SetScenarios(map[identity.Key]model_scenario.Scenario{
		altKey:   model_scenario.NewScenario(altKey, "Alt Path", ""),
		happyKey: happy,
	})
	rush := model_use_case.NewUseCase(rushKey, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelMud}, model_use_case.GeneralizationRefs{SubclassOfKey: &genKey}, model_use_case.UseCaseDetails{
		Name:            "Rush Order",
		UnfinishedNotes: "later",
	})
	signIn := model_use_case.NewUseCase(signInKey, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelMud, ReadOnly: true}, model_use_case.GeneralizationRefs{}, model_use_case.UseCaseDetails{Name: "Sign In"})

	subdomain := model_domain.NewSubdomain(subKey, "Default", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{clerkKey: clerk, customerKey: customer, orderKey: order}
	subdomain.UseCaseGeneralizations = map[identity.Key]model_use_case.Generalization{
		genKey: model_use_case.NewGeneralization(genKey, model_use_case.GeneralizationDetails{Name: "Checkout", Details: "How checkout varies."}, "", model_use_case.GeneralizationTraits{IsComplete: true}, "kinds"),
	}
	subdomain.UseCases = map[identity.Key]model_use_case.UseCase{placeKey: place, rushKey: rush, signInKey: signIn}
	subdomain.UseCaseShares = map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		placeKey: {
			rushKey:   model_use_case.NewUseCaseShared(model_use_case.ShareTypeInclude, "adds rush"),
			signInKey: model_use_case.NewUseCaseShared(model_use_case.ShareTypeExtend, ""),
		},
	}
	domain := model_domain.NewDomain(domainKey, "Commerce", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subKey: subdomain}

	model := core.NewModel("sample", core.ModelDetails{Name: "Sample"}, "", nil, nil, nil)
	model.Actors = map[identity.Key]model_actor.Actor{
		clerkActorKey:    model_actor.NewActor(clerkActorKey, "person", model_actor.GeneralizationRefs{}, model_actor.ActorDetails{Name: "Clerk"}),
		customerActorKey: model_actor.NewActor(customerActorKey, "person", model_actor.GeneralizationRefs{}, model_actor.ActorDetails{Name: "Customer"}),
	}
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	return model
}

func scenarioStepKey(t *testing.T, scenario identity.Key, sub string) identity.Key {
	t.Helper()
	return mustKey(t, func() (identity.Key, error) { return identity.NewScenarioStepKey(scenario, sub) })
}

func leafStep(t *testing.T, scenario identity.Key, sub, kind, description string, from, to, event, query, other *identity.Key) model_scenario.Step {
	t.Helper()
	leaf := kind
	return model_scenario.Step{
		Key:           scenarioStepKey(t, scenario, sub),
		StepType:      model_scenario.STEP_TYPE_LEAF,
		LeafType:      &leaf,
		Description:   description,
		FromObjectKey: from,
		ToObjectKey:   to,
		EventKey:      event,
		QueryKey:      query,
		ScenarioKey:   other,
	}
}
