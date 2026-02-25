package parser_ai

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// convErr creates a ParseError for conversion-time errors with the given error code, message, and file context.
func convErr(code int, msg, file string) *ParseError {
	return NewParseError(code, msg, file)
}

// ConvertToModel converts an inputModel to a req_model.Model.
// The input model is assumed to have been validated by readModelTree.
// This function performs the conversion and validates the resulting req_model.Model.
func ConvertToModel(input *inputModel, modelKey string) (*req_model.Model, error) {
	result := &req_model.Model{
		Key:                  strings.TrimSpace(strings.ToLower(modelKey)),
		Name:                 input.Name,
		Details:              input.Details,
		Actors:               make(map[identity.Key]model_actor.Actor),
		ActorGeneralizations: make(map[identity.Key]model_actor.Generalization),
		GlobalFunctions:      make(map[identity.Key]model_logic.GlobalFunction),
		Domains:              make(map[identity.Key]model_domain.Domain),
		DomainAssociations:   make(map[identity.Key]model_domain.Association),
		ClassAssociations:    make(map[identity.Key]model_class.Association),
	}

	// Convert invariants
	result.Invariants = convertInvariantsToModel(input.Invariants)

	// Convert global functions
	for key, gf := range input.GlobalFunctions {
		converted, err := convertGlobalFunctionToModel(key, gf)
		if err != nil {
			return nil, err
		}
		result.GlobalFunctions[converted.Key] = converted
	}

	// Convert actors
	for key, actor := range input.Actors {
		converted, err := convertActorToModel(key, actor, input.ActorGeneralizations)
		if err != nil {
			return nil, err
		}
		result.Actors[converted.Key] = converted
	}

	// Convert actor generalizations
	for key, gen := range input.ActorGeneralizations {
		converted, err := convertActorGeneralizationToModel(key, gen)
		if err != nil {
			return nil, err
		}
		result.ActorGeneralizations[converted.Key] = converted
	}

	// Convert domains
	for key, domain := range input.Domains {
		converted, err := convertDomainToModel(key, domain)
		if err != nil {
			return nil, err
		}
		result.Domains[converted.Key] = converted
	}

	// Convert domain associations
	for key, assoc := range input.DomainAssociations {
		converted, err := convertDomainAssocToModel(key, assoc)
		if err != nil {
			return nil, err
		}
		result.DomainAssociations[converted.Key] = converted
	}

	// Convert model-level class associations
	for key, assoc := range input.ClassAssociations {
		converted, err := convertModelAssociationToModel(key, assoc, result.Domains)
		if err != nil {
			return nil, err
		}
		result.ClassAssociations[converted.Key] = converted
	}

	// Validate the resulting model
	if err := result.Validate(); err != nil {
		return nil, convErr(
			ErrConvModelValidation,
			fmt.Sprintf("resulting model validation failed: %s", err.Error()),
			"model.json",
		)
	}

	return result, nil
}

// convertInvariantsToModel converts a slice of inputLogic to model_logic.Logic for model invariants.
func convertInvariantsToModel(invariants []inputLogic) []model_logic.Logic {
	if len(invariants) == 0 {
		return nil
	}
	result := make([]model_logic.Logic, len(invariants))
	for i, inv := range invariants {
		invKey, _ := identity.NewInvariantKey(fmt.Sprintf("%d", i))
		result[i] = model_logic.Logic{
			Key:           invKey,
			Type:          model_logic.LogicTypeAssessment,
			Description:   inv.Description,
			Notation:      inv.Notation,
			Specification: inv.Specification,
		}
	}
	return result
}

// convertGlobalFunctionToModel converts an inputGlobalFunction to a model_logic.GlobalFunction.
func convertGlobalFunctionToModel(keyStr string, gf *inputGlobalFunction) (model_logic.GlobalFunction, error) {
	key, err := identity.NewGlobalFunctionKey(keyStr)
	if err != nil {
		return model_logic.GlobalFunction{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create global function key '%s': %s", keyStr, err.Error()),
			fmt.Sprintf("global_functions/%s.gfunc.json", keyStr),
		).WithField("key")
	}

	logic := model_logic.Logic{
		Key:           key,
		Type:          model_logic.LogicTypeValue,
		Description:   gf.Logic.Description,
		Notation:      gf.Logic.Notation,
		Specification: gf.Logic.Specification,
	}

	return model_logic.GlobalFunction{
		Key:        key,
		Name:       gf.Name,
		Parameters: gf.Parameters,
		Logic:      logic,
	}, nil
}

// convertActorToModel converts an inputActor to a model_actor.Actor.
func convertActorToModel(keyStr string, actor *inputActor, actorGeneralizations map[string]*inputActorGeneralization) (model_actor.Actor, error) {
	key, err := identity.NewActorKey(keyStr)
	if err != nil {
		return model_actor.Actor{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create actor key '%s': %s", keyStr, err.Error()),
			fmt.Sprintf("actors/%s.actor.json", keyStr),
		).WithField("key")
	}

	result := model_actor.Actor{
		Key:        key,
		Name:       actor.Name,
		Type:       actor.Type,
		Details:    actor.Details,
		UmlComment: actor.UMLComment,
	}

	// Set SuperclassOfKey and SubclassOfKey from actor generalizations
	for genKeyStr, gen := range actorGeneralizations {
		genKey, err := identity.NewActorGeneralizationKey(genKeyStr)
		if err != nil {
			continue
		}
		if gen.SuperclassKey == keyStr {
			result.SuperclassOfKey = &genKey
		}
		for _, subKey := range gen.SubclassKeys {
			if subKey == keyStr {
				result.SubclassOfKey = &genKey
				break
			}
		}
	}

	return result, nil
}

// convertActorGeneralizationToModel converts an inputActorGeneralization to a model_actor.Generalization.
func convertActorGeneralizationToModel(keyStr string, gen *inputActorGeneralization) (model_actor.Generalization, error) {
	key, err := identity.NewActorGeneralizationKey(keyStr)
	if err != nil {
		return model_actor.Generalization{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create actor generalization key '%s': %s", keyStr, err.Error()),
			fmt.Sprintf("actor_generalizations/%s.agen.json", keyStr),
		).WithField("key")
	}

	return model_actor.Generalization{
		Key:        key,
		Name:       gen.Name,
		Details:    gen.Details,
		IsComplete: gen.IsComplete,
		IsStatic:   gen.IsStatic,
		UmlComment: gen.UMLComment,
	}, nil
}

// convertDomainAssocToModel converts an inputDomainAssociation to a model_domain.Association.
func convertDomainAssocToModel(keyStr string, assoc *inputDomainAssociation) (model_domain.Association, error) {
	assocFile := fmt.Sprintf("domain_associations/%s.domain_assoc.json", keyStr)

	problemDomainKey, err := identity.NewDomainKey(assoc.ProblemDomainKey)
	if err != nil {
		return model_domain.Association{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create problem domain key '%s': %s", assoc.ProblemDomainKey, err.Error()),
			assocFile,
		).WithField("problem_domain_key")
	}
	solutionDomainKey, err := identity.NewDomainKey(assoc.SolutionDomainKey)
	if err != nil {
		return model_domain.Association{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create solution domain key '%s': %s", assoc.SolutionDomainKey, err.Error()),
			assocFile,
		).WithField("solution_domain_key")
	}
	key, err := identity.NewDomainAssociationKey(problemDomainKey, solutionDomainKey)
	if err != nil {
		return model_domain.Association{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create domain association key: %s", err.Error()),
			assocFile,
		).WithField("key")
	}

	return model_domain.Association{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        assoc.UmlComment,
	}, nil
}

// convertDomainToModel converts an inputDomain to a model_domain.Domain.
func convertDomainToModel(keyStr string, domain *inputDomain) (model_domain.Domain, error) {
	domainFile := fmt.Sprintf("domains/%s/domain.json", keyStr)

	domainKey, err := identity.NewDomainKey(keyStr)
	if err != nil {
		return model_domain.Domain{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create domain key '%s': %s", keyStr, err.Error()),
			domainFile,
		).WithField("key")
	}

	result := model_domain.Domain{
		Key:               domainKey,
		Name:              domain.Name,
		Details:           domain.Details,
		Realized:          domain.Realized,
		UmlComment:        domain.UMLComment,
		Subdomains:        make(map[identity.Key]model_domain.Subdomain),
		ClassAssociations: make(map[identity.Key]model_class.Association),
	}

	// Convert subdomains
	for key, subdomain := range domain.Subdomains {
		converted, err := convertSubdomainToModel(key, subdomain, domainKey, keyStr)
		if err != nil {
			return model_domain.Domain{}, err
		}
		result.Subdomains[converted.Key] = converted
	}

	// Convert domain-level class associations
	for key, assoc := range domain.ClassAssociations {
		converted, err := convertDomainClassAssociationToModel(key, assoc, domainKey, result.Subdomains, keyStr)
		if err != nil {
			return model_domain.Domain{}, err
		}
		result.ClassAssociations[converted.Key] = converted
	}

	return result, nil
}

// convertSubdomainToModel converts an inputSubdomain to a model_domain.Subdomain.
func convertSubdomainToModel(keyStr string, subdomain *inputSubdomain, domainKey identity.Key, domainKeyStr string) (model_domain.Subdomain, error) {
	subdomainFile := fmt.Sprintf("domains/%s/subdomains/%s/subdomain.json", domainKeyStr, keyStr)

	subdomainKey, err := identity.NewSubdomainKey(domainKey, keyStr)
	if err != nil {
		return model_domain.Subdomain{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create subdomain key '%s': %s", keyStr, err.Error()),
			subdomainFile,
		).WithField("key")
	}

	result := model_domain.Subdomain{
		Key:                    subdomainKey,
		Name:                   subdomain.Name,
		Details:                subdomain.Details,
		UmlComment:             subdomain.UMLComment,
		Classes:                make(map[identity.Key]model_class.Class),
		Generalizations:        make(map[identity.Key]model_class.Generalization),
		ClassAssociations:      make(map[identity.Key]model_class.Association),
		UseCases:               make(map[identity.Key]model_use_case.UseCase),
		UseCaseGeneralizations: make(map[identity.Key]model_use_case.Generalization),
		UseCaseShares:          make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared),
	}

	// Convert class generalizations first to get the key mappings
	genKeyMap := make(map[string]identity.Key)
	for key, gen := range subdomain.ClassGeneralizations {
		converted, err := convertClassGeneralizationToModel(key, gen, subdomainKey, domainKeyStr, keyStr)
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		result.Generalizations[converted.Key] = converted
		genKeyMap[key] = converted.Key
	}

	// Convert classes
	for key, class := range subdomain.Classes {
		converted, err := convertClassToModel(key, class, subdomainKey, subdomain.ClassGeneralizations, genKeyMap, domainKeyStr, keyStr)
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		result.Classes[converted.Key] = converted
	}

	// Convert use case generalizations first to get the key mappings
	ucGenKeyMap := make(map[string]identity.Key)
	for key, gen := range subdomain.UseCaseGeneralizations {
		converted, err := convertUseCaseGeneralizationToModel(key, gen, subdomainKey, domainKeyStr, keyStr)
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		result.UseCaseGeneralizations[converted.Key] = converted
		ucGenKeyMap[key] = converted.Key
	}

	// Convert use cases
	for key, useCase := range subdomain.UseCases {
		converted, err := convertUseCaseToModel(key, useCase, subdomainKey, subdomain.UseCaseGeneralizations, ucGenKeyMap, result.Classes, domainKeyStr, keyStr)
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		result.UseCases[converted.Key] = converted
	}

	// Convert use case shares
	for seaKeyStr, mudMap := range subdomain.UseCaseShares {
		seaKey, err := identity.NewUseCaseKey(subdomainKey, seaKeyStr)
		if err != nil {
			return model_domain.Subdomain{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create use case key '%s': %s", seaKeyStr, err.Error()),
				subdomainFile,
			).WithField("use_case_shares")
		}
		innerMap := make(map[identity.Key]model_use_case.UseCaseShared)
		for mudKeyStr, shared := range mudMap {
			mudKey, err := identity.NewUseCaseKey(subdomainKey, mudKeyStr)
			if err != nil {
				return model_domain.Subdomain{}, convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create use case key '%s': %s", mudKeyStr, err.Error()),
					subdomainFile,
				).WithField("use_case_shares")
			}
			innerMap[mudKey] = model_use_case.UseCaseShared{
				ShareType:  shared.ShareType,
				UmlComment: shared.UmlComment,
			}
		}
		result.UseCaseShares[seaKey] = innerMap
	}

	// Convert subdomain-level class associations
	for key, assoc := range subdomain.ClassAssociations {
		converted, err := convertSubdomainAssociationToModel(key, assoc, subdomainKey, result.Classes, domainKeyStr, keyStr)
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		result.ClassAssociations[converted.Key] = converted
	}

	return result, nil
}

// convertUseCaseToModel converts an inputUseCase to a model_use_case.UseCase.
func convertUseCaseToModel(keyStr string, uc *inputUseCase, subdomainKey identity.Key, ucGeneralizations map[string]*inputUseCaseGeneralization, ucGenKeyMap map[string]identity.Key, classes map[identity.Key]model_class.Class, domainKeyStr, subdomainKeyStr string) (model_use_case.UseCase, error) {
	useCaseFile := fmt.Sprintf("domains/%s/subdomains/%s/use_cases/%s/use_case.json", domainKeyStr, subdomainKeyStr, keyStr)

	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, keyStr)
	if err != nil {
		return model_use_case.UseCase{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create use case key '%s': %s", keyStr, err.Error()),
			useCaseFile,
		).WithField("key")
	}

	result := model_use_case.UseCase{
		Key:        useCaseKey,
		Name:       uc.Name,
		Details:    uc.Details,
		Level:      uc.Level,
		ReadOnly:   uc.ReadOnly,
		UmlComment: uc.UMLComment,
		Actors:     make(map[identity.Key]model_use_case.Actor),
		Scenarios:  make(map[identity.Key]model_scenario.Scenario),
	}

	// Set SuperclassOfKey and SubclassOfKey from use case generalizations
	for genKeyStr, gen := range ucGeneralizations {
		genKey := ucGenKeyMap[genKeyStr]
		if gen.SuperclassKey == keyStr {
			result.SuperclassOfKey = &genKey
		}
		for _, subKey := range gen.SubclassKeys {
			if subKey == keyStr {
				result.SubclassOfKey = &genKey
				break
			}
		}
	}

	// Convert actors (keyed by class key string)
	for classKeyStr, actor := range uc.Actors {
		// Find the class key in our classes map
		var classKey identity.Key
		for key := range classes {
			if key.SubKey == classKeyStr {
				classKey = key
				break
			}
		}
		if classKey.SubKey == "" {
			// Try creating a class key from the subdomain
			classKey, err = identity.NewClassKey(subdomainKey, classKeyStr)
			if err != nil {
				return model_use_case.UseCase{}, convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create class key for use case actor '%s': %s", classKeyStr, err.Error()),
					useCaseFile,
				).WithField("actors")
			}
		}
		result.Actors[classKey] = model_use_case.Actor{
			UmlComment: actor.UmlComment,
		}
	}

	// Convert scenarios
	for scenKeyStr, scenario := range uc.Scenarios {
		converted, err := convertScenarioToModel(scenKeyStr, scenario, useCaseKey, subdomainKey, domainKeyStr, subdomainKeyStr, keyStr)
		if err != nil {
			return model_use_case.UseCase{}, err
		}
		result.Scenarios[converted.Key] = converted
	}

	return result, nil
}

// convertScenarioToModel converts an inputScenario to a model_scenario.Scenario.
func convertScenarioToModel(keyStr string, scenario *inputScenario, useCaseKey, subdomainKey identity.Key, domainKeyStr, subdomainKeyStr, useCaseKeyStr string) (model_scenario.Scenario, error) {
	scenarioFile := fmt.Sprintf("domains/%s/subdomains/%s/use_cases/%s/scenarios/%s.scenario.json", domainKeyStr, subdomainKeyStr, useCaseKeyStr, keyStr)

	scenarioKey, err := identity.NewScenarioKey(useCaseKey, keyStr)
	if err != nil {
		return model_scenario.Scenario{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create scenario key '%s': %s", keyStr, err.Error()),
			scenarioFile,
		).WithField("key")
	}

	result := model_scenario.Scenario{
		Key:     scenarioKey,
		Name:    scenario.Name,
		Details: scenario.Details,
		Objects: make(map[identity.Key]model_scenario.Object),
	}

	// Convert objects
	for objKeyStr, obj := range scenario.Objects {
		converted, err := convertScenarioObjectToModel(objKeyStr, obj, scenarioKey, subdomainKey, scenarioFile)
		if err != nil {
			return model_scenario.Scenario{}, err
		}
		result.Objects[converted.Key] = converted
	}

	// Convert steps
	if scenario.Steps != nil {
		converted, err := convertStepToModel(scenario.Steps, scenarioKey, useCaseKey, subdomainKey, scenario.Objects, 0, scenarioFile)
		if err != nil {
			return model_scenario.Scenario{}, err
		}
		result.Steps = converted
	}

	return result, nil
}

// convertScenarioObjectToModel converts an inputObject to a model_scenario.Object.
func convertScenarioObjectToModel(keyStr string, obj *inputObject, scenarioKey, subdomainKey identity.Key, scenarioFile string) (model_scenario.Object, error) {
	objKey, err := identity.NewScenarioObjectKey(scenarioKey, keyStr)
	if err != nil {
		return model_scenario.Object{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create scenario object key '%s': %s", keyStr, err.Error()),
			scenarioFile,
		).WithField(fmt.Sprintf("objects.%s", keyStr))
	}

	classKey, err := identity.NewClassKey(subdomainKey, obj.ClassKey)
	if err != nil {
		return model_scenario.Object{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create class key '%s' for scenario object '%s': %s", obj.ClassKey, keyStr, err.Error()),
			scenarioFile,
		).WithField(fmt.Sprintf("objects.%s.class_key", keyStr))
	}

	return model_scenario.Object{
		Key:          objKey,
		ObjectNumber: obj.ObjectNumber,
		Name:         obj.Name,
		NameStyle:    obj.NameStyle,
		ClassKey:     classKey,
		Multi:        obj.Multi,
		UmlComment:   obj.UmlComment,
	}, nil
}

// convertStepToModel recursively converts an inputStep to a model_scenario.Step.
func convertStepToModel(step *inputStep, scenarioKey, useCaseKey, subdomainKey identity.Key, objects map[string]*inputObject, stepIndex int, scenarioFile string) (*model_scenario.Step, error) {
	stepKey, err := identity.NewScenarioStepKey(scenarioKey, fmt.Sprintf("step_%d", stepIndex))
	if err != nil {
		return nil, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create scenario step key 'step_%d': %s", stepIndex, err.Error()),
			scenarioFile,
		).WithField(fmt.Sprintf("steps[%d]", stepIndex))
	}

	result := &model_scenario.Step{
		Key:         stepKey,
		StepType:    step.StepType,
		LeafType:    step.LeafType,
		Condition:   step.Condition,
		Description: step.Description,
	}

	// Convert reference keys for leaf steps
	if step.FromObjectKey != nil {
		objKey, err := identity.NewScenarioObjectKey(scenarioKey, *step.FromObjectKey)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create from object key '%s': %s", *step.FromObjectKey, err.Error()),
				scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].from_object_key", stepIndex))
		}
		result.FromObjectKey = &objKey
	}
	if step.ToObjectKey != nil {
		objKey, err := identity.NewScenarioObjectKey(scenarioKey, *step.ToObjectKey)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create to object key '%s': %s", *step.ToObjectKey, err.Error()),
				scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].to_object_key", stepIndex))
		}
		result.ToObjectKey = &objKey
	}
	if step.EventKey != nil {
		// Resolve the class key from the to_object (receiver) or from_object.
		classKey, err := resolveClassKeyFromStep(step, objects, subdomainKey, scenarioFile, stepIndex)
		if err != nil {
			return nil, err
		}
		eventKey, err := identity.NewEventKey(classKey, *step.EventKey)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create event key '%s': %s", *step.EventKey, err.Error()),
				scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].event_key", stepIndex))
		}
		result.EventKey = &eventKey
	}
	if step.QueryKey != nil {
		// Resolve the class key from the to_object (target) or from_object.
		classKey, err := resolveClassKeyFromStep(step, objects, subdomainKey, scenarioFile, stepIndex)
		if err != nil {
			return nil, err
		}
		queryKey, err := identity.NewQueryKey(classKey, *step.QueryKey)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create query key '%s': %s", *step.QueryKey, err.Error()),
				scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].query_key", stepIndex))
		}
		result.QueryKey = &queryKey
	}
	if step.ScenarioKey != nil {
		scenKey, err := identity.NewScenarioKey(useCaseKey, *step.ScenarioKey)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create scenario key '%s': %s", *step.ScenarioKey, err.Error()),
				scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].scenario_key", stepIndex))
		}
		result.ScenarioKey = &scenKey
	}

	// Convert sub-statements
	for i, subStep := range step.Statements {
		subStepCopy := subStep
		converted, err := convertStepToModel(&subStepCopy, scenarioKey, useCaseKey, subdomainKey, objects, stepIndex*100+i+1, scenarioFile)
		if err != nil {
			return nil, err
		}
		result.Statements = append(result.Statements, *converted)
	}

	return result, nil
}

// resolveClassKeyFromStep resolves the class identity.Key from a step's object references.
// It checks to_object first (receiver/target), then from_object.
func resolveClassKeyFromStep(step *inputStep, objects map[string]*inputObject, subdomainKey identity.Key, scenarioFile string, stepIndex int) (identity.Key, error) {
	var classKeyStr string
	if step.ToObjectKey != nil {
		if obj, ok := objects[*step.ToObjectKey]; ok {
			classKeyStr = obj.ClassKey
		}
	}
	if classKeyStr == "" && step.FromObjectKey != nil {
		if obj, ok := objects[*step.FromObjectKey]; ok {
			classKeyStr = obj.ClassKey
		}
	}
	if classKeyStr == "" {
		return identity.Key{}, convErr(
			ErrConvObjectResolveFailed,
			"no object reference to resolve class key - step must have from_object_key or to_object_key",
			scenarioFile,
		).WithField(fmt.Sprintf("steps[%d]", stepIndex))
	}
	classKey, err := identity.NewClassKey(subdomainKey, classKeyStr)
	if err != nil {
		return identity.Key{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create class key '%s' from step object: %s", classKeyStr, err.Error()),
			scenarioFile,
		).WithField(fmt.Sprintf("steps[%d]", stepIndex))
	}
	return classKey, nil
}

// convertUseCaseGeneralizationToModel converts an inputUseCaseGeneralization to a model_use_case.Generalization.
func convertUseCaseGeneralizationToModel(keyStr string, gen *inputUseCaseGeneralization, subdomainKey identity.Key, domainKeyStr, subdomainKeyStr string) (model_use_case.Generalization, error) {
	genFile := fmt.Sprintf("domains/%s/subdomains/%s/use_case_generalizations/%s.ucgen.json", domainKeyStr, subdomainKeyStr, keyStr)

	key, err := identity.NewUseCaseGeneralizationKey(subdomainKey, keyStr)
	if err != nil {
		return model_use_case.Generalization{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create use case generalization key '%s': %s", keyStr, err.Error()),
			genFile,
		).WithField("key")
	}

	return model_use_case.Generalization{
		Key:        key,
		Name:       gen.Name,
		Details:    gen.Details,
		IsComplete: gen.IsComplete,
		IsStatic:   gen.IsStatic,
		UmlComment: gen.UMLComment,
	}, nil
}

// convertClassToModel converts an inputClass to a model_class.Class.
func convertClassToModel(keyStr string, class *inputClass, subdomainKey identity.Key, generalizations map[string]*inputClassGeneralization, genKeyMap map[string]identity.Key, domainKeyStr, subdomainKeyStr string) (model_class.Class, error) {
	classFile := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKeyStr, subdomainKeyStr, keyStr)

	classKey, err := identity.NewClassKey(subdomainKey, keyStr)
	if err != nil {
		return model_class.Class{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create class key '%s': %s", keyStr, err.Error()),
			classFile,
		).WithField("key")
	}

	result := model_class.Class{
		Key:         classKey,
		Name:        class.Name,
		Details:     class.Details,
		UmlComment:  class.UMLComment,
		Attributes:  make(map[identity.Key]model_class.Attribute),
		States:      make(map[identity.Key]model_state.State),
		Events:      make(map[identity.Key]model_state.Event),
		Guards:      make(map[identity.Key]model_state.Guard),
		Actions:     make(map[identity.Key]model_state.Action),
		Queries:     make(map[identity.Key]model_state.Query),
		Transitions: make(map[identity.Key]model_state.Transition),
	}

	// Set actor key if present
	if class.ActorKey != "" {
		actorKey, err := identity.NewActorKey(class.ActorKey)
		if err != nil {
			return model_class.Class{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create actor key reference '%s': %s", class.ActorKey, err.Error()),
				classFile,
			).WithField("actor_key")
		}
		result.ActorKey = &actorKey
	}

	// Find generalization references for this class
	for genKeyStr, gen := range generalizations {
		genKey := genKeyMap[genKeyStr]
		if gen.SuperclassKey == keyStr {
			result.SuperclassOfKey = &genKey
		}
		for _, subclassKey := range gen.SubclassKeys {
			if subclassKey == keyStr {
				result.SubclassOfKey = &genKey
				break
			}
		}
	}

	// Convert attributes with index tracking
	for attrKeyStr, attr := range class.Attributes {
		converted, err := convertAttributeToModel(attrKeyStr, attr, classKey, class.Indexes, classFile)
		if err != nil {
			return model_class.Class{}, err
		}
		result.Attributes[converted.Key] = converted
	}

	// Convert class invariants
	result.SetInvariants(convertLogicsToModel(class.Invariants, model_logic.LogicTypeAssessment, classKey, identity.NewClassInvariantKey))

	// Convert state machine if present
	if class.StateMachine != nil {
		if err := convertStateMachineToModel(class.StateMachine, class.Actions, &result, classKey, domainKeyStr, subdomainKeyStr, keyStr); err != nil {
			return model_class.Class{}, err
		}
	}

	// Convert actions
	for actionKeyStr, action := range class.Actions {
		converted, err := convertActionToModel(actionKeyStr, action, classKey, domainKeyStr, subdomainKeyStr, keyStr)
		if err != nil {
			return model_class.Class{}, err
		}
		result.Actions[converted.Key] = converted
	}

	// Convert queries
	for queryKeyStr, query := range class.Queries {
		converted, err := convertQueryToModel(queryKeyStr, query, classKey, domainKeyStr, subdomainKeyStr, keyStr)
		if err != nil {
			return model_class.Class{}, err
		}
		result.Queries[converted.Key] = converted
	}

	return result, nil
}

// convertAttributeToModel converts an inputAttribute to a model_class.Attribute.
func convertAttributeToModel(keyStr string, attr *inputAttribute, classKey identity.Key, indexes [][]string, classFile string) (model_class.Attribute, error) {
	attrKey, err := identity.NewAttributeKey(classKey, keyStr)
	if err != nil {
		return model_class.Attribute{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create attribute key '%s': %s", keyStr, err.Error()),
			classFile,
		).WithField(fmt.Sprintf("attributes.%s", keyStr))
	}

	// Find which indexes this attribute is part of
	var indexNums []uint
	for i, index := range indexes {
		for _, attrKeyInIndex := range index {
			if attrKeyInIndex == keyStr {
				indexNums = append(indexNums, uint(i))
				break
			}
		}
	}

	result := model_class.Attribute{
		Key:           attrKey,
		Name:          attr.Name,
		DataTypeRules: attr.DataTypeRules,
		Details:       attr.Details,
		Nullable:      attr.Nullable,
		UmlComment:    attr.UMLComment,
		IndexNums:     indexNums,
	}
	if attr.DerivationPolicy != nil {
		dpKey, err := identity.NewAttributeDerivationKey(attrKey, "derivation")
		if err != nil {
			return model_class.Attribute{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create derivation key for attribute '%s': %s", keyStr, err.Error()),
				classFile,
			).WithField(fmt.Sprintf("attributes.%s.derivation_policy", keyStr))
		}
		dp := model_logic.Logic{
			Key:           dpKey,
			Type:          model_logic.LogicTypeValue,
			Description:   attr.DerivationPolicy.Description,
			Notation:      attr.DerivationPolicy.Notation,
			Specification: attr.DerivationPolicy.Specification,
		}
		result.DerivationPolicy = &dp
	}
	return result, nil
}

// convertStateMachineToModel converts an inputStateMachine to populate a Class's state machine fields.
func convertStateMachineToModel(sm *inputStateMachine, actions map[string]*inputAction, class *model_class.Class, classKey identity.Key, domainKeyStr, subdomainKeyStr, classKeyStr string) error {
	smFile := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKeyStr, subdomainKeyStr, classKeyStr)

	// Convert states
	for stateKeyStr, state := range sm.States {
		stateKey, err := identity.NewStateKey(classKey, stateKeyStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create state key '%s': %s", stateKeyStr, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("states.%s", stateKeyStr))
		}

		converted := model_state.State{
			Key:        stateKey,
			Name:       state.Name,
			Details:    state.Details,
			UmlComment: state.UMLComment,
			Actions:    []model_state.StateAction{},
		}

		// Convert state actions
		for _, stateAction := range state.Actions {
			actionKey, err := identity.NewActionKey(classKey, stateAction.ActionKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create action key reference '%s' in state '%s': %s", stateAction.ActionKey, stateKeyStr, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("states.%s.actions", stateKeyStr))
			}
			stateActionKey, err := identity.NewStateActionKey(stateKey, stateAction.When, stateAction.ActionKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create state action key for '%s/%s' in state '%s': %s", stateAction.When, stateAction.ActionKey, stateKeyStr, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("states.%s.actions", stateKeyStr))
			}
			converted.Actions = append(converted.Actions, model_state.StateAction{
				Key:       stateActionKey,
				ActionKey: actionKey,
				When:      stateAction.When,
			})
		}

		class.States[converted.Key] = converted
	}

	// Convert events
	for eventKeyStr, event := range sm.Events {
		eventKey, err := identity.NewEventKey(classKey, eventKeyStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create event key '%s': %s", eventKeyStr, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("events.%s", eventKeyStr))
		}

		converted := model_state.Event{
			Key:        eventKey,
			Name:       event.Name,
			Details:    event.Details,
			Parameters: []model_state.Parameter{},
		}

		// Convert event parameters
		for _, param := range event.Parameters {
			converted.Parameters = append(converted.Parameters, model_state.Parameter{
				Name:          param.Name,
				DataTypeRules: param.DataTypeRules,
			})
		}

		class.Events[converted.Key] = converted
	}

	// Convert guards
	for guardKeyStr, guard := range sm.Guards {
		guardKey, err := identity.NewGuardKey(classKey, guardKeyStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create guard key '%s': %s", guardKeyStr, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("guards.%s", guardKeyStr))
		}

		converted := model_state.Guard{
			Key:  guardKey,
			Name: guard.Name,
			Logic: model_logic.Logic{
				Key:           guardKey,
				Type:          model_logic.LogicTypeAssessment,
				Description:   guard.Logic.Description,
				Notation:      guard.Logic.Notation,
				Specification: guard.Logic.Specification,
			},
		}

		class.Guards[converted.Key] = converted
	}

	// Convert transitions
	for i, transition := range sm.Transitions {
		// Determine from and to state keys
		var fromStr, toStr string
		if transition.FromStateKey != nil {
			fromStr = *transition.FromStateKey
		}
		if transition.ToStateKey != nil {
			toStr = *transition.ToStateKey
		}

		// Get guard and action keys as strings
		var guardStr, actionStr string
		if transition.GuardKey != nil {
			guardStr = *transition.GuardKey
		}
		if transition.ActionKey != nil {
			actionStr = *transition.ActionKey
		}

		transitionKey, err := identity.NewTransitionKey(classKey, fromStr, transition.EventKey, guardStr, actionStr, toStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create transition key for transition[%d]: %s", i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d]", i))
		}

		converted := model_state.Transition{
			Key:        transitionKey,
			UmlComment: transition.UMLComment,
		}

		// Set event key (required)
		eventKey, err := identity.NewEventKey(classKey, transition.EventKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create event key reference '%s' for transition[%d]: %s", transition.EventKey, i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d].event_key", i))
		}
		converted.EventKey = eventKey

		// Set from state key (optional)
		if transition.FromStateKey != nil {
			stateKey, err := identity.NewStateKey(classKey, *transition.FromStateKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create from state key reference '%s' for transition[%d]: %s", *transition.FromStateKey, i, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("transitions[%d].from_state_key", i))
			}
			converted.FromStateKey = &stateKey
		}

		// Set to state key (optional)
		if transition.ToStateKey != nil {
			stateKey, err := identity.NewStateKey(classKey, *transition.ToStateKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create to state key reference '%s' for transition[%d]: %s", *transition.ToStateKey, i, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("transitions[%d].to_state_key", i))
			}
			converted.ToStateKey = &stateKey
		}

		// Set guard key (optional)
		if transition.GuardKey != nil {
			guardKey, err := identity.NewGuardKey(classKey, *transition.GuardKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create guard key reference '%s' for transition[%d]: %s", *transition.GuardKey, i, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("transitions[%d].guard_key", i))
			}
			converted.GuardKey = &guardKey
		}

		// Set action key (optional)
		if transition.ActionKey != nil {
			actionKey, err := identity.NewActionKey(classKey, *transition.ActionKey)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create action key reference '%s' for transition[%d]: %s", *transition.ActionKey, i, err.Error()),
					smFile,
				).WithField(fmt.Sprintf("transitions[%d].action_key", i))
			}
			converted.ActionKey = &actionKey
		}

		class.Transitions[converted.Key] = converted
	}

	return nil
}

// convertActionToModel converts an inputAction to a model_state.Action.
func convertActionToModel(keyStr string, action *inputAction, classKey identity.Key, domainKeyStr, subdomainKeyStr, classKeyStr string) (model_state.Action, error) {
	actionFile := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/actions/%s.json", domainKeyStr, subdomainKeyStr, classKeyStr, keyStr)

	actionKey, err := identity.NewActionKey(classKey, keyStr)
	if err != nil {
		return model_state.Action{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create action key '%s': %s", keyStr, err.Error()),
			actionFile,
		).WithField("key")
	}

	return model_state.Action{
		Key:         actionKey,
		Name:        action.Name,
		Details:     action.Details,
		Parameters:  convertParametersToModel(action.Parameters),
		Requires:    convertLogicsToModel(action.Requires, model_logic.LogicTypeAssessment, actionKey, identity.NewActionRequireKey),
		Guarantees:  convertLogicsToModel(action.Guarantees, model_logic.LogicTypeStateChange, actionKey, identity.NewActionGuaranteeKey),
		SafetyRules: convertLogicsToModel(action.SafetyRules, model_logic.LogicTypeSafetyRule, actionKey, identity.NewActionSafetyKey),
	}, nil
}

// convertQueryToModel converts an inputQuery to a model_state.Query.
func convertQueryToModel(keyStr string, query *inputQuery, classKey identity.Key, domainKeyStr, subdomainKeyStr, classKeyStr string) (model_state.Query, error) {
	queryFile := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/queries/%s.json", domainKeyStr, subdomainKeyStr, classKeyStr, keyStr)

	queryKey, err := identity.NewQueryKey(classKey, keyStr)
	if err != nil {
		return model_state.Query{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create query key '%s': %s", keyStr, err.Error()),
			queryFile,
		).WithField("key")
	}

	return model_state.Query{
		Key:        queryKey,
		Name:       query.Name,
		Details:    query.Details,
		Parameters: convertParametersToModel(query.Parameters),
		Requires:   convertLogicsToModel(query.Requires, model_logic.LogicTypeAssessment, queryKey, identity.NewQueryRequireKey),
		Guarantees: convertLogicsToModel(query.Guarantees, model_logic.LogicTypeQuery, queryKey, identity.NewQueryGuaranteeKey),
	}, nil
}

// convertLogicToModel converts an inputLogic to a model_logic.Logic with the given key.
func convertLogicToModel(input *inputLogic, logicType string, parentKey identity.Key) model_logic.Logic {
	return model_logic.Logic{
		Key:           parentKey,
		Type:          logicType,
		Description:   input.Description,
		Notation:      input.Notation,
		Specification: input.Specification,
	}
}

// convertLogicsToModel converts a slice of inputLogic to a slice of model_logic.Logic.
// keyFactory creates the identity key for each logic entry using the parent key and an index-based sub-key.
func convertLogicsToModel(logics []inputLogic, logicType string, parentKey identity.Key, keyFactory func(identity.Key, string) (identity.Key, error)) []model_logic.Logic {
	if len(logics) == 0 {
		return nil
	}
	result := make([]model_logic.Logic, len(logics))
	for i, logic := range logics {
		logicKey, _ := keyFactory(parentKey, fmt.Sprintf("%d", i))
		result[i] = model_logic.Logic{
			Key:           logicKey,
			Type:          logicType,
			Description:   logic.Description,
			Notation:      logic.Notation,
			Specification: logic.Specification,
		}
	}
	return result
}

// convertParametersToModel converts a slice of inputParameter to a slice of model_state.Parameter.
func convertParametersToModel(params []inputParameter) []model_state.Parameter {
	if len(params) == 0 {
		return nil
	}
	result := make([]model_state.Parameter, len(params))
	for i, param := range params {
		result[i] = model_state.Parameter{
			Name:          param.Name,
			DataTypeRules: param.DataTypeRules,
		}
	}
	return result
}

// convertClassGeneralizationToModel converts an inputClassGeneralization to a model_class.Generalization.
func convertClassGeneralizationToModel(keyStr string, gen *inputClassGeneralization, subdomainKey identity.Key, domainKeyStr, subdomainKeyStr string) (model_class.Generalization, error) {
	genFile := fmt.Sprintf("domains/%s/subdomains/%s/class_generalizations/%s.cgen.json", domainKeyStr, subdomainKeyStr, keyStr)

	genKey, err := identity.NewGeneralizationKey(subdomainKey, keyStr)
	if err != nil {
		return model_class.Generalization{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create class generalization key '%s': %s", keyStr, err.Error()),
			genFile,
		).WithField("key")
	}

	return model_class.Generalization{
		Key:        genKey,
		Name:       gen.Name,
		Details:    gen.Details,
		IsComplete: gen.IsComplete,
		IsStatic:   gen.IsStatic,
		UmlComment: gen.UMLComment,
	}, nil
}

// convertSubdomainAssociationToModel converts an inputClassAssociation at subdomain level to a model_class.Association.
func convertSubdomainAssociationToModel(keyStr string, assoc *inputClassAssociation, subdomainKey identity.Key, classes map[identity.Key]model_class.Class, domainKeyStr, subdomainKeyStr string) (model_class.Association, error) {
	assocFile := fmt.Sprintf("domains/%s/subdomains/%s/associations/%s.assoc.json", domainKeyStr, subdomainKeyStr, keyStr)

	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for key := range classes {
		if key.SubKey == assoc.FromClassKey {
			fromClassKey = key
		}
		if key.SubKey == assoc.ToClassKey {
			toClassKey = key
		}
	}

	if fromClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("from_class_key '%s' not found in subdomain", assoc.FromClassKey),
			assocFile,
		).WithField("from_class_key")
	}
	if toClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("to_class_key '%s' not found in subdomain", assoc.ToClassKey),
			assocFile,
		).WithField("to_class_key")
	}

	assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvAssocKeyConstruction,
			fmt.Sprintf("failed to create association key for '%s': %s", keyStr, err.Error()),
			assocFile,
		).WithField("key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse from_multiplicity '%s': %s", assoc.FromMultiplicity, err.Error()),
			assocFile,
		).WithField("from_multiplicity")
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse to_multiplicity '%s': %s", assoc.ToMultiplicity, err.Error()),
			assocFile,
		).WithField("to_multiplicity")
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	// Handle association class key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		for key := range classes {
			if key.SubKey == *assoc.AssociationClassKey {
				result.AssociationClassKey = &key
				break
			}
		}
	}

	return result, nil
}

// convertDomainClassAssociationToModel converts an inputClassAssociation at domain level to a model_class.Association.
func convertDomainClassAssociationToModel(keyStr string, assoc *inputClassAssociation, domainKey identity.Key, subdomains map[identity.Key]model_domain.Subdomain, domainKeyStr string) (model_class.Association, error) {
	assocFile := fmt.Sprintf("domains/%s/associations/%s.assoc.json", domainKeyStr, keyStr)

	// Parse subdomain/class format
	fromSubdomain, fromClass, err := parseDomainScopedKey(assoc.FromClassKey)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvScopedKeyInvalid,
			fmt.Sprintf("failed to parse from_class_key '%s': %s", assoc.FromClassKey, err.Error()),
			assocFile,
		).WithField("from_class_key")
	}
	toSubdomain, toClass, err := parseDomainScopedKey(assoc.ToClassKey)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvScopedKeyInvalid,
			fmt.Sprintf("failed to parse to_class_key '%s': %s", assoc.ToClassKey, err.Error()),
			assocFile,
		).WithField("to_class_key")
	}

	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for subKey, subdomain := range subdomains {
		if subKey.SubKey == fromSubdomain {
			for classKey := range subdomain.Classes {
				if classKey.SubKey == fromClass {
					fromClassKey = classKey
					break
				}
			}
		}
		if subKey.SubKey == toSubdomain {
			for classKey := range subdomain.Classes {
				if classKey.SubKey == toClass {
					toClassKey = classKey
					break
				}
			}
		}
	}

	if fromClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("from_class_key '%s' not found in domain", assoc.FromClassKey),
			assocFile,
		).WithField("from_class_key")
	}
	if toClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("to_class_key '%s' not found in domain", assoc.ToClassKey),
			assocFile,
		).WithField("to_class_key")
	}

	assocKey, err := identity.NewClassAssociationKey(domainKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvAssocKeyConstruction,
			fmt.Sprintf("failed to create association key for '%s': %s", keyStr, err.Error()),
			assocFile,
		).WithField("key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse from_multiplicity '%s': %s", assoc.FromMultiplicity, err.Error()),
			assocFile,
		).WithField("from_multiplicity")
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse to_multiplicity '%s': %s", assoc.ToMultiplicity, err.Error()),
			assocFile,
		).WithField("to_multiplicity")
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	return result, nil
}

// convertModelAssociationToModel converts an inputClassAssociation at model level to a model_class.Association.
func convertModelAssociationToModel(keyStr string, assoc *inputClassAssociation, domains map[identity.Key]model_domain.Domain) (model_class.Association, error) {
	assocFile := fmt.Sprintf("associations/%s.assoc.json", keyStr)

	// Parse domain/subdomain/class format
	fromDomain, fromSubdomain, fromClass, err := parseModelScopedKey(assoc.FromClassKey)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvScopedKeyInvalid,
			fmt.Sprintf("failed to parse from_class_key '%s': %s", assoc.FromClassKey, err.Error()),
			assocFile,
		).WithField("from_class_key")
	}
	toDomain, toSubdomain, toClass, err := parseModelScopedKey(assoc.ToClassKey)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvScopedKeyInvalid,
			fmt.Sprintf("failed to parse to_class_key '%s': %s", assoc.ToClassKey, err.Error()),
			assocFile,
		).WithField("to_class_key")
	}

	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for domKey, domain := range domains {
		if domKey.SubKey == fromDomain {
			for subKey, subdomain := range domain.Subdomains {
				if subKey.SubKey == fromSubdomain {
					for classKey := range subdomain.Classes {
						if classKey.SubKey == fromClass {
							fromClassKey = classKey
							break
						}
					}
				}
			}
		}
		if domKey.SubKey == toDomain {
			for subKey, subdomain := range domain.Subdomains {
				if subKey.SubKey == toSubdomain {
					for classKey := range subdomain.Classes {
						if classKey.SubKey == toClass {
							toClassKey = classKey
							break
						}
					}
				}
			}
		}
	}

	if fromClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("from_class_key '%s' not found in model", assoc.FromClassKey),
			assocFile,
		).WithField("from_class_key")
	}
	if toClassKey.SubKey == "" {
		return model_class.Association{}, convErr(
			ErrConvClassNotFound,
			fmt.Sprintf("to_class_key '%s' not found in model", assoc.ToClassKey),
			assocFile,
		).WithField("to_class_key")
	}

	// For model-level associations, parent key is empty
	emptyKey := identity.Key{}
	assocKey, err := identity.NewClassAssociationKey(emptyKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvAssocKeyConstruction,
			fmt.Sprintf("failed to create association key for '%s': %s", keyStr, err.Error()),
			assocFile,
		).WithField("key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse from_multiplicity '%s': %s", assoc.FromMultiplicity, err.Error()),
			assocFile,
		).WithField("from_multiplicity")
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, convErr(
			ErrConvMultiplicityInvalid,
			fmt.Sprintf("failed to parse to_multiplicity '%s': %s", assoc.ToMultiplicity, err.Error()),
			assocFile,
		).WithField("to_multiplicity")
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	return result, nil
}

// normalizeMultiplicity converts user-friendly multiplicity strings to the format expected by model_class.NewMultiplicity.
// "*" -> "any", "1..*" -> "1..many", etc.
func normalizeMultiplicity(mult string) string {
	// Handle standalone "*"
	if mult == "*" {
		return "any"
	}
	// Handle "n..*" patterns -> "n..many"
	if strings.HasSuffix(mult, "..*") {
		return strings.TrimSuffix(mult, "..*") + "..many"
	}
	return mult
}
