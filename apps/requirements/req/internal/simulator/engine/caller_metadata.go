package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
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
	populateAssociationSetAddSenders(model, catalog)
	populateAssociationSetMapSenders(model, catalog)
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

func populateAssociationSetAddSenders(model *core.Model, catalog *ClassCatalog) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			assocByKey := subdomain.ClassAssociations
			for _, class := range subdomain.Classes {
				recordAssociationSetAddSenders(class, assocByKey, catalog)
			}
		}
	}
}

func recordAssociationSetAddSenders(class model_class.Class, associations map[identity.Key]model_class.Association, catalog *ClassCatalog) {
	for _, action := range class.Actions {
		for _, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet || guar.Target == "" {
				continue
			}
			if !model_class.IsAssociationSetAddSpecification(guar.Spec.Specification) {
				continue
			}
			toClassKey, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations)
			if !ok {
				continue
			}
			toInfo := catalog.GetClassInfo(toClassKey)
			if toInfo == nil {
				continue
			}
			for _, ev := range toInfo.CreationEvents {
				catalog.addEventSender(ev.Key, class.Key)
			}
		}
	}
}

func populateAssociationSetMapSenders(model *core.Model, catalog *ClassCatalog) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			assocByKey := subdomain.ClassAssociations
			for _, class := range subdomain.Classes {
				recordAssociationSetMapSenders(class, assocByKey, catalog)
			}
		}
	}
}

func recordAssociationSetMapSenders(class model_class.Class, associations map[identity.Key]model_class.Association, catalog *ClassCatalog) {
	for _, action := range class.Actions {
		for _, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet || guar.Target == "" {
				continue
			}
			if guar.Type == model_logic.LogicTypeDelete {
				eventKey, ok := model_class.AssociationDestroyEventKey(guar)
				if !ok {
					continue
				}
				if _, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations); !ok {
					continue
				}
				catalog.addEventSender(eventKey, class.Key)
				continue
			}
			spec := guar.Spec.Specification
			if !model_class.IsAssociationSetMapSpecification(spec) && !model_class.IsAssociationAddOrUpdateSpecification(spec) {
				continue
			}
			if _, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations); !ok {
				continue
			}
			eventKey, ok := model_class.AssociationSetMapEventKey(guar.Spec.Expression)
			if !ok {
				continue
			}
			catalog.addEventSender(eventKey, class.Key)
		}
	}
}

func associationToClassForSetAddTarget(
	fromClassKey identity.Key,
	target string,
	associations map[identity.Key]model_class.Association,
) (identity.Key, bool) {
	for _, assoc := range associations {
		if assoc.FromClassKey != fromClassKey {
			continue
		}
		if model_class.AssociationTLAFieldName(assoc.Name) != target {
			continue
		}
		return assoc.ToClassKey, true
	}
	return identity.Key{}, false
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
