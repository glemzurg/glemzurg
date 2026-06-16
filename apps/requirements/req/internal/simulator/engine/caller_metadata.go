package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// PopulateCallerDataFromModel records SentBy/CalledBy metadata from use-case
// scenarios and mandatory association creation chains.
func PopulateCallerDataFromModel(model *core.Model, catalog *ClassCatalog) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, useCase := range subdomain.UseCases {
				populateCallerDataFromUseCase(useCase, catalog)
			}
		}
	}
	populateMandatoryAssociationSenders(catalog)
}

func populateCallerDataFromUseCase(useCase model_use_case.UseCase, catalog *ClassCatalog) {
	for _, scenario := range useCase.Scenarios {
		if scenario.Steps == nil {
			continue
		}
		walkScenarioStep(scenario.Steps, scenario.Objects, catalog)
	}
}

func walkScenarioStep(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	catalog *ClassCatalog,
) {
	if step == nil {
		return
	}

	if step.StepType == model_scenario.STEP_TYPE_LEAF && step.LeafType != nil {
		switch *step.LeafType {
		case model_scenario.LEAF_TYPE_EVENT:
			recordScenarioEventSender(step, objects, catalog)
		case model_scenario.LEAF_TYPE_QUERY:
			recordScenarioQueryCaller(step, objects, catalog)
		}
		return
	}

	for i := range step.Statements {
		walkScenarioStep(&step.Statements[i], objects, catalog)
	}
}

func recordScenarioEventSender(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	catalog *ClassCatalog,
) {
	if step.FromObjectKey == nil || step.EventKey == nil {
		return
	}
	obj, ok := objects[*step.FromObjectKey]
	if !ok {
		return
	}
	catalog.addEventSender(*step.EventKey, obj.ClassKey)
}

func recordScenarioQueryCaller(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	catalog *ClassCatalog,
) {
	if step.FromObjectKey == nil || step.QueryKey == nil {
		return
	}
	obj, ok := objects[*step.FromObjectKey]
	if !ok {
		return
	}
	catalog.addQueryCaller(*step.QueryKey, obj.ClassKey)
}

func populateMandatoryAssociationSenders(catalog *ClassCatalog) {
	for _, ai := range catalog.AllAssociations() {
		if !ai.MandatoryTo {
			continue
		}
		toInfo := catalog.GetClassInfo(ai.ToClassKey)
		if toInfo == nil {
			continue
		}
		for _, ev := range toInfo.CreationEvents {
			catalog.addEventSender(ev.Key, ai.FromClassKey)
		}
	}
}
