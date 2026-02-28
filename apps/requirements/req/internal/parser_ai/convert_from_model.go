package parser_ai

import (
	"fmt"

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

// ConvertFromModel converts a req_model.Model to an inputModel.
// It first validates the source model, then performs the conversion.
func ConvertFromModel(model *req_model.Model) (*inputModel, error) {
	// Validate the source model
	if err := model.Validate(); err != nil {
		return nil, convErr(
			ErrConvSourceModelValidation,
			fmt.Sprintf("source model validation failed: %s", err.Error()),
			"model.json",
		)
	}

	result := &inputModel{
		Name:                 model.Name,
		Details:              model.Details,
		Invariants:           convertLogicsFromModel(model.Invariants),
		Actors:               make(map[string]*inputActor),
		ActorGeneralizations: make(map[string]*inputActorGeneralization),
		GlobalFunctions:      make(map[string]*inputGlobalFunction),
		Domains:              make(map[string]*inputDomain),
		DomainAssociations:   make(map[string]*inputDomainAssociation),
		ClassAssociations:    make(map[string]*inputClassAssociation),
	}

	// Convert actors
	for key, actor := range model.Actors {
		converted := convertActorFromModel(&actor)
		result.Actors[key.SubKey] = converted
	}

	// Convert actor generalizations
	for key, gen := range model.ActorGeneralizations {
		converted := convertActorGeneralizationFromModel(&gen, model.Actors)
		result.ActorGeneralizations[key.SubKey] = converted
	}

	// Convert global functions (SubKey has underscore stripped, add it back)
	for key, gf := range model.GlobalFunctions {
		converted := convertGlobalFunctionFromModel(&gf)
		result.GlobalFunctions["_"+key.SubKey] = converted
	}

	// Convert domains
	for key, domain := range model.Domains {
		converted, err := convertDomainFromModel(&domain)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to convert domain '%s': %s", key.SubKey, err.Error()),
				fmt.Sprintf("domains/%s/domain.json", key.SubKey),
			)
		}
		result.Domains[key.SubKey] = converted
	}

	// Convert domain associations
	for key, assoc := range model.DomainAssociations {
		converted := convertDomainAssocFromModel(&assoc)
		result.DomainAssociations[key.SubKey+"."+key.SubKey2] = converted
	}

	// Convert model-level class associations
	for key, assoc := range model.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, "")
		result.ClassAssociations[key.SubKey3] = converted
	}

	return result, nil
}

// convertActorFromModel converts a model_actor.Actor to an inputActor.
func convertActorFromModel(actor *model_actor.Actor) *inputActor {
	result := &inputActor{
		Name:       actor.Name,
		Type:       actor.Type,
		Details:    actor.Details,
		UMLComment: actor.UmlComment,
	}
	return result
}

// convertActorGeneralizationFromModel converts a model_actor.Generalization to an inputActorGeneralization.
func convertActorGeneralizationFromModel(gen *model_actor.Generalization, actors map[identity.Key]model_actor.Actor) *inputActorGeneralization {
	result := &inputActorGeneralization{
		Name:         gen.Name,
		Details:      gen.Details,
		IsComplete:   gen.IsComplete,
		IsStatic:     gen.IsStatic,
		UMLComment:   gen.UmlComment,
		SubclassKeys: []string{},
	}

	// Find superclass and subclasses by examining actor references
	for key, actor := range actors {
		if actor.SuperclassOfKey != nil && actor.SuperclassOfKey.SubKey == gen.Key.SubKey {
			result.SuperclassKey = key.SubKey
		}
		if actor.SubclassOfKey != nil && actor.SubclassOfKey.SubKey == gen.Key.SubKey {
			result.SubclassKeys = append(result.SubclassKeys, key.SubKey)
		}
	}

	return result
}

// convertGlobalFunctionFromModel converts a model_logic.GlobalFunction to an inputGlobalFunction.
func convertGlobalFunctionFromModel(gf *model_logic.GlobalFunction) *inputGlobalFunction {
	return &inputGlobalFunction{
		Name:       gf.Name,
		Parameters: gf.Parameters,
		Logic:      convertLogicFromModel(&gf.Logic),
	}
}

// convertDomainAssocFromModel converts a model_domain.Association to an inputDomainAssociation.
func convertDomainAssocFromModel(assoc *model_domain.Association) *inputDomainAssociation {
	return &inputDomainAssociation{
		ProblemDomainKey:  assoc.ProblemDomainKey.SubKey,
		SolutionDomainKey: assoc.SolutionDomainKey.SubKey,
		UmlComment:        assoc.UmlComment,
	}
}

// convertDomainFromModel converts a model_domain.Domain to an inputDomain.
func convertDomainFromModel(domain *model_domain.Domain) (*inputDomain, error) {
	result := &inputDomain{
		Name:              domain.Name,
		Details:           domain.Details,
		Realized:          domain.Realized,
		UMLComment:        domain.UmlComment,
		Subdomains:        make(map[string]*inputSubdomain),
		ClassAssociations: make(map[string]*inputClassAssociation),
	}

	// Convert subdomains
	for key, subdomain := range domain.Subdomains {
		converted, err := convertSubdomainFromModel(&subdomain, domain.Key)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to convert subdomain '%s': %s", key.SubKey, err.Error()),
				fmt.Sprintf("domains/%s/subdomains/%s/subdomain.json", domain.Key.SubKey, key.SubKey),
			)
		}
		result.Subdomains[key.SubKey] = converted
	}

	// Convert domain-level class associations
	for key, assoc := range domain.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, identity.KEY_TYPE_DOMAIN)
		result.ClassAssociations[key.SubKey3] = converted
	}

	return result, nil
}

// convertSubdomainFromModel converts a model_domain.Subdomain to an inputSubdomain.
func convertSubdomainFromModel(subdomain *model_domain.Subdomain, domainKey identity.Key) (*inputSubdomain, error) {
	result := &inputSubdomain{
		Name:                   subdomain.Name,
		Details:                subdomain.Details,
		UMLComment:             subdomain.UmlComment,
		Classes:                make(map[string]*inputClass),
		ClassGeneralizations:   make(map[string]*inputClassGeneralization),
		ClassAssociations:      make(map[string]*inputClassAssociation),
		UseCases:               make(map[string]*inputUseCase),
		UseCaseGeneralizations: make(map[string]*inputUseCaseGeneralization),
		UseCaseShares:          make(map[string]map[string]*inputUseCaseShared),
	}

	// Convert classes
	for key, class := range subdomain.Classes {
		converted, err := convertClassFromModel(&class)
		if err != nil {
			return nil, convErr(
				ErrConvKeyConstruction,
				fmt.Sprintf("failed to convert class '%s': %s", key.SubKey, err.Error()),
				fmt.Sprintf("domains/%s/subdomains/%s/classes/%s/class.json", domainKey.SubKey, subdomain.Key.SubKey, key.SubKey),
			)
		}
		result.Classes[key.SubKey] = converted
	}

	// Convert generalizations
	for key, gen := range subdomain.Generalizations {
		converted := convertClassGeneralizationFromModel(&gen, subdomain.Classes)
		result.ClassGeneralizations[key.SubKey] = converted
	}

	// Convert use case generalizations
	for key, gen := range subdomain.UseCaseGeneralizations {
		converted := convertUseCaseGeneralizationFromModel(&gen, subdomain.UseCases)
		result.UseCaseGeneralizations[key.SubKey] = converted
	}

	// Convert use cases
	for key, useCase := range subdomain.UseCases {
		converted := convertUseCaseFromModel(&useCase)
		result.UseCases[key.SubKey] = converted
	}

	// Convert use case shares
	for seaKey, mudMap := range subdomain.UseCaseShares {
		innerMap := make(map[string]*inputUseCaseShared)
		for mudKey, shared := range mudMap {
			innerMap[mudKey.SubKey] = &inputUseCaseShared{
				ShareType:  shared.ShareType,
				UmlComment: shared.UmlComment,
			}
		}
		result.UseCaseShares[seaKey.SubKey] = innerMap
	}

	// Convert subdomain-level class associations
	for key, assoc := range subdomain.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, identity.KEY_TYPE_SUBDOMAIN)
		result.ClassAssociations[key.SubKey3] = converted
	}

	return result, nil
}

// convertUseCaseFromModel converts a model_use_case.UseCase to an inputUseCase.
func convertUseCaseFromModel(uc *model_use_case.UseCase) *inputUseCase {
	result := &inputUseCase{
		Name:       uc.Name,
		Details:    uc.Details,
		Level:      uc.Level,
		ReadOnly:   uc.ReadOnly,
		UMLComment: uc.UmlComment,
		Actors:     make(map[string]*inputUseCaseActor),
		Scenarios:  make(map[string]*inputScenario),
	}

	// Convert actors (keyed by class key)
	for classKey, actor := range uc.Actors {
		result.Actors[classKey.SubKey] = &inputUseCaseActor{
			UmlComment: actor.UmlComment,
		}
	}

	// Convert scenarios
	for key, scenario := range uc.Scenarios {
		result.Scenarios[key.SubKey] = convertScenarioFromModel(&scenario)
	}

	return result
}

// convertScenarioFromModel converts a model_scenario.Scenario to an inputScenario.
func convertScenarioFromModel(scenario *model_scenario.Scenario) *inputScenario {
	result := &inputScenario{
		Name:    scenario.Name,
		Details: scenario.Details,
		Objects: make(map[string]*inputObject),
	}

	// Convert objects
	for key, obj := range scenario.Objects {
		result.Objects[key.SubKey] = convertScenarioObjectFromModel(&obj)
	}

	// Convert steps
	if scenario.Steps != nil {
		step := convertStepFromModel(scenario.Steps)
		result.Steps = &step
	}

	return result
}

// convertScenarioObjectFromModel converts a model_scenario.Object to an inputObject.
func convertScenarioObjectFromModel(obj *model_scenario.Object) *inputObject {
	return &inputObject{
		ObjectNumber: obj.ObjectNumber,
		Name:         obj.Name,
		NameStyle:    obj.NameStyle,
		ClassKey:     obj.ClassKey.SubKey,
		Multi:        obj.Multi,
		UmlComment:   obj.UmlComment,
	}
}

// convertStepFromModel recursively converts a model_scenario.Step to an inputStep.
func convertStepFromModel(step *model_scenario.Step) inputStep {
	result := inputStep{
		StepType:    step.StepType,
		LeafType:    step.LeafType,
		Condition:   step.Condition,
		Description: step.Description,
	}

	if step.FromObjectKey != nil {
		key := step.FromObjectKey.SubKey
		result.FromObjectKey = &key
	}
	if step.ToObjectKey != nil {
		key := step.ToObjectKey.SubKey
		result.ToObjectKey = &key
	}
	if step.EventKey != nil {
		key := step.EventKey.SubKey
		result.EventKey = &key
	}
	if step.QueryKey != nil {
		key := step.QueryKey.SubKey
		result.QueryKey = &key
	}
	if step.ScenarioKey != nil {
		key := step.ScenarioKey.SubKey
		result.ScenarioKey = &key
	}

	// Convert sub-statements
	for _, subStep := range step.Statements {
		result.Statements = append(result.Statements, convertStepFromModel(&subStep))
	}

	return result
}

// convertUseCaseGeneralizationFromModel converts a model_use_case.Generalization to an inputUseCaseGeneralization.
func convertUseCaseGeneralizationFromModel(gen *model_use_case.Generalization, useCases map[identity.Key]model_use_case.UseCase) *inputUseCaseGeneralization {
	result := &inputUseCaseGeneralization{
		Name:         gen.Name,
		Details:      gen.Details,
		IsComplete:   gen.IsComplete,
		IsStatic:     gen.IsStatic,
		UMLComment:   gen.UmlComment,
		SubclassKeys: []string{},
	}

	// Find superclass and subclasses by examining use case references
	for key, uc := range useCases {
		if uc.SuperclassOfKey != nil && uc.SuperclassOfKey.SubKey == gen.Key.SubKey {
			result.SuperclassKey = key.SubKey
		}
		if uc.SubclassOfKey != nil && uc.SubclassOfKey.SubKey == gen.Key.SubKey {
			result.SubclassKeys = append(result.SubclassKeys, key.SubKey)
		}
	}

	return result
}

// convertClassFromModel converts a model_class.Class to an inputClass.
func convertClassFromModel(class *model_class.Class) (*inputClass, error) {
	result := &inputClass{
		Name:       class.Name,
		Details:    class.Details,
		UMLComment: class.UmlComment,
		Attributes: make(map[string]*inputAttribute),
		Indexes:    [][]string{},
		Actions:    make(map[string]*inputAction),
		Queries:    make(map[string]*inputQuery),
	}

	// Set actor key if present
	if class.ActorKey != nil {
		result.ActorKey = class.ActorKey.SubKey
	}

	// Convert attributes
	for key, attr := range class.Attributes {
		converted := convertAttributeFromModel(&attr)
		result.Attributes[key.SubKey] = converted
	}

	// Build indexes from attribute IndexNums
	indexMap := make(map[uint][]string)
	for key, attr := range class.Attributes {
		for _, indexNum := range attr.IndexNums {
			indexMap[indexNum] = append(indexMap[indexNum], key.SubKey)
		}
	}
	// Convert map to slice
	for i := uint(0); i < uint(len(indexMap)); i++ {
		if attrs, ok := indexMap[i]; ok {
			result.Indexes = append(result.Indexes, attrs)
		}
	}

	// Convert class invariants
	result.Invariants = convertLogicsFromModel(class.Invariants)

	// Convert state machine if present
	if len(class.States) > 0 || len(class.Events) > 0 {
		result.StateMachine = convertStateMachineFromModel(class)
	}

	// Convert actions
	for key, action := range class.Actions {
		converted := convertActionFromModel(&action)
		result.Actions[key.SubKey] = converted
	}

	// Convert queries
	for key, query := range class.Queries {
		converted := convertQueryFromModel(&query)
		result.Queries[key.SubKey] = converted
	}

	return result, nil
}

// convertAttributeFromModel converts a model_class.Attribute to an inputAttribute.
func convertAttributeFromModel(attr *model_class.Attribute) *inputAttribute {
	result := &inputAttribute{
		Name:          attr.Name,
		DataTypeRules: attr.DataTypeRules,
		Details:       attr.Details,
		Nullable:      attr.Nullable,
		UMLComment:    attr.UmlComment,
	}
	if attr.DerivationPolicy != nil {
		dp := convertLogicFromModel(attr.DerivationPolicy)
		result.DerivationPolicy = &dp
	}
	return result
}

// convertStateMachineFromModel converts state machine components from a Class to an inputStateMachine.
func convertStateMachineFromModel(class *model_class.Class) *inputStateMachine {
	sm := &inputStateMachine{
		States:      make(map[string]*inputState),
		Events:      make(map[string]*inputEvent),
		Guards:      make(map[string]*inputGuard),
		Transitions: []inputTransition{},
	}

	// Convert states
	for key, state := range class.States {
		converted := convertStateFromModel(&state)
		sm.States[key.SubKey] = converted
	}

	// Convert events
	for key, event := range class.Events {
		converted := convertEventFromModel(&event)
		sm.Events[key.SubKey] = converted
	}

	// Convert guards
	for key, guard := range class.Guards {
		converted := convertGuardFromModel(&guard)
		sm.Guards[key.SubKey] = converted
	}

	// Convert transitions
	for _, transition := range class.Transitions {
		converted := convertTransitionFromModel(&transition)
		sm.Transitions = append(sm.Transitions, converted)
	}

	return sm
}

// convertStateFromModel converts a model_state.State to an inputState.
func convertStateFromModel(state *model_state.State) *inputState {
	result := &inputState{
		Name:       state.Name,
		Details:    state.Details,
		UMLComment: state.UmlComment,
		Actions:    []inputStateAction{},
	}

	// Convert state actions
	for _, stateAction := range state.Actions {
		converted := inputStateAction{
			ActionKey: stateAction.ActionKey.SubKey,
			When:      stateAction.When,
		}
		result.Actions = append(result.Actions, converted)
	}

	return result
}

// convertEventFromModel converts a model_state.Event to an inputEvent.
func convertEventFromModel(event *model_state.Event) *inputEvent {
	result := &inputEvent{
		Name:       event.Name,
		Details:    event.Details,
		Parameters: []inputParameter{},
	}

	// Convert event parameters
	for _, param := range event.Parameters {
		converted := inputParameter{
			Name:          param.Name,
			DataTypeRules: param.DataTypeRules,
		}
		result.Parameters = append(result.Parameters, converted)
	}

	return result
}

// convertGuardFromModel converts a model_state.Guard to an inputGuard.
func convertGuardFromModel(guard *model_state.Guard) *inputGuard {
	return &inputGuard{
		Name:  guard.Name,
		Logic: convertLogicFromModel(&guard.Logic),
	}
}

// convertTransitionFromModel converts a model_state.Transition to an inputTransition.
func convertTransitionFromModel(transition *model_state.Transition) inputTransition {
	result := inputTransition{
		EventKey:   transition.EventKey.SubKey,
		UMLComment: transition.UmlComment,
	}

	// Handle from state (nil for initial transitions)
	if transition.FromStateKey != nil {
		fromKey := transition.FromStateKey.SubKey
		// Check if it's "initial" (meaning no from state)
		if fromKey != "initial" {
			result.FromStateKey = &fromKey
		}
	}

	// Handle to state (nil for final transitions)
	if transition.ToStateKey != nil {
		toKey := transition.ToStateKey.SubKey
		// Check if it's "final" (meaning no to state)
		if toKey != "final" {
			result.ToStateKey = &toKey
		}
	}

	// Handle guard key
	if transition.GuardKey != nil {
		guardKey := transition.GuardKey.SubKey
		result.GuardKey = &guardKey
	}

	// Handle action key
	if transition.ActionKey != nil {
		actionKey := transition.ActionKey.SubKey
		result.ActionKey = &actionKey
	}

	return result
}

// convertActionFromModel converts a model_state.Action to an inputAction.
func convertActionFromModel(action *model_state.Action) *inputAction {
	result := &inputAction{
		Name:    action.Name,
		Details: action.Details,
	}
	result.Parameters = convertParametersFromModel(action.Parameters)
	result.Requires = convertLogicsFromModel(action.Requires)
	result.Guarantees = convertLogicsFromModel(action.Guarantees)
	result.SafetyRules = convertLogicsFromModel(action.SafetyRules)
	return result
}

// convertQueryFromModel converts a model_state.Query to an inputQuery.
func convertQueryFromModel(query *model_state.Query) *inputQuery {
	result := &inputQuery{
		Name:    query.Name,
		Details: query.Details,
	}
	result.Parameters = convertParametersFromModel(query.Parameters)
	result.Requires = convertLogicsFromModel(query.Requires)
	result.Guarantees = convertLogicsFromModel(query.Guarantees)
	return result
}

// convertLogicFromModel converts a model_logic.Logic to an inputLogic.
func convertLogicFromModel(logic *model_logic.Logic) inputLogic {
	return inputLogic{
		Type:          logic.Type,
		Description:   logic.Description,
		Target:        logic.Target,
		Notation:      logic.Spec.Notation,
		Specification: logic.Spec.Specification,
	}
}

// convertLogicsFromModel converts a slice of model_logic.Logic to a slice of inputLogic.
func convertLogicsFromModel(logics []model_logic.Logic) []inputLogic {
	if len(logics) == 0 {
		return nil
	}
	result := make([]inputLogic, len(logics))
	for i, logic := range logics {
		result[i] = convertLogicFromModel(&logic)
	}
	return result
}

// convertParametersFromModel converts a slice of model_state.Parameter to a slice of inputParameter.
func convertParametersFromModel(params []model_state.Parameter) []inputParameter {
	if len(params) == 0 {
		return nil
	}
	result := make([]inputParameter, len(params))
	for i, param := range params {
		result[i] = inputParameter{
			Name:          param.Name,
			DataTypeRules: param.DataTypeRules,
		}
	}
	return result
}

// convertClassGeneralizationFromModel converts a model_class.Generalization to an inputClassGeneralization.
// It needs the classes map to find which classes reference this class generalization.
func convertClassGeneralizationFromModel(gen *model_class.Generalization, classes map[identity.Key]model_class.Class) *inputClassGeneralization {
	result := &inputClassGeneralization{
		Name:         gen.Name,
		Details:      gen.Details,
		IsComplete:   gen.IsComplete,
		IsStatic:     gen.IsStatic,
		UMLComment:   gen.UmlComment,
		SubclassKeys: []string{},
	}

	// Find superclass and subclasses by examining class references
	for key, class := range classes {
		if class.SuperclassOfKey != nil && class.SuperclassOfKey.SubKey == gen.Key.SubKey {
			// This class is the superclass of this generalization
			result.SuperclassKey = key.SubKey
		}
		if class.SubclassOfKey != nil && class.SubclassOfKey.SubKey == gen.Key.SubKey {
			// This class is a subclass of this generalization
			result.SubclassKeys = append(result.SubclassKeys, key.SubKey)
		}
	}

	return result
}

// convertAssociationFromModel converts a model_class.Association to an inputClassAssociation.
// The parentType indicates the scope level: "", "domain", or "subdomain".
func convertAssociationFromModel(assoc *model_class.Association, parentType string) *inputClassAssociation {
	result := &inputClassAssociation{
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromMultiplicity: assoc.FromMultiplicity.String(),
		ToMultiplicity:   assoc.ToMultiplicity.String(),
		UmlComment:       assoc.UmlComment,
	}

	// Format class keys based on scope
	switch parentType {
	case identity.KEY_TYPE_SUBDOMAIN:
		// Subdomain level - just class name
		result.FromClassKey = assoc.FromClassKey.SubKey
		result.ToClassKey = assoc.ToClassKey.SubKey
		if assoc.AssociationClassKey != nil {
			key := assoc.AssociationClassKey.SubKey
			result.AssociationClassKey = &key
		}
	case identity.KEY_TYPE_DOMAIN:
		// Domain level - subdomain/class format
		result.FromClassKey = extractDomainScopedKey(assoc.FromClassKey)
		result.ToClassKey = extractDomainScopedKey(assoc.ToClassKey)
		if assoc.AssociationClassKey != nil {
			key := extractDomainScopedKey(*assoc.AssociationClassKey)
			result.AssociationClassKey = &key
		}
	default:
		// Model level - domain/subdomain/class format
		result.FromClassKey = extractModelScopedKey(assoc.FromClassKey)
		result.ToClassKey = extractModelScopedKey(assoc.ToClassKey)
		if assoc.AssociationClassKey != nil {
			key := extractModelScopedKey(*assoc.AssociationClassKey)
			result.AssociationClassKey = &key
		}
	}

	return result
}

// extractDomainScopedKey extracts subdomain/class from a full class key.
func extractDomainScopedKey(classKey identity.Key) string {
	// Class key format: domain/domainName/subdomain/subdomainName/class/className
	// We want: subdomainName/className
	keyStr := classKey.String()
	// Parse to find subdomain and class
	parts := splitKeyPath(keyStr)
	var subdomainName, className string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == identity.KEY_TYPE_SUBDOMAIN && i+1 < len(parts) {
			subdomainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_CLASS && i+1 < len(parts) {
			className = parts[i+1]
		}
	}
	return subdomainName + "/" + className
}

// extractModelScopedKey extracts domain/subdomain/class from a full class key.
func extractModelScopedKey(classKey identity.Key) string {
	// Class key format: domain/domainName/subdomain/subdomainName/class/className
	// We want: domainName/subdomainName/className
	keyStr := classKey.String()
	parts := splitKeyPath(keyStr)
	var domainName, subdomainName, className string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == identity.KEY_TYPE_DOMAIN && i+1 < len(parts) {
			domainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_SUBDOMAIN && i+1 < len(parts) {
			subdomainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_CLASS && i+1 < len(parts) {
			className = parts[i+1]
		}
	}
	return domainName + "/" + subdomainName + "/" + className
}

// splitKeyPath splits a key string by "/".
func splitKeyPath(keyStr string) []string {
	result := []string{}
	current := ""
	for _, c := range keyStr {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
