package parser_ai

import (
	"fmt"
	"slices"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// convErr creates a ParseError for conversion-time errors with the given error code, message, and file context.
func convErr(code int, msg, file string) *ParseError {
	return NewParseError(code, msg, file)
}

// ConvertToModel converts an inputModel to a core.Model.
// The input model is assumed to have been validated by readModelTree.
// This function performs the conversion and validates the resulting core.Model.
func ConvertToModel(input *inputModel, modelKey string) (*core.Model, error) {
	result, err := convertModelScalars(input, modelKey)
	if err != nil {
		return nil, err
	}

	if err := convertActorsToModel(input, result); err != nil {
		return nil, err
	}

	if err := convertDomainsAndAssociationsToModel(input, result); err != nil {
		return nil, err
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

// convertModelScalars converts invariants, global functions, and named sets to create the initial core.Model.
func convertModelScalars(input *inputModel, modelKey string) (*core.Model, error) {
	invariants, err := convertInvariantsToModel(input.Invariants)
	if err != nil {
		return nil, err
	}

	globalFunctions, err := convertGlobalFunctionsMap(input.GlobalFunctions)
	if err != nil {
		return nil, err
	}

	namedSets, err := convertNamedSetsMap(input.NamedSets)
	if err != nil {
		return nil, err
	}

	return &core.Model{
		Key:                  strings.TrimSpace(strings.ToLower(modelKey)),
		Name:                 input.Name,
		Details:              input.Details,
		Invariants:           invariants,
		GlobalFunctions:      globalFunctions,
		NamedSets:            namedSets,
		Actors:               make(map[identity.Key]model_actor.Actor),
		ActorGeneralizations: make(map[identity.Key]model_actor.Generalization),
		Domains:              make(map[identity.Key]model_domain.Domain),
		DomainAssociations:   make(map[identity.Key]model_domain.Association),
		ClassAssociations:    make(map[identity.Key]model_class.Association),
	}, nil
}

// convertGlobalFunctionsMap converts the global functions map.
func convertGlobalFunctionsMap(input map[string]*inputGlobalFunction) (map[identity.Key]model_logic.GlobalFunction, error) {
	globalFunctions := make(map[identity.Key]model_logic.GlobalFunction)
	for key, gf := range input {
		converted, err := convertGlobalFunctionToModel(key, gf)
		if err != nil {
			return nil, err
		}
		globalFunctions[converted.Key] = converted
	}
	return globalFunctions, nil
}

// convertNamedSetsMap converts the named sets map.
func convertNamedSetsMap(input map[string]*inputNamedSet) (map[identity.Key]model_named_set.NamedSet, error) {
	if len(input) == 0 {
		return nil, nil //nolint:nilnil // empty input is valid, not an error
	}
	namedSets := make(map[identity.Key]model_named_set.NamedSet)
	for key, ns := range input {
		converted, err := convertNamedSetToModel(key, ns)
		if err != nil {
			return nil, err
		}
		namedSets[converted.Key] = converted
	}
	return namedSets, nil
}

// convertActorsToModel converts actors and actor generalizations into the result model.
func convertActorsToModel(input *inputModel, result *core.Model) error {
	for key, actor := range input.Actors {
		converted, err := convertActorToModel(key, actor, input.ActorGeneralizations)
		if err != nil {
			return err
		}
		result.Actors[converted.Key] = converted
	}
	for key, gen := range input.ActorGeneralizations {
		converted, err := convertActorGeneralizationToModel(key, gen)
		if err != nil {
			return err
		}
		result.ActorGeneralizations[converted.Key] = converted
	}
	return nil
}

// convertDomainsAndAssociationsToModel converts domains, domain associations, and class associations into the result model.
func convertDomainsAndAssociationsToModel(input *inputModel, result *core.Model) error {
	for key, domain := range input.Domains {
		converted, err := convertDomainToModel(key, domain)
		if err != nil {
			return err
		}
		result.Domains[converted.Key] = converted
	}
	for key, assoc := range input.DomainAssociations {
		converted, err := convertDomainAssocToModel(key, assoc)
		if err != nil {
			return err
		}
		result.DomainAssociations[converted.Key] = converted
	}
	for key, assoc := range input.ClassAssociations {
		converted, err := convertModelAssociationToModel(key, assoc, result.Domains)
		if err != nil {
			return err
		}
		result.ClassAssociations[converted.Key] = converted
	}
	return nil
}

// convertInvariantsToModel converts a slice of inputLogic to model_logic.Logic for model invariants.
func convertInvariantsToModel(invariants []inputLogic) ([]model_logic.Logic, error) {
	if len(invariants) == 0 {
		return nil, nil
	}
	result := make([]model_logic.Logic, len(invariants))
	for i, inv := range invariants {
		invKey, err := identity.NewInvariantKey(fmt.Sprintf("%d", i))
		if err != nil {
			return nil, convErr(ErrConvKeyConstruction, fmt.Sprintf("failed to create invariant key '%d': %s", i, err.Error()), "model.json")
		}
		logic, err := convertLogicToModel(&inv, resolveLogicType(&inv, model_logic.LogicTypeAssessment), invKey)
		if err != nil {
			return nil, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert invariant %d: %s", i, err.Error()), "model.json")
		}
		result[i] = logic
	}
	return result, nil
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

	gfFile := fmt.Sprintf("global_functions/%s.gfunc.json", keyStr)

	logic, err := convertLogicToModel(&gf.Logic, model_logic.LogicTypeValue, key)
	if err != nil {
		return model_logic.GlobalFunction{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert global function logic: %s", err.Error()), gfFile)
	}

	result, err := model_logic.NewGlobalFunction(key, gf.Name, gf.Parameters, logic)
	if err != nil {
		return model_logic.GlobalFunction{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to create global function: %s", err.Error()), gfFile)
	}
	return result, nil
}

// convertNamedSetToModel converts an inputNamedSet to a model_named_set.NamedSet.
func convertNamedSetToModel(keyStr string, ns *inputNamedSet) (model_named_set.NamedSet, error) {
	nsFile := fmt.Sprintf("named_sets/%s.nset.json", keyStr)

	key, err := identity.NewNamedSetKey(keyStr)
	if err != nil {
		return model_named_set.NamedSet{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create named set key '%s': %s", keyStr, err.Error()),
			nsFile,
		).WithField("key")
	}

	spec, err := model_spec.NewExpressionSpec(ns.Notation, ns.Specification, nil)
	if err != nil {
		return model_named_set.NamedSet{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to create named set spec: %s", err.Error()), nsFile)
	}

	var typeSpec *model_spec.TypeSpec
	if ns.TypeSpec != "" {
		ts, err := model_spec.NewTypeSpec(model_logic.NotationTLAPlus, ns.TypeSpec, nil)
		if err != nil {
			return model_named_set.NamedSet{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to create named set type spec: %s", err.Error()), nsFile)
		}
		typeSpec = &ts
	}

	result, err := model_named_set.NewNamedSet(key, ns.Name, ns.Description, spec, typeSpec)
	if err != nil {
		return model_named_set.NamedSet{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to create named set: %s", err.Error()), nsFile)
	}
	return result, nil
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
		if slices.Contains(gen.SubclassKeys, keyStr) {
			result.SubclassOfKey = &genKey
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

// subdomainConvContext holds the context needed for subdomain-level conversions.
type subdomainConvContext struct {
	subdomainKey    identity.Key
	domainKeyStr    string
	subdomainKeyStr string
	subdomainFile   string
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

	ctx := subdomainConvContext{
		subdomainKey:    subdomainKey,
		domainKeyStr:    domainKeyStr,
		subdomainKeyStr: keyStr,
		subdomainFile:   subdomainFile,
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

	if err := convertSubdomainClassesAndGeneralizations(ctx, subdomain, &result); err != nil {
		return model_domain.Subdomain{}, err
	}

	if err := convertSubdomainUseCases(ctx, subdomain, &result); err != nil {
		return model_domain.Subdomain{}, err
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

// convertSubdomainClassesAndGeneralizations converts class generalizations and classes for a subdomain.
func convertSubdomainClassesAndGeneralizations(ctx subdomainConvContext, subdomain *inputSubdomain, result *model_domain.Subdomain) error {
	genKeyMap := make(map[string]identity.Key)
	for key, gen := range subdomain.ClassGeneralizations {
		converted, err := convertClassGeneralizationToModel(key, gen, ctx.subdomainKey, ctx.domainKeyStr, ctx.subdomainKeyStr)
		if err != nil {
			return err
		}
		result.Generalizations[converted.Key] = converted
		genKeyMap[key] = converted.Key
	}

	for key, class := range subdomain.Classes {
		converted, err := convertClassToModel(key, class, ctx.subdomainKey, subdomain.ClassGeneralizations, genKeyMap, ctx.domainKeyStr, ctx.subdomainKeyStr)
		if err != nil {
			return err
		}
		result.Classes[converted.Key] = converted
	}
	return nil
}

// convertSubdomainUseCases converts use case generalizations, use cases, and use case shares for a subdomain.
func convertSubdomainUseCases(ctx subdomainConvContext, subdomain *inputSubdomain, result *model_domain.Subdomain) error {
	ucGenKeyMap := make(map[string]identity.Key)
	for key, gen := range subdomain.UseCaseGeneralizations {
		converted, err := convertUseCaseGeneralizationToModel(key, gen, ctx.subdomainKey, ctx.domainKeyStr, ctx.subdomainKeyStr)
		if err != nil {
			return err
		}
		result.UseCaseGeneralizations[converted.Key] = converted
		ucGenKeyMap[key] = converted.Key
	}

	ucCtx := useCaseConvContext{
		subdomainKey:    ctx.subdomainKey,
		domainKeyStr:    ctx.domainKeyStr,
		subdomainKeyStr: ctx.subdomainKeyStr,
	}
	for key, useCase := range subdomain.UseCases {
		converted, err := convertUseCaseToModel(key, useCase, ucCtx, subdomain.UseCaseGeneralizations, ucGenKeyMap, result.Classes)
		if err != nil {
			return err
		}
		result.UseCases[converted.Key] = converted
	}

	// Convert use case shares
	if err := convertUseCaseSharesMap(ctx, subdomain.UseCaseShares, result); err != nil {
		return err
	}

	return nil
}

// convertUseCaseSharesMap converts the use case shares map for a subdomain.
func convertUseCaseSharesMap(ctx subdomainConvContext, shares map[string]map[string]*inputUseCaseShared, result *model_domain.Subdomain) error {
	for seaKeyStr, mudMap := range shares {
		seaKey, err := identity.NewUseCaseKey(ctx.subdomainKey, seaKeyStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create use case key '%s': %s", seaKeyStr, err.Error()),
				ctx.subdomainFile,
			).WithField("use_case_shares")
		}
		innerMap := make(map[identity.Key]model_use_case.UseCaseShared)
		for mudKeyStr, shared := range mudMap {
			mudKey, err := identity.NewUseCaseKey(ctx.subdomainKey, mudKeyStr)
			if err != nil {
				return convErr(
					ErrConvKeyConstruction,
					fmt.Sprintf("failed to create use case key '%s': %s", mudKeyStr, err.Error()),
					ctx.subdomainFile,
				).WithField("use_case_shares")
			}
			innerMap[mudKey] = model_use_case.UseCaseShared{
				ShareType:  shared.ShareType,
				UmlComment: shared.UmlComment,
			}
		}
		result.UseCaseShares[seaKey] = innerMap
	}
	return nil
}

// useCaseConvContext holds the context needed for use-case-level conversions.
type useCaseConvContext struct {
	subdomainKey    identity.Key
	domainKeyStr    string
	subdomainKeyStr string
}

// convertUseCaseToModel converts an inputUseCase to a model_use_case.UseCase.
func convertUseCaseToModel(keyStr string, uc *inputUseCase, ctx useCaseConvContext, ucGeneralizations map[string]*inputUseCaseGeneralization, ucGenKeyMap map[string]identity.Key, classes map[identity.Key]model_class.Class) (model_use_case.UseCase, error) {
	useCaseFile := fmt.Sprintf("domains/%s/subdomains/%s/use_cases/%s/use_case.json", ctx.domainKeyStr, ctx.subdomainKeyStr, keyStr)

	useCaseKey, err := identity.NewUseCaseKey(ctx.subdomainKey, keyStr)
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
		if slices.Contains(gen.SubclassKeys, keyStr) {
			result.SubclassOfKey = &genKey
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
			classKey, err = identity.NewClassKey(ctx.subdomainKey, classKeyStr)
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
		converted, err := convertScenarioToModel(scenKeyStr, scenario, useCaseKey, ctx.subdomainKey, ctx.domainKeyStr, ctx.subdomainKeyStr, keyStr)
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
		stepCtx := stepConvContext{
			scenarioKey:  scenarioKey,
			useCaseKey:   useCaseKey,
			subdomainKey: subdomainKey,
			objects:      scenario.Objects,
			scenarioFile: scenarioFile,
		}
		converted, err := convertStepToModel(scenario.Steps, stepCtx, 0)
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

// stepConvContext holds the context needed for step-level conversions.
type stepConvContext struct {
	scenarioKey  identity.Key
	useCaseKey   identity.Key
	subdomainKey identity.Key
	objects      map[string]*inputObject
	scenarioFile string
}

// convertStepToModel recursively converts an inputStep to a model_scenario.Step.
func convertStepToModel(step *inputStep, ctx stepConvContext, stepIndex int) (*model_scenario.Step, error) {
	stepKey, err := identity.NewScenarioStepKey(ctx.scenarioKey, fmt.Sprintf("step_%d", stepIndex))
	if err != nil {
		return nil, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create scenario step key 'step_%d': %s", stepIndex, err.Error()),
			ctx.scenarioFile,
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
	if err := convertStepObjectKeys(step, result, ctx, stepIndex); err != nil {
		return nil, err
	}
	if err := convertStepBehaviorKeys(step, result, ctx, stepIndex); err != nil {
		return nil, err
	}

	// Convert sub-statements
	for i, subStep := range step.Statements {
		subStepCopy := subStep
		converted, err := convertStepToModel(&subStepCopy, ctx, stepIndex*100+i+1)
		if err != nil {
			return nil, err
		}
		result.Statements = append(result.Statements, *converted)
	}

	return result, nil
}

// convertStepObjectKeys converts from_object_key and to_object_key references in a step.
func convertStepObjectKeys(step *inputStep, result *model_scenario.Step, ctx stepConvContext, stepIndex int) error {
	if step.FromObjectKey != nil {
		objKey, err := identity.NewScenarioObjectKey(ctx.scenarioKey, *step.FromObjectKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create from object key '%s': %s", *step.FromObjectKey, err.Error()),
				ctx.scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].from_object_key", stepIndex))
		}
		result.FromObjectKey = &objKey
	}
	if step.ToObjectKey != nil {
		objKey, err := identity.NewScenarioObjectKey(ctx.scenarioKey, *step.ToObjectKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create to object key '%s': %s", *step.ToObjectKey, err.Error()),
				ctx.scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].to_object_key", stepIndex))
		}
		result.ToObjectKey = &objKey
	}
	return nil
}

// convertStepBehaviorKeys converts event_key, query_key, and scenario_key references in a step.
func convertStepBehaviorKeys(step *inputStep, result *model_scenario.Step, ctx stepConvContext, stepIndex int) error {
	if step.EventKey != nil {
		classKey, err := resolveClassKeyFromStep(step, ctx.objects, ctx.subdomainKey, ctx.scenarioFile, stepIndex)
		if err != nil {
			return err
		}
		eventKey, err := identity.NewEventKey(classKey, *step.EventKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create event key '%s': %s", *step.EventKey, err.Error()),
				ctx.scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].event_key", stepIndex))
		}
		result.EventKey = &eventKey
	}
	if step.QueryKey != nil {
		classKey, err := resolveClassKeyFromStep(step, ctx.objects, ctx.subdomainKey, ctx.scenarioFile, stepIndex)
		if err != nil {
			return err
		}
		queryKey, err := identity.NewQueryKey(classKey, *step.QueryKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create query key '%s': %s", *step.QueryKey, err.Error()),
				ctx.scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].query_key", stepIndex))
		}
		result.QueryKey = &queryKey
	}
	if step.ScenarioKey != nil {
		scenKey, err := identity.NewScenarioKey(ctx.useCaseKey, *step.ScenarioKey)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create scenario key '%s': %s", *step.ScenarioKey, err.Error()),
				ctx.scenarioFile,
			).WithField(fmt.Sprintf("steps[%d].scenario_key", stepIndex))
		}
		result.ScenarioKey = &scenKey
	}
	return nil
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
		if slices.Contains(gen.SubclassKeys, keyStr) {
			result.SubclassOfKey = &genKey
		}
	}

	// Convert attributes and invariants
	if err := convertClassAttributesAndInvariants(class, &result, classKey, classFile); err != nil {
		return model_class.Class{}, err
	}

	// Convert state machine if present
	if class.StateMachine != nil {
		if err := convertStateMachineToModel(class.StateMachine, &result, classKey, domainKeyStr, subdomainKeyStr, keyStr); err != nil {
			return model_class.Class{}, err
		}
	}

	// Convert actions and queries
	if err := convertClassActionsAndQueries(class, &result, classKey, domainKeyStr, subdomainKeyStr, keyStr); err != nil {
		return model_class.Class{}, err
	}

	return result, nil
}

// convertClassAttributesAndInvariants converts attributes and class-level invariants into the result class.
func convertClassAttributesAndInvariants(class *inputClass, result *model_class.Class, classKey identity.Key, classFile string) error {
	for attrKeyStr, attr := range class.Attributes {
		converted, err := convertAttributeToModel(attrKeyStr, attr, classKey, class.Indexes, classFile)
		if err != nil {
			return err
		}
		result.Attributes[converted.Key] = converted
	}

	classInvariants, err := convertLogicsToModel(class.Invariants, model_logic.LogicTypeAssessment, classKey, identity.NewClassInvariantKey)
	if err != nil {
		return convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert class invariants: %s", err.Error()), classFile)
	}
	result.SetInvariants(classInvariants)
	return nil
}

// convertClassActionsAndQueries converts actions and queries into the result class.
func convertClassActionsAndQueries(class *inputClass, result *model_class.Class, classKey identity.Key, domainKeyStr, subdomainKeyStr, classKeyStr string) error {
	for actionKeyStr, action := range class.Actions {
		converted, err := convertActionToModel(actionKeyStr, action, classKey, domainKeyStr, subdomainKeyStr, classKeyStr)
		if err != nil {
			return err
		}
		result.Actions[converted.Key] = converted
	}
	for queryKeyStr, query := range class.Queries {
		converted, err := convertQueryToModel(queryKeyStr, query, classKey, domainKeyStr, subdomainKeyStr, classKeyStr)
		if err != nil {
			return err
		}
		result.Queries[converted.Key] = converted
	}
	return nil
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
		if slices.Contains(index, keyStr) {
			indexNums = append(indexNums, uint(i)) //nolint:gosec // index i is bounded by slice length, no overflow possible
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
		dp, err := convertLogicToModel(attr.DerivationPolicy, model_logic.LogicTypeValue, dpKey)
		if err != nil {
			return model_class.Attribute{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert derivation policy for attribute '%s': %s", keyStr, err.Error()), classFile)
		}
		result.DerivationPolicy = &dp
	}

	// Convert attribute invariants
	attrInvariants, err := convertLogicsToModel(attr.Invariants, model_logic.LogicTypeAssessment, attrKey, identity.NewAttributeInvariantKey)
	if err != nil {
		return model_class.Attribute{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert attribute '%s' invariants: %s", keyStr, err.Error()), classFile)
	}
	result.SetInvariants(attrInvariants)

	return result, nil
}

// convertStateMachineToModel converts an inputStateMachine to populate a Class's state machine fields.
func convertStateMachineToModel(sm *inputStateMachine, class *model_class.Class, classKey identity.Key, domainKeyStr, subdomainKeyStr, classKeyStr string) error {
	smFile := fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/state_machine.json", domainKeyStr, subdomainKeyStr, classKeyStr)

	if err := convertSMStatesToModel(sm, class, classKey, smFile); err != nil {
		return err
	}
	if err := convertSMEventsToModel(sm, class, classKey, smFile); err != nil {
		return err
	}
	if err := convertSMGuardsToModel(sm, class, classKey, smFile); err != nil {
		return err
	}
	if err := convertSMTransitionsToModel(sm, class, classKey, smFile); err != nil {
		return err
	}
	return nil
}

// convertSMStatesToModel converts state machine states into the class.
func convertSMStatesToModel(sm *inputStateMachine, class *model_class.Class, classKey identity.Key, smFile string) error {
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
	return nil
}

// convertSMEventsToModel converts state machine events into the class.
func convertSMEventsToModel(sm *inputStateMachine, class *model_class.Class, classKey identity.Key, smFile string) error {
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

		for _, param := range event.Parameters {
			converted.Parameters = append(converted.Parameters, model_state.Parameter{
				Name:          param.Name,
				DataTypeRules: param.DataTypeRules,
			})
		}

		class.Events[converted.Key] = converted
	}
	return nil
}

// convertSMGuardsToModel converts state machine guards into the class.
func convertSMGuardsToModel(sm *inputStateMachine, class *model_class.Class, classKey identity.Key, smFile string) error {
	for guardKeyStr, guard := range sm.Guards {
		guardKey, err := identity.NewGuardKey(classKey, guardKeyStr)
		if err != nil {
			return convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create guard key '%s': %s", guardKeyStr, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("guards.%s", guardKeyStr))
		}

		guardLogic, err := convertLogicToModel(&guard.Logic, model_logic.LogicTypeAssessment, guardKey)
		if err != nil {
			return convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert guard '%s' logic: %s", guardKeyStr, err.Error()), smFile)
		}

		converted, err := model_state.NewGuard(guardKey, guard.Name, guardLogic)
		if err != nil {
			return convErr(ErrConvModelValidation, fmt.Sprintf("failed to create guard '%s': %s", guardKeyStr, err.Error()), smFile)
		}

		class.Guards[converted.Key] = converted
	}
	return nil
}

// convertSMTransitionsToModel converts state machine transitions into the class.
func convertSMTransitionsToModel(sm *inputStateMachine, class *model_class.Class, classKey identity.Key, smFile string) error {
	for i, transition := range sm.Transitions {
		converted, err := convertSingleTransitionToModel(transition, i, classKey, smFile)
		if err != nil {
			return err
		}
		class.Transitions[converted.Key] = converted
	}
	return nil
}

// convertSingleTransitionToModel converts a single inputTransition to a model_state.Transition.
func convertSingleTransitionToModel(transition inputTransition, i int, classKey identity.Key, smFile string) (model_state.Transition, error) {
	var fromStr, toStr string
	if transition.FromStateKey != nil {
		fromStr = *transition.FromStateKey
	}
	if transition.ToStateKey != nil {
		toStr = *transition.ToStateKey
	}

	var guardStr, actionStr string
	if transition.GuardKey != nil {
		guardStr = *transition.GuardKey
	}
	if transition.ActionKey != nil {
		actionStr = *transition.ActionKey
	}

	transitionKey, err := identity.NewTransitionKey(classKey, fromStr, transition.EventKey, guardStr, actionStr, toStr)
	if err != nil {
		return model_state.Transition{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create transition key for transition[%d]: %s", i, err.Error()),
			smFile,
		).WithField(fmt.Sprintf("transitions[%d]", i))
	}

	converted := model_state.Transition{
		Key:        transitionKey,
		UmlComment: transition.UMLComment,
	}

	eventKey, err := identity.NewEventKey(classKey, transition.EventKey)
	if err != nil {
		return model_state.Transition{}, convErr(
			ErrConvKeyConstruction,
			fmt.Sprintf("failed to create event key reference '%s' for transition[%d]: %s", transition.EventKey, i, err.Error()),
			smFile,
		).WithField(fmt.Sprintf("transitions[%d].event_key", i))
	}
	converted.EventKey = eventKey

	if transition.FromStateKey != nil {
		stateKey, err := identity.NewStateKey(classKey, *transition.FromStateKey)
		if err != nil {
			return model_state.Transition{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create from state key reference '%s' for transition[%d]: %s", *transition.FromStateKey, i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d].from_state_key", i))
		}
		converted.FromStateKey = &stateKey
	}

	if transition.ToStateKey != nil {
		stateKey, err := identity.NewStateKey(classKey, *transition.ToStateKey)
		if err != nil {
			return model_state.Transition{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create to state key reference '%s' for transition[%d]: %s", *transition.ToStateKey, i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d].to_state_key", i))
		}
		converted.ToStateKey = &stateKey
	}

	if transition.GuardKey != nil {
		guardKey, err := identity.NewGuardKey(classKey, *transition.GuardKey)
		if err != nil {
			return model_state.Transition{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create guard key reference '%s' for transition[%d]: %s", *transition.GuardKey, i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d].guard_key", i))
		}
		converted.GuardKey = &guardKey
	}

	if transition.ActionKey != nil {
		actionKey, err := identity.NewActionKey(classKey, *transition.ActionKey)
		if err != nil {
			return model_state.Transition{}, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to create action key reference '%s' for transition[%d]: %s", *transition.ActionKey, i, err.Error()),
				smFile,
			).WithField(fmt.Sprintf("transitions[%d].action_key", i))
		}
		converted.ActionKey = &actionKey
	}

	return converted, nil
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

	requires, err := convertLogicsToModel(action.Requires, model_logic.LogicTypeAssessment, actionKey, identity.NewActionRequireKey)
	if err != nil {
		return model_state.Action{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert action requires: %s", err.Error()), actionFile)
	}
	guarantees, err := convertLogicsToModel(action.Guarantees, model_logic.LogicTypeStateChange, actionKey, identity.NewActionGuaranteeKey)
	if err != nil {
		return model_state.Action{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert action guarantees: %s", err.Error()), actionFile)
	}
	safetyRules, err := convertLogicsToModel(action.SafetyRules, model_logic.LogicTypeSafetyRule, actionKey, identity.NewActionSafetyKey)
	if err != nil {
		return model_state.Action{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert action safety rules: %s", err.Error()), actionFile)
	}

	return model_state.Action{
		Key:         actionKey,
		Name:        action.Name,
		Details:     action.Details,
		Parameters:  convertParametersToModel(action.Parameters),
		Requires:    requires,
		Guarantees:  guarantees,
		SafetyRules: safetyRules,
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

	requires, err := convertLogicsToModel(query.Requires, model_logic.LogicTypeAssessment, queryKey, identity.NewQueryRequireKey)
	if err != nil {
		return model_state.Query{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert query requires: %s", err.Error()), queryFile)
	}
	guarantees, err := convertLogicsToModel(query.Guarantees, model_logic.LogicTypeQuery, queryKey, identity.NewQueryGuaranteeKey)
	if err != nil {
		return model_state.Query{}, convErr(ErrConvModelValidation, fmt.Sprintf("failed to convert query guarantees: %s", err.Error()), queryFile)
	}

	return model_state.Query{
		Key:        queryKey,
		Name:       query.Name,
		Details:    query.Details,
		Parameters: convertParametersToModel(query.Parameters),
		Requires:   requires,
		Guarantees: guarantees,
	}, nil
}

// resolveLogicType returns the logic type from the input if specified as "let",
// otherwise returns the default type for the context.
func resolveLogicType(input *inputLogic, defaultType string) string {
	if input.Type == model_logic.LogicTypeLet {
		return model_logic.LogicTypeLet
	}
	return defaultType
}

// convertLogicToModel converts an inputLogic to a model_logic.Logic with the given key.
func convertLogicToModel(input *inputLogic, logicType string, logicKey identity.Key) (model_logic.Logic, error) {
	spec, err := model_spec.NewExpressionSpec(input.Notation, input.Specification, nil)
	if err != nil {
		return model_logic.Logic{}, fmt.Errorf("failed to create expression spec: %w", err)
	}

	var targetTypeSpec *model_spec.TypeSpec
	if input.TargetTypeSpec != "" {
		ts, err := model_spec.NewTypeSpec(model_logic.NotationTLAPlus, input.TargetTypeSpec, nil)
		if err != nil {
			return model_logic.Logic{}, fmt.Errorf("failed to create target type spec: %w", err)
		}
		targetTypeSpec = &ts
	}

	logic, err := model_logic.NewLogic(logicKey, logicType, input.Description, input.Target, spec, targetTypeSpec)
	if err != nil {
		return model_logic.Logic{}, err
	}
	return logic, nil
}

// convertLogicsToModel converts a slice of inputLogic to a slice of model_logic.Logic.
// keyFactory creates the identity key for each logic entry using the parent key and an index-based sub-key.
func convertLogicsToModel(logics []inputLogic, logicType string, parentKey identity.Key, keyFactory func(identity.Key, string) (identity.Key, error)) ([]model_logic.Logic, error) {
	if len(logics) == 0 {
		return nil, nil
	}
	result := make([]model_logic.Logic, len(logics))
	for i, logic := range logics {
		logicKey, err := keyFactory(parentKey, fmt.Sprintf("%d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create logic key %d: %w", i, err)
		}
		converted, err := convertLogicToModel(&logic, resolveLogicType(&logic, logicType), logicKey)
		if err != nil {
			return nil, fmt.Errorf("logic %d: %w", i, err)
		}
		result[i] = converted
	}
	return result, nil
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
	if trimmed, ok := strings.CutSuffix(mult, "..*"); ok {
		return trimmed + "..many"
	}
	return mult
}
